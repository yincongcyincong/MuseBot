import React, { useEffect, useState } from "react";
import Pagination from "../components/Pagination";
import BotSelector from "../components/BotSelector";
import ReactMarkdown from 'react-markdown'; // Import ReactMarkdown

function BotRecordsPage() {
    const [botId, setBotId] = useState(null);
    const [userIdSearch, setUserIdSearch] = useState("");
    const [records, setRecords] = useState([]);
    const [page, setPage] = useState(1);
    const [pageSize] = useState(10);
    const [total, setTotal] = useState(0);

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
            console.error("Failed to fetch bot records:", err);
        }
    };

    const handleUserIdSearchChange = (e) => {
        setUserIdSearch(e.target.value);
        setPage(1);
    };

    const handlePageChange = (newPage) => {
        setPage(newPage);
    };

    return (
        <div className="p-6 bg-gray-100 min-h-screen">
            <h2 className="text-2xl font-bold text-gray-800 mb-6">Bot Record History</h2>

            <div className="flex space-x-4 mb-6 max-w-4xl flex-wrap items-end">
                {/* Bot Selector */}
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

            {/* 记录表格 */}
            <div className="overflow-x-auto rounded-lg shadow">
                <table className="min-w-full bg-white divide-y divide-gray-200">
                    <thead className="bg-gray-50">
                    <tr>
                        {["UserID", "Question", "Answer", "Token", "Status", "Created At"].map(title => (
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
                                    {/* Use ReactMarkdown for the answer field */}
                                    <ReactMarkdown>{record.answer}</ReactMarkdown>
                                </td>
                                <td className="px-6 py-4 text-sm text-gray-800">{record.token}</td>
                                <td className="px-6 py-4 text-sm text-gray-600">{record.is_deleted ? "Deleted" : "Active"}</td>
                                <td className="px-6 py-4 text-sm text-gray-800">
                                    {new Date(record.create_time * 1000).toLocaleString()}
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
        </div>
    );
}

export default BotRecordsPage;