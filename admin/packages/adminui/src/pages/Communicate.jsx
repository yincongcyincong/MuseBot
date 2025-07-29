import React, { useEffect, useRef, useState } from "react";
import ReactMarkdown from "react-markdown";
import BotSelector from "../components/BotSelector";
import { Copy, Image as ImageIcon } from 'lucide-react';
import Toast from "../components/Toast";

function Communicate() {
    const [botId, setBotId] = useState(null);
    const [input, setInput] = useState("");
    const [messages, setMessages] = useState([]);
    const [loading, setLoading] = useState(false);
    const [chatPage, setChatPage] = useState(1);
    const [hasMoreHistory, setHasMoreHistory] = useState(true);
    const [toast, setToast] = useState(null);
    const [mediaFile, setMediaFile] = useState(null); // base64 image or video
    const [mediaPreview, setMediaPreview] = useState(null);

    const messageEndRef = useRef(null);
    const chatContainerRef = useRef(null);

    useEffect(() => {
        if (botId !== null) {
            setMessages([]);
            setChatPage(1);
            setHasMoreHistory(true);
            fetchChatMessages(botId, 1, true);
        }
    }, [botId]);

    useEffect(() => {
        if (botId !== null && messages.length > 0 && chatPage === 1) {
            const timer = setTimeout(() => {
                scrollToBottom();
            }, 50);
            return () => clearTimeout(timer);
        }
    }, [messages, botId, chatPage]);

    const fetchChatMessages = async (currentBotId, page, isInitialLoad = false) => {
        if (loading || !hasMoreHistory) return;
        setLoading(true);
        try {
            const res = await fetch(`/bot/admin/chat?id=${currentBotId}&page=${page}`);
            const data = await res.json();
            if (data.code !== 0) {
                setToast({ message: data.message || "Failed to fetch chat record", type: "error" });
                return;
            }
            const historyList = data?.data?.list || [];
            setHasMoreHistory(historyList.length > 0);
            const formattedHistory = historyList.reverse().flatMap(msg => [
                { role: "user", content: msg.question },
                { role: "assistant", content: msg.answer }
            ]);
            setMessages(prev => [...formattedHistory, ...prev]);

            if (!isInitialLoad && chatContainerRef.current) {
                const prevScrollHeight = chatContainerRef.current.scrollHeight;
                setTimeout(() => {
                    const newScrollHeight = chatContainerRef.current.scrollHeight;
                    chatContainerRef.current.scrollTop = newScrollHeight - prevScrollHeight;
                }, 0);
            }
        } catch (err) {
            setHasMoreHistory(false);
            setToast({ message: "Error fetching chat history!", type: "error" });
        } finally {
            setLoading(false);
        }
    };

    const scrollToBottom = () => {
        messageEndRef.current?.scrollIntoView({ behavior: "smooth" });
    };

    const handleChatScroll = () => {
        const container = chatContainerRef.current;
        if (!container || loading || !hasMoreHistory) return;
        if (container.scrollTop === 0) {
            const nextPage = chatPage + 1;
            setChatPage(nextPage);
            fetchChatMessages(botId, nextPage);
        }
    };

    const handleSendPrompt = async () => {
        if (!input.trim() && !mediaFile) return;
        const userPrompt = input.trim();
        const formData = new FormData();
        console.log(mediaFile);
        if (mediaFile) formData.append("file", mediaFile);

        setMessages(prev => [...prev, { role: "user", content: userPrompt}]);
        setInput("");
        setMediaFile(null);
        setLoading(true);

        try {
            const response = await fetch(`/bot/communicate?id=${botId}&prompt=${encodeURIComponent(userPrompt)}`, {
                method: "POST",
                body: formData,
            });

            if (!response.ok) throw new Error("Request failed");

            const reader = response.body.getReader();
            const decoder = new TextDecoder("utf-8");

            let assistantReply = "";
            setMessages(prev => [...prev, { role: "assistant", content: "" }]);

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
            }
        } catch (err) {
            setMessages(prev => [...prev, { role: "system", content: "Error: Could not get a response." }]);
            setToast({ message: "Failed to get bot response.", type: "error" });
        } finally {
            setLoading(false);
            scrollToBottom();
        }
    };

    const handleFileUpload = (e) => {
        const file = e.target.files?.[0];
        if (!file) return;

        setMediaFile(file);

        const reader = new FileReader();
        reader.onload = () => {
            setMediaPreview(reader.result);
        };
        reader.readAsDataURL(file);
    };

    const handleKeyDown = (e) => {
        if (e.key === "Enter" && !e.shiftKey && !e.nativeEvent.isComposing) {
            e.preventDefault();
            handleSendPrompt();
        }
    };

    const handleCopyClick = (text) => {
        navigator.clipboard.writeText(text).then(() => {
            setToast({ message: "Copied to clipboard!", type: "success" });
        }).catch(err => {
            setToast({ message: "Failed to copy!", type: "error" });
        });
    };

    const handleCloseToast = () => setToast(null);

    return (
        <div className="p-6 bg-gray-100 min-h-screen">
            {toast && <Toast message={toast.message} type={toast.type} onClose={handleCloseToast} />}
            <div className="flex space-x-4 mb-4 max-w-4xl min-w-[90px]">
                <BotSelector value={botId} onChange={(bot) => setBotId(bot.id)} />
            </div>
            <div className="flex h-[70vh] bg-white shadow rounded-lg overflow-hidden">
                <div className="w-full flex flex-col">
                    <div ref={chatContainerRef} onScroll={handleChatScroll} className="flex-1 p-4 overflow-y-auto space-y-4 flex flex-col">
                        {messages.map((msg, idx) => (
                            <div key={idx} className={`relative max-w-xl px-4 py-2 rounded-lg shadow flex flex-col ${msg.role === "user" ? "bg-blue-100 self-end ml-auto" : "bg-gray-100 self-start mr-auto"}`}>
                                {msg.content && msg.content.startsWith("data:image/") ? (
                                    <img src={msg.content} alt="media" className="rounded max-w-xs" />
                                ) : msg.content.startsWith("data:video/") ? (
                                    <video controls src={msg.content} className="rounded max-w-xs" />
                                ) : (
                                    <ReactMarkdown className="text-sm prose prose-sm max-w-none whitespace-pre-wrap mt-1">
                                        {msg.content}
                                    </ReactMarkdown>
                                )}
                            </div>
                        ))}
                        {loading && chatPage > 1 && <div className="text-center text-gray-500 py-2">Loading more history...</div>}
                        <div ref={messageEndRef} />
                    </div>
                    <div className="border-t p-4 relative">
                        <label className="absolute left-4 top-6 cursor-pointer">
                            <ImageIcon size={20} />
                            <input type="file" accept="image/*,video/*" hidden onChange={handleFileUpload} />
                        </label>
                        <textarea
                            rows={2}
                            className="w-full pl-10 border rounded p-2 focus:outline-none focus:ring resize-none"
                            placeholder="Type your message..."
                            value={input}
                            onChange={(e) => setInput(e.target.value)}
                            onKeyDown={handleKeyDown}
                        />
                        <button
                            onClick={handleSendPrompt}
                            disabled={loading || (!input.trim() && !mediaFile)}
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
