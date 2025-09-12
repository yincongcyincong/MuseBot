import React, {useEffect, useState} from "react";
import BotSelector from "../components/BotSelector";
import {
    CategoryScale,
    Chart as ChartJS,
    Legend,
    LinearScale,
    LineElement,
    PointElement,
    Title,
    Tooltip
} from "chart.js";
import {Line} from "react-chartjs-2";
import {Bot, ClipboardList, Users} from "lucide-react";
import { useTranslation } from 'react-i18next';

ChartJS.register(
    CategoryScale,
    LinearScale,
    PointElement,
    LineElement,
    Title,
    Tooltip,
    Legend
);


function formatHourMinute(timestamp) {
    const date = new Date(parseInt(timestamp, 10) * 1000);
    const year = date.getFullYear();
    const month = (date.getMonth() + 1).toString().padStart(2, '0');
    const day = date.getDate().toString().padStart(2, '0');
    const hours = date.getHours().toString().padStart(2, '0');
    const minutes = date.getMinutes().toString().padStart(2, '0');
    return `${year}-${month}-${day} ${hours}:${minutes}`;
}

export default function DashboardPage() {
    const [selectedBot, setSelectedBot] = useState(null);
    const [dayRange, setDayRange] = useState(7);
    const [dashboardData, setDashboardData] = useState(null);
    const [loading, setLoading] = useState(false);

    const { t } = useTranslation();

    useEffect(() => {
        if (selectedBot) {
            fetchDashboard(selectedBot.id, dayRange);
        }
    }, [selectedBot, dayRange]);

    const fetchDashboard = async (botId, day) => {
        setLoading(true);
        try {
            const res = await fetch(`/bot/dashboard?id=${botId}&day=${day}`);
            const json = await res.json();
            if (json.code === 0) {
                setDashboardData(json.data);
            } else {
                setDashboardData(null);
                console.error("Dashboard API error:", json.message);
            }
        } catch (err) {
            setDashboardData(null);
            console.error("Fetch dashboard failed:", err);
        }
        setLoading(false);
    };

    function formatDurationFromTimestamp(startTimestamp) {
        if (!startTimestamp) return "-";
        const now = Math.floor(Date.now() / 1000);
        const diff = now - parseInt(startTimestamp, 10);
        if (diff < 0) return "-";

        const days = Math.floor(diff / (3600 * 24));
        const hours = Math.floor((diff % (3600 * 24)) / 3600);
        const minutes = Math.floor((diff % 3600) / 60);
        const seconds = diff % 60;

        const parts = [];
        if (days > 0) parts.push(`${days}d`);
        if (hours > 0) parts.push(`${hours}h`);
        if (minutes > 0) parts.push(`${minutes}m`);
        parts.push(`${seconds}s`); // 总是展示秒数

        return parts.join(' ');
    }

    const buildChartData = (dayCountArray, color = "rgb(59 130 246)") => {
        if (!dayCountArray || dayCountArray.length === 0) {
            return {
                labels: [],
                datasets: []
            };
        }
        const sorted = [...dayCountArray].sort(
            (a, b) => parseInt(a.date, 10) - parseInt(b.date, 10)
        );
        return {
            labels: sorted.map(item => formatHourMinute(item.date)),
            datasets: [
                {
                    data: sorted.map(item => item.new_count),
                    fill: false,
                    borderColor: color,
                    backgroundColor: color,
                    tension: 0.2
                }
            ]
        };
    };


    const chartOptions = {
        plugins: {
            legend: {display: false}
        },
        scales: {
            x: {
                ticks: {
                    autoSkip: true,
                    maxTicksLimit: 10
                }
            }
        }
    };

    return (
        <div className="p-6 bg-gray-100 min-h-screen">


            <h2 className="text-2xl font-bold text-gray-800 mb-6">{t('dashboard_name')}</h2>

            <div className="flex space-x-4 mb-6 flex-wrap items-end">
                <div className="flex-1 min-w-[240px] max-w-[400px]">
                    <BotSelector
                        value={selectedBot}
                        onChange={setSelectedBot}
                    />
                </div>
            </div>

            {selectedBot && (
                <>
                    <div className="flex space-x-6 mb-4">
                        <div className="flex-1 bg-white rounded shadow p-4 text-center flex flex-col items-center">
                            <div className="text-gray-500 mb-2 flex items-center justify-center space-x-2">
                                <ClipboardList className="text-blue-600 w-6 h-6"/>
                                <span>{t("message_num")}</span>
                            </div>
                            <div className="text-3xl font-semibold text-blue-700">
                                {loading ? "Loading..." : dashboardData?.record_count ?? "-"}
                            </div>
                        </div>

                        <div className="flex-1 bg-white rounded shadow p-4 text-center flex flex-col items-center">
                            <div className="text-gray-500 mb-2 flex items-center justify-center space-x-2">
                                <Users className="text-green-600 w-6 h-6"/>
                                <span>{t("user_num")}</span>
                            </div>
                            <div className="text-3xl font-semibold text-green-700">
                                {loading ? "Loading..." : dashboardData?.user_count ?? "-"}
                            </div>
                        </div>

                        <div className="flex-1 bg-white rounded shadow p-4 text-center flex flex-col items-center">
                            <div className="text-gray-500 mb-2 flex items-center justify-center space-x-2">
                                <Bot className="text-black-600 w-6 h-6"/>
                                <span>{t("running_time")}</span>
                            </div>
                            <div className="text-3xl font-semibold text-black-700">
                                {loading ? "Loading..." : formatDurationFromTimestamp(dashboardData?.start_time)}
                            </div>
                        </div>
                    </div>

                    <div className="mb-6">
                        <span className="mr-4 font-medium text-gray-700">{t("date_range")}:</span>
                        {[1, 3, 7, 30].map((d) => (
                            <button
                                key={d}
                                onClick={() => setDayRange(d)}
                                className={`mr-2 px-4 py-1 rounded ${
                                    dayRange === d
                                        ? "bg-blue-600 text-white"
                                        : "bg-gray-200 text-gray-700 hover:bg-gray-300"
                                }`}
                            >
                                {d} {t("day")}
                            </button>
                        ))}
                    </div>

                    <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                        <div className="bg-white rounded shadow p-4">
                            <h3 className="text-center font-semibold mb-2 text-gray-800">{t("message_new_num")}</h3>
                            {loading || !dashboardData ? (
                                <div className="text-center text-gray-500 py-16">Loading chart...</div>
                            ) : (
                                <Line
                                    data={buildChartData(dashboardData.record_day_count, "rgb(59 130 246)")}
                                    options={chartOptions}
                                />
                            )}
                        </div>

                        <div className="bg-white rounded shadow p-4">
                            <h3 className="text-center font-semibold mb-2 text-gray-800">{t("user_new_num")}</h3>
                            {loading || !dashboardData ? (
                                <div className="text-center text-gray-500 py-16">Loading chart...</div>
                            ) : (
                                <Line
                                    data={buildChartData(dashboardData.user_day_count, "rgb(22 163 74)")}
                                    options={chartOptions}
                                />
                            )}
                        </div>
                    </div>
                </>
            )}
        </div>
    );
}
