import React from "react";
import LoginForm from "../components/LoginForm";

export default function LoginPage() {
    return (
        <div style={{
            height: "100vh",
            display: "flex",
            justifyContent: "center",
            alignItems: "center"
        }}>
            <LoginForm />
        </div>
    );
}
