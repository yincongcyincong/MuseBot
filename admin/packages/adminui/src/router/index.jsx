import React from "react";
import { Navigate, Route, Routes } from "react-router-dom";
import Layout from "../components/Layout";
import LoginPage from "../pages/LoginPage";
import Dashboard from "../pages/Dashboard";
import Users from "../pages/Users";
import Bot from "../pages/Bot";
import { useUser } from "../context/UserContext.jsx";
import TestPage from "../pages/TestPage.jsx";
import BotUser from "../pages/BotUser.jsx";
import BotChat from "../pages/BotChat.jsx";

export default function Router() {
    const { isAuthenticated, isLoading } = useUser(); // ✨ 1. 获取 isLoading 状态

    if (isLoading) {
        return (
            <div className="fixed inset-0 flex items-center justify-center bg-white">
                <div className="flex space-x-2">
                    <div className="w-4 h-4 bg-gray-400 rounded-full animate-pulse [animation-delay:-0.3s]"></div>
                    <div className="w-4 h-4 bg-gray-400 rounded-full animate-pulse [animation-delay:-0.15s]"></div>
                    <div className="w-4 h-4 bg-gray-400 rounded-full animate-pulse"></div>
                </div>
            </div>
        );
    }
    return (
        <Routes>
            {/* 登录页面的逻辑保持不变 */}
            <Route
                path="/login"
                element={!isAuthenticated ? <LoginPage /> : <Navigate to="/dashboard" />}
            />
            {/* 只有在 isAuthenticated 为 true 时才渲染受保护的路由 */}
            {isAuthenticated && (
                <Route path="/" element={<Layout />}>
                    <Route path="dashboard" element={<Dashboard />} />
                    <Route path="admins" element={<Users />} />
                    <Route path="bot" element={<Bot />} />
                    <Route path="users" element={<BotUser />} />
                    <Route path="chats" element={<BotChat />} />
                    <Route path="test" element={<TestPage />} />
                    {/* 从根路径 / 跳转到看板页 */}
                    <Route index element={<Navigate to="/dashboard" />} />
                </Route>
            )}
            {/* 兜底路由：根据登录状态重定向到正确页面 */}
            <Route
                path="*"
                element={<Navigate to={isAuthenticated ? "/dashboard" : "/login"} />}
            />
        </Routes>
    );
}