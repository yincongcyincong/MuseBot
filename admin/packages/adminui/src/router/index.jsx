import React from "react";
import { Navigate, Route, Routes } from "react-router-dom";
import Layout from "../components/Layout";
import LoginPage from "../pages/LoginPage";
import Dashboard from "../pages/Dashboard";
import Users from "../pages/Users";
import Bot from "../pages/Bot";
import { useUser } from "../context/UserContext.jsx";
import TestPage from "../pages/TestPage.jsx";

export default function Router() {
    const { isAuthenticated } = useUser();

    // if (!isAuthenticated) {
    //     // 还没确定登录状态，可以返回加载中页面或者null
    //     return <div>Loading...</div>;
    // }

    return (
        <Routes>
            <Route
                path="/login"
                element={!isAuthenticated ? <LoginPage /> : <Navigate to="/dashboard" />}
            />
            {isAuthenticated && (
                <Route path="/" element={<Layout />}>
                    <Route path="dashboard" element={<Dashboard />} />
                    <Route path="users" element={<Users />} />
                    <Route path="bot" element={<Bot />} />
                    <Route path="test" element={<TestPage />} />
                    <Route index element={<Dashboard />} />
                </Route>
            )}
            {!isAuthenticated && <Route path="*" element={<Navigate to="/login" />} />}
            {isAuthenticated && <Route path="*" element={<Navigate to="/dashboard" />} />}
        </Routes>
    );
}
