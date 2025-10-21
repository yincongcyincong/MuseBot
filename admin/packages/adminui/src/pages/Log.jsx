import React, { useEffect, useRef, useState } from "react";
import BotSelector from "../components/BotSelector";
import Toast from "../components/Toast";
import { useTranslation } from "react-i18next";

function BotLogPage() {
    const [botId, setBotId] = useState(null);
    const [typ, setTyp] = useState("");
    const [logs, setLogs] = useState([]);
    const [toast, setToast] = useState({ show: false, message: "", type: "error" });
    const eventSourceRef = useRef(null);
    const logRef = useRef(null);
    const shouldAutoScroll = useRef(true);
    const hasFirstScroll = useRef(false);

    const { t } = useTranslation();

    const showToast = (message, type = "error") => {
        setToast({ show: true, message, type });
    };

    useEffect(() => {
        if (!botId) return;

        if (eventSourceRef.current) {
            eventSourceRef.current.close();
        }

        // ✅ 将 typ 参数加入请求 URL
        const url = `http://127.0.0.1:18080/bot/log?id=${botId}&type=${typ}`;
        const es = new EventSource(url);

        es.onmessage = (event) => {
            setLogs((prevLogs) => {
                const newLogs = [...prevLogs, event.data];
                if (newLogs.length > 5000) {
                    return newLogs.slice(newLogs.length - 5000);
                }
                return newLogs;
            });
        };

        es.onerror = (err) => {
            console.error("SSE error:", err);
            showToast("SSE connection error");
            es.close();
        };

        eventSourceRef.current = es;
        hasFirstScroll.current = false;

        return () => es.close();
    }, [botId, typ]);

    // 用户滚动事件
    const handleScroll = () => {
        const el = logRef.current;
        if (!el) return;
        const isAtBottom = el.scrollHeight - el.scrollTop - el.clientHeight < 50;
        shouldAutoScroll.current = isAtBottom;
    };

    // 添加滚动监听
    useEffect(() => {
        const el = logRef.current;
        if (!el) return;
        el.addEventListener("scroll", handleScroll);
        return () => el.removeEventListener("scroll", handleScroll);
    }, []);

    // 自动滚动到底部逻辑
    useEffect(() => {
        const el = logRef.current;
        if (!el) return;

        const scrollToBottom = () => {
            el.scrollTop = el.scrollHeight;
        };

        if (shouldAutoScroll.current || !hasFirstScroll.current) {
            requestAnimationFrame(() => {
                requestAnimationFrame(scrollToBottom);
                hasFirstScroll.current = true;
            });
        }
    }, [logs]);

    // 日志颜色
    const getLevelColor = (level) => {
        switch (level) {
            case "info":
                return "text-green-400";
            case "warn":
                return "text-yellow-400";
            case "error":
                return "text-red-500";
            default:
                return "text-white";
        }
    };

    const renderLogLine = (line, index) => {
        try {
            const obj = JSON.parse(line);
            const colorClass = getLevelColor(obj.level);
            return (
                <div key={index} className={`font-mono ${colorClass}`}>
                    {line}
                </div>
            );
        } catch {
            return (
                <div key={index} className="font-mono text-white">
                    {line}
                </div>
            );
        }
    };

    return (
        <div className="p-6 bg-gray-100 min-h-screen">
            {/* Toast */}
            {toast.show && (
                <Toast
                    message={toast.message}
                    type={toast.type}
                    onClose={() => setToast({ ...toast, show: false })}
                />
            )}

            {/* 标题 */}
            <div className="flex justify-between items-center mb-6">
                <h2 className="text-2xl font-bold text-gray-800">{t("log")}</h2>
            </div>

            <div className="flex space-x-4 mb-6 max-w-4xl flex-wrap">
                <div className="flex-1 min-w-[200px]">
                    <BotSelector
                        value={botId}
                        onChange={(bot) => {
                            setBotId(bot.id);
                            setLogs([]);
                            hasFirstScroll.current = false;
                        }}
                    />
                </div>

                <div className="flex-1 min-w-[200px]">
                    <label className="block font-medium text-gray-700 mb-1">{t("type")}:</label>

                    <select
                        id="logType"
                        className="w-full px-4 py-2 border border-gray-300 rounded shadow-sm focus:outline-none focus:ring focus:border-blue-400"
                        value={typ}
                        onChange={(e) => {
                            const value = e.target.value === "all" ? "" : e.target.value;
                            setTyp(value);
                            setLogs([]);
                        }}
                    >
                        <option value="">All</option>
                        <option value="info">Info</option>
                        <option value="warn">Warn</option>
                        <option value="error">Error</option>
                    </select>
                </div>

            </div>


            <div
                ref={logRef}
                onScroll={handleScroll}
                className="rounded-lg shadow border border-gray-700 overflow-y-auto h-[70vh] p-2 bg-gray-900"
            >
                {logs.map(renderLogLine)}
            </div>
        </div>
    );
}

export default BotLogPage;
