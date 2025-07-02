import React, { useEffect, useState } from "react";
import Sidebar from "./Sidebar";
import Header from "./Header";
import { Outlet, useNavigate } from "react-router-dom";
import {useUser} from "../context/UserContext.jsx";

export default function Layout() {
    const [userInfo, setUserInfo] = useState({ username: "" });
    const { user } = useUser();
    useEffect(() => {
        if (user) {
            setUserInfo(user);
        }
    }, [user]);

    return (
        <div style={{ display: "flex", flexDirection: "column", height: "100vh" }}>
            {/* 顶部 Header */}
            <Header username={userInfo.username} />

            {/* 底部 Sidebar + Content */}
            <div style={{ display: "flex", flex: 1 }}>
                <Sidebar />
                <div style={{ flex: 1, padding: "20px", overflowY: "auto" }}>
                    <Outlet />
                </div>
            </div>
        </div>
    );
}
