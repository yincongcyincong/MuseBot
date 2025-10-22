import i18n from 'i18next';
import {initReactI18next} from 'react-i18next';
import LanguageDetector from 'i18next-browser-languagedetector';

i18n
    .use(LanguageDetector) // 自动检测语言
    .use(initReactI18next) // React 绑定
    .init({
        resources: {
            en: {
                translation: {
                    bot_choose: "Select Bot",

                    dashboard_name: "Bot Dashboard",
                    admin_name: "admin",
                    message_num: "Message Number",
                    user_num: "User Number",
                    running_time: "Running Time",
                    date_range: "Select Date Range",
                    day: "Days",
                    message_new_num: "Message New Number",
                    user_new_num: "User New Number",

                    user_manage: "User Management",
                    add_user: "Add User",

                    bot_manage: "Bot Management",
                    add_bot: "Add Bot",

                    mcp_manage: "MCP Management",
                    add_mcp: "Add MCP",
                    sync_mcp: "Sync MCP",

                    bot_user_manage: "Bot User Management",
                    search_user_id: "Search User ID",
                    add_token: "Add Token",

                    bot_record_manage: "Bot Record Management",
                    add_record: "Add Record",

                    search: "Search",

                    id: "ID",
                    name: "Name",
                    user_id: "User ID",
                    mode: "Mode",
                    model: "Model",
                    question: "Question",
                    answer: "Answer",
                    token: "Token",
                    available_token: "Available Token",
                    description: "Description",
                    address: "Address",
                    status: "status",
                    username: "Username",
                    create_time: "Create Time",
                    update_time: "Update Time",
                    action: "Action",

                    submit: "Submit",
                    cancel: "Cancel",
                    edit: "Edit",
                    delete: "Delete",
                    disable: "Disable",
                    enable: "Enable",
                    save: "Save",
                    add: "Add",
                    command: "Command",
                    start: "Start",
                    stop: "Stop",
                    local_start: "Local Start",

                    address_placeholder: "Search By Address",
                    username_placeholder: "Search By Username",
                    user_id_placeholder: "Search By User ID",
                    message_placeholder: "Type your message here...",

                    dashboard: "Dashboard",
                    admin_users: "AdminUsers",
                    bots: "Bots",
                    mcp: "MCP",
                    bot_users: "BotUsers",
                    bot_chats: "BotChats",
                    chat: "Chat",
                    log: "Log",

                    welcome_back: "Welcome Back",
                    password: "Password",
                    login: "Login",

                    type: "Type",
                    rag: "RAG",
                }
            },
            zh: {
                translation: {
                    bot_choose: "选择机器人",

                    dashboard_name: "机器人面板",
                    admin_name: "管理员",
                    message_num: "消息总数",
                    user_num: "用户总数",
                    running_time: "运行时间",
                    date_range: "选择日期范围",
                    day: "天",
                    message_new_num: "新增消息数",
                    user_new_num: "新增用户数",

                    user_manage: "用户管理",
                    add_user: "添加用户",

                    bot_manage: "机器人管理",
                    add_bot: "添加机器人",

                    mcp_manage: "MCP 管理",
                    add_mcp: "添加 MCP",
                    sync_mcp: "同步 MCP",

                    bot_user_manage: "机器人用户管理",
                    search_user_id: "搜索用户 ID",
                    add_token: "添加令牌",

                    bot_record_manage: "机器人记录管理",
                    add_record: "添加记录",

                    search: "搜索",

                    id: "编号",
                    name: "名称",
                    user_id: "用户 ID",
                    mode: "模式",
                    model: "模型",
                    question: "问题",
                    answer: "回答",
                    token: "令牌",
                    available_token: "可用令牌",
                    description: "描述",
                    address: "地址",
                    status: "状态",
                    username: "用户名",
                    create_time: "创建时间",
                    update_time: "更新时间",
                    action: "操作",

                    submit: "提交",
                    cancel: "取消",
                    edit: "编辑",
                    delete: "删除",
                    disable: "禁用",
                    enable: "启用",
                    save: "保存",
                    add: "添加",
                    command: "命令",
                    start: "启动",
                    stop: "停止",
                    local_start: "本地启动",

                    address_placeholder: "通过地址搜索",
                    username_placeholder: "通过用户名搜索",
                    user_id_placeholder: "通过用户 ID 搜索",
                    message_placeholder: "在此输入您的消息...",

                    dashboard: "仪表盘",
                    admin_users: "管理员用户",
                    bots: "机器人",
                    mcp: "MCP",
                    bot_users: "机器人用户",
                    bot_chats: "机器人聊天记录",
                    chat: "聊天",
                    log: "日志",

                    welcome_back: "欢迎回来",
                    password: "密码",
                    login: "登录",

                    type: "类型",
                    rag: "RAG",
                }
            }
        },
        fallbackLng: 'en',
        interpolation: {
            escapeValue: false,
        }
    });

export default i18n;
