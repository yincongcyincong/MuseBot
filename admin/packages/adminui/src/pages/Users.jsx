import React, {useEffect, useState} from "react";
import Modal from "../components/Modal";

function Users() {
    const [users, setUsers] = useState([]);
    const [search, setSearch] = useState("");
    const [modalVisible, setModalVisible] = useState(false);
    const [editingUser, setEditingUser] = useState(null);
    const [form, setForm] = useState({id: 0, username: "", password: ""});

    useEffect(() => {
        fetchUsers();
    }, []);

    const fetchUsers = async () => {
        const res = await fetch("/user/list");
        const data = await res.json();
        setUsers(data.data.list);
    };

    const handleAddClick = () => {
        setForm({id: 0, username: "", password: ""});
        setEditingUser(null);
        setModalVisible(true);
    };

    const handleEditClick = (user) => {
        setForm({id: user.id, username: user.username, password: ""});
        setEditingUser(user);
        setModalVisible(true);
    };

    const handleDeleteClick = async (userId) => {
        if (!window.confirm("Are you sure you want to delete this user?")) return;

        try {
            const res = await fetch(`/user/delete?id=${userId}`, {
                method: "GET",
            });

            if (!res.ok) throw new Error("Delete failed");

            await fetchUsers(); // 刷新列表
        } catch (error) {
            console.error("Failed to delete user:", error);
        }
    };


    const handleSave = async () => {
        const url = editingUser
            ? "/user/update"
            : "/user/create";

        await fetch(url, {
            method: "POST",
            headers: {"Content-Type": "application/json"},
            body: JSON.stringify(form),
        });
        await fetchUsers();
        setModalVisible(false);
    };

    return (
        <div>
            <h2>User Management</h2>

            <div style={{marginBottom: "20px"}}>
                <input
                    type="text"
                    placeholder="Search by username"
                    value={search}
                    onChange={(e) => setSearch(e.target.value)}
                    style={{padding: "8px", marginRight: "10px"}}
                />
                <button style={{padding: "8px 16px"}} onClick={handleAddClick}>
                    Add User
                </button>
            </div>

            <table border="1" cellPadding="8" cellSpacing="0" width="100%">
                <thead style={{background: "#f0f0f0"}}>
                <tr>
                    <th>ID</th>
                    <th>Username</th>
                    <th>CreateTime</th>
                    <th>UpdateTime</th>
                    <th>Actions</th>
                </tr>
                </thead>
                <tbody>
                {users.map((user) => (
                    <tr key={user.id}>
                        <td>{user.id}</td>
                        <td>{user.username}</td>
                        <td>{new Date(user.create_time).toLocaleString()}</td>
                        <td>{new Date(user.update_time).toLocaleString()}</td>
                        <td>
                            <button onClick={() => handleEditClick(user)}>Edit</button>
                            <button onClick={() => handleDeleteClick(user.id)}
                                    style={{marginLeft: "10px", color: "red"}}>
                                Delete
                            </button>
                        </td>
                    </tr>
                ))}
                </tbody>
            </table>

            <Modal
                visible={modalVisible}
                title={editingUser ? "Edit User" : "Add User"}
                onClose={() => setModalVisible(false)}
            >
                <input type="hidden" value={form.id}/>
                <div style={{marginBottom: "10px"}}>
                    <input
                        type="text"
                        placeholder="Username"
                        value={form.username}
                        onChange={(e) => setForm({...form, username: e.target.value})}
                        style={{width: "100%", padding: "8px"}}
                        disabled={!!editingUser}
                    />
                </div>
                <div style={{marginBottom: "10px"}}>
                    <input
                        type="password"
                        placeholder="Password"
                        value={form.password}
                        onChange={(e) => setForm({...form, password: e.target.value})}
                        style={{width: "100%", padding: "8px"}}
                    />
                </div>
                <div style={{textAlign: "right"}}>
                    <button onClick={() => setModalVisible(false)} style={{marginRight: "10px"}}>
                        Cancel
                    </button>
                    <button onClick={handleSave}>Save</button>
                </div>
            </Modal>
        </div>
    );
}

export default Users;
