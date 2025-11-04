import React, {useEffect, useRef, useState} from "react";
import ReactMarkdown from "react-markdown";
import BotSelector from "../components/BotSelector";
import Toast from "../components/Toast";
import Modal from "../components/Modal";
import {ArrowUp, Circle, Copy, Mic, Check, Image as ImageIcon} from "lucide-react";
import {useTranslation} from "react-i18next";

function Communicate() {
    const [botId, setBotId] = useState(null);
    const [input, setInput] = useState("");
    const [messages, setMessages] = useState([]);
    const [loading, setLoading] = useState(false);
    const [chatPage, setChatPage] = useState(1);
    const [hasMoreHistory, setHasMoreHistory] = useState(true);
    const [toast, setToast] = useState(null);
    const [File, setMediaFile] = useState(null);
    const [mediaPreview, setMediaPreview] = useState(null);
    const [modalVisible, setModalVisible] = useState(false);
    const [modalMedia, setModalMedia] = useState(null);

    const [recording, setRecording] = useState(false);
    const mediaRecorderRef = useRef(null);
    const recordedChunksRef = useRef([]);

    const messageEndRef = useRef(null);
    const chatContainerRef = useRef(null);

    const { t } = useTranslation();

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

    useEffect(() => {
        const handleEsc = (e) => {
            if (e.key === "Escape") {
                setModalVisible(false);
            }
        };
        window.addEventListener("keydown", handleEsc);
        return () => window.removeEventListener("keydown", handleEsc);
    }, []);

    const scrollToBottom = () => {
        messageEndRef.current?.scrollIntoView({behavior: "smooth"});
    };

    const fetchChatMessages = async (currentBotId, page, isInitialLoad = false) => {
        if (loading || !hasMoreHistory) return;
        setLoading(true);
        try {
            const res = await fetch(`/bot/admin/chat?id=${currentBotId}&page=${page}`);
            const data = await res.json();
            if (data.code !== 0) {
                setToast({message: data.message || "Failed to fetch chat record", type: "error"});
                return;
            }
            const historyList = data?.data?.list || [];
            setHasMoreHistory(historyList.length > 0);
            const formattedHistory = historyList.reverse().flatMap(msg => [
                {role: "user", content: msg.question, media: msg.content},
                {role: "assistant", content: msg.answer, media: ""}
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
            setToast({message: "Error fetching chat history!", type: "error"});
        } finally {
            setLoading(false);
        }
    };

    const handleSendPrompt = async () => {
        if (loading) return;
        if (!input.trim() && !mediaFile) return;

        const userPrompt = input.trim();
        const formData = new FormData();
        if (mediaFile) formData.append("file", mediaFile);

        if (userPrompt === "/clear") {
            setMessages([]);
        }

        setMessages(prev => [...prev, {role: "user", content: userPrompt, media: mediaPreview}]);
        setInput("");
        setMediaFile(null);
        setMediaPreview(null);
        setLoading(true);
        let assistantReply = "";
        setMessages(prev => [...prev, {role: "assistant", content: "🤔thinking...", media: ""}]);

        try {
            const response = await fetch(`/bot/communicate?id=${botId}&prompt=${encodeURIComponent(userPrompt)}`, {
                method: "POST",
                body: formData,
            });

            if (!response.ok) throw new Error("Request failed");

            const reader = response.body.getReader();
            const decoder = new TextDecoder("utf-8");

            while (true) {
                const {done, value} = await reader.read();
                if (done) break;
                const chunk = decoder.decode(value, {stream: true});
                assistantReply += chunk;
                setMessages(prev => {
                    const updated = [...prev];
                    updated[updated.length - 1] = {role: "assistant", content: assistantReply, media: ""};
                    return updated;
                });
            }
        } catch (err) {
            setMessages(prev => [...prev, {role: "system", content: "Error: Could not get a response."}]);
            setToast({message: "Failed to get bot response.", type: "error"});
        } finally {
            setLoading(false);
            scrollToBottom();
        }
    };

    const startRecording = async () => {
        try {
            const stream = await navigator.mediaDevices.getUserMedia({audio: true});
            mediaRecorderRef.current = new MediaRecorder(stream);
            recordedChunksRef.current = [];

            mediaRecorderRef.current.ondataavailable = (event) => {
                if (event.data.size > 0) {
                    recordedChunksRef.current.push(event.data);
                }
            };

            mediaRecorderRef.current.onstop = () => {
                const audioBlob = new Blob(recordedChunksRef.current, {type: "audio/webm"});
                const audioFile = new File([audioBlob], "recording.webm", {type: "audio/webm"});
                setMediaFile(audioFile);

                const reader = new FileReader();
                reader.onloadend = () => {
                    setMediaPreview(reader.result);
                };
                reader.readAsDataURL(audioBlob);
            };

            mediaRecorderRef.current.start();
            setRecording(true);
        } catch (error) {
            setToast({message: "Microphone access denied.", type: "error"});
        }
    };

    const stopRecording = () => {
        if (mediaRecorderRef.current) {
            mediaRecorderRef.current.stop();
            setRecording(false);
        }
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

    const handleCopyClick = async (text) => {
        try {
            if (text.startsWith("data:image/")) {
                const res = await fetch(text);
                const blob = await res.blob();
                await navigator.clipboard.write([
                    new ClipboardItem({[blob.type]: blob})
                ]);
                setToast({message: "Image copied to clipboard!", type: "success"});
            } else {
                await navigator.clipboard.writeText(text);
                setToast({message: "Text copied to clipboard!", type: "success"});
            }
        } catch (err) {
            setToast({message: "Failed to copy!", type: "error"});
        }
    };

    const handleMediaClick = (media) => {
        setModalMedia(media);
        setModalVisible(true);
    };

    const renderContent = (msg) => {
        if (!msg.content) return null;

        if (msg.content.startsWith("data:image/")) {
            return (
                <img
                    src={msg.content}
                    alt="image"
                    className="max-w-[100px] max-h-[100px] cursor-pointer"
                    onClick={() => handleMediaClick(msg.content)}
                />
            );
        }

        if (msg.content.startsWith("data:video/")) {
            return (
                <video
                    controls
                    src={msg.content}
                    className="max-w-[100px] max-h-[100px] cursor-pointer"
                    onClick={() => handleMediaClick(msg.content)}
                />
            );
        }

        return (
            <ReactMarkdown className="text-sm prose prose-sm max-w-none whitespace-pre-wrap mt-1">
                {msg.content}
            </ReactMarkdown>
        );
    };

    const renderMedia = (msg) => {
        if (!msg.media) return null;

        if (msg.media.startsWith("data:image/")) {
            return (
                <img
                    src={msg.media}
                    alt="media"
                    className="max-w-[100px] max-h-[100px] cursor-pointer"
                    onClick={() => handleMediaClick(msg.media)}
                />
            );
        }

        if (msg.media.startsWith("data:video/") || msg.media.startsWith("data:audio/")) {
            return (
                <video
                    controls
                    src={msg.media}
                    className="max-w-[100px] max-h-[100px] cursor-pointer"
                    onClick={() => handleMediaClick(msg.media)}
                />
            );
        }

        return null;
    };

    const handleCloseToast = () => setToast(null);

    return (
        <div className="p-6 bg-gray-100 min-h-screen">
            {toast && <Toast message={toast.message} type={toast.type} onClose={handleCloseToast}/>}
            <div className="flex space-x-4 mb-4 max-w-4xl min-w-[90px]">
                <BotSelector value={botId} onChange={(bot) => setBotId(bot.id)}/>
            </div>
            <div className="flex h-[70vh] bg-white shadow rounded-lg overflow-hidden">
                <div className="w-full flex flex-col">
                    <div
                        ref={chatContainerRef}
                        onScroll={handleChatScroll}
                        className="flex-1 p-4 overflow-y-auto space-y-4 flex flex-col"
                    >
                        {messages.map((msg, idx) => (
                            <div key={idx} className="relative flex flex-col items-start">
                                <div
                                    className={`max-w-xl px-4 py-2 rounded-lg shadow flex flex-col ${
                                        msg.role === "user"
                                            ? "bg-blue-100 self-end ml-auto"
                                            : "bg-gray-100 self-start mr-auto"
                                    }`}
                                >
                                    {renderContent(msg)}
                                    {renderMedia(msg)}
                                </div>

                                {(msg.content || msg.media) && (
                                    <button
                                        onClick={() => handleCopyClick(msg.content || msg.media)}
                                        className={`ml-2 mt-1 text-gray-400 hover:text-gray-600 ${
                                            msg.role === "user" ? "self-end" : "self-start"
                                        }`}
                                        title="Copy"
                                    >
                                        <Copy size={16}/>
                                    </button>
                                )}
                            </div>
                        ))}
                        {loading && chatPage > 1 && (
                            <div className="text-center text-gray-500 py-2">Loading more history...</div>
                        )}
                        <div ref={messageEndRef}/>
                    </div>

                    <div className="relative">
                        <div className="border-t p-8">
                            {mediaPreview && (
                                <div className="mb-2">
                                    {mediaPreview.startsWith("data:image/") ? (
                                        <img src={mediaPreview} alt="preview"
                                             className="max-w-[50px] max-h-[50px] rounded"/>
                                    ) : mediaPreview.startsWith("data:video/") || mediaPreview.startsWith("data:audio/") ? (
                                        <video controls src={mediaPreview}
                                               className="max-w-[50px] max-h-[50px] rounded"/>
                                    ) : null}
                                </div>
                            )}
                            <textarea
                                rows={2}
                                className="w-full border rounded p-2 focus:outline-none focus:ring resize-none"
                                placeholder={t("message_placeholder")}
                                value={input}
                                onChange={(e) => setInput(e.target.value)}
                                onKeyDown={handleKeyDown}
                                disabled={loading}
                            />
                        </div>

                        <label className="absolute bottom-0 left-4 p-2 z-10 cursor-pointer">
                            <ImageIcon size={22}/>
                            <input type="file" accept="image/*,video/*" hidden onChange={handleFileUpload}/>
                        </label>

                        <button
                            onClick={recording ? stopRecording : startRecording}
                            className="absolute bottom-0 left-16 p-2 z-10 cursor-pointer"
                            title={recording ? "Stop Recording" : "Start Recording"}
                        >
                            {recording ? <Check size={22} /> : <Mic size={22} />}
                        </button>

                        <button
                            onClick={handleSendPrompt}
                            disabled={loading}
                            className={`absolute bottom-0 right-4 p-2 rounded-full z-10 ${
                                loading ? "opacity-50 cursor-not-allowed" : ""
                            }`}
                        >
                            {loading ? <Circle size={22}/> : <ArrowUp size={22}/>}
                        </button>
                    </div>
                </div>
            </div>

            <Modal visible={modalVisible} title="Preview" onClose={() => setModalVisible(false)}>
                {modalMedia?.startsWith("data:image/") && (
                    <img src={modalMedia} alt="preview" className="max-w-full max-h-[80vh] mx-auto"/>
                )}
                {(modalMedia?.startsWith("data:video/") || modalMedia?.startsWith("data:audio/")) && (
                    <video src={modalMedia} controls className="max-w-full max-h-[80vh] mx-auto"/>
                )}
            </Modal>
        </div>
    );
}

export default Communicate;
