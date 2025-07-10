import React, { useEffect, useState } from "react";
import Modal from "../components/Modal";
import Pagination from "../components/Pagination";
import Toast from "../components/Toast";
import ConfirmModal from "../components/ConfirmModal";

function Users() {
    const [users, setUsers] = useState([]);
    const [search, setSearch] = useState("");
    const [modalVisible, setModalVisible] = useState(false);
    const [editingUser, setEditingUser] = useState(null);
    const [form, setForm] = useState({ id: 0, username: "", password: "" });

    const [page, setPage] = useState(1);
    const pageSize = 10;
    const [total, setTotal] = useState(0);

    const [toast, setToast] = useState({ show: false, message: "", type: "error" });

    // 确认删除相关状态
    const [confirmVisible, setConfirmVisible] = useState(false);
    const [userToDelete, setUserToDelete] = useState(null);

    const showToast = (message, type = "error") => {
        setToast({ show: true, message, type });
    };

    useEffect(() => {
        fetchUsers();
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [page]);

    const fetchUsers = async () => {
        try {
            const res = await fetch(
                `/user/list?page=${page}&page_size=${pageSize}&username=${encodeURIComponent(search)}`
            );
            const data = await res.json();
            if (data.code !== 0) {
                showToast(data.message || "Failed to fetch users");
                return;
            }
            setUsers(data.data.list);
            setTotal(data.data.total);
        } catch (error) {
            showToast("Request failed: " + error.message);
        }
    };

    const handleAddClick = () => {
        setForm({ id: 0, username: "", password: "" });
        setEditingUser(null);
        setModalVisible(true);
    };

    const handleEditClick = (user) => {
        setForm({ id: user.id, username: user.username, password: "" });
        setEditingUser(user);
        setModalVisible(true);
    };

    // 触发删除弹窗
    const handleDeleteClick = (userId) => {
        setUserToDelete(userId);
        setConfirmVisible(true);
    };

    // 取消删除弹窗
    const cancelDelete = () => {
        setUserToDelete(null);
        setConfirmVisible(false);
    };

    // 确认删除
    const confirmDelete = async () => {
        if (!userToDelete) return;
        try {
            const res = await fetch(`/user/delete?id=${userToDelete}`, {
                method: "GET",
            });
            const data = await res.json();
            if (data.code !== 0) {
                showToast(data.message || "Failed to delete user");
                return;
            }
            showToast("User deleted", "success");
            setConfirmVisible(false);
            setUserToDelete(null);
            // 删除成功刷新数据，注意不要直接调用 fetchUsers() 导致死循环
            fetchUsers();
        } catch (error) {
            showToast("Delete failed: " + error.message);
        }
    };

    const handleSave = async () => {
        const url = editingUser ? "/user/update" : "/user/create";
        try {
            const res = await fetch(url, {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify(form),
            });

            const data = await res.json();
            if (data.code !== 0) {
                showToast(data.message || "Failed to save user");
                return;
            }

            await fetchUsers();
            showToast("User saved", "success");
            setModalVisible(false);
        } catch (error) {
            showToast("Save failed: " + error.message);
        }
    };

    const handlePageChange = (newPage) => {
        setPage(newPage);
    };

    const handleSearch = () => {
        setPage(1);
        fetchUsers();
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

            <div className="flex justify-between items-center mb-6">
                <h2 className="text-2xl font-bold text-gray-800">User Management</h2>
                <button
                    onClick={handleAddClick}
                    className="bg-blue-600 hover:bg-blue-700 text-white px-4 py-2 rounded shadow"
                >
                    + Add User
                </button>
            </div>

            <div className="flex mb-4 space-x-2">
                <input
                    type="text"
                    placeholder="Search by username"
                    value={search}
                    onChange={(e) => setSearch(e.target.value)}
                    className="w-full sm:w-64 px-4 py-2 border border-gray-300 rounded shadow-sm focus:outline-none focus:ring focus:border-blue-400"
                />
                <button
                    onClick={handleSearch}
                    className="px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-600"
                >
                    Search
                </button>
            </div>

            <div className="overflow-x-auto rounded-lg shadow">
                <table className="min-w-full bg-white divide-y divide-gray-200">
                    <thead className="bg-gray-50">
                    <tr>
                        {["ID", "Username", "Create Time", "Update Time", "Actions"].map((title) => (
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
                            <td className="px-6 py-4 text-sm text-gray-800">{user.username}</td>
                            <td className="px-6 py-4 text-sm text-gray-600">
                                {new Date(user.create_time * 1000).toLocaleString()}
                            </td>
                            <td className="px-6 py-4 text-sm text-gray-600">
                                {new Date(user.update_time * 1000).toLocaleString()}
                            </td>
                            <td className="px-6 py-4 space-x-2">
                                <button
                                    onClick={() => handleEditClick(user)}
                                    className="text-blue-600 hover:underline text-sm"
                                >
                                    Edit
                                </button>
                                <button
                                    onClick={() => handleDeleteClick(user.id)}
                                    className="text-red-600 hover:underline text-sm"
                                >
                                    Delete
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
                title={editingUser ? "Edit User" : "Add User"}
                onClose={() => setModalVisible(false)}
            >
                <input type="hidden" value={form.id} />

                <div className="mb-4">
                    <input
                        type="text"
                        placeholder="Username"
                        value={form.username}
                        onChange={(e) => setForm({ ...form, username: e.target.value })}
                        disabled={!!editingUser}
                        className="w-full px-4 py-2 border border-gray-300 rounded focus:outline-none focus:ring focus:border-blue-400"
                    />
                </div>

                <div className="mb-4">
                    <input
                        type="password"
                        placeholder="Password"
                        value={form.password}
                        onChange={(e) => setForm({ ...form, password: e.target.value })}
                        className="w-full px-4 py-2 border border-gray-300 rounded focus:outline-none focus:ring focus:border-blue-400"
                    />
                </div>

                <div className="flex justify-end space-x-2">
                    <button
                        onClick={() => setModalVisible(false)}
                        className="bg-gray-300 hover:bg-gray-400 text-gray-800 px-4 py-2 rounded"
                    >
                        Cancel
                    </button>
                    <button
                        onClick={handleSave}
                        className="bg-blue-600 hover:bg-blue-700 text-white px-4 py-2 rounded"
                    >
                        Save
                    </button>
                </div>
            </Modal>

            <ConfirmModal
                visible={confirmVisible}
                message="Are you sure you want to delete this user?"
                onConfirm={confirmDelete}
                onCancel={cancelDelete}
            />
        </div>
    );
}

export default Users;
