import React, { useEffect, useState } from "react";
import Pagination from "../components/Pagination";
import Modal from "../components/Modal"; // 确保路径正确

function BotUserListPage() {
    const [bots, setBots] = useState([]);
    const [botId, setBotId] = useState(null);
    const [searchText, setSearchText] = useState("");
    const [dropdownOpen, setDropdownOpen] = useState(false);
    const [userIdSearch, setUserIdSearch] = useState("");
    const [users, setUsers] = useState([]);
    const [page, setPage] = useState(1);
    const [pageSize] = useState(10);
    const [total, setTotal] = useState(0);

    // Modal 状态
    const [showModal, setShowModal] = useState(false);
    const [newUserId, setNewUserId] = useState("");
    const [newToken, setNewToken] = useState("");

    useEffect(() => {
        fetchOnlineBots();
    }, []);

    useEffect(() => {
        if (botId !== null) {
            fetchBotUsers();
        }
    }, [botId, page, userIdSearch]);

    const fetchOnlineBots = async () => {
        try {
            const res = await fetch("/bot/online");
            const data = await res.json();
            if (data.data && data.data.length > 0) {
                setBots(data.data);
                const firstBot = data.data[0];
                setBotId(firstBot.id);
                setSearchText(firstBot.address);
            }
        } catch (err) {
            console.error("Failed to fetch online bots:", err);
        }
    };

    const fetchBotUsers = async () => {
        try {
            const params = new URLSearchParams({
                id: botId,
                page,
                pageSize,
            });
            if (userIdSearch.trim() !== "") {
                params.append("userId", userIdSearch.trim());
            }
            const res = await fetch(`/bot/user/list?${params.toString()}`);
            const data = await res.json();
            setUsers(data.data.list);
            setTotal(data.data.total);
        } catch (err) {
            console.error("Failed to fetch bot users:", err);
        }
    };

    const handleAddToken = async (userId) => {
        try {
            const res = await fetch("/bot/user/add_token", {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({ botId, userId }),
            });

            if (!res.ok) throw new Error("Failed to add token");

            await fetchBotUsers(); // 刷新列表
        } catch (err) {
            console.error("Add token failed:", err);
        }
    };

    const handleSubmitNewToken = async () => {
        try {
            const res = await fetch(`/bot/add/token?id=${botId}`, {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({ botId, user_id: Number(newUserId), token: Number(newToken) }),
            });

            if (!res.ok) throw new Error("Failed to add token");

            setShowModal(false);
            setNewUserId("");
            setNewToken("");
            await fetchBotUsers(); // 刷新
        } catch (err) {
            console.error("Submit new token failed:", err);
        }
    };

    const filteredBots = bots.filter((bot) =>
        bot.address.toLowerCase().includes(searchText.toLowerCase())
    );

    const handleSelectBot = (bot) => {
        setBotId(bot.id);
        setSearchText(bot.address);
        setDropdownOpen(false);
        setPage(1);
        setUserIdSearch("");
    };

    const handleUserIdSearchChange = (e) => {
        setUserIdSearch(e.target.value);
        setPage(1);
    };

    return (
        <div className="p-6 bg-gray-100 min-h-screen">
            <div className="flex justify-between items-center mb-6">
                <h2 className="text-2xl font-bold text-gray-800">Bot User List</h2>
                <button
                    onClick={() => setShowModal(true)}
                    className="bg-blue-600 text-white px-4 py-2 rounded hover:bg-blue-700"
                >
                    + Add Token
                </button>
            </div>

            <div className="flex space-x-4 mb-6 max-w-4xl flex-wrap">
                <div className="relative flex-1 min-w-[200px]">
                    <label className="block font-medium text-gray-700 mb-1">Search & Select Bot:</label>
                    <input
                        type="text"
                        value={searchText}
                        onChange={(e) => {
                            setSearchText(e.target.value);
                            setDropdownOpen(true);
                        }}
                        onFocus={() => setDropdownOpen(true)}
                        className="w-full px-4 py-2 border border-gray-300 rounded shadow-sm focus:outline-none focus:ring focus:border-blue-400"
                        placeholder="Search by address..."
                    />
                    {dropdownOpen && (
                        <ul className="absolute z-10 mt-1 w-full max-h-48 overflow-auto bg-white border border-gray-300 rounded shadow-lg">
                            {filteredBots.length > 0 ? (
                                filteredBots.map((bot) => (
                                    <li
                                        key={bot.id}
                                        onClick={() => handleSelectBot(bot)}
                                        className="px-4 py-2 cursor-pointer hover:bg-blue-100"
                                    >
                                        {bot.address}
                                    </li>
                                ))
                            ) : (
                                <li className="px-4 py-2 text-gray-500">No match</li>
                            )}
                        </ul>
                    )}
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
                        {["ID", "User ID", "Mode", "Token", "Available Token"].map((title) => (
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
                    {users.map((user) => (
                        <tr key={user.id} className="hover:bg-gray-50">
                            <td className="px-6 py-4 text-sm text-gray-800">{user.id}</td>
                            <td className="px-6 py-4 text-sm text-gray-800">{user.user_id}</td>
                            <td className="px-6 py-4 text-sm text-gray-800">{user.mode}</td>
                            <td className="px-6 py-4 text-sm text-gray-800">{user.token}</td>
                            <td className="px-6 py-4 text-sm text-gray-800">{user.avail_token}</td>
                        </tr>
                    ))}
                    </tbody>
                </table>
            </div>

            <Pagination page={page} pageSize={pageSize} total={total} onPageChange={setPage} />

            {/* Add Token Modal */}
            <Modal visible={showModal} title="Add New Token" onClose={() => setShowModal(false)}>
                <div className="space-y-4">
                    <div>
                        <label className="block text-sm font-medium text-gray-700 mb-1">User ID:</label>
                        <input
                            type="text"
                            value={newUserId}
                            onChange={(e) => setNewUserId(e.target.value)}
                            className="w-full px-3 py-2 border border-gray-300 rounded"
                        />
                    </div>
                    <div>
                        <label className="block text-sm font-medium text-gray-700 mb-1">Token:</label>
                        <input
                            type="text"
                            value={newToken}
                            onChange={(e) => setNewToken(e.target.value)}
                            className="w-full px-3 py-2 border border-gray-300 rounded"
                        />
                    </div>
                    <div className="text-right">
                        <button
                            onClick={handleSubmitNewToken}
                            className="bg-blue-600 hover:bg-blue-700 text-white px-4 py-2 rounded"
                        >
                            Submit
                        </button>
                    </div>
                </div>
            </Modal>
        </div>
    );
}

export default BotUserListPage;
