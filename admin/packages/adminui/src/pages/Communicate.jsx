import React, { useEffect, useRef, useState } from "react";
import ReactMarkdown from "react-markdown";
import BotSelector from "../components/BotSelector";

function Communicate() {
    const [botId, setBotId] = useState(null);
    const [chatHistory, setChatHistory] = useState([]);
    const [chatPage, setChatPage] = useState(1);
    const [hasMoreHistory, setHasMoreHistory] = useState(true);

    const [input, setInput] = useState("");
    const [messages, setMessages] = useState([]);
    const [loading, setLoading] = useState(false);

    const messageEndRef = useRef(null);
    const historyRef = useRef(null);

    useEffect(() => {
        if (botId !== null) {
            setChatHistory([]);
            setChatPage(1);
            fetchChatHistory(1);
        }
    }, [botId]);

    const fetchChatHistory = async (page = 1) => {
        try {
            const res = await fetch(`/bot/admin/chat?id=${botId}&page=${page}`);
            const data = await res.json();
            if (data.code !== 0) return alert(data.message || "Failed to fetch chat record");

            const list = data?.data?.list || [];
            if (page === 1) {
                setChatHistory(list);
            } else {
                setChatHistory(prev => [...prev, ...list]);
            }
            setHasMoreHistory(list.length > 0);
        } catch (err) {
            console.error("Failed to fetch chat history:", err);
        }
    };

    const handleHistoryScroll = () => {
        const div = historyRef.current;
        if (!div || loading) return;

        if (div.scrollTop + div.clientHeight >= div.scrollHeight - 20 && hasMoreHistory) {
            const nextPage = chatPage + 1;
            setChatPage(nextPage);
            fetchChatHistory(nextPage);
        }
    };

    const scrollToBottom = () => {
        messageEndRef.current?.scrollIntoView({ behavior: "smooth" });
    };

    const handleSendPrompt = async () => {
        if (!input.trim()) return;
        const userPrompt = input.trim();

        setMessages(prev => [...prev, { role: "user", content: userPrompt }]);
        setInput("");
        setLoading(true);

        try {
            const response = await fetch(`/bot/communicate?id=${botId}&prompt=${encodeURIComponent(userPrompt)}`, {
                method: "POST",
                headers: {
                    "Content-Type": "application/json",
                },
                body: JSON.stringify({ prompt: userPrompt }),
            });

            if (!response.ok) throw new Error("SSE failed");

            const reader = response.body.getReader();
            const decoder = new TextDecoder("utf-8");

            let assistantReply = "";
            const newMsg = { role: "assistant", content: "" };
            setMessages(prev => [...prev, newMsg]);

            while (true) {
                const { done, value } = await reader.read();
                if (done) break;

                const chunk = decoder.decode(value, { stream: true });
                assistantReply += chunk;

                setMessages(prev => {
                    const updated = [...prev];
                    updated[updated.length - 1] = { role: "assistant", content: assistantReply };
                    return updated;
                });

                scrollToBottom();
            }
        } catch (err) {
            console.error("SSE error:", err);
        } finally {
            setLoading(false);
            scrollToBottom();
        }
    };

    const handleKeyDown = (e) => {
        if (e.key === "Enter" && !e.shiftKey) {
            e.preventDefault();
            handleSendPrompt();
        }
    };

    return (
        <div className="p-6 bg-gray-100 min-h-screen">
            <div className="flex space-x-4 mb-4 max-w-4xl">
                <BotSelector
                    value={botId}
                    onChange={(bot) => {
                        setBotId(bot.id);
                        setMessages([]);
                    }}
                />
            </div>

            <div className="flex h-[70vh] bg-white shadow rounded-lg overflow-hidden">
                {/* 左侧历史 */}
                <div
                    ref={historyRef}
                    className="w-1/3 border-r p-4 overflow-y-auto bg-gray-50"
                    onScroll={handleHistoryScroll}
                >
                    <h3 className="text-lg font-semibold mb-2">History</h3>
                    {chatHistory.length > 0 ? (
                        <ul className="space-y-2">
                            {chatHistory.map((msg, idx) => (
                                <li key={idx} className="text-sm text-gray-700 border p-2 rounded">
                                    <strong>User:</strong> {msg.question}<br />
                                    <strong>Bot:</strong> <ReactMarkdown className="prose prose-sm">{msg.answer}</ReactMarkdown>
                                </li>
                            ))}
                        </ul>
                    ) : (
                        <p className="text-gray-500 text-sm">No history available.</p>
                    )}
                </div>

                {/* 右侧对话区 */}
                <div className="w-2/3 flex flex-col">
                    <div className="flex-1 p-4 overflow-y-auto space-y-4">
                        {messages.map((msg, idx) => (
                            <div
                                key={idx}
                                className={`max-w-xl px-4 py-2 rounded-lg shadow ${
                                    msg.role === "user" ? "bg-blue-100 self-end" : "bg-gray-100 self-start"
                                }`}
                            >
                                <ReactMarkdown className="text-sm prose prose-sm max-w-none whitespace-pre-wrap">
                                    {msg.content}
                                </ReactMarkdown>
                            </div>
                        ))}
                        <div ref={messageEndRef} />
                    </div>

                    <div className="border-t p-4">
                        <textarea
                            rows={2}
                            className="w-full border rounded p-2 focus:outline-none focus:ring resize-none"
                            placeholder="Type your message..."
                            value={input}
                            onChange={(e) => setInput(e.target.value)}
                            onKeyDown={handleKeyDown}
                        />
                        <button
                            onClick={handleSendPrompt}
                            disabled={loading || !input.trim()}
                            className="mt-2 px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 disabled:opacity-50"
                        >
                            {loading ? "Sending..." : "Send"}
                        </button>
                    </div>
                </div>
            </div>
        </div>
    );
}

export default Communicate;
