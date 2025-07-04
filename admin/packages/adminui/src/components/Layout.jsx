import React, { useEffect, useState } from "react";
import Sidebar from "./Sidebar";
import Header from "./Header";
import { Outlet } from "react-router-dom";
import { useUser } from "../context/UserContext.jsx";

export default function Layout() {
    const [userInfo, setUserInfo] = useState({ username: "" });
    const { user } = useUser();

    useEffect(() => {
        if (user) {
            setUserInfo(user);
        }
    }, [user]);

    return (
        <div className="flex flex-col h-screen">
            {/* Header，固定高度，底部有分割线 */}
            <div className="h-24">
                <Header username={userInfo.username} />
            </div>

            {/* 主体区域：侧边栏 + 内容 */}
            <div className="flex flex-1">
                {/* 侧边栏，固定宽度，右边有分割线 */}
                <div className="w-56">
                    <Sidebar />
                </div>

                {/* 内容区域：加 overflow-auto 防止溢出 */}
                <div className="flex-1 overflow-auto">
                    {/* 横向滚动包装 */}
                    <div className="overflow-x-auto">
                        <Outlet />
                    </div>
                </div>
            </div>
        </div>
    );
}
