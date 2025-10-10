import React, { useEffect, useState } from "react";
import Pagination from "../components/Pagination";
import Modal from "../components/Modal";
import Toast from "../components/Toast";
import BotSelector from "../components/BotSelector";
import {useTranslation} from "react-i18next";

function BotUserListPage() {
    const [botId, setBotId] = useState(null);
    const [userIdSearch, setUserIdSearch] = useState("");
    const [users, setUsers] = useState([]);
    const [page, setPage] = useState(1);
    const [pageSize] = useState(10);
    const [total, setTotal] = useState(0);

    const [showModal, setShowModal] = useState(false);
    const [newUserId, setNewUserId] = useState("");
    const [newToken, setNewToken] = useState("");

    const [toast, setToast] = useState({ show: false, message: "", type: "error" });
    const showToast = (message, type = "error") => {
        setToast({ show: true, message, type });
    };

    const { t } = useTranslation();

    useEffect(() => {
        if (botId !== null) {
            fetchBotUsers();
        }
    }, [botId, page, userIdSearch]);

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
            if (data.code !== 0) return showToast(data.message || "Failed to fetch users");
            setUsers(data.data.list || []);
            setTotal(data.data.total || 0);
        } catch (err) {
            showToast("Request error: " + err.message);
        }
    };

    const handleSubmitNewToken = async () => {
        try {
            const res = await fetch(`/bot/add/token?id=${botId}`, {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({
                    botId,
                    user_id: newUserId,
                    token: Number(newToken),
                }),
            });

            const data = await res.json();
            if (data.code !== 0) return showToast(data.message || "Failed to submit token");

            setShowModal(false);
            setNewUserId("");
            setNewToken("");
            await fetchBotUsers();
            showToast("New token submitted", "success");
        } catch (err) {
            showToast("Submit new token failed: " + err.message);
        }
    };

    const handleUserIdSearchChange = (e) => {
        setUserIdSearch(e.target.value);
        setPage(1);
    };

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
                <h2 className="text-2xl font-bold text-gray-800">{t("bot_user_manage")}</h2>
                <button
                    onClick={() => setShowModal(true)}
                    className="bg-blue-600 text-white px-4 py-2 rounded hover:bg-blue-700"
                >
                    + {t("add_token")}
                </button>
            </div>

            <div className="flex space-x-4 mb-6 max-w-4xl flex-wrap">
                <div className="flex-1 min-w-[200px]">
                    <BotSelector
                        value={botId}
                        onChange={(bot) => {
                            setBotId(bot.id);
                            setUserIdSearch("");
                            setPage(1);
                        }}
                    />
                </div>

                <div className="flex-1 min-w-[200px]">
                    <label className="block font-medium text-gray-700 mb-1">{t("search_user_id")}:</label>
                    <input
                        type="text"
                        value={userIdSearch}
                        onChange={handleUserIdSearchChange}
                        placeholder={t("user_id_placeholder")}
                        className="w-full px-4 py-2 border border-gray-300 rounded shadow-sm focus:outline-none focus:ring focus:border-blue-400"
                    />
                </div>
            </div>

            <div className="overflow-x-auto rounded-lg shadow">
                <table className="min-w-full bg-white divide-y divide-gray-200">
                    <thead className="bg-gray-50">
                    <tr>
                        {[t("id"), t("user_id"), t("mode"), t("token"), t("available_token"), t("create_time"), t("update_time")].map((title) => (
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
                    {users.length > 0 ? (
                        users.map((user) => (
                            <tr key={user.id} className="hover:bg-gray-50">
                                <td className="px-6 py-4 text-sm text-gray-800">{user.id}</td>
                                <td className="px-6 py-4 text-sm text-gray-800">{user.user_id}</td>
                                <td className="px-6 py-4 text-sm text-gray-800">{user.mode}</td>
                                <td className="px-6 py-4 text-sm text-gray-800">{user.token}</td>
                                <td className="px-6 py-4 text-sm text-gray-800">{user.avail_token}</td>
                                <td className="px-6 py-4 text-sm text-gray-800">
                                    {new Date(user.create_time * 1000).toLocaleString()}
                                </td>
                                <td className="px-6 py-4 text-sm text-gray-800">
                                    {new Date(user.update_time * 1000).toLocaleString()}
                                </td>
                            </tr>
                        ))
                    ) : (
                        <tr>
                            <td colSpan={5} className="text-center py-6 text-gray-500">
                                No users found.
                            </td>
                        </tr>
                    )}
                    </tbody>
                </table>
            </div>

            <Pagination page={page} pageSize={pageSize} total={total} onPageChange={setPage} />

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
                            {t("submit")}
                        </button>
                    </div>
                </div>
            </Modal>
        </div>
    );
}

export default BotUserListPage;
