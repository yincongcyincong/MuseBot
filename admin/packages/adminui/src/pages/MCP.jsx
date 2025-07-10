import React, { useEffect, useState } from "react";
import Toast from "../components/Toast";
import Modal from "../components/Modal";

function BotMcpListPage() {
    const [bots, setBots] = useState([]);
    const [botId, setBotId] = useState(null);
    const [searchText, setSearchText] = useState("");
    const [dropdownOpen, setDropdownOpen] = useState(false);
    const [mcpServices, setMcpServices] = useState([]);

    const [confirmingService, setConfirmingService] = useState(null);
    const [showAddModal, setShowAddModal] = useState(false);

    const [newServiceName, setNewServiceName] = useState("");
    const [newCommand, setNewCommand] = useState("");
    const [newArgs, setNewArgs] = useState("");
    const [newDescription, setNewDescription] = useState("");

    const [toast, setToast] = useState({ show: false, message: "", type: "error" });
    const showToast = (message, type = "error") => {
        setToast({ show: true, message, type });
    };

    useEffect(() => {
        fetchOnlineBots();
    }, []);

    useEffect(() => {
        if (botId !== null) {
            fetchMcpServices();
        }
    }, [botId]);

    const fetchOnlineBots = async () => {
        try {
            const res = await fetch("/bot/online");
            const data = await res.json();
            if (data.code !== 0) {
                showToast(data.message || "Failed to fetch bots");
                return;
            }
            if (data.data && data.data.length > 0) {
                setBots(data.data);
                const firstBot = data.data[0];
                setBotId(firstBot.id);
                setSearchText(firstBot.address);
            }
        } catch (err) {
            showToast("Request error: " + err.message);
        }
    };

    const fetchMcpServices = async () => {
        try {
            const res = await fetch(`/bot/mcp/get?id=${botId}`);
            const data = await res.json();
            if (data.code !== 0) {
                showToast(data.message || "Failed to fetch services");
                return;
            }

            const mcpObj = data.data.mcpServers || {};
            const entries = Object.entries(mcpObj).map(([name, config]) => ({
                name,
                command: config.command,
                args: config.args.join(" "),
                description: config.description,
            }));
            setMcpServices(entries);
        } catch (err) {
            showToast("Request error: " + err.message);
        }
    };

    const handleDeleteClick = (serviceName) => {
        setConfirmingService(serviceName);
    };

    const confirmDelete = async () => {
        try {
            const params = new URLSearchParams({ id: botId, name: confirmingService });
            const res = await fetch(`/bot/mcp/delete?${params.toString()}`);
            const data = await res.json();

            if (data.code !== 0) {
                showToast(data.message || "Failed to delete service");
                return;
            }

            showToast("Service deleted", "success");
            setConfirmingService(null);
            await fetchMcpServices();
        } catch (err) {
            showToast("Delete service failed: " + err.message);
        }
    };

    const handleAddClick = () => {
        setShowAddModal(true);
    };

    const handleAddMcp = async () => {
        try {
            const params = new URLSearchParams({
                id: botId,
                name: newServiceName,
                command: newCommand,
                args: newArgs,
                description: newDescription,
            });

            const res = await fetch(`/bot/mcp/add?${params.toString()}`);
            const data = await res.json();

            if (data.code !== 0) {
                showToast(data.message || "Failed to add MCP service");
                return;
            }

            showToast("Service added successfully", "success");
            setShowAddModal(false);
            setNewServiceName("");
            setNewCommand("");
            setNewArgs("");
            setNewDescription("");
            await fetchMcpServices();
        } catch (err) {
            showToast("Add MCP failed: " + err.message);
        }
    };

    const cancelDelete = () => setConfirmingService(null);

    const filteredBots = bots.filter((bot) =>
        bot.address.toLowerCase().includes(searchText.toLowerCase())
    );

    const handleSelectBot = (bot) => {
        setBotId(bot.id);
        setSearchText(bot.address);
        setDropdownOpen(false);
    };

    return (
        <div className="p-6 bg-gray-100 min-h-screen relative">
            {toast.show && (
                <Toast
                    message={toast.message}
                    type={toast.type}
                    onClose={() => setToast({ ...toast, show: false })}
                />
            )}

            {/* 顶部标题 + 添加按钮 */}
            <div className="flex justify-between items-center mb-6">
                <h2 className="text-2xl font-bold text-gray-800">MCP Service List</h2>
                <button
                    onClick={handleAddClick}
                    className="bg-blue-600 hover:bg-blue-700 text-white px-4 py-2 rounded shadow"
                >
                    + Add MCP
                </button>
            </div>

            {/* 选择 Bot */}
            <div className="mb-6 max-w-2xl">
                <label className="block font-medium text-gray-700 mb-1">Search & Select Bot:</label>
                <div className="relative">
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
            </div>

            {/* 表格 */}
            <div className="overflow-x-auto rounded-lg shadow">
                <table className="min-w-full bg-white divide-y divide-gray-200">
                    <thead className="bg-gray-50">
                    <tr>
                        <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Name</th>
                        <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Description</th>
                        <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Action</th>
                    </tr>
                    </thead>
                    <tbody className="divide-y divide-gray-100">
                    {mcpServices.map((svc) => (
                        <tr key={svc.name} className="hover:bg-gray-50">
                            <td className="px-6 py-4 text-sm text-gray-800">{svc.name}</td>
                            <td className="px-6 py-4 text-sm text-gray-800 whitespace-pre-line">{svc.description}</td>
                            <td className="px-6 py-4 text-sm">
                                <button
                                    onClick={() => handleDeleteClick(svc.name)}
                                    className="text-red-600 hover:underline"
                                >
                                    Delete
                                </button>
                            </td>
                        </tr>
                    ))}
                    {mcpServices.length === 0 && (
                        <tr>
                            <td colSpan={5} className="text-center py-6 text-gray-400">No MCP services found.</td>
                        </tr>
                    )}
                    </tbody>
                </table>
            </div>

            {/* 删除确认弹窗 */}
            {confirmingService && (
                <div className="fixed inset-0 bg-black bg-opacity-30 flex items-center justify-center z-50">
                    <div className="bg-white p-6 rounded shadow-lg max-w-sm w-full">
                        <h3 className="text-lg font-semibold text-gray-800 mb-4">Confirm Delete</h3>
                        <p className="text-gray-700 mb-4">Are you sure you want to delete service <strong>{confirmingService}</strong>?</p>
                        <div className="flex justify-end space-x-4">
                            <button onClick={cancelDelete} className="px-4 py-2 text-gray-600 hover:text-gray-800">Cancel</button>
                            <button onClick={confirmDelete} className="px-4 py-2 bg-red-600 text-white rounded hover:bg-red-700">Delete</button>
                        </div>
                    </div>
                </div>
            )}

            {/* 添加 MCP 弹窗 */}
            <Modal visible={showAddModal} title="Add MCP Service" onClose={() => setShowAddModal(false)}>
                <div className="space-y-4">
                    <div>
                        <label className="block text-sm font-medium text-gray-700 mb-1">Service Name:</label>
                        <input
                            type="text"
                            value={newServiceName}
                            onChange={(e) => setNewServiceName(e.target.value)}
                            className="w-full px-3 py-2 border border-gray-300 rounded"
                        />
                    </div>
                    <div>
                        <label className="block text-sm font-medium text-gray-700 mb-1">Command:</label>
                        <input
                            type="text"
                            value={newCommand}
                            onChange={(e) => setNewCommand(e.target.value)}
                            className="w-full px-3 py-2 border border-gray-300 rounded"
                        />
                    </div>
                    <div>
                        <label className="block text-sm font-medium text-gray-700 mb-1">Args (space-separated):</label>
                        <input
                            type="text"
                            value={newArgs}
                            onChange={(e) => setNewArgs(e.target.value)}
                            className="w-full px-3 py-2 border border-gray-300 rounded"
                        />
                    </div>
                    <div>
                        <label className="block text-sm font-medium text-gray-700 mb-1">Description:</label>
                        <textarea
                            value={newDescription}
                            onChange={(e) => setNewDescription(e.target.value)}
                            className="w-full px-3 py-2 border border-gray-300 rounded"
                        />
                    </div>
                    <div className="text-right">
                        <button
                            onClick={handleAddMcp}
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

export default BotMcpListPage;
