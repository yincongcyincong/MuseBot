import React from "react";
import { BrowserRouter } from "react-router-dom";
import Router from "./router";
import { UserProvider } from "./context/UserContext.jsx";

function AppWithAuthCheck() {
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
