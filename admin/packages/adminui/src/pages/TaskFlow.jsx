import React, {useCallback, useEffect, useRef, useState} from "react";
import {useSearchParams} from "react-router-dom";
import BotSelector from "../components/BotSelector";
import Toast from "../components/Toast";
import Pagination from "../components/Pagination.jsx";
import {useTranslation} from "react-i18next";
import ReactMarkdown from "react-markdown";

// 状态机节点顺序（与后端 taskflow 保持一致）
const STAGE_NODES = ["PLAN", "EXECUTE", "REFLECT", "SUMMARIZE"];
const TERMINAL_STATES = ["DONE", "FAILED", "CANCELED"];

// 状态 → 展示配置
const STATE_STYLE = {
    PLAN: {label: "PLAN", color: "bg-blue-500", text: "text-blue-600"},
    EXECUTE: {label: "EXECUTE", color: "bg-indigo-500", text: "text-indigo-600"},
    REFLECT: {label: "REFLECT", color: "bg-purple-500", text: "text-purple-600"},
    SUMMARIZE: {label: "SUMMARIZE", color: "bg-cyan-500", text: "text-cyan-600"},
    DONE: {label: "DONE", color: "bg-green-500", text: "text-green-600"},
    FAILED: {label: "FAILED", color: "bg-red-500", text: "text-red-600"},
    CANCELED: {label: "CANCELED", color: "bg-gray-400", text: "text-gray-500"},
};

const STEP_STYLE = {
    pending: {border: "border-gray-200 bg-gray-50", dot: "bg-gray-300", label: "text-gray-400"},
    running: {border: "border-blue-400 bg-blue-50", dot: "bg-blue-500 animate-pulse", label: "text-blue-600"},
    done: {border: "border-green-200 bg-green-50", dot: "bg-green-500", label: "text-green-700"},
    failed: {border: "border-red-300 bg-red-50", dot: "bg-red-500", label: "text-red-600"},
};

const EVENT_ICON = {
    run_started: "▶",
    stage: "⇢",
    loop_started: "↻",
    plan_created: "☰",
    step_started: "▸",
    step_finished: "▪",
    tool_called: "🔧",
    approval_requested: "🔐",
    run_finished: "🏁",
};

function isTerminal(state) {
    return TERMINAL_STATES.includes(state);
}

function formatTime(ts) {
    if (!ts) return "-";
    const d = new Date(ts * 1000);
    return d.toLocaleString();
}

function formatDuration(start, end) {
    if (!start || !end) return "-";
    const secs = Math.max(0, end - start);
    if (secs < 60) return `${secs}s`;
    return `${Math.floor(secs / 60)}m${secs % 60}s`;
}

// 状态流转图：PLAN → EXECUTE → REFLECT (↔ 循环) → SUMMARIZE → 终态
function FlowDiagram({run}) {
    const {t} = useTranslation();
    const state = run.state || "PLAN";
    const terminal = isTerminal(state);
    // 已完成的阶段 = 流转到当前节点所经过的节点
    const currentIdx = STAGE_NODES.indexOf(state);
    const passed = (node) => {
        if (terminal) return true; // 终态时全流程已走完
        const idx = STAGE_NODES.indexOf(node);
        return currentIdx >= 0 && idx < currentIdx;
    };
    const active = (node) => !terminal && node === state;

    const nodeCls = (node) => {
        const s = STATE_STYLE[node];
        if (active(node)) return `${s.color} text-white shadow-lg scale-110`;
        if (passed(node)) return `${s.color} text-white opacity-60`;
        return "bg-gray-100 text-gray-400";
    };

    return (
        <div className="bg-white rounded-lg shadow p-6 overflow-x-auto">
            <div className="flex items-center gap-2 min-w-[560px] justify-center">
                {STAGE_NODES.map((node, i) => (
                    <React.Fragment key={node}>
                        {i > 0 && (
                            <div className="flex flex-col items-center">
                                {i === 2 && (
                                    // REFLECT → PLAN 回环提示
                                    <span className="text-[10px] text-purple-500 mb-0.5">
                                        ↻ {t("taskflow_loop", {loop: run.loop || 1, max: run.max_loops || "-"})}
                                    </span>
                                )}
                                <div
                                    className={`w-8 h-0.5 ${passed(node) || active(STAGE_NODES[i - 1]) ? "bg-indigo-400" : "bg-gray-200"}`}/>
                            </div>
                        )}
                        <div
                            className={`flex flex-col items-center transition-all duration-500 ${active(node) ? "animate-pulse" : ""}`}>
                            <div
                                className={`w-16 h-16 rounded-full flex items-center justify-center text-xs font-bold transition-all duration-500 ${nodeCls(node)}`}>
                                {STATE_STYLE[node].label}
                            </div>
                            {active(node) && (
                                <span className="text-[10px] text-indigo-500 mt-1">{t("taskflow_running")}</span>
                            )}
                        </div>
                    </React.Fragment>
                ))}
                {/* 终态节点 */}
                <div className="w-8 h-0.5 bg-gray-200"/>
                <div
                    className={`w-16 h-16 rounded-full flex items-center justify-center text-xs font-bold ${
                        terminal ? `${STATE_STYLE[state].color} text-white` : "bg-gray-100 text-gray-300"
                    }`}>
                    {terminal ? STATE_STYLE[state].label : t("taskflow_end")}
                </div>
            </div>
        </div>
    );
}

// 计划步骤卡片
function StepList({steps}) {
    const {t} = useTranslation();
    if (!steps || steps.length === 0) {
        return <div className="text-sm text-gray-400 p-4">{t("taskflow_no_plan")}</div>;
    }
    return (
        <div className="space-y-3">
            {steps.map((step, i) => {
                const style = STEP_STYLE[step.status] || STEP_STYLE.pending;
                return (
                    <div key={i} className={`rounded-lg border-2 p-4 transition-all ${style.border}`}>
                        <div className="flex items-center gap-2">
                            <span className={`w-2.5 h-2.5 rounded-full ${style.dot}`}/>
                            <span className="font-semibold text-gray-800">{step.name}</span>
                            <span className={`text-xs font-medium ml-auto ${style.label}`}>
                                {t(`taskflow_step_${step.status || "pending"}`)}
                            </span>
                        </div>
                        {step.description && (
                            <p className="text-sm text-gray-500 mt-1 ml-4.5">{step.description}</p>
                        )}
                        {step.result && (
                            <div className="mt-2 ml-4 text-sm text-gray-600 bg-white/60 rounded p-2 max-h-40 overflow-y-auto">
                                <ReactMarkdown>{step.result}</ReactMarkdown>
                            </div>
                        )}
                    </div>
                );
            })}
        </div>
    );
}

// 事件时间线（倒序显示最新在上）
function EventTimeline({events}) {
    const {t} = useTranslation();
    if (!events || events.length === 0) {
        return <div className="text-sm text-gray-400 p-4">{t("taskflow_no_events")}</div>;
    }
    return (
        <div className="space-y-1.5 max-h-96 overflow-y-auto">
            {[...events].reverse().map((ev, i) => (
                <div key={i} className="flex items-start gap-2 text-sm font-mono">
                    <span className="text-gray-300 text-xs whitespace-nowrap pt-0.5">
                        {new Date(ev.ts * 1000).toLocaleTimeString()}
                    </span>
                    <span className="whitespace-nowrap">{EVENT_ICON[ev.type] || "·"}</span>
                    <span className={
                        ev.type === "run_finished" ? "font-bold text-gray-700" :
                            ev.type === "tool_called" ? "text-amber-700" :
                                ev.type === "approval_requested" ? "text-red-500" : "text-gray-600"
                    }>{ev.detail}</span>
                </div>
            ))}
        </div>
    );
}

function TaskFlow() {
    const {t} = useTranslation();
    const [botId, setBotId] = useState(null);
    const [page, setPage] = useState(1);
    const [pageSize] = useState(10);
    const [total, setTotal] = useState(0);
    const [running, setRunning] = useState([]);
    const [history, setHistory] = useState([]);
    // 详情的 run_id 挂在 URL（?run=...）：刷新页面或直接打开链接仍停留在详情页，
    // 浏览器返回键也能自然回到列表
    const [searchParams, setSearchParams] = useSearchParams();
    const selectedRunId = searchParams.get("run");
    const [detail, setDetail] = useState(null);
    const [toast, setToast] = useState({show: false, message: "", type: "error"});

    const pollTimer = useRef(null);

    const showToast = (message, type = "error") => setToast({show: true, message, type});

    const fetchList = useCallback(async () => {
        if (!botId) return;
        try {
            const params = new URLSearchParams({page, page_size: pageSize, id: botId});
            const res = await fetch(`/bot/task/list?${params.toString()}`);
            const data = await res.json();
            if (data.code !== 0) return showToast(data.message || t("request_error"));
            setRunning(data.data?.running || []);
            setHistory(data.data?.list || []);
            setTotal(data.data?.total || 0);
        } catch (err) {
            showToast(t("request_error") + ": " + err.message);
        }
    }, [botId, page, pageSize, t]);

    const fetchDetail = useCallback(async (runId) => {
        if (!botId || !runId) return;
        try {
            const params = new URLSearchParams({run_id: runId, id: botId});
            const res = await fetch(`/bot/task/get?${params.toString()}`);
            const data = await res.json();
            if (data.code !== 0) return showToast(data.message || t("request_error"));
            setDetail(data.data);
        } catch (err) {
            showToast(t("request_error") + ": " + err.message);
        }
    }, [botId, t]);

    // 列表加载 + 运行中 2s 轮询
    useEffect(() => {
        if (botId !== null) fetchList();
    }, [botId, page, pageSize]);

    useEffect(() => {
        if (botId === null) return;
        const hasRunning = running.length > 0 || (detail && !isTerminal(detail.state));
        if (!hasRunning) return;
        pollTimer.current = setInterval(() => {
            fetchList();
            if (selectedRunId) fetchDetail(selectedRunId);
        }, 2000);
        return () => clearInterval(pollTimer.current);
        // fetchList/fetchDetail must be deps: their identity changes with page,
        // otherwise the interval keeps the stale closure and re-fetches the old page
    }, [running.length, detail?.state, selectedRunId, fetchList, fetchDetail]);

    // 详情加载。deps 必须含 botId：直接打开/刷新 ?run= 链接时，首次触发时
    // botId 还是 null（fetchDetail 会短路），等 BotSelector 自动选中后再真正拉取
    useEffect(() => {
        if (selectedRunId) {
            setDetail(null);
            if (botId !== null) fetchDetail(selectedRunId);
        } else {
            setDetail(null);
        }
    }, [selectedRunId, botId, fetchDetail]);

    const stateBadge = (state) => {
        const s = STATE_STYLE[state] || STATE_STYLE.PLAN;
        const terminal = isTerminal(state);
        return (
            <span className={`inline-flex items-center gap-1.5 text-xs font-semibold ${s.text}`}>
                {!terminal && <span className={`w-2 h-2 rounded-full ${s.color} animate-pulse`}/>}
                {terminal && <span className={`w-2 h-2 rounded-full ${s.color}`}/>}
                {s.label}
            </span>
        );
    };

    const openDetail = (runId) => setSearchParams(runId ? {run: runId} : {});

    // ---------------- 详情视图 ----------------
    if (selectedRunId) {
        return (
            <div className="p-6 bg-gray-100 min-h-screen">
                {toast.show && (
                    <Toast message={toast.message} type={toast.type} onClose={() => setToast({...toast, show: false})}/>
                )}

                <div className="flex items-center gap-3 mb-6">
                    <button onClick={() => setSearchParams({})}
                            className="text-indigo-600 hover:underline text-sm">← {t("taskflow_back")}</button>
                    <h2 className="text-2xl font-bold text-gray-800">{t("taskflow_detail")}</h2>
                    {detail && stateBadge(detail.state)}
                    {!detail && <span className="text-sm text-gray-400">{t("taskflow_loading")}</span>}
                    {/* 详情视图也要渲染 BotSelector：直接打开/刷新 ?run= 链接时，
                        只有它能（通过缓存）初始化 botId，否则详情请求永远发不出去。
                        切换 bot 不清 run 参数，避免初始化时的自动选中把用户弹回列表 */}
                    <div className="w-72 ml-auto">
                        <BotSelector
                            value={botId}
                            onChange={(bot) => {
                                setBotId(bot.id);
                                setPage(1);
                            }}
                        />
                    </div>
                </div>

                {detail && (
                    <div className="space-y-4">
                        {/* 用户请求 */}
                        <div className="bg-white rounded-lg shadow p-4">
                            <div className="text-xs text-gray-400 mb-1">
                                {t("taskflow_request")} · {detail.user_id} · {formatTime(detail.started_at)}
                                {detail.finished_at && ` · ${formatDuration(detail.started_at, detail.finished_at)}`}
                            </div>
                            <div className="text-gray-800 whitespace-pre-wrap break-words">{detail.content}</div>
                        </div>

                        {/* 流程图 */}
                        <FlowDiagram run={detail}/>

                        <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
                            {/* 计划步骤 */}
                            <div className="bg-white rounded-lg shadow p-4">
                                <h3 className="font-bold text-gray-700 mb-3">{t("taskflow_plan")}</h3>
                                <StepList steps={detail.plan}/>
                            </div>
                            {/* 事件时间线 */}
                            <div className="bg-white rounded-lg shadow p-4">
                                <h3 className="font-bold text-gray-700 mb-3">{t("taskflow_events")}</h3>
                                <EventTimeline events={detail.events}/>
                            </div>
                        </div>

                        {/* 最终结果 / 错误信息 */}
                        {detail.final_result && (
                            <div className="bg-white rounded-lg shadow p-4">
                                <h3 className="font-bold text-gray-700 mb-3">{t("taskflow_final_result")}</h3>
                                <div className="prose prose-sm max-w-none text-gray-800">
                                    <ReactMarkdown>{detail.final_result}</ReactMarkdown>
                                </div>
                            </div>
                        )}
                        {detail.err_msg && (
                            <div className="bg-red-50 border border-red-200 rounded-lg p-4">
                                <h3 className="font-bold text-red-600 mb-2">{t("taskflow_error")}</h3>
                                <p className="text-sm text-red-500 font-mono break-all">{detail.err_msg}</p>
                            </div>
                        )}
                    </div>
                )}
            </div>
        );
    }

    // ---------------- 列表视图 ----------------
    return (
        <div className="p-6 bg-gray-100 min-h-screen">
            {toast.show && (
                <Toast message={toast.message} type={toast.type} onClose={() => setToast({...toast, show: false})}/>
            )}

            <div className="flex justify-between items-center mb-6">
                <h2 className="text-2xl font-bold text-gray-800">{t("taskflow_manage")}</h2>
                <div className="w-72">
                    <BotSelector
                        value={botId}
                        onChange={(bot) => {
                            setBotId(bot.id);
                            setPage(1);
                            setSearchParams({});
                        }}
                    />
                </div>
            </div>

            {/* 运行中 */}
            <div className="bg-white rounded-lg shadow mb-6 overflow-x-auto">
                <div className="px-4 py-3 border-b border-gray-100 font-semibold text-gray-700 flex items-center gap-2">
                    <span className="w-2 h-2 rounded-full bg-green-500 animate-pulse"/>
                    {t("taskflow_running_tasks")} ({running.length})
                </div>
                <table className="min-w-full divide-y divide-gray-100">
                    <thead className="bg-gray-50">
                    <tr>
                        <th className="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase">Run ID</th>
                        <th className="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase">{t("taskflow_request")}</th>
                        <th className="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase">{t("status")}</th>
                        <th className="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase">{t("taskflow_loop_count")}</th>
                        <th className="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase">{t("create_time")}</th>
                        <th className="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase">{t("action")}</th>
                    </tr>
                    </thead>
                    <tbody className="divide-y divide-gray-50">
                    {running.length > 0 ? running.map((run) => (
                        <tr key={run.run_id} className="hover:bg-indigo-50/40">
                            <td className="px-4 py-3 text-xs font-mono text-gray-400">{run.run_id.slice(0, 8)}</td>
                            <td className="px-4 py-3 text-sm text-gray-800 max-w-md truncate">{run.content}</td>
                            <td className="px-4 py-3">{stateBadge(run.state)}</td>
                            <td className="px-4 py-3 text-sm text-gray-600">{run.loop}/{run.max_loops || "-"}</td>
                            <td className="px-4 py-3 text-sm text-gray-500">{formatTime(run.started_at)}</td>
                            <td className="px-4 py-3">
                                <button onClick={() => openDetail(run.run_id)}
                                        className="text-indigo-600 hover:underline text-sm">{t("taskflow_view")}</button>
                            </td>
                        </tr>
                    )) : (
                        <tr>
                            <td colSpan="6" className="px-4 py-6 text-center text-sm text-gray-400">
                                {t("taskflow_no_running")}
                            </td>
                        </tr>
                    )}
                    </tbody>
                </table>
            </div>

            {/* 历史 */}
            <div className="bg-white rounded-lg shadow overflow-x-auto">
                <div className="px-4 py-3 border-b border-gray-100 font-semibold text-gray-700">
                    {t("taskflow_history")}
                </div>
                <table className="min-w-full divide-y divide-gray-100">
                    <thead className="bg-gray-50">
                    <tr>
                        <th className="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase">Run ID</th>
                        <th className="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase">{t("taskflow_request")}</th>
                        <th className="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase">{t("status")}</th>
                        <th className="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase">{t("taskflow_loop_count")}</th>
                        <th className="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase">{t("create_time")}</th>
                        <th className="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase">{t("taskflow_duration")}</th>
                        <th className="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase">{t("action")}</th>
                    </tr>
                    </thead>
                    <tbody className="divide-y divide-gray-50">
                    {history.length > 0 ? history.map((run) => (
                        <tr key={run.run_id} className="hover:bg-indigo-50/40">
                            <td className="px-4 py-3 text-xs font-mono text-gray-400">{run.run_id.slice(0, 8)}</td>
                            <td className="px-4 py-3 text-sm text-gray-800 max-w-md truncate">{run.content}</td>
                            <td className="px-4 py-3">{stateBadge(run.state)}</td>
                            <td className="px-4 py-3 text-sm text-gray-600">{run.loop}</td>
                            <td className="px-4 py-3 text-sm text-gray-500">{formatTime(run.started_at)}</td>
                            <td className="px-4 py-3 text-sm text-gray-500">
                                {formatDuration(run.started_at, run.finished_at)}
                            </td>
                            <td className="px-4 py-3">
                                <button onClick={() => openDetail(run.run_id)}
                                        className="text-indigo-600 hover:underline text-sm">{t("taskflow_view")}</button>
                            </td>
                        </tr>
                    )) : (
                        <tr>
                            <td colSpan="7" className="px-4 py-6 text-center text-sm text-gray-400">
                                {t("taskflow_no_history")}
                            </td>
                        </tr>
                    )}
                    </tbody>
                </table>
            </div>

            <Pagination page={page} pageSize={pageSize} total={total} onPageChange={setPage}/>
        </div>
    );
}

export default TaskFlow;
