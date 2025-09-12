import React, { useState, useRef, useEffect } from "react";
import { useTranslation } from "react-i18next";

export default function Header({ username = "USER", avatarUrl = "" }) {
    const [menuOpen, setMenuOpen] = useState(false);
    const [langOpen, setLangOpen] = useState(false);
    const menuRef = useRef(null);
    const langRef = useRef(null);
    const { i18n } = useTranslation();

    const handleLogout = () => {
        fetch("/user/logout", {
            method: "POST",
            credentials: "include",
        }).then(() => {
            window.location.href = "/login";
        });
    };

    const changeLanguage = (lng) => {
        i18n.changeLanguage(lng);
        setLangOpen(false);
    };

    useEffect(() => {
        function handleClickOutside(event) {
            if (
                (menuRef.current && !menuRef.current.contains(event.target)) &&
                (langRef.current && !langRef.current.contains(event.target))
            ) {
                setMenuOpen(false);
                setLangOpen(false);
            }
        }
        document.addEventListener("mousedown", handleClickOutside);
        return () => document.removeEventListener("mousedown", handleClickOutside);
    }, []);

    return (
        <header className="flex justify-between items-center px-6 py-4 bg-gradient-to-r from-indigo-600 via-purple-600 to-pink-600 shadow-lg z-1000">
            {/* 左侧标题 */}
            <div className="text-xl font-bold text-white drop-shadow-md">
                <a
                    href="https://github.com/yincongcyincong/MuseBot"
                    target="_blank"
                    rel="noopener noreferrer"
                >
                    MuseBot
                </a>
            </div>

            <div className="flex items-center space-x-4">
                {/* 🌐 语言选择 */}
                <div className="relative" ref={langRef}>
                    <button
                        onClick={() => setLangOpen(!langOpen)}
                        className="flex items-center space-x-1 bg-white bg-opacity-90 hover:bg-opacity-100 active:bg-opacity-80
                                   px-3 py-1.5 rounded-full text-sm font-semibold text-gray-800
                                   focus:outline-none focus:ring-2 focus:ring-indigo-300"
                    >
                        🌐 {i18n.language.toUpperCase()}
                        <svg
                            className={`w-3 h-3 text-indigo-600 transition-transform duration-200 ${
                                langOpen ? "rotate-180" : "rotate-0"
                            }`}
                            fill="none"
                            stroke="currentColor"
                            strokeWidth="2"
                            viewBox="0 0 24 24"
                        >
                            <path strokeLinecap="round" strokeLinejoin="round" d="M19 9l-7 7-7-7" />
                        </svg>
                    </button>

                    {/* 下拉语言菜单 */}
                    <div
                        className={`absolute right-0 mt-2 w-28 bg-white border border-gray-200 rounded shadow-lg
                                    transition-all duration-200 ease-in-out
                                    ${langOpen ? "opacity-100 scale-100 max-h-40" : "opacity-0 scale-95 max-h-0 pointer-events-none"}`}
                    >
                        <button
                            onClick={() => changeLanguage("en")}
                            className="block w-full text-left px-4 py-2 text-sm text-gray-700 hover:bg-indigo-100"
                        >
                            English
                        </button>
                        <button
                            onClick={() => changeLanguage("zh")}
                            className="block w-full text-left px-4 py-2 text-sm text-gray-700 hover:bg-indigo-100"
                        >
                            中文
                        </button>
                    </div>
                </div>

                {/* 用户菜单 */}
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

                    <div
                        className={`absolute right-0 mt-2 w-36 bg-white border border-gray-200 rounded shadow-lg
                                    transition-all duration-200 ease-in-out
                                    ${menuOpen ? "opacity-100 scale-100 max-h-40" : "opacity-0 scale-95 max-h-0 pointer-events-none"}`}
                    >
                        <button
                            onClick={handleLogout}
                            className="block w-full text-left px-4 py-2 text-sm text-gray-700 hover:bg-indigo-100 transition-colors"
                        >
                            LOGOUT
                        </button>
                    </div>
                </div>
            </div>
        </header>
    );
}
