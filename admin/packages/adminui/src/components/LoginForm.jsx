import React, { useState } from "react";
import { useNavigate } from "react-router-dom";
import { login } from "../api/auth";
import { useUser } from "../context/UserContext.jsx";
import {useTranslation} from "react-i18next";

export default function LoginForm() {
    const [username, setUsername] = useState("");
    const [password, setPassword] = useState("");
    const [error, setError] = useState(null);
    const navigate = useNavigate();
    const { setUser } = useUser();

    const { t } = useTranslation();

    const handleSubmit = async (e) => {
        e.preventDefault();
        setError(null);

        try {
            const data = await login(username, password);
            setUser(data);
            navigate("/dashboard");
        } catch (err) {
            setError(err.message || "Login failed");
        }
    };

    return (
        <div className="min-h-screen flex items-center justify-center bg-gradient-to-tr from-indigo-900 via-purple-900 to-pink-900 px-6">
            <form
                onSubmit={handleSubmit}
                className="bg-white bg-opacity-90 backdrop-blur-md rounded-2xl shadow-xl p-10 max-w-md w-full animate-fadeIn"
            >
                <h2 className="text-4xl font-extrabold text-center mb-8 text-indigo-700 drop-shadow-lg">
                    {t("welcome_back")}
                </h2>

                {error && (
                    <div className="mb-4 text-center text-red-600 font-medium animate-pulse">
                        {error}
                    </div>
                )}

                <div className="mb-6 relative">
                    <input
                        type="text"
                        placeholder={t("username")}
                        value={username}
                        onChange={(e) => setUsername(e.target.value)}
                        required
                        className="peer w-full px-5 py-3 rounded-xl border border-gray-300
              focus:outline-none focus:ring-4 focus:ring-indigo-400
              transition transform duration-200
              focus:shadow-lg focus:scale-105"
                    />
                </div>

                <div className="mb-8 relative">
                    <input
                        type="password"
                        placeholder={t("password")}
                        value={password}
                        onChange={(e) => setPassword(e.target.value)}
                        required
                        className="peer w-full px-5 py-3 rounded-xl border border-gray-300
              focus:outline-none focus:ring-4 focus:ring-indigo-400
              transition transform duration-200
              focus:shadow-lg focus:scale-105"
                    />
                </div>

                <button
                    type="submit"
                    className="w-full py-3 bg-indigo-600 text-white font-semibold rounded-xl shadow-lg
            hover:bg-indigo-700 active:scale-95 transition-transform duration-150"
                >
                    {t("login")}
                </button>

                <p className="mt-6 text-center text-gray-600 text-sm select-none">
                    Powered by <a href="https://github.com/yincongcyincong/MuseBot" className="font-bold text-indigo-600">Jack Yin</a>
                </p>
            </form>
        </div>
    );
}
