import React from "react";
import { BrowserRouter } from "react-router-dom";
import Router from "./router";
import { UserProvider } from "./context/UserContext.jsx";
import { useAuthCheck } from "./utils/auth.js"; // 假设路径

function AppWithAuthCheck() {
    useAuthCheck();
    return <Router />;
}

export default function App() {
    return (
        <BrowserRouter>
            <UserProvider>
                <AppWithAuthCheck />
            </UserProvider>
        </BrowserRouter>
    );
}
