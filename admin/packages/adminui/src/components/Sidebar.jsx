import React from "react";
import { Link, useLocation } from "react-router-dom";
import {
    LayoutDashboard,
    Users,
    Bot,
    Database,
    UserCircle,
    MessageSquare,
} from "lucide-react";

export default function Sidebar() {
    const location = useLocation();

    const links = [
        { path: "/dashboard", label: "Dashboard", icon: LayoutDashboard },
        { path: "/admins", label: "Users", icon: Users },
        { path: "/bot", label: "Bots", icon: Bot },
        { path: "/mcp", label: "MCP", icon: Database },
        { path: "/users", label: "BotUsers", icon: UserCircle },
        { path: "/chats", label: "BotChats", icon: MessageSquare },
    ];

    return (
        <div className="w-60 h-full bg-gradient-to-b from-indigo-700 via-indigo-800 to-indigo-900 p-6 shadow-lg text-gray-100">
            <nav className="space-y-3">
                {links.map(({ path, label, icon: Icon }) => {
                    const isActive = location.pathname === path;
                    return (
                        <Link
                            key={path}
                            to={path}
                            className={`flex items-center gap-3 px-5 py-3 rounded-lg text-sm font-semibold transition-colors
                                ${
                                isActive
                                    ? "bg-white bg-opacity-20 text-white shadow-md"
                                    : "text-indigo-300 hover:bg-white hover:bg-opacity-30 hover:text-white"
                            }`}
                        >
                            <Icon size={18} />
                            {label}
                        </Link>
                    );
                })}
            </nav>
        </div>
    );
}
