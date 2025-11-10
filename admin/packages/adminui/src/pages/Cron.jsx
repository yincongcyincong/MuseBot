import React, { useEffect, useState } from "react";
import Toast from "../components/Toast";
import Modal from "../components/Modal";
import BotSelector from "../components/BotSelector"; // 保留 BotSelector
import ConfirmModal from "../components/ConfirmModal.jsx";
import Editor from "@monaco-editor/react";
import { useTranslation } from "react-i18next";
import Pagination from "../components/Pagination.jsx";
import InputField from "../components/InputField.jsx";
import TextareaField from "../components/TextAreaField.jsx";

// 对应后端的 db.Cron 结构体（简化版用于前端）
const initialNewCronTask = {
    cron_name: "",
    cron_spec: "",
    target_id: "",
    group_id: "",
    command: "",
    prompt: "",
    type: "",
};

function Cron() {
    const [botId, setBotId] = useState(null);
    const [cronTasks, setCronTasks] = useState([]); // 存储 Cron 任务列表
    const [showEditModal, setShowEditModal] = useState(false);
    const [showCreateModal, setShowCreateModal] = useState(false); // 用于新增任务

    // 用于新增或编辑任务的表单数据
    const [editingTask, setEditingTask] = useState(initialNewCronTask);

    // 分页状态 (从原代码推断需要分页)
    const [page, setPage] = useState(1);
    const [pageSize] = useState(10);
    const [total, setTotal] = useState(0);
    const [searchName, setSearchName] = useState("");

    const [CRONToDelete, setCRONToDelete] = useState(null);
    const [confirmVisible, setConfirmVisible] = useState(false);
    const [toast, setToast] = useState({ show: false, message: "", type: "error" });

    const { t } = useTranslation();

    const showToast = (message, type = "error") => {
        setToast({ show: true, message, type });
    };

    useEffect(() => {
        // botId 变化或分页/搜索条件变化时，重新获取任务列表
        if (botId !== null) {
            fetchCronTasks();
        }
    }, [botId, page, pageSize, searchName]);

    // ----------------------------------------------------
    // --- API 调用函数 ---
    // ----------------------------------------------------

    const fetchCronTasks = async () => {
        if (!botId) return;
        try {
            // 注意：BotId 字段在后端被用作 from_bot 过滤
            const params = new URLSearchParams({
                page: page,
                page_size: pageSize,
                name: searchName,
                id: botId,
            });

            const res = await fetch(`/bot/cron/list?${params.toString()}`);
            const data = await res.json();

            if (data.code !== 0) return showToast(data.message || t("failed_to_fetch_tasks"));

            data.data.list.forEach((task) => {
                task.cron_spec = task.cron
            });
            console.log(data.data.list);

            setCronTasks(data.data.list || []);
            setTotal(data.data.total || 0);
        } catch (err) {
            showToast(t("request_error") + ": " + err.message);
        }
    };

    // 处理新增/打开新增模态框
    const handleAddCronClick = () => {
        setEditingTask(initialNewCronTask);
        setShowCreateModal(true);
    };

    // 处理提交新增任务
    const handleSubmitNewCron = async () => {
        if (!editingTask.cron_name || !editingTask.cron_spec || !editingTask.target_id) {
            return showToast(t("fields_required"));
        }

        try {
            const res = await fetch(`/bot/cron/create?id=${botId}`, {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify(editingTask),
            });

            const data = await res.json();
            if (data.code !== 0) return showToast(data.message || t("failed_to_add_cron"));

            showToast(t("cron_added"), "success");
            setShowCreateModal(false);
            await fetchCronTasks();
        } catch (err) {
            showToast(t("request_error") + ": " + err.message);
        }
    };

    // 打开编辑模态框
    const openEditModal = (task) => {
        setEditingTask({
            id: task.id,
            cron_name: task.cron_name,
            cron_spec: task.cron_spec,
            target_id: task.target_id,
            group_id: task.group_id,
            command: task.command,
            prompt: task.prompt,
            type: task.type,
        });
        setShowEditModal(true);
    };

    // 处理更新任务
    const handleUpdateCron = async () => {
        if (!editingTask.id || !editingTask.cron_name || !editingTask.cron_spec || !editingTask.target_id) {
            return showToast(t("fields_required"));
        }

        try {
            const res = await fetch(`/bot/cron/update?id=${botId}`, {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify(editingTask),
            });

            const data = await res.json();
            if (data.code !== 0) return showToast(data.message || t("failed_to_update_cron"));

            showToast(t("cron_updated"), "success");
            setShowEditModal(false);
            await fetchCronTasks();
        } catch (err) {
            showToast(t("request_error") + ": " + err.message);
        }
    };

    // 处理启用/禁用任务 (UpdateCronStatus)
    const toggleDisableService = async (id, currentStatus) => {
        const newStatus = currentStatus === 1 ? 0 : 1;
        try {
            const res = await fetch(`/bot/cron/update/status?id=${botId}`, {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({ id: id, status: newStatus }),
            });

            const data = await res.json();
            if (data.code !== 0) return showToast(data.message || t("failed_to_toggle"));

            showToast(newStatus === 1 ? t("enabled") : t("disabled"), "success");
            await fetchCronTasks();
        } catch (err) {
            showToast(t("toggle_failed") + ": " + err.message);
        }
    };

    // 显示删除确认模态框
    const handleDeleteClick = (id) => {
        setCRONToDelete(id);
        setConfirmVisible(true);
    };

    // 取消删除
    const cancelDelete = () => {
        setCRONToDelete(null);
        setConfirmVisible(false);
    };

    // 确认删除 (DeleteCron)
    const confirmDelete = async () => {
        if (!CRONToDelete) return;
        try {
            const res = await fetch(`/bot/cron/delete?id=${botId}&cron_id=${CRONToDelete}`, { method: "GET" });
            const data = await res.json();

            if (data.code !== 0) {
                showToast(data.message || t("failed_to_delete_cron"));
                return;
            }
            showToast(t("cron_deleted"), "success");
            setConfirmVisible(false);
            setCRONToDelete(null);
            await fetchCronTasks();
        } catch (error) {
            showToast(t("request_error") + ": " + error.message);
        }
    };

    // ----------------------------------------------------
    // --- 渲染部分 ---
    // ----------------------------------------------------

    // 任务状态显示
    const renderStatus = (status) => {
        return status === 1 ? <span className="text-green-600 font-semibold">{t("enable")}</span> : <span className="text-red-600">{t("disable")}</span>;
    };

    const totalPages = Math.ceil(total / pageSize);

    const handlePageChange = (newPage) => {
        if (newPage >= 1 && newPage <= totalPages) {
            setPage(newPage);
        }
    };

    const handleEditFormChange = (e) => {
        const { name, value } = e.target;
        setEditingTask(prev => ({ ...prev, [name]: value }));
    };

    return (
        <div className="p-6 bg-gray-100 min-h-screen">
            {toast.show && (
                <Toast message={toast.message} type={toast.type} onClose={() => setToast({ ...toast, show: false })} />
            )}

            <div className="flex justify-between items-center mb-6">
                <h2 className="text-2xl font-bold text-gray-800">{t("cron_task_manage")}</h2>
                <div className="flex gap-2">
                    <button
                        onClick={handleAddCronClick}
                        className="bg-blue-600 text-white px-4 py-2 rounded hover:bg-blue-700"
                    >
                        + {t("add_cron_task")}
                    </button>
                    {/* 移除 Sync 按钮，Cron 任务通常是加载而非同步 */}
                </div>
            </div>

            <div className="flex space-x-4 mb-6 max-w-4xl flex-wrap items-center">
                <div className="flex-1 min-w-[200px]">
                    <BotSelector
                        value={botId}
                        onChange={(bot) => {
                            setBotId(bot.id);
                            setPage(1); // 切换 Bot 时重置分页
                        }}
                    />
                </div>
                <div className="flex-1 min-w-[200px]">
                    <label className="block font-medium text-gray-700 mb-1">{t("cron_name")}:</label>
                    <input
                        type="text"
                        onChange={(e) => {
                            setSearchName(e.target.value);
                            setPage(1); // 搜索时重置分页
                        }}
                        placeholder={t("user_id_placeholder")}
                        className="w-full px-4 py-2 border border-gray-300 rounded shadow-sm focus:outline-none focus:ring focus:border-blue-400"
                    />
                </div>
            </div>

            <div className="overflow-x-auto rounded-lg shadow mb-4">
                <table className="min-w-full bg-white divide-y divide-gray-200">
                    <thead className="bg-gray-50">
                    <tr>
                        <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">ID</th>
                        <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">{t("name")}</th>
                        <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">{t("type")}</th>
                        <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">{t("cron_spec")}</th>
                        <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">{t("target_id")}</th>
                        <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">{t("status")}</th>
                        <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">{t("action")}</th>
                    </tr>
                    </thead>
                    <tbody className="divide-y divide-gray-100">
                    {cronTasks.length > 0 ? (
                        cronTasks.map((task) => (
                            <tr key={task.id} className="hover:bg-gray-50">
                                <td className="px-6 py-4 text-sm text-gray-800">{task.id}</td>
                                <td className="px-6 py-4 text-sm text-gray-800">{task.type}</td>
                                <td className="px-6 py-4 text-sm text-gray-800">{task.cron_name}</td>
                                <td className="px-6 py-4 text-sm text-gray-800">{task.cron_spec}</td>
                                <td className="px-6 py-4 text-sm text-gray-800">{task.target_id}</td>
                                <td className="px-6 py-4 text-sm text-gray-800">{renderStatus(task.status)}</td>
                                <td className="px-6 py-4 text-sm space-x-3">
                                    <button onClick={() => openEditModal(task)} className="text-blue-600 hover:underline">{t("edit")}</button>
                                    {task.status === 1 ? (
                                        <button onClick={() => toggleDisableService(task.id, 1)} className="text-yellow-600 hover:underline">{t("disable")}</button>
                                    ) : (
                                        <button onClick={() => toggleDisableService(task.id, 0)} className="text-green-600 hover:underline">{t("enable")}</button>
                                    )}
                                    <button onClick={() => handleDeleteClick(task.id)} className="text-red-600 hover:underline">{t("delete")}</button>
                                </td>
                            </tr>
                        ))
                    ) : (
                        <tr>
                            <td colSpan="6" className="px-6 py-4 text-center text-sm text-gray-500">{t("no_cron_tasks")}</td>
                        </tr>
                    )}
                    </tbody>
                </table>
            </div>

            <Pagination page={page} pageSize={pageSize} total={total} onPageChange={handlePageChange} />

            {/* 编辑/新增任务模态框 (统一使用一个结构，通过 showEditModal/showCreateModal 控制显示) */}
            <Modal
                visible={showEditModal || showCreateModal}
                title={showCreateModal ? t("create_cron_task") : t("edit_cron_task")}
                onClose={() => {
                    setShowEditModal(false);
                    setShowCreateModal(false);
                }}
            >
                <div className="space-y-4">
                    <InputField label={t("name")} name="cron_name" value={editingTask.cron_name} onChange={handleEditFormChange} placeholder="daily-report" readOnly={!showCreateModal} />
                    <InputField label={t("type")} name="type" value={editingTask.type} onChange={handleEditFormChange} placeholder="telegram/wechat/personal_qq ..." />
                    <InputField label={t("cron_spec")} name="cron_spec" value={editingTask.cron_spec} onChange={handleEditFormChange} placeholder="0 0 10 * * *" />

                    <div className="flex space-x-4">
                        <div className="flex-1">
                            <InputField label={t("target_id")} name="target_id" value={editingTask.target_id} onChange={handleEditFormChange} placeholder="QQ/WeChat User ID" />
                        </div>
                        <div className="flex-1">
                            <InputField label={t("group_id")} name="group_id" value={editingTask.group_id} onChange={handleEditFormChange} placeholder="QQ/WeChat Group ID (Optional)" />
                        </div>
                    </div>

                    {/* Command 和 Prompt */}
                    <InputField label={t("command")} name="command" value={editingTask.command} onChange={handleEditFormChange} placeholder="/photo (for bot command)" />
                    <TextareaField label={t("prompt")} name="prompt" value={editingTask.prompt} onChange={handleEditFormChange} placeholder="Generate yesterday's sales report." />

                    <div className="text-right pt-4">
                        <button
                            onClick={showCreateModal ? handleSubmitNewCron : handleUpdateCron}
                            className="bg-blue-600 text-white px-4 py-2 rounded hover:bg-blue-700"
                        >
                            {showCreateModal ? t("create") : t("update")}
                        </button>
                    </div>
                </div>
            </Modal>


            <ConfirmModal
                visible={confirmVisible}
                message={t("confirm_delete_cron")}
                onConfirm={confirmDelete}
                onCancel={cancelDelete}
            />
        </div>
    );
}

export default Cron;