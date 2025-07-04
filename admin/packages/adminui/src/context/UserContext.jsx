import React, { createContext, useContext, useState, useEffect } from "react";

const UserContext = createContext(null);

export function UserProvider({ children }) {
    const [user, setUser] = useState(null);
    // 1. 管理一个加载状态
    const [isLoading, setIsLoading] = useState(true);

    // 2. 将 auth.js 的逻辑移到这里
    useEffect(() => {
        async function checkAuth() {
            try {
                const response = await fetch("/user/me", {
                    credentials: "include", // 确保发送 cookie
                });

                if (response.ok) {
                    const data = await response.json();
                    // 假设 code === 0 表示成功
                    if (data?.code === 0) {
                        setUser(data.data);
                    } else {
                        setUser(null);
                    }
                } else {
                    setUser(null);
                }
            } catch {
                setUser(null);
            } finally {
                // 3. 无论认证成功与否，检查结束后都设置加载完成
                setIsLoading(false);
            }
        }

        checkAuth();
    }, []); // 空依赖数组确保只在应用加载时执行一次

    // 4. isAuthenticated 派生自 user 状态
    const isAuthenticated = !!user;

    // 5. 将所有相关状态提供给子组件
    return (
        <UserContext.Provider value={{ user, setUser, isAuthenticated, isLoading }}>
            {/* 只有在加载完成后才渲染子组件，避免闪烁 */}
            {!isLoading && children}
        </UserContext.Provider>
    );
}

export function useUser() {
    const context = useContext(UserContext);
    if (!context) {
        throw new Error("useUser must be used within a UserProvider");
    }
    return context;
}