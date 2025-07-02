import React from "react";
import { Navigate, Route, Routes } from "react-router-dom";
import Layout from "../components/Layout";
import LoginPage from "../pages/LoginPage";
import Dashboard from "../pages/Dashboard";
import Users from "../pages/Users";
import Bot from "../pages/Bot";
import { useUser } from "../context/UserContext.jsx";

export default function Router() {
    const { isAuthenticated } = useUser();

    if (isAuthenticated === null) {
        // 还没确定登录状态，可以返回加载中页面或者null
        return <div>Loading...</div>;
    }

    return (
        <Routes>
            <Route
                path="/login"
                element={!isAuthenticated ? <LoginPage /> : <Navigate to="/dashboard" />}
            />
            {isAuthenticated && (
                <Route path="/" element={<Layout />}>
                    <Route index element={<Dashboard />} />
                    <Route path="dashboard" element={<Dashboard />} />
                    <Route path="users" element={<Users />} />
                    <Route path="bot" element={<Bot />} />
                </Route>
            )}
            {!isAuthenticated && <Route path="*" element={<Navigate to="/login" />} />}
        </Routes>
    );
}
