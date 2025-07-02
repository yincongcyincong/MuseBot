import React, { useState } from "react";
import "../styles/Header.css"; // 如果你用的是 CSS 模块或 Tailwind 可以替换掉

export default function Header({ username = "用户", avatarUrl = "" }) {
    const [menuOpen, setMenuOpen] = useState(false);

    const handleLogout = () => {
        fetch("/user/logout", {
            method: "POST",
            credentials: "include",
        }).then(() => {
            window.location.href = "/login";
        });
    };

    return (
        <div className="header">
            <div className="title">仪表盘</div>
            <div className="user-area" onClick={() => setMenuOpen(!menuOpen)}>
                <img className="avatar" src={avatarUrl || "/default-avatar.png"} alt="avatar" />
                <span className="username">{username}</span>
                <span className="arrow">▼</span>
                {menuOpen && (
                    <div className="dropdown">
                        <button onClick={handleLogout}>退出登录</button>
                    </div>
                )}
            </div>
        </div>
    );
}
