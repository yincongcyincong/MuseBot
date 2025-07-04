import React from "react";
import { Link, useLocation } from "react-router-dom";

export default function Sidebar() {
    const location = useLocation();

    const links = [
        { path: "/dashboard", label: "Dashboard" },
        { path: "/users", label: "Users" },
        { path: "/bot", label: "Bot" },
    ];

    return (
        <div className="w-60 h-screen bg-gradient-to-b from-indigo-700 via-indigo-800 to-indigo-900 p-6 shadow-lg text-gray-100">
            <nav className="space-y-3">
                {links.map((link) => {
                    const isActive = location.pathname === link.path;
                    return (
                        <Link
                            key={link.path}
                            to={link.path}
                            className={`block px-5 py-3 rounded-lg text-sm font-semibold transition-colors
                                ${
                                isActive
                                    ? "bg-white bg-opacity-20 text-white shadow-md"
                                    : "text-indigo-300 hover:bg-white hover:bg-opacity-30 hover:text-white"
                            }`}
                        >
                            {link.label}
                        </Link>
                    );
                })}
            </nav>
        </div>
    );
}
