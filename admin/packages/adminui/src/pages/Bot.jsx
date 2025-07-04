import React, {useEffect, useState} from "react";
import Modal from "../components/Modal";
import ConfigForm from "./ConfigForm"; // 假设你已经拆出来

function Bots() {
    const [bots, setBots] = useState([]);
    const [search, setSearch] = useState("");
    const [modalVisible, setModalVisible] = useState(false);
    const [editingBot, setEditingBot] = useState(null);
    const [form, setForm] = useState({id: 0, address: "", crt_file: ""});

    const [rawConfigVisible, setRawConfigVisible] = useState(false);
    const [structuredConfigVisible, setStructuredConfigVisible] = useState(false);
    const [rawConfigText, setRawConfigText] = useState("");
    const [selectId, setSelectId] = useState(0);

    useEffect(() => {
        fetchBots();
    }, []);

    const fetchBots = async () => {
        const res = await fetch("/bot/list");
        const data = await res.json();
        setBots(data.data.list);
    };

    const handleAddClick = () => {
        setForm({id: 0, address: "", crt_file: ""});
        setEditingBot(null);
        setModalVisible(true);
    };

    const handleEditClick = (bot) => {
        setForm({id: bot.id, address: bot.address, crt_file: bot.crt_file});
        setEditingBot(bot);
        setModalVisible(true);
    };

    const handleDeleteClick = async (botId) => {
        if (!window.confirm("Are you sure you want to delete this bot?")) return;

        try {
            const res = await fetch(`/bot/delete?id=${botId}`, {
                method: "DELETE",
            });

            if (!res.ok) throw new Error("Delete failed");
            await fetchBots();
        } catch (error) {
            console.error("Failed to delete bot:", error);
        }
    };

    const handleSave = async () => {
        const url = editingBot ? `/bot/update` : "/bot/create";

        await fetch(url, {
            method: "POST",
            headers: {"Content-Type": "application/json"},
            body: JSON.stringify(form),
        });

        await fetchBots();
        setModalVisible(false);
    };

    const handleShowRawConfig = async (botId) => {
        try {
            const res = await fetch(`/bot/command/get?id=${botId}`); // 后端返回 JSON 的 data 是 string
            const data = await res.json();
            setRawConfigText(data.data);
            setRawConfigVisible(true);
        } catch (err) {
            console.error("Failed to fetch raw config:", err);
        }
    };

    const handleShowStructuredConfig = (botId) => {
        setStructuredConfigVisible(true);
        setSelectId(botId)
    };

    return (
        <div style={{padding: "20px"}}>
            <h2>Bot Management</h2>

            <div style={{marginBottom: "20px"}}>
                <input
                    type="text"
                    placeholder="Search by address"
                    value={search}
                    onChange={(e) => setSearch(e.target.value)}
                    style={{padding: "8px", marginRight: "10px"}}
                />
                <button style={{padding: "8px 16px", marginRight: "10px"}} onClick={handleAddClick}>
                    Add Bot
                </button>
            </div>

            <table border="1" cellPadding="8" cellSpacing="0" width="100%">
                <thead style={{background: "#f0f0f0"}}>
                <tr>
                    <th>ID</th>
                    <th>Address</th>
                    <th>CRT File</th>
                    <th>Status</th>
                    <th>CreateTime</th>
                    <th>UpdateTime</th>
                    <th>Actions</th>
                </tr>
                </thead>
                <tbody>
                {bots
                    .filter((bot) => bot.address.includes(search))
                    .map((bot) => (
                        <tr key={bot.id}>
                            <td>{bot.id}</td>
                            <td>{bot.address}</td>
                            <td>{bot.crt_file}</td>
                            <td>{bot.status}</td>
                            <td>{new Date(bot.create_time * 1000).toLocaleString()}</td>
                            <td>{new Date(bot.update_time * 1000).toLocaleString()}</td>
                            <td>
                                <button onClick={() => handleEditClick(bot)}>
                                    Edit
                                </button>
                                <button onClick={() => handleShowRawConfig(bot.id)}>
                                    Command
                                </button>
                                <button onClick={() => handleShowStructuredConfig(bot.id)}>
                                    Config
                                </button>
                                <button
                                    onClick={() => handleDeleteClick(bot.id)}
                                    style={{marginLeft: "10px", color: "red"}}
                                >
                                    Delete
                                </button>
                            </td>
                        </tr>
                    ))}
                </tbody>
            </table>

            {/* Bot Add/Edit Modal */}
            <Modal
                visible={modalVisible}
                title={editingBot ? "Edit Bot" : "Add Bot"}
                onClose={() => setModalVisible(false)}
            >
                <input type="hidden" value={form.id}/>
                <div style={{marginBottom: "10px"}}>
                    <input
                        type="text"
                        placeholder="Address"
                        value={form.address}
                        onChange={(e) => setForm({...form, address: e.target.value})}
                        style={{width: "100%", padding: "8px"}}
                    />
                </div>
                <div style={{marginBottom: "10px"}}>
                    <textarea
                        name="crt_file"
                        placeholder="CRT File"
                        onChange={(e) => setForm({...form, crt_file: e.target.value})}
                        style={{width: "100%", padding: "8px"}}
                    >{form.crt_file}</textarea>
                </div>
                <div style={{textAlign: "right"}}>
                    <button onClick={() => setModalVisible(false)} style={{marginRight: "10px"}}>
                        Cancel
                    </button>
                    <button onClick={handleSave}>Save</button>
                </div>
            </Modal>

            {/* Raw Config Modal */}
            <Modal visible={rawConfigVisible} title="Command" onClose={() => setRawConfigVisible(false)}>
                <pre style={{whiteSpace: "pre-wrap", wordBreak: "break-word", maxHeight: "500px", overflowY: "auto"}}>
                    {rawConfigText}
                </pre>
            </Modal>

            {/* Structured Config Modal with embedded form */}
            <Modal
                visible={structuredConfigVisible}
                title="Edit Config"
                onClose={() => setStructuredConfigVisible(false)}
            >
                <ConfigForm botId={selectId}/>
            </Modal>
        </div>
    );
}

export default Bots;
