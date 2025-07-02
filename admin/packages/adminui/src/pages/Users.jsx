import React, { useState } from "react";

function Users() {
    const [search, setSearch] = useState("");
    const [users] = useState([
        { id: 1, name: "Alice", role: "Admin" },
        { id: 2, name: "Bob", role: "User" },
        { id: 3, name: "Charlie", role: "Moderator" },
    ]);

    const filteredUsers = users.filter((user) =>
        user.name.toLowerCase().includes(search.toLowerCase())
    );

    return (
        <div>
            <h2>User Management</h2>

            {/* 搜索区域 */}
            <div style={{ marginBottom: "20px" }}>
                <input
                    type="text"
                    placeholder="Search by name"
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
                    <th>ID</th>
                    <th>Name</th>
                    <th>Role</th>
                </tr>
                </thead>
                <tbody>
                {filteredUsers.map((user) => (
                    <tr key={user.id}>
                        <td>{user.id}</td>
                        <td>{user.name}</td>
                        <td>{user.role}</td>
                    </tr>
                ))}
                {filteredUsers.length === 0 && (
                    <tr>
                        <td colSpan="3" style={{ textAlign: "center" }}>
                            No users found
                        </td>
                    </tr>
                )}
                </tbody>
            </table>
        </div>
    );
}

export default Users;
