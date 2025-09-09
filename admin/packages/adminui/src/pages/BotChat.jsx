import React, { useEffect, useState } from "react";
import Pagination from "../components/Pagination";
import BotSelector from "../components/BotSelector";
import ReactMarkdown from 'react-markdown';
import Modal from "../components/Modal";
import Editor from "@monaco-editor/react";
import Toast from "../components/Toast.jsx";

function BotRecordsPage() {
    const [botId, setBotId] = useState(null);
    const [userIdSearch, setUserIdSearch] = useState("");
    const [records, setRecords] = useState([]);
    const [page, setPage] = useState(1);
    const [pageSize] = useState(10);
    const [total, setTotal] = useState(0);

    const [toast, setToast] = useState({show: false, message: "", type: "error"});
    const showToast = (message, type = "error") => {
        setToast({show: true, message, type});
    };

    const [isModalOpen, setIsModalOpen] = useState(false);
    const [rawConfigText, setRawConfigText] = useState(
        JSON.stringify(
            { user_id: "", records: [{ question: "", answer: "" }] },
            null,
            2
        )
    );

    useEffect(() => {
        if (botId !== null) {
            fetchBotRecords();
        }
    }, [botId, page, userIdSearch]);

    const fetchBotRecords = async () => {
        try {
            const params = new URLSearchParams({
                id: botId,
                page,
                pageSize,
            });
            if (userIdSearch.trim() !== "") {
                params.append("userId", userIdSearch.trim());
            }
            const res = await fetch(`/bot/record/list?${params.toString()}`);
            const data = await res.json();
            setRecords(data.data.list || []);
            setTotal(data.data.total || 0);
        } catch (err) {
            showToast("Failed to fetch bot records: " + err.message);
        }
    };

    const insertRecords = async () => {
        try {
            const res = await fetch(`/bot/user/insert/records?id=${botId}`, {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: rawConfigText,
            });
            if (!res.ok) throw new Error("Failed to insert records");
            const data = await res.json();
            setIsModalOpen(false);
            if (data.code === 0) {
                showToast("Records inserted successfully", "success");
            } else {
                showToast(data.message);
            }
            await fetchBotRecords();
        } catch (err) {
            showToast("Failed to insert records: " + err.message);
        }
    };

    const handleUserIdSearchChange = (e) => {
        setUserIdSearch(e.target.value);
        setPage(1);
    };

    const handlePageChange = (newPage) => {
        setPage(newPage);
    };

    function renderAnswerContent(answer) {
        if (typeof answer !== 'string' || answer.trim() === '') {
            return null;
        }

        if (answer.startsWith('data:video/')) {
            return (
                <video
                    controls
                    className="max-w-[300px] w-full rounded shadow"
                    src={answer}
                />
            );
        }

        if (answer.startsWith('data:image/')) {
            return (
                <img
                    src={answer}
                    alt="base64 image"
                    className="max-w-[300px] w-full rounded shadow"
                />
            );
        }

        return (
            <ReactMarkdown
                components={{
                    img: ({ node, ...props }) => (
                        <img {...props} className="max-w-[300px] w-full rounded shadow" />
                    ),
                }}
            >
                {answer}
            </ReactMarkdown>
        );
    }

    return (
        <div className="p-6 bg-gray-100 min-h-screen relative">
            {toast.show && (
                <Toast
                    message={toast.message}
                    type={toast.type}
                    onClose={() => setToast({...toast, show: false})}
                />
            )}
            <div className="flex justify-between items-center mb-6">
                <h2 className="text-2xl font-bold text-gray-800">Bot Record History</h2>
                <button
                    onClick={() => setIsModalOpen(true)}
                    className="px-4 py-2 bg-blue-600 text-white rounded-lg shadow hover:bg-blue-700"
                >
                    Insert Record
                </button>
            </div>

            <div className="flex space-x-4 mb-6 max-w-4xl flex-wrap items-end">
                <div className="flex-1 min-w-[200px]">
                    <BotSelector
                        value={botId}
                        onChange={(bot) => {
                            setBotId(bot.id);
                            setPage(1);
                            setUserIdSearch("");
                        }}
                    />
                </div>

                <div className="flex-1 min-w-[200px]">
                    <label className="block font-medium text-gray-700 mb-1">Search User ID:</label>
                    <input
                        type="text"
                        value={userIdSearch}
                        onChange={handleUserIdSearchChange}
                        placeholder="Enter user ID"
                        className="w-full px-4 py-2 border border-gray-300 rounded shadow-sm focus:outline-none focus:ring focus:border-blue-400"
                    />
                </div>
            </div>

            <div className="overflow-x-auto rounded-lg shadow">
                <table className="min-w-full bg-white divide-y divide-gray-200">
                    <thead className="bg-gray-50">
                    <tr>
                        {["UserID", "Question", "Answer", "Token", "Status", "Model", "Created", "Updated"].map(title => (
                            <th
                                key={title}
                                className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider"
                            >
                                {title}
                            </th>
                        ))}
                    </tr>
                    </thead>
                    <tbody className="divide-y divide-gray-100">
                    {records.length > 0 ? (
                        records.map(record => (
                            <tr key={record.id} className="hover:bg-gray-50">
                                <td className="px-6 py-4 text-sm text-gray-800">{record.user_id}</td>
                                <td className="px-6 py-4 text-sm text-gray-800">{record.question}</td>
                                <td className="px-6 py-4 text-sm text-gray-800">
                                    {renderAnswerContent(record.answer)}
                                </td>
                                <td className="px-6 py-4 text-sm text-gray-800">{record.token}</td>
                                <td className="px-6 py-4 text-sm text-gray-600">{record.is_deleted ? "Deleted" : "Active"}</td>
                                <td className="px-6 py-4 text-sm text-gray-800">
                                    {record.mode}
                                </td>
                                <td className="px-6 py-4 text-sm text-gray-800">
                                    {new Date(record.create_time * 1000).toLocaleString()}
                                </td>
                                <td className="px-6 py-4 text-sm text-gray-800">
                                    {record.update_time != 0 ? new Date(record.update_time * 1000).toLocaleString() : "-"}
                                </td>
                            </tr>
                        ))
                    ) : (
                        <tr>
                            <td colSpan={6} className="text-center py-6 text-gray-500">
                                No records found.
                            </td>
                        </tr>
                    )}
                    </tbody>
                </table>
            </div>

            <Pagination page={page} pageSize={pageSize} total={total} onPageChange={handlePageChange} />

            <Modal visible={isModalOpen} onClose={() => setIsModalOpen(false)} title={"Insert Record"}>
                <div className="mb-4">
                    <Editor
                        height="300px"
                        defaultLanguage="json"
                        value={rawConfigText}
                        onChange={(value) => setRawConfigText(value ?? "")}
                        options={{
                            minimap: { enabled: false },
                            fontSize: 14,
                            automaticLayout: true,
                            formatOnPaste: true,
                            formatOnType: true,
                        }}
                    />
                </div>
                <div className="flex justify-end space-x-2">
                    <button
                        onClick={() => setIsModalOpen(false)}
                        className="bg-gray-300 hover:bg-gray-400 text-gray-800 px-4 py-2 rounded"
                    >
                        Cancel
                    </button>
                    <button
                        onClick={insertRecords}
                        className="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700"
                    >
                        Insert
                    </button>
                </div>
            </Modal>
        </div>
    );
}

export default BotRecordsPage;
