import React, { useState } from "react";
import { useNavigate } from "react-router-dom";
import { login } from "../api/auth";
import {useUser} from "../context/UserContext.jsx";

export default function LoginForm() {
    const [username, setUsername] = useState("");
    const [password, setPassword] = useState("");
    const [error, setError] = useState(null);
    const navigate = useNavigate();
    const {setUser} = useUser();

    const handleSubmit = async (e) => {
        e.preventDefault();
        setError(null);

        try {
            const data = await login(username, password);
            setUser(data);
            navigate("/dashboard");
        } catch (err) {
            setError(err.message || "登录失败");
        }
    };

    return (
        <form onSubmit={handleSubmit} style={{ width: 300 }}>
            <h2>登录后台</h2>
            {error && <div style={{ color: "red", marginBottom: 10 }}>{error}</div>}
            <div>
                <input
                    type="text"
                    placeholder="用户名"
                    value={username}
                    onChange={(e) => setUsername(e.target.value)}
                    required
                    style={{ width: "100%", marginBottom: 10 }}
                />
            </div>
            <div>
                <input
                    type="password"
                    placeholder="密码"
                    value={password}
                    onChange={(e) => setPassword(e.target.value)}
                    required
                    style={{ width: "100%", marginBottom: 10 }}
                />
            </div>
            <button type="submit" style={{ width: "100%" }}>登录</button>
        </form>
    );
}
