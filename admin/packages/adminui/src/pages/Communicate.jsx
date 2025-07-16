import React, { useEffect, useRef, useState } from "react";
import ReactMarkdown from "react-markdown";
import BotSelector from "../components/BotSelector";
import { Copy } from 'lucide-react';
import Toast from "../components/Toast"; // Make sure the path to Toast.js is correct

function Communicate() {
    const [botId, setBotId] = useState(null);
    const [input, setInput] = useState("");
    const [messages, setMessages] = useState([]);
    const [loading, setLoading] = useState(false);
    const [chatPage, setChatPage] = useState(1);
    const [hasMoreHistory, setHasMoreHistory] = useState(true);
    const [toast, setToast] = useState(null);

    const messageEndRef = useRef(null);
    const chatContainerRef = useRef(null);

    // Effect to load initial chat messages when botId changes
    useEffect(() => {
        if (botId !== null) {
            setMessages([]); // Clear current messages
            setChatPage(1); // Reset page to 1
            setHasMoreHistory(true); // Assume there's more history initially
            // Fetch the first page of history for the new bot
            fetchChatMessages(botId, 1, true);
        }
    }, [botId]);

    // Scroll to bottom whenever messages are updated, but only if not loading more history
    useEffect(() => {
        // Only scroll if messages are updated and it's not a history load (which adjusts scroll differently)
        // And ensure it's a valid botId (meaning chat is active)
        if (botId !== null && messages.length > 0 && chatPage === 1) {
            // Add a small delay to ensure DOM has fully updated
            const timer = setTimeout(() => {
                scrollToBottom();
            }, 50); // A slightly longer delay might be more reliable
            return () => clearTimeout(timer);
        }
    }, [messages, botId, chatPage]); // Depend on messages, botId, and chatPage

    // Function to fetch chat messages (can be initial load or more pages)
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
            // Update hasMoreHistory based on whether any new items were fetched
            setHasMoreHistory(historyList.length > 0);

            // Format history messages and add them to the beginning of the messages array
            // The API returns most recent first, so reverse() makes them chronological.
            const formattedHistory = historyList.reverse().flatMap(msg => [
                { role: "user", content: msg.question },
                { role: "assistant", content: msg.answer }
            ]);

            setMessages(prevMessages => {
                // Prepend new history to existing messages
                const newMessages = [...formattedHistory, ...prevMessages];
                return newMessages;
            });

            // Removed initialLoad scrollToBottom here, handled by new useEffect
            if (!isInitialLoad) {
                // For subsequent history loads (pulling up), maintain scroll position
                if (chatContainerRef.current) {
                    const prevScrollHeight = chatContainerRef.current.scrollHeight;
                    setTimeout(() => {
                        const newScrollHeight = chatContainerRef.current.scrollHeight;
                        // Adjust scroll position to keep the top of the previous content in view
                        chatContainerRef.current.scrollTop = newScrollHeight - prevScrollHeight;
                    }, 0);
                }
            }
        } catch (err) {
            console.error("Failed to fetch chat history:", err);
            setHasMoreHistory(false); // Assume no more history on error
            setToast({ message: "Error fetching chat history!", type: "error" });
        } finally {
            setLoading(false); // Reset loading state
        }
    };

    // Scroll to the bottom of the chat
    const scrollToBottom = () => {
        messageEndRef.current?.scrollIntoView({ behavior: "smooth" });
    };

    // Handle scroll event for the chat container to load more history
    const handleChatScroll = () => {
        const container = chatContainerRef.current;
        if (!container || loading || !hasMoreHistory) return;

        // Check if scrolled exactly to the top (or very close to 0)
        // If the current scroll position is at or very near the top
        if (container.scrollTop === 0) {
            const nextPage = chatPage + 1;
            setChatPage(nextPage);
            fetchChatMessages(botId, nextPage);
        }
    };

    // Handler for sending new prompts
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
            console.error("SSE error:", err);
            setMessages(prev => [...prev, { role: "system", content: "Error: Could not get a response." }]);
            setToast({ message: "Failed to get bot response.", type: "error" });
        } finally {
            setLoading(false);
            scrollToBottom(); // Ensure scroll to bottom after final message
        }
    };

    // Keyboard event handler for sending messages with Enter key
    const handleKeyDown = (e) => {
        // Only send message if Enter is pressed, Shift is NOT pressed, AND input method editor (IME) is NOT composing
        if (e.key === "Enter" && !e.shiftKey && !e.nativeEvent.isComposing) {
            e.preventDefault(); // Prevent default new line behavior in textarea
            handleSendPrompt();
        }
    };

    // Function to handle copying message content to clipboard
    const handleCopyClick = (text) => {
        navigator.clipboard.writeText(text)
            .then(() => {
                setToast({ message: "Copied to clipboard!", type: "success" });
            })
            .catch(err => {
                console.error('Failed to copy text: ', err);
                setToast({ message: "Failed to copy!", type: "error" });
            });
    };

    // Function to close the toast
    const handleCloseToast = () => {
        setToast(null);
    };

    return (
        <div className="p-6 bg-gray-100 min-h-screen">
            {toast && (
                <Toast
                    message={toast.message}
                    type={toast.type}
                    onClose={handleCloseToast}
                />
            )}
            <div className="flex space-x-4 mb-4 max-w-4xl min-w-[90px]">
                <BotSelector
                    value={botId}
                    onChange={(bot) => {
                        setBotId(bot.id);
                    }}
                />
            </div>

            <div className="flex h-[70vh] bg-white shadow rounded-lg overflow-hidden">
                <div className="w-full flex flex-col">
                    <div
                        ref={chatContainerRef}
                        className="flex-1 p-4 overflow-y-auto space-y-4 flex flex-col"
                        onScroll={handleChatScroll}
                    >
                        {messages.map((msg, idx) => (
                            <div
                                key={idx}
                                className={`relative max-w-xl px-4 py-2 rounded-lg shadow flex flex-col ${
                                    msg.role === "user" ? "bg-blue-100 self-end ml-auto" : "bg-gray-100 self-start mr-auto"
                                }`}
                            >
                                {/* Message Content */}
                                <div className="mb-1">
                                    {msg.role === "system" ? (
                                        <p className="text-red-500 text-sm">{msg.content}</p>
                                    ) : (
                                        <ReactMarkdown className="text-sm prose prose-sm max-w-none whitespace-pre-wrap">
                                            {msg.content}
                                        </ReactMarkdown>
                                    )}
                                </div>
                                {/* Copy button below the message content */}
                                {! (msg.role === "system") && (
                                    <button
                                        onClick={() => handleCopyClick(msg.content)}
                                        className={`self-end p-1 text-gray-500 hover:text-gray-700 focus:outline-none focus:ring focus:border-blue-300 rounded text-xs`}
                                        title="Copy to clipboard"
                                    >
                                        <Copy size={16} />
                                    </button>
                                )}
                            </div>
                        ))}
                        {/* Loading indicator for history */}
                        {loading && chatPage > 1 && (
                            <div className="text-center text-gray-500 py-2">Loading more history...</div>
                        )}
                        {/* This div is the target for scrolling to the bottom */}
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