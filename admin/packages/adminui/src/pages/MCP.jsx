import React, { useEffect, useState } from "react";
import Toast from "../components/Toast";
import Modal from "../components/Modal";
import BotSelector from "../components/BotSelector";

function BotMcpListPage() {
    const [botId, setBotId] = useState(null);
    const [mcpServices, setMcpServices] = useState([]);
    const [showEditModal, setShowEditModal] = useState(false);
    const [showPrepareModal, setShowPrepareModal] = useState(false);
    const [prepareServices, setPrepareServices] = useState([]);
    const [prepareTab, setPrepareTab] = useState("list");
    const [selectedPreparedService, setSelectedPreparedService] = useState(null);
    const [prepareEditJson, setPrepareEditJson] = useState("");
    const [editingService, setEditingService] = useState(null);
    const [editJson, setEditJson] = useState("");
    const [prepareSearch, setPrepareSearch] = useState("");
    const [toast, setToast] = useState({ show: false, message: "", type: "error" });

    const showToast = (message, type = "error") => {
        setToast({ show: true, message, type });
    };

    useEffect(() => {
        if (botId !== null) {
            fetchMcpServices();
        }
    }, [botId]);

    const fetchMcpServices = async () => {
        try {
            const res = await fetch(`/bot/mcp/get?id=${botId}`);
            const data = await res.json();
            if (data.code !== 0) return showToast(data.message || "Failed to fetch services");
            const mcpObj = data.data.mcpServers || {};
            const entries = Object.entries(mcpObj).map(([name, config]) => ({ name, config }));
            setMcpServices(entries);
        } catch (err) {
            showToast("Request error: " + err.message);
        }
    };

    const handlePrepareClick = async () => {
        try {
            const res = await fetch(`/bot/mcp/prepare?id=${botId}`);
            const data = await res.json();
            if (data.code !== 0) return showToast(data.message || "Failed to prepare");
            const mcpObj = data.data.mcpServers || {};
            const entries = Object.entries(mcpObj).map(([name, config]) => ({ name, config }));
            setPrepareServices(entries);
            setPrepareTab("list");
            setShowPrepareModal(true);
        } catch (err) {
            showToast("Prepare failed: " + err.message);
        }
    };

    const handleAddPreparedService = (name, config) => {
        setSelectedPreparedService(name);
        setPrepareEditJson(JSON.stringify(config, null, 2));
        setPrepareTab("json");
    };

    const handleSubmitPreparedService = async () => {
        try {
            const config = JSON.parse(prepareEditJson);
            const params = new URLSearchParams({ id: botId, name: selectedPreparedService });

            const res = await fetch(`/bot/mcp/update?${params.toString()}`, {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify(config),
            });

            const data = await res.json();
            if (data.code !== 0) return showToast(data.message || "Failed to add");
            showToast("Service added", "success");
            setShowPrepareModal(false);
            await fetchMcpServices();
        } catch (err) {
            showToast("Invalid JSON or request error: " + err.message);
        }
    };

    const openEditModal = (svc) => {
        setEditingService(svc.name);
        setEditJson(JSON.stringify(svc.config, null, 2));
        setShowEditModal(true);
    };

    const handleUpdateService = async () => {
        try {
            const config = JSON.parse(editJson);
            const params = new URLSearchParams({ id: botId, name: editingService });

            const res = await fetch(`/bot/mcp/update?${params.toString()}`, {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify(config),
            });

            const data = await res.json();
            if (data.code !== 0) return showToast(data.message || "Failed to update MCP");

            showToast("Service updated", "success");
            setShowEditModal(false);
            await fetchMcpServices();
        } catch (err) {
            showToast("Invalid JSON or request error: " + err.message);
        }
    };

    const toggleDisableService = async (name, disable) => {
        try {
            const params = new URLSearchParams({ id: botId, name, disable: disable ? "1" : "0" });
            const res = await fetch(`/bot/mcp/disable?${params.toString()}`);
            const data = await res.json();
            if (data.code !== 0) return showToast(data.message || "Failed to toggle");
            showToast(disable ? "Disabled" : "Enabled", "success");
            await fetchMcpServices();
        } catch (err) {
            showToast("Toggle failed: " + err.message);
        }
    };

    const filteredPrepareServices = prepareServices.filter(svc =>
        svc.name.toLowerCase().includes(prepareSearch.toLowerCase())
    );

    return (
        <div className="p-6 bg-gray-100 min-h-screen relative">
            {toast.show && (
                <Toast message={toast.message} type={toast.type} onClose={() => setToast({ ...toast, show: false })} />
            )}

            <div className="flex justify-between items-center mb-6">
                <h2 className="text-2xl font-bold text-gray-800">MCP Services</h2>
                <button
                    onClick={handlePrepareClick}
                    className="bg-blue-600 text-white px-4 py-2 rounded hover:bg-blue-700"
                >
                    + ADD MCP
                </button>
            </div>

            <div className="flex space-x-4 mb-6 max-w-4xl flex-wrap">
                <div className="flex-1 min-w-[200px]">
                    <BotSelector
                        value={botId}
                        onChange={(bot) => {
                            setBotId(bot.id);
                        }}
                    />
                </div>
            </div>

            <div className="overflow-x-auto rounded-lg shadow">
                <table className="min-w-full bg-white divide-y divide-gray-200">
                    <thead className="bg-gray-50">
                    <tr>
                        <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Name</th>
                        <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Description</th>
                        <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Status</th>
                        <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Action</th>
                    </tr>
                    </thead>
                    <tbody className="divide-y divide-gray-100">
                    {mcpServices.map((svc) => (
                        <tr key={svc.name} className="hover:bg-gray-50">
                            <td className="px-6 py-4 text-sm text-gray-800">{svc.name}</td>
                            <td className="px-6 py-4 text-sm text-gray-800 whitespace-pre-line">{svc.config.description}</td>
                            <td className="px-6 py-4 text-sm text-gray-800">{svc.config.disabled ? "Disabled" : "Enabled"}</td>
                            <td className="px-6 py-4 text-sm space-x-3">
                                <button onClick={() => openEditModal(svc)} className="text-blue-600 hover:underline">Edit</button>
                                {svc.config.disabled ? (
                                    <button onClick={() => toggleDisableService(svc.name, false)} className="text-green-600 hover:underline">Enable</button>
                                ) : (
                                    <button onClick={() => toggleDisableService(svc.name, true)} className="text-yellow-600 hover:underline">Disable</button>
                                )}
                            </td>
                        </tr>
                    ))}
                    </tbody>
                </table>
            </div>

            <Modal visible={showEditModal} title="Edit MCP Service" onClose={() => setShowEditModal(false)}>
                <div className="space-y-4">
                    <div>
                        <label className="block text-sm font-medium text-gray-700">Service Name</label>
                        <input type="text" value={editingService || ""} readOnly className="w-full px-3 py-2 border border-gray-300 rounded bg-gray-100 text-gray-700" />
                    </div>
                    <div>
                        <label className="block text-sm font-medium text-gray-700">Config JSON</label>
                        <textarea value={editJson} onChange={(e) => setEditJson(e.target.value)} rows={10} className="w-full px-3 py-2 border border-gray-300 rounded font-mono" />
                    </div>
                    <div className="text-right">
                        <button onClick={handleUpdateService} className="bg-blue-600 text-white px-4 py-2 rounded hover:bg-blue-700">Update</button>
                    </div>
                </div>
            </Modal>

            <Modal visible={showPrepareModal} title="Prepared MCP Services" onClose={() => setShowPrepareModal(false)}>
                <div className="max-h-[80vh] overflow-y-auto">
                    <div className="flex space-x-4 mb-4 border-b">
                        <button className={`pb-2 ${prepareTab === "list" ? "border-b-2 border-blue-500 font-semibold" : "text-gray-500"}`} onClick={() => setPrepareTab("list")}>Service List</button>
                        <button className={`pb-2 ${prepareTab === "json" ? "border-b-2 border-blue-500 font-semibold" : "text-gray-500"}`} onClick={() => setPrepareTab("json")}>JSON Edit</button>
                    </div>

                    {prepareTab === "list" && (
                        <>
                            <div className="mb-4">
                                <input
                                    type="text"
                                    placeholder="Search service name"
                                    value={prepareSearch}
                                    onChange={(e) => setPrepareSearch(e.target.value)}
                                    className="w-full px-4 py-2 border border-gray-300 rounded shadow-sm focus:outline-none focus:ring focus:border-blue-400"
                                />
                            </div>
                            <table className="min-w-full bg-white divide-y divide-gray-200">
                                <thead className="bg-gray-50">
                                <tr>
                                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Name</th>
                                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Description</th>
                                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Action</th>
                                </tr>
                                </thead>
                                <tbody className="divide-y divide-gray-100">
                                {filteredPrepareServices.map((svc) => (
                                    <tr key={svc.name} className="hover:bg-gray-50">
                                        <td className="px-6 py-4 text-sm text-gray-800">{svc.name}</td>
                                        <td className="px-6 py-4 text-sm text-gray-800 whitespace-pre-line">{svc.config.description}</td>
                                        <td className="px-6 py-4 text-sm">
                                            <button onClick={() => handleAddPreparedService(svc.name, svc.config)} className="bg-blue-600 hover:bg-blue-700 text-white px-3 py-1 rounded">Add</button>
                                        </td>
                                    </tr>
                                ))}
                                </tbody>
                            </table>
                        </>
                    )}

                    {prepareTab === "json" && (
                        <div className="space-y-4">
                            <div>
                                <label className="block text-sm font-medium text-gray-700">Service Name</label>
                                <input type="text" value={selectedPreparedService || ""} readOnly className="w-full px-3 py-2 border border-gray-300 rounded bg-gray-100 text-gray-700" />
                            </div>
                            <div>
                                <label className="block text-sm font-medium text-gray-700">Config JSON</label>
                                <textarea value={prepareEditJson} onChange={(e) => setPrepareEditJson(e.target.value)} rows={10} className="w-full px-3 py-2 border border-gray-300 rounded font-mono" />
                            </div>
                            <div className="text-right">
                                <button onClick={handleSubmitPreparedService} className="bg-green-600 hover:bg-green-700 text-white px-4 py-2 rounded">Submit</button>
                            </div>
                        </div>
                    )}
                </div>
            </Modal>
        </div>
    );
}

export default BotMcpListPage;