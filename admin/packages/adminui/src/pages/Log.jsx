import React, { useEffect, useRef, useState } from "react";
import BotSelector from "../components/BotSelector";
import Toast from "../components/Toast";

function BotLogPage() {
    const [botId, setBotId] = useState(null);
    const [logs, setLogs] = useState([]);
    const [toast, setToast] = useState({ show: false, message: "", type: "error" });
    const eventSourceRef = useRef(null);
    const logRef = useRef(null);

    const showToast = (message, type = "error") => {
        setToast({ show: true, message, type });
    };

    useEffect(() => {
        if (!botId) return;

        if (eventSourceRef.current) {
            eventSourceRef.current.close();
        }

        const url = `http://127.0.0.1:18080/bot/log?id=${botId}`;
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
        return () => es.close();
    }, [botId]);

    useEffect(() => {
        if (logRef.current) {
            logRef.current.scrollTop = logRef.current.scrollHeight;
        }
    }, [logs]);

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
                <h2 className="text-2xl font-bold text-gray-800">Log</h2>
            </div>

            {/* BotSelector */}
            <div className="mb-6 max-w-4xl">
                <BotSelector
                    value={botId}
                    onChange={(bot) => {
                        setBotId(bot.id);
                        setLogs([]);
                    }}
                />
            </div>

            {/* 日志展示部分，黑色背景 */}
            <div
                ref={logRef}
                className="rounded-lg shadow border border-gray-700 overflow-y-auto h-[70vh] p-2 bg-gray-900"
            >
                {logs.map(renderLogLine)}
            </div>
        </div>
    );
}

export default BotLogPage;
