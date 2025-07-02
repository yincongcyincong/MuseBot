import {useEffect} from "react";
import {useUser} from "../context/UserContext.jsx";

export function useAuthCheck() {
    const {setUser} = useUser();

    useEffect(() => {
        async function checkAuth() {
            try {
                const response = await fetch("/user/me", {
                    credentials: "include",
                });

                if (!response.ok) {
                    setUser(null);
                    return;
                }

                const data = await response.json();
                if (data?.code === 0) {
                    setUser(data.data);
                } else {
                    setUser(null);
                }
            } catch {
                setUser(null);
            }
        }

        checkAuth();
    }, [setUser]);
}
