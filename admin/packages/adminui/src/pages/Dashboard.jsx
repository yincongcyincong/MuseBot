import React, { useState, useEffect } from "react";
import BotSelector from "../components/BotSelector";
import {
    Chart as ChartJS,
    CategoryScale,
    LinearScale,
    PointElement,
    LineElement,
    Title,
    Tooltip,
    Legend
} from "chart.js";
import { Line } from "react-chartjs-2";
import { ClipboardList, Users } from "lucide-react";  // 从lucide-react导入图标

ChartJS.register(
    CategoryScale,
    LinearScale,
    PointElement,
    LineElement,
    Title,
    Tooltip,
    Legend
);

function formatDate(timestamp) {
    const date = new Date(parseInt(timestamp, 10) * 1000);
    return date.toISOString().slice(0, 10);
}

export default function DashboardPage() {
    const [selectedBot, setSelectedBot] = useState(null);
    const [dayRange, setDayRange] = useState(7);
    const [dashboardData, setDashboardData] = useState(null);
    const [loading, setLoading] = useState(false);

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

    const buildChartData = (dayCountArray) => {
        const sorted = [...dayCountArray].sort(
            (a, b) => parseInt(a.date, 10) - parseInt(b.date, 10)
        );
        return {
            labels: sorted.map(item => formatDate(item.date)),
            datasets: [
                {
                    label: "Count",
                    data: sorted.map(item => item.new_count),
                    fill: false,
                    borderColor: "rgb(59 130 246)", // blue-500
                    backgroundColor: "rgb(59 130 246)",
                    tension: 0.2
                }
            ]
        };
    };

    return (
        <div className="p-6 bg-gray-100 min-h-screen">
            <h2 className="text-2xl font-bold text-gray-800 mb-6">Bot Dashboard</h2>

            <div className="flex space-x-4 mb-6 max-w-4xl flex-wrap items-end">
                <div className="flex-1 min-w-[240px] max-w-[400px]">
                    <BotSelector
                        value={selectedBot}
                        onChange={setSelectedBot}
                    />
                </div>
            </div>

            {selectedBot && (
                <>
                    <div className="max-w-4xl flex space-x-6 mb-4">
                        {/* Record Count */}
                        <div className="flex-1 bg-white rounded shadow p-4 text-center flex flex-col items-center">
                            <div className="text-gray-500 mb-2 flex items-center justify-center space-x-2">
                                <ClipboardList className="text-blue-600 w-6 h-6" />
                                <span>Record Count</span>
                            </div>
                            <div className="text-3xl font-semibold text-blue-700">
                                {loading ? "Loading..." : dashboardData?.record_count ?? "-"}
                            </div>
                        </div>

                        {/* User Count */}
                        <div className="flex-1 bg-white rounded shadow p-4 text-center flex flex-col items-center">
                            <div className="text-gray-500 mb-2 flex items-center justify-center space-x-2">
                                <Users className="text-green-600 w-6 h-6" />
                                <span>User Count</span>
                            </div>
                            <div className="text-3xl font-semibold text-green-700">
                                {loading ? "Loading..." : dashboardData?.user_count ?? "-"}
                            </div>
                        </div>
                    </div>

                    <div className="max-w-4xl mb-6">
                        <span className="mr-4 font-medium text-gray-700">Select Date Range:</span>
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
                                {d} Day{d > 1 ? "s" : ""}
                            </button>
                        ))}
                    </div>

                    <div className="max-w-4xl grid grid-cols-1 md:grid-cols-2 gap-6">
                        <div className="bg-white rounded shadow p-4">
                            <h3 className="text-center font-semibold mb-2 text-gray-800">Record New Count</h3>
                            {loading || !dashboardData ? (
                                <div className="text-center text-gray-500 py-16">Loading chart...</div>
                            ) : (
                                <Line data={buildChartData(dashboardData.record_day_count)} />
                            )}
                        </div>

                        <div className="bg-white rounded shadow p-4">
                            <h3 className="text-center font-semibold mb-2 text-gray-800">User New Count</h3>
                            {loading || !dashboardData ? (
                                <div className="text-center text-gray-500 py-16">Loading chart...</div>
                            ) : (
                                <Line data={buildChartData(dashboardData.user_day_count)} />
                            )}
                        </div>
                    </div>
                </>
            )}
        </div>
    );
}
