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
            <div className="h-24">
                <Header username={userInfo.username} />
            </div>

            <div className="flex flex-1">
                <Sidebar />

                <div className="flex-1 overflow-auto">
                    <div className="overflow-x-auto">
                        <Outlet />
                    </div>
                </div>
            </div>
        </div>
    );
}
