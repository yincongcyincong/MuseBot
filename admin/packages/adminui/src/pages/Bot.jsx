import React, { useState } from "react";

function Bot() {
    const [search, setSearch] = useState("");
    const [settings] = useState([
        { key: "theme", value: "light" },
        { key: "language", value: "en" },
        { key: "notifications", value: "enabled" },
    ]);

    const filteredBot = settings.filter((item) =>
        item.key.toLowerCase().includes(search.toLowerCase())
    );

    return (
        <div>
            <h2>System Bot</h2>

            {/* 搜索区域 */}
            <div style={{ marginBottom: "20px" }}>
                <input
                    type="text"
                    placeholder="Search setting key"
                    value={search}
                    onChange={(e) => setSearch(e.target.value)}
                    style={{ padding: "8px", marginRight: "10px" }}
                />
                <button onClick={() => {}} style={{ padding: "8px 16px" }}>
                    Search
                </button>
            </div>

            {/* 表格区域 */}
            <table border="1" cellPadding="8" cellSpacing="0" width="100%">
                <thead style={{ background: "#f0f0f0" }}>
                <tr>
                    <th>Key</th>
                    <th>Value</th>
                </tr>
                </thead>
                <tbody>
                {filteredBot.map((item, index) => (
                    <tr key={index}>
                        <td>{item.key}</td>
                        <td>{item.value}</td>
                    </tr>
                ))}
                {filteredBot.length === 0 && (
                    <tr>
                        <td colSpan="2" style={{ textAlign: "center" }}>
                            No settings found
                        </td>
                    </tr>
                )}
                </tbody>
            </table>
        </div>
    );
}

export default Bot;
