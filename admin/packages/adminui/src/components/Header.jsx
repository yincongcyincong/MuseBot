import React, { useState, useRef, useEffect } from "react";

export default function Header({ username = "用户", avatarUrl = "" }) {
    const [menuOpen, setMenuOpen] = useState(false);
    const menuRef = useRef(null);

    const handleLogout = () => {
        fetch("/user/logout", {
            method: "POST",
            credentials: "include",
        }).then(() => {
            window.location.href = "/login";
        });
    };

    useEffect(() => {
        function handleClickOutside(event) {
            if (menuRef.current && !menuRef.current.contains(event.target)) {
                setMenuOpen(false);
            }
        }
        if (menuOpen) {
            document.addEventListener("mousedown", handleClickOutside);
        } else {
            document.removeEventListener("mousedown", handleClickOutside);
        }
        return () => {
            document.removeEventListener("mousedown", handleClickOutside);
        };
    }, [menuOpen]);

    return (
        <header className="flex justify-between items-center px-6 py-4 bg-gradient-to-r from-indigo-600 via-purple-600 to-pink-600 shadow-lg">
            {/* 左侧标题 */}
            <div className="text-xl font-bold text-white drop-shadow-md">
                <a href="https://github.com/yincongcyincong/telegram-deepseek-bot"
                    target="_blank"
                    rel="noopener noreferrer">telegram-deepseek-bot</a>
            </div>

            {/* 右侧用户信息 */}
            <div className="relative" ref={menuRef}>
                <button
                    onClick={() => setMenuOpen(!menuOpen)}
                    className="flex items-center space-x-2 cursor-pointer select-none
                               bg-white bg-opacity-90 hover:bg-opacity-100 active:bg-opacity-80
                               transition-colors rounded-full px-3 py-1.5
                               focus:outline-none focus:ring-2 focus:ring-indigo-300"
                    aria-haspopup="true"
                    aria-expanded={menuOpen}
                >
                    <img
                        src={avatarUrl || "/avatar.jpeg"}
                        alt="avatar"
                        className="w-8 h-8 rounded-full border-2 border-indigo-500"
                    />
                    <span className="text-gray-800 text-sm font-semibold">{username}</span>
                    <svg
                        className={`w-3 h-3 text-indigo-600 transition-transform duration-200 ${
                            menuOpen ? "rotate-180" : "rotate-0"
                        }`}
                        fill="none"
                        stroke="currentColor"
                        strokeWidth="2"
                        viewBox="0 0 24 24"
                    >
                        <path strokeLinecap="round" strokeLinejoin="round" d="M19 9l-7 7-7-7" />
                    </svg>
                </button>

                {/* 下拉菜单 */}
                <div
                    className={`absolute right-0 mt-2 w-36 bg-white border border-gray-200 rounded shadow-lg
                                overflow-visible transform origin-top-right
                                transition-all duration-200 ease-in-out
                                ${menuOpen ? "opacity-100 scale-100 max-h-40" : "opacity-0 scale-95 max-h-0 pointer-events-none"}`}
                >
                    <button
                        onClick={handleLogout}
                        className="block w-full text-left px-4 py-2 text-sm text-gray-700 hover:bg-indigo-100 transition-colors"
                    >
                        退出登录
                    </button>
                </div>
            </div>
        </header>
    );
}
