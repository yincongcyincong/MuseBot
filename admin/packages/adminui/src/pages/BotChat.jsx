import React, { useEffect, useState, useRef } from "react";
import Pagination from "../components/Pagination";

function BotRecordsPage() {
    const [bots, setBots] = useState([]);
    const [filteredBots, setFilteredBots] = useState([]);
    const [botId, setBotId] = useState(null);
    const [botSearchText, setBotSearchText] = useState("");
    const [userIdSearch, setUserIdSearch] = useState("");
    const [dropdownOpen, setDropdownOpen] = useState(false);
    const [records, setRecords] = useState([]);
    const [page, setPage] = useState(1);
    const [pageSize] = useState(10);
    const [total, setTotal] = useState(0);

    const wrapperRef = useRef(null);

    useEffect(() => {
        fetchOnlineBots();
    }, []);

    useEffect(() => {
        const lower = botSearchText.toLowerCase();
        setFilteredBots(bots.filter(bot => bot.address.toLowerCase().includes(lower)));
    }, [botSearchText, bots]);

    useEffect(() => {
        if (botId !== null) {
            fetchBotRecords();
        }
    }, [botId, page, userIdSearch]);

    useEffect(() => {
        // 点击外部关闭下拉
        function handleClickOutside(event) {
            if (wrapperRef.current && !wrapperRef.current.contains(event.target)) {
                setDropdownOpen(false);
            }
        }
        document.addEventListener("mousedown", handleClickOutside);
        return () => document.removeEventListener("mousedown", handleClickOutside);
    }, []);

    const fetchOnlineBots = async () => {
        try {
            const res = await fetch("/bot/online");
            const data = await res.json();
            if (data.data && data.data.length > 0) {
                setBots(data.data);
                setFilteredBots(data.data);
                setBotId(data.data[0].id);
                setBotSearchText(data.data[0].address);
            }
        } catch (err) {
            console.error("Failed to fetch online bots:", err);
        }
    };

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
            setRecords(data.data.list);
            setTotal(data.data.total);
        } catch (err) {
            console.error("Failed to fetch bot records:", err);
        }
    };

    const handleSelectBot = (bot) => {
        setBotId(bot.id);
        setBotSearchText(bot.address);
        setDropdownOpen(false);
        setPage(1);
        setUserIdSearch("");
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

            {/* 一行布局：Bot搜索+选择 + UserId搜索 */}
            <div className="flex space-x-4 mb-6 max-w-4xl flex-wrap items-end">
                {/* Bot搜索+选择合一 */}
                <div className="relative flex-1 min-w-[200px]" ref={wrapperRef}>
                    <label className="block font-medium text-gray-700 mb-1">Select Bot:</label>
                    <input
                        type="text"
                        value={botSearchText}
                        onChange={e => {
                            setBotSearchText(e.target.value);
                            setDropdownOpen(true);
                        }}
                        onFocus={() => setDropdownOpen(true)}
                        placeholder="Search and select bot"
                        className="w-full px-4 py-2 border border-gray-300 rounded shadow-sm focus:outline-none focus:ring focus:border-blue-400"
                        autoComplete="off"
                    />
                    {dropdownOpen && (
                        <ul className="absolute z-10 mt-1 w-full max-h-48 overflow-auto bg-white border border-gray-300 rounded shadow-lg">
                            {filteredBots.length > 0 ? (
                                filteredBots.map(bot => (
                                    <li
                                        key={bot.id}
                                        onClick={() => handleSelectBot(bot)}
                                        className="px-4 py-2 cursor-pointer hover:bg-blue-100"
                                    >
                                        {bot.address}
                                    </li>
                                ))
                            ) : (
                                <li className="px-4 py-2 text-gray-500">No bots found</li>
                            )}
                        </ul>
                    )}
                </div>

                {/* UserId搜索框 */}
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
                        {["UserID", "Question", "Answer", "Token", "Status"].map(title => (
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
                                <td className="px-6 py-4 text-sm text-gray-800">{record.answer}</td>
                                <td className="px-6 py-4 text-sm text-gray-800">{record.token}</td>
                                <td className="px-6 py-4 text-sm text-gray-800">{record.is_deleted ? "Deleted" : "Active"}</td>
                            </tr>
                        ))
                    ) : (
                        <tr>
                            <td colSpan={5} className="text-center py-6 text-gray-500">
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
