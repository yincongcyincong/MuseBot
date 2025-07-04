import React, { useEffect, useState } from "react";

function ConfigForm({ botId }) {
    const [configData, setConfigData] = useState(null);
    const [loading, setLoading] = useState(true);
    const [saveStatusMap, setSaveStatusMap] = useState({}); // key: `${section}.${key}` -> status

    useEffect(() => {
        fetchConfig();
    }, []);

    const fetchConfig = async () => {
        try {
            const res = await fetch(`/bot/conf/get?id=${botId}`);
            const data = await res.json();
            setConfigData(data.data);
        } catch (err) {
            console.error("Failed to fetch config:", err);
        } finally {
            setLoading(false);
        }
    };

    const handleChange = (section, key, value) => {
        setConfigData((prev) => ({
            ...prev,
            [section]: {
                ...prev[section],
                [key]: value,
            },
        }));
    };

    const handleSaveSingle = async (section, key) => {
        const value = configData[section][key];
        const statusKey = `${section}.${key}`;
        setSaveStatusMap((prev) => ({ ...prev, [statusKey]: "Saving..." }));

        try {
            const res = await fetch(`/bot/conf/update?id=${botId}`, {
                method: "POST",
                headers: {
                    "Content-Type": "application/json",
                },
                body: JSON.stringify({ type: section, key, value }),
            });

            if (!res.ok) throw new Error("Failed");

            setSaveStatusMap((prev) => ({ ...prev, [statusKey]: "✔ Saved" }));
        } catch (err) {
            setSaveStatusMap((prev) => ({ ...prev, [statusKey]: "❌ Failed" }));
        }

        setTimeout(() => {
            setSaveStatusMap((prev) => {
                const copy = { ...prev };
                delete copy[statusKey];
                return copy;
            });
        }, 3000); // 清除状态提示
    };

    if (loading) return <div>Loading config...</div>;
    if (!configData) return <div>No config data</div>;

    return (
        <div
            style={{
                padding: "20px",
                maxHeight: "70vh",
                overflowY: "auto",
                border: "1px solid #ccc",
                borderRadius: "8px",
            }}
        >
            <h2>System Configurations</h2>
            {Object.entries(configData).map(([sectionName, sectionValues]) => (
                <div key={sectionName} style={{ marginBottom: "30px" }}>
                    <h3>{sectionName.toUpperCase()}</h3>
                    {Object.entries(sectionValues).map(([key, value]) => {
                        const statusKey = `${sectionName}.${key}`;
                        const statusText = saveStatusMap[statusKey] || "";
                        return (
                            <div
                                key={key}
                                style={{
                                    display: "flex",
                                    alignItems: "center",
                                    marginBottom: "10px",
                                }}
                            >
                                <label
                                    style={{
                                        width: "200px",
                                        fontWeight: "bold",
                                        wordBreak: "break-word",
                                    }}
                                >
                                    {key}
                                </label>
                                <input
                                    type="text"
                                    value={value ?? ""}
                                    onChange={(e) =>
                                        handleChange(sectionName, key, e.target.value)
                                    }
                                    style={{ flex: 1, padding: "8px", fontSize: "14px" }}
                                />
                                <button
                                    style={{
                                        marginLeft: "10px",
                                        padding: "6px 12px",
                                    }}
                                    onClick={() => handleSaveSingle(sectionName, key)}
                                >
                                    Save
                                </button>
                                <span style={{ marginLeft: "10px", color: "#666" }}>
                                    {statusText}
                                </span>
                            </div>
                        );
                    })}
                </div>
            ))}
        </div>
    );
}

export default ConfigForm;
