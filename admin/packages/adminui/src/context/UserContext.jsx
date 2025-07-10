import React, { createContext, useContext, useState, useEffect } from "react";

const UserContext = createContext(null);

export function UserProvider({ children }) {
    const [user, setUser] = useState(null);
    const [isLoading, setIsLoading] = useState(true);

    useEffect(() => {
        async function checkAuth() {
            try {
                const response = await fetch("/user/me", {
                    credentials: "include", // 确保发送 cookie
                });

                if (response.ok) {
                    const data = await response.json();
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
                setIsLoading(false);
            }
        }

        checkAuth();
    }, []);

    const isAuthenticated = !!user;

    return (
        <UserContext.Provider value={{ user, setUser, isAuthenticated, isLoading }}>
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