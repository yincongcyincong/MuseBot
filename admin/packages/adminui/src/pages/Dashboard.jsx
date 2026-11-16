import {createElement, useEffect, useState} from "react";
import {Chart as ChartJS, Filler, LinearScale, LineElement, PointElement, Tooltip} from "chart.js";
import {Line} from "react-chartjs-2";
import {Activity, BarChart3, Clock3, MessageSquare, Users} from "lucide-react";
import {useTranslation} from "react-i18next";
import BotSelector from "../components/BotSelector";

ChartJS.register(Filler, LinearScale, PointElement, LineElement, Tooltip);

const cardStyle = "rounded-2xl border border-slate-200 bg-white shadow-sm";
const dayRanges = [1, 3, 7, 30];

function formatUptime(timestamp) {
    const seconds = Math.floor(Date.now() / 1000 - Number(timestamp));
    if (!timestamp || !Number.isFinite(seconds) || seconds < 0) return "—";
    if (seconds >= 86400) return `${Math.floor(seconds / 86400)}d ${Math.floor(seconds % 86400 / 3600)}h`;
    if (seconds >= 3600) return `${Math.floor(seconds / 3600)}h ${Math.floor(seconds % 3600 / 60)}m`;
    return `${Math.floor(seconds / 60)}m ${seconds % 60}s`;
}

function StatCard({icon, label, value, tone, loading}) {
    return (
        <div className={`${cardStyle} flex min-w-0 items-start gap-3 p-4 sm:gap-4 sm:p-5`}>
            <div className={`flex h-10 w-10 shrink-0 items-center justify-center rounded-xl ${tone}`}>
                {createElement(icon, {size: 19, "aria-hidden": true})}
            </div>
            <div className="min-w-0">
                <dt className="text-sm font-medium text-slate-500">{label}</dt>
                <dd className="mt-1 break-words text-2xl font-semibold tracking-tight text-slate-900 tabular-nums xl:text-3xl">
                    {loading ? <span className="block h-7 w-20 animate-pulse rounded bg-slate-100 motion-reduce:animate-none"/> : value}
                </dd>
            </div>
        </div>
    );
}

function TrendCard({title, entries, color, locale, days, loading, label}) {
    const formatCount = new Intl.NumberFormat(locale);
    const formatTick = new Intl.DateTimeFormat(locale, days === 1
        ? {hour: "2-digit", minute: "2-digit"}
        : {month: "short", day: "numeric"});
    const formatDate = new Intl.DateTimeFormat(locale, days === 30
        ? {month: "short", day: "numeric"}
        : {month: "short", day: "numeric", hour: "2-digit", minute: "2-digit"});
    const step = days <= 3 ? 3600 : days === 7 ? 10800 : 86400;
    const counts = new Map((entries ?? []).map(({date, new_count}) => [Number(date), Number(new_count)]));
    const dates = [...counts.keys()].sort((a, b) => a - b);
    const points = [];
    if (dates.length) {
        const lastBucket = dates.at(-1);
        const firstBucket = Math.max(dates[0], lastBucket - (days * 24 + 24) * 3600);
        for (let date = firstBucket; date <= lastBucket; date += step) {
            points.push({x: date, y: counts.get(date) ?? 0});
        }
    }
    const total = [...counts.values()].reduce((sum, count) => sum + count, 0);
    const data = {datasets: [{
        data: points,
        borderColor: color,
        backgroundColor: ({chart}) => {
            const gradient = chart.ctx.createLinearGradient(0, 0, 0, chart.height);
            gradient.addColorStop(0, `${color}33`);
            gradient.addColorStop(1, `${color}00`);
            return gradient;
        },
        fill: true,
        borderWidth: 2,
        pointRadius: points.length === 1 ? 4 : 0,
        pointHoverRadius: 4,
        pointHitRadius: 12,
    }]};
    const options = {
        animation: false,
        maintainAspectRatio: false,
        interaction: {mode: "nearest", axis: "x", intersect: false},
        plugins: {tooltip: {
            displayColors: false,
            padding: 10,
            callbacks: {
                title: ([item]) => formatDate.format(item.parsed.x * 1000),
                label: (item) => formatCount.format(item.parsed.y),
            },
        }},
        scales: {
            x: {
                type: "linear",
                bounds: "data",
                grid: {display: false},
                border: {display: false},
                ticks: {
                    stepSize: step, autoSkip: true, maxTicksLimit: 5, maxRotation: 0,
                    color: "#64748b", callback: (value) => formatTick.format(value * 1000),
                },
            },
            y: {
                beginAtZero: true,
                border: {display: false},
                grid: {color: "#edf2f7"},
                ticks: {precision: 0, maxTicksLimit: 5, color: "#64748b"},
            },
        },
    };

    return (
        <article className={`${cardStyle} min-w-0 p-4 sm:p-6`}>
            <div className="mb-5 flex items-center justify-between gap-3">
                <h3 className="flex items-center gap-2 text-sm font-semibold text-slate-700">
                    <span className="h-2 w-2 shrink-0 rounded-full" style={{backgroundColor: color}} aria-hidden="true"/>
                    {title}
                </h3>
                {!loading && <span className="text-xl font-semibold text-slate-900 tabular-nums">{formatCount.format(total)}</span>}
            </div>
            <div className="relative h-56 sm:h-64">
                {loading ? (
                    <div className="h-full animate-pulse rounded-xl bg-slate-100 motion-reduce:animate-none"/>
                ) : points.length ? (
                    <Line data={data} options={options} role="img" aria-label={`${title} · ${formatCount.format(total)}`}/>
                ) : (
                    <div className="flex h-full items-center justify-center rounded-xl border border-dashed border-slate-200 text-sm text-slate-500">
                        {label}
                    </div>
                )}
            </div>
        </article>
    );
}

export default function DashboardPage() {
    const [selectedBot, setSelectedBot] = useState(null);
    const [dayRange, setDayRange] = useState(7);
    const [result, setResult] = useState(null);
    const [attempt, setAttempt] = useState(0);
    const {t, i18n} = useTranslation();
    const locale = i18n.resolvedLanguage;
    const botId = selectedBot?.id;
    const query = new URLSearchParams({id: botId, day: dayRange}).toString();
    const requestKey = `${query}&attempt=${attempt}`;
    const current = result?.key === requestKey ? result : null;
    const loading = botId != null && !current;
    const data = current?.data;

    useEffect(() => {
        if (botId == null) return;
        const controller = new AbortController();
        async function loadDashboard() {
            try {
                const response = await fetch(`/bot/dashboard?${query}`, {signal: controller.signal});
                if (!response.ok) throw new Error(`HTTP ${response.status}`);
                const json = await response.json();
                if (json.code !== 0 || !json.data) throw new Error(json.message);
                if (!controller.signal.aborted) setResult({key: requestKey, data: json.data});
            } catch {
                if (!controller.signal.aborted) setResult({key: requestKey, error: true});
            }
        }
        loadDashboard();
        return () => controller.abort();
    }, [botId, query, requestKey]);

    const formatCount = new Intl.NumberFormat(locale);
    const count = (value) => value == null ? "—" : formatCount.format(value);

    return (
        <div className="min-h-full bg-slate-50 px-4 py-5 sm:px-7 sm:py-8">
            <div className="mx-auto max-w-7xl space-y-6">
                <header className={`${cardStyle} flex flex-wrap items-center justify-between gap-4 px-5 py-5 sm:px-6`}>
                    <div className="flex items-center gap-3">
                        <span className="rounded-xl bg-indigo-50 p-2.5 text-indigo-600"><BarChart3 size={23} aria-hidden="true"/></span>
                        <h2 className="text-xl font-semibold tracking-tight text-slate-900 sm:text-2xl">{t("dashboard_name")}</h2>
                    </div>
                    <div className="w-full sm:w-64"><BotSelector value={selectedBot} onChange={setSelectedBot}/></div>
                </header>

                {botId == null ? (
                    <div className={`${cardStyle} px-5 py-12 text-center text-sm text-slate-500`}>{t("dashboard_select_bot")}</div>
                ) : (
                    <div className="space-y-6" aria-busy={loading}>
                        {loading && <span className="sr-only" role="status">{t("dashboard_loading")}</span>}
                        {current?.error && (
                            <div role="alert" className="flex flex-wrap items-center justify-between gap-3 rounded-xl border border-rose-200 bg-rose-50 px-4 py-3 text-sm text-rose-800">
                                {t("dashboard_error")}
                                <button type="button" onClick={() => setAttempt(attempt + 1)} className="rounded-md border border-rose-200 bg-white px-3 py-1 font-medium hover:bg-rose-100">
                                    {t("dashboard_retry")}
                                </button>
                            </div>
                        )}
                        <dl className="grid gap-3 md:grid-cols-3">
                            <StatCard icon={MessageSquare} label={t("message_num")} value={count(data?.record_count)} tone="bg-indigo-50 text-indigo-600" loading={loading}/>
                            <StatCard icon={Users} label={t("user_num")} value={count(data?.user_count)} tone="bg-teal-50 text-teal-700" loading={loading}/>
                            <StatCard icon={Clock3} label={t("running_time")} value={formatUptime(data?.start_time)} tone="bg-slate-100 text-slate-600" loading={loading}/>
                        </dl>
                        <section className="space-y-3">
                            <div className="flex flex-wrap items-center justify-between gap-3">
                                <h2 className="flex items-center gap-2 text-base font-semibold text-slate-900">
                                    <Activity size={18} className="text-slate-500" aria-hidden="true"/>
                                    {t("dashboard_trends")}
                                </h2>
                                <div role="group" aria-label={t("date_range")} className="inline-flex gap-0.5 rounded-xl bg-slate-200/70 p-1">
                                    {dayRanges.map((days) => (
                                        <button key={days} type="button" aria-pressed={dayRange === days} onClick={() => setDayRange(days)}
                                                className={`rounded-lg px-2.5 py-1.5 text-sm font-medium focus-visible:outline focus-visible:outline-2 focus-visible:outline-indigo-600 sm:px-3 ${dayRange === days ? "bg-white text-indigo-700 shadow-sm" : "text-slate-600 hover:bg-white/60"}`}>
                                            {days} {t("day")}
                                        </button>
                                    ))}
                                </div>
                            </div>
                            {!current?.error && (
                                <div className="grid gap-4 lg:grid-cols-2">
                                    <TrendCard title={t("message_new_num")} entries={data?.record_day_count} color="#4f46e5" locale={locale} days={dayRange} loading={loading} label={t("dashboard_empty")}/>
                                    <TrendCard title={t("user_new_num")} entries={data?.user_day_count} color="#0f766e" locale={locale} days={dayRange} loading={loading} label={t("dashboard_empty")}/>
                                </div>
                            )}
                        </section>
                    </div>
                )}
            </div>
        </div>
    );
}
