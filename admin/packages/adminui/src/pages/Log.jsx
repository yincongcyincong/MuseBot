import React, { useEffect, useRef, useState } from "react";
import Editor from "@monaco-editor/react";
import BotSelector from "../components/BotSelector";
import Toast from "../components/Toast";
import { useTranslation } from "react-i18next";

function BotLogPage() {
    const [botId, setBotId] = useState(null);
    const [logs, setLogs] = useState([]);
    const [toast, setToast] = useState({ show: false, message: "", type: "error" });
    const eventSourceRef = useRef(null);
    const editorRef = useRef(null); // 保存 monaco editor 实例

    const { t } = useTranslation();

    const showToast = (message, type = "error") => {
        setToast({ show: true, message, type });
    };

    useEffect(() => {
        if (!botId) return;

        // 关闭旧连接
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

        return () => {
            es.close();
        };
    }, [botId]);

    // 日志变化后自动滚动到底部
    useEffect(() => {
        if (editorRef.current && logs.length > 0) {
            const editor = editorRef.current;
            const model = editor.getModel();
            const lineCount = model.getLineCount();
            editor.revealLine(lineCount);
        }
    }, [logs]);

    return (
        <div className="p-6 bg-gray-100 min-h-screen">
            {toast.show && (
                <Toast
                    message={toast.message}
                    type={toast.type}
                    onClose={() => setToast({ ...toast, show: false })}
                />
            )}

            <div className="flex justify-between items-center mb-6">
                <h2 className="text-2xl font-bold text-gray-800">Log</h2>
            </div>

            <div className="mb-6 max-w-4xl">
                <BotSelector
                    value={botId}
                    onChange={(bot) => {
                        setBotId(bot.id);
                        setLogs([]); // 切换 bot 时清空日志
                    }}
                />
            </div>

            <div className="rounded-lg shadow border border-gray-300 overflow-hidden">
                <Editor
                    height="70vh"
                    theme="vs-dark"   // 黑色背景
                    defaultLanguage="log"
                    value={logs.join("\n")}
                    options={{
                        readOnly: true,
                        minimap: { enabled: false },
                        fontSize: 14,
                        wordWrap: "on",
                    }}
                    onMount={(editor) => {
                        editorRef.current = editor;
                        if (editor) {
                            const model = editor.getModel();
                            if (model) {
                                const lineCount = model.getLineCount();
                                editor.revealLine(lineCount);
                            }
                        }
                    }}
                />
            </div>
        </div>
    );
}

export default BotLogPage;
