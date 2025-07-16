import React, { useState } from "react";
import { Link, useLocation } from "react-router-dom";
import {
    LayoutDashboard,
    Users,
    Bot,
    Database,
    UserCircle,
    MessageCircle,
    MessageSquare,
    ChevronFirst,
    ChevronLast,
} from "lucide-react";

export default function Sidebar() {
    const location = useLocation();
    const [collapsed, setCollapsed] = useState(false);

    const links = [
        { path: "/dashboard", label: "Dashboard", icon: LayoutDashboard },
        { path: "/admins", label: "Users", icon: Users },
        { path: "/bot", label: "Bots", icon: Bot },
        { path: "/mcp", label: "MCP", icon: Database },
        { path: "/users", label: "BotUsers", icon: UserCircle },
        { path: "/chats", label: "BotChats", icon: MessageCircle },
        { path: "/communicate", label: "Chat", icon: MessageSquare }, // 改了图标
    ];

    return (
        <div
            className={`h-full bg-gradient-to-b from-indigo-700 via-indigo-800 to-indigo-900 p-4 shadow-lg text-gray-100 transition-all duration-300 ${
                collapsed ? "w-20" : "w-60"
            }`}
        >
            <div className="flex justify-center mb-6">
                <button
                    onClick={() => setCollapsed(!collapsed)}
                    className="text-white p-1 rounded hover:bg-indigo-600 transition"
                >
                    {collapsed ? <ChevronLast size={20} /> : <ChevronFirst size={20} />}
                </button>
            </div>

            <nav className="space-y-3">
                {links.map(({ path, label, icon: Icon }) => {
                    const isActive = location.pathname === path;
                    return (
                        <Link
                            key={path}
                            to={path}
                            className={`flex items-center ${
                                collapsed ? "justify-center" : "gap-3"
                            } px-3 py-3 rounded-lg text-sm font-semibold transition-colors ${
                                isActive
                                    ? "bg-white bg-opacity-20 text-white shadow-md"
                                    : "text-indigo-300 hover:bg-white hover:bg-opacity-30 hover:text-white"
                            }`}
                        >
                            <Icon size={20} />
                            {!collapsed && <span>{label}</span>}
                        </Link>
                    );
                })}
            </nav>
        </div>
    );
}
