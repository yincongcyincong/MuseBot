import React, {useEffect, useState} from "react";
import Pagination from "../components/Pagination";
import BotSelector from "../components/BotSelector";
import Modal from "../components/Modal";
import Editor from "@monaco-editor/react";
import Toast from "../components/Toast.jsx";
import {useTranslation} from "react-i18next";
import ConfirmModal from "../components/ConfirmModal.jsx";

function Rag() {
    const {t} = useTranslation();
    const [botId, setBotId] = useState(null);
    const [ragList, setRagList] = useState([]);
    const [page, setPage] = useState(1);
    const [pageSize] = useState(10);
    const [total, setTotal] = useState(0);
    const [toast, setToast] = useState({show: false, message: "", type: "error"});
    const [isModalOpen, setIsModalOpen] = useState(false);
    const [editorValue, setEditorValue] = useState("");
    const [fileName, setFileName] = useState("");
    const [isEditing, setIsEditing] = useState(false);
    const [confirmVisible, setConfirmVisible] = useState(false);
    const [fileToDelete, setFileToDelete] = useState(null);

    const showToast = (message, type = "error") => {
        setToast({show: true, message, type});
    };

    useEffect(() => {
        if (botId !== null) {
            fetchRagList();
        }
    }, [botId, page]);

    // 获取 RAG 文件列表
    const fetchRagList = async () => {
        try {
            const params = new URLSearchParams({
                id: botId,
                page,
                pageSize,
            });
            const res = await fetch(`/bot/rag/list?${params.toString()}`);
            const data = await res.json();
            if (data.code === 0) {
                setRagList(data.data.list || []);
                setTotal(data.data.total || 0);
            } else {
                showToast(data.message);
            }
        } catch (err) {
            showToast("Failed to fetch RAG list: " + err.message);
        }
    };

    // 打开编辑器窗口
    const handleAdd = () => {
        setFileName("");
        setEditorValue("");
        setIsEditing(false);
        setIsModalOpen(true);
    };

    const handleEdit = async (name) => {
        try {
            const res = await fetch(`/bot/rag/get?id=${botId}&file_name=${encodeURIComponent(name)}`);
            const data = await res.json();
            if (data.code === 0) {
                setFileName(name);
                setEditorValue(data.data.content || "");
                setIsEditing(true);
                setIsModalOpen(true);
            } else {
                showToast(data.message);
            }
        } catch (err) {
            showToast("Failed to get file content: " + err.message);
        }
    };

    // 保存新增或修改
    const handleSave = async () => {
        if (!fileName.trim()) {
            showToast("Please input file name");
            return;
        }
        try {
            const res = await fetch(`/bot/rag/create?id=${botId}`, {
                method: "POST",
                headers: {"Content-Type": "application/json"},
                body: JSON.stringify({file_name: fileName,content: editorValue}),
            });
            const data = await res.json();
            if (data.code === 0) {
                showToast(isEditing ? "File updated successfully" : "File added successfully", "success");
                setIsModalOpen(false);
                await fetchRagList();
            } else {
                showToast(data.message);
            }
        } catch (err) {
            showToast("Failed to save file: " + err.message);
        }
    };

    const handleDeleteClick = (fileName) => {
        setFileToDelete(fileName);
        setConfirmVisible(true);
    };

    // 取消删除弹窗
    const cancelDelete = () => {
        setFileToDelete(null);
        setConfirmVisible(false);
    };

    const handleDelete = async () => {
        try {
            const res = await fetch(`/bot/rag/delete?id=${botId}&file_name=${encodeURIComponent(fileToDelete)}`, {method: "POST"});
            const data = await res.json();
            if (data.code === 0) {
                showToast("File deleted successfully", "success");
                await fetchRagList();
            } else {
                showToast(data.message);
            }
            setFileToDelete(null);
            setConfirmVisible(false);
        } catch (err) {
            showToast("Failed to delete file: " + err.message);
        }
    };

    return (
        <div className="p-6 bg-gray-100 min-h-screen">
            {toast.show && (
                <Toast
                    message={toast.message}
                    type={toast.type}
                    onClose={() => setToast({...toast, show: false})}
                />
            )}

            <div className="flex justify-between items-center mb-6">
                <h2 className="text-2xl font-bold text-gray-800">{t("rag_manage") || "RAG File Manage"}</h2>
                <button
                    onClick={handleAdd}
                    className="px-4 py-2 bg-blue-600 text-white rounded-lg shadow hover:bg-blue-700"
                >
                    {t("add_file") || "Add File"}
                </button>
            </div>

            <div className="flex space-x-4 mb-6 max-w-4xl flex-wrap items-end">
                <div className="flex-1 min-w-[200px]">
                    <BotSelector
                        value={botId}
                        onChange={(bot) => {
                            setBotId(bot.id);
                            setPage(1);
                        }}
                    />
                </div>
            </div>

            <div className="overflow-x-auto rounded-lg shadow">
                <table className="min-w-full bg-white divide-y divide-gray-200">
                    <thead className="bg-gray-50">
                    <tr>
                        {["File Name", "Created", "Updated", "Actions"].map(title => (
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
                    {ragList.length > 0 ? (
                        ragList.map(file => (
                            <tr key={file.id} className="hover:bg-gray-50">
                                <td className="px-6 py-4 text-sm text-gray-800">{file.file_name}</td>
                                <td className="px-6 py-4 text-sm text-gray-800">
                                    {new Date(file.create_time * 1000).toLocaleString()}
                                </td>
                                <td className="px-6 py-4 text-sm text-gray-800">
                                    {new Date(file.update_time * 1000).toLocaleString()}
                                </td>
                                <td className="px-6 py-4 text-sm text-gray-800 space-x-2">
                                    <button
                                        onClick={() => handleEdit(file.file_name)}
                                        className="text-blue-600 hover:underline"
                                    >
                                        Edit
                                    </button>
                                    <button
                                        onClick={() => handleDeleteClick(file.file_name)}
                                        className="text-red-600 hover:underline"
                                    >
                                        Delete
                                    </button>
                                </td>
                            </tr>
                        ))
                    ) : (
                        <tr>
                            <td colSpan={6} className="text-center py-6 text-gray-500">
                                No files found.
                            </td>
                        </tr>
                    )}
                    </tbody>
                </table>
            </div>

            <Pagination page={page} pageSize={pageSize} total={total} onPageChange={setPage}/>

            <Modal
                visible={isModalOpen}
                onClose={() => setIsModalOpen(false)}
                title={isEditing ? "Edit File" : "Add File"}
            >
                <div className="mb-4">
                    <label className="block text-sm font-medium text-gray-700 mb-1">File Name</label>
                    <input
                        type="text"
                        value={fileName}
                        onChange={(e) => setFileName(e.target.value)}
                        placeholder="Enter file name"
                        className="w-full px-3 py-2 border border-gray-300 rounded mb-3"
                        disabled={isEditing}
                    />
                    <Editor
                        height="400px"
                        defaultLanguage="markdown"
                        value={editorValue}
                        onChange={(value) => setEditorValue(value ?? "")}
                        options={{
                            minimap: {enabled: false},
                            fontSize: 14,
                            automaticLayout: true,
                            formatOnPaste: true,
                            formatOnType: true,
                        }}
                    />
                </div>
                <div className="flex justify-end space-x-2">
                    <button
                        onClick={() => setIsModalOpen(false)}
                        className="bg-gray-300 hover:bg-gray-400 text-gray-800 px-4 py-2 rounded"
                    >
                        Cancel
                    </button>
                    <button
                        onClick={handleSave}
                        className="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700"
                    >
                        Save
                    </button>
                </div>
            </Modal>

            <ConfirmModal
                visible={confirmVisible}
                message="Are you sure you want to delete this user?"
                onConfirm={handleDelete}
                onCancel={cancelDelete}
            />
        </div>
    );
}

export default Rag;
