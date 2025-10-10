import React, { useEffect, useState } from "react";
import Modal from "../components/Modal";
import Pagination from "../components/Pagination";
import Toast from "../components/Toast";
import ConfirmModal from "../components/ConfirmModal";
import Editor from "@monaco-editor/react";
import { useTranslation } from "react-i18next";

function Bots() {
    const [bots, setBots] = useState([]);
    const [search, setSearch] = useState("");
    const [modalVisible, setModalVisible] = useState(false);
    const [editingBot, setEditingBot] = useState(null);
    const [form, setForm] = useState({
        id: "",
        name: "",
        address: "",
        crt_file: "",
        key_file: "",
        ca_file: "",
        command: "",
        is_start: true,
    });

    const { t } = useTranslation();

    const [rawConfigVisible, setRawConfigVisible] = useState(false);
    const [rawConfigText, setRawConfigText] = useState("");

    const [page, setPage] = useState(1);
    const [pageSize] = useState(10);
    const [total, setTotal] = useState(0);
    const [isRegister, setIsRegister] = useState(false);

    const [toast, setToast] = useState({ show: false, message: "", type: "error" });
    const [confirmVisible, setConfirmVisible] = useState(false);
    const [confirmStopVisible, setConfirmStopVisible] = useState(false);
    const [botToDelete, setBotToDelete] = useState(null);
    const [botToStop, setBotToStop] = useState(null);
    const [botToRestart, setBotToRestart] = useState(null);

    const [httpsExpanded, setHttpsExpanded] = useState(false);

    const toggleHttps = () => {
        setHttpsExpanded(!httpsExpanded);
    };

    const showToast = (message, type = "error") => {
        setToast({ show: true, message, type });
    };

    useEffect(() => {
        fetchBots();
    }, [page]);

    const fetchBots = async () => {
        try {
            const res = await fetch(
                `/bot/list?page=${page}&page_size=${pageSize}&address=${encodeURIComponent(search)}`
            );
            const data = await res.json();
            if (data.code !== 0) {
                showToast(data.message || "Failed to fetch bots");
                return;
            }
            setBots(data.data.list);
            setTotal(data.data.total);
            setIsRegister(data.data.is_register);
        } catch (err) {
            showToast("Request error: " + err.message);
        }
    };

    const handleRestart = async () => {
        try {
            const res = await fetch(
                `/bot/restart?id=${botToRestart}&params=${encodeURIComponent(rawConfigText)}`,
                { method: "GET" }
            );
            const data = await res.json();
            if (data.code !== 0) {
                showToast("Request error: " + (data.message || "Restart failed"));
                return;
            }
            showToast("restart Bot", "success");
            setRawConfigVisible(false);
        } catch (err) {
            showToast("Request error: " + err.message);
        }
    };

    const handleSearch = () => {
        setPage(1);
        fetchBots();
    };

    const handleAddClick = () => {
        setForm({
            id: 0,
            address: "",
            name: "",
            crt_file: "",
            key_file: "",
            ca_file: "",
            command: "-bot_name=MuseBot\n-http_host=127.0.0.1:36060",
            is_start: true,
        });
        setEditingBot(null);
        setModalVisible(true);
    };

    const handleEditClick = (bot) => {
        setForm({
            id: bot.id,
            name: bot.name || "",
            address: bot.address,
            crt_file: bot.crt_file,
            key_file: bot.key_file,
            ca_file: bot.ca_file,
            command: bot.command || "-bot_name=MuseBot\n-http_host=127.0.0.1:36060",
            is_start: true,
        });
        setEditingBot(bot);
        setModalVisible(true);
    };

    const handleDeleteClick = (botId) => {
        setBotToDelete(botId);
        setConfirmVisible(true);
    };

    const cancelDelete = () => {
        setBotToDelete(null);
        setConfirmVisible(false);
    };

    const confirmDelete = async () => {
        if (!botToDelete) return;
        try {
            const res = await fetch(`/bot/delete?id=${botToDelete}`, { method: "DELETE" });
            const data = await res.json();
            if (data.code !== 0) {
                showToast(data.message || "Failed to delete bot");
                return;
            }
            showToast("Bot deleted", "success");
            setConfirmVisible(false);
            setBotToDelete(null);
            await fetchBots();
        } catch (error) {
            showToast("Request error: " + error.message);
        }
    };

    const handleStopClick = (botId) => {
        setBotToStop(botId);
        setConfirmStopVisible(true);
    };

    const cancelStop = () => {
        setBotToStop(null);
        setConfirmStopVisible(false);
    };

    const confirmStop = async () => {
        if (!botToStop) return;
        try {
            const res = await fetch(`/bot/stop?id=${botToStop}`, { method: "DELETE" });
            const data = await res.json();
            if (data.code !== 0) {
                showToast(data.message || "Failed to stop bot");
                return;
            }
            showToast("Bot stoped", "success");
            setConfirmStopVisible(false);
            setBotToStop(null);
            await fetchBots();
        } catch (error) {
            showToast("Request error: " + error.message);
        }
    };

    const handleSave = async () => {
        try {
            const url = editingBot ? "/bot/update" : "/bot/create";
            const res = await fetch(url, {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify(form),
            });
            const data = await res.json();
            if (data.code !== 0) {
                showToast(data.message || "Failed to save bot");
                return;
            }
            await fetchBots();
            setModalVisible(false);
        } catch (err) {
            showToast("Request error: " + err.message);
        }
    };

    const handleShowRawConfig = async (botId) => {
        try {
            const res = await fetch(`/bot/command/get?id=${botId}`);
            const data = await res.json();
            if (data.code !== 0) {
                showToast(data.message || "Failed to fetch command config");
                return;
            }
            setRawConfigText(data.data);
            setRawConfigVisible(true);
            setBotToRestart(botId);
        } catch (err) {
            showToast("Request error: " + err.message);
        }
    };

    const handlePageChange = (newPage) => {
        setPage(newPage);
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
                <h2 className="text-2xl font-bold text-gray-800">{t("bot_manage")}</h2>
                {!isRegister && (
                    <button
                        onClick={handleAddClick}
                        className="bg-blue-600 hover:bg-blue-700 text-white px-4 py-2 rounded shadow"
                    >
                        + {t("add_bot")}
                    </button>
                )}
            </div>

            <div className="flex mb-4 space-x-2">
                <input
                    type="text"
                    placeholder={t("address_placeholder")}
                    value={search}
                    onChange={(e) => setSearch(e.target.value)}
                    className="w-full sm:w-64 px-4 py-2 border border-gray-300 rounded shadow-sm focus:outline-none focus:ring focus:border-blue-400"
                />
                <button
                    onClick={handleSearch}
                    className="px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-600"
                >
                    {t("search")}
                </button>
            </div>

            <div className="overflow-x-auto rounded-lg shadow">
                <table className="min-w-full bg-white divide-y divide-gray-200">
                    <thead className="bg-gray-50">
                    <tr>
                        {[t("name"), t("address"), t("status"), t("create_time"), t("update_time"), t("action")].map(
                            (title) => (
                                <th
                                    key={title}
                                    className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider"
                                >
                                    {title}
                                </th>
                            )
                        )}
                    </tr>
                    </thead>
                    <tbody className="divide-y divide-gray-100">
                    {bots.map((bot) => (
                        <tr key={bot.id} className="hover:bg-gray-50">
                            <td className="px-6 py-4 text-sm text-gray-800">{bot.name}</td>
                            <td className="px-6 py-4 text-sm text-gray-800">{bot.address}</td>
                            <td className="px-6 py-4 text-sm text-gray-800">{bot.status}</td>
                            <td className="px-6 py-4 text-sm text-gray-600">
                                {new Date(bot.create_time * 1000).toLocaleString()}
                            </td>
                            <td className="px-6 py-4 text-sm text-gray-600">
                                {new Date(bot.update_time * 1000).toLocaleString()}
                            </td>
                            <td className="px-6 py-4 space-x-2 text-sm">
                                {!isRegister && (
                                    <>
                                        <button
                                            onClick={() => handleEditClick(bot)}
                                            className="text-blue-600 hover:underline"
                                        >
                                            {t("edit")}
                                        </button>
                                        <button
                                            onClick={() => handleDeleteClick(bot.id)}
                                            className="text-red-600 hover:underline"
                                        >
                                            {t("delete")}
                                        </button>
                                        <button
                                            onClick={() => handleStopClick(bot.id)}
                                            className="text-purple-600 hover:underline"
                                        >
                                            {t("stop")}
                                        </button>
                                    </>
                                )}
                                <button
                                    onClick={() => handleShowRawConfig(bot.id)}
                                    className="text-purple-600 hover:underline"
                                >
                                    {t("command")}
                                </button>
                            </td>
                        </tr>
                    ))}
                    </tbody>
                </table>
            </div>

            <Pagination page={page} pageSize={pageSize} total={total} onPageChange={handlePageChange} />

            <Modal
                visible={modalVisible}
                title={editingBot ? "Edit Bot" : "Add Bot"}
                onClose={() => setModalVisible(false)}
            >
                <input type="hidden" value={form.id} />
                {/* Command Editor */}
                <div className="mb-4">
                    <label className="block text-sm font-medium text-gray-700 mb-1">Command</label>
                    <Editor
                        height="200px"
                        value={form.command}
                        onChange={(value) => setForm({ ...form, command: value ?? "" })}
                        options={{
                            minimap: { enabled: false },
                            fontSize: 14,
                            automaticLayout: true,
                        }}
                    />
                </div>

                {/* Is Start 单选改为勾选方框，和标签在同一行 */}
                <div className="mb-4 flex items-center space-x-2">
                    <label className="text-sm font-medium text-gray-700">{t('local_start')}:</label>
                    <div
                        onClick={() => setForm({ ...form, is_start: !form.is_start })}
                        className={`w-6 h-6 border rounded flex items-center justify-center cursor-pointer 
            ${form.is_start ? "bg-blue-600 border-blue-600" : "bg-white border-gray-400"}`}
                    >
                        {form.is_start && (
                            <svg
                                className="w-4 h-4 text-white"
                                fill="none"
                                stroke="currentColor"
                                strokeWidth="2"
                                viewBox="0 0 24 24"
                            >
                                <path strokeLinecap="round" strokeLinejoin="round" d="M5 13l4 4L19 7" />
                            </svg>
                        )}
                    </div>
                </div>


                {/* HTTPS Config 折叠 */}
                <div className="mb-4 border rounded">
                    <div
                        onClick={toggleHttps}
                        className="cursor-pointer bg-gray-100 px-4 py-2 flex justify-between items-center"
                    >
                        <span>HTTPS Config</span>
                        <span>{httpsExpanded ? "▲" : "▼"}</span>
                    </div>
                    {httpsExpanded && (
                        <div className="px-4 py-2 space-y-2">
                            <textarea
                                placeholder="CA File"
                                value={form.ca_file}
                                onChange={(e) => setForm({ ...form, ca_file: e.target.value })}
                                className="w-full px-2 py-1 border border-gray-300 rounded focus:outline-none focus:ring focus:border-blue-400"
                                rows={3}
                            />
                            <textarea
                                placeholder="KEY File"
                                value={form.key_file}
                                onChange={(e) => setForm({ ...form, key_file: e.target.value })}
                                className="w-full px-2 py-1 border border-gray-300 rounded focus:outline-none focus:ring focus:border-blue-400"
                                rows={3}
                            />
                            <textarea
                                placeholder="CRT File"
                                value={form.crt_file}
                                onChange={(e) => setForm({ ...form, crt_file: e.target.value })}
                                className="w-full px-2 py-1 border border-gray-300 rounded focus:outline-none focus:ring focus:border-blue-400"
                                rows={3}
                            />
                        </div>
                    )}
                </div>

                <div className="flex justify-end space-x-2">
                    <button
                        onClick={() => setModalVisible(false)}
                        className="bg-gray-300 hover:bg-gray-400 text-gray-800 px-4 py-2 rounded"
                    >
                        {t("cancel")}
                    </button>
                    <button
                        onClick={handleSave}
                        className="bg-blue-600 hover:bg-blue-700 text-white px-4 py-2 rounded"
                    >
                        {t("save")}
                    </button>
                </div>
            </Modal>

            <Modal visible={rawConfigVisible} title="Command" onClose={() => setRawConfigVisible(false)}>
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
                        onClick={() => setRawConfigVisible(false)}
                        className="bg-gray-300 hover:bg-gray-400 text-gray-800 px-4 py-2 rounded"
                    >
                        {t("cancel")}
                    </button>
                    <button
                        onClick={handleRestart}
                        className="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700"
                    >
                        {t("start")}
                    </button>
                </div>
            </Modal>

            <ConfirmModal
                visible={confirmVisible}
                message="Are you sure you want to delete this bot?"
                onConfirm={confirmDelete}
                onCancel={cancelDelete}
            />

            <ConfirmModal
                visible={confirmStopVisible}
                message="Are you sure you want to stop this bot?"
                onConfirm={confirmStop}
                onCancel={cancelStop}
            />
        </div>
    );
}

export default Bots;
