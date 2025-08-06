import React, { useEffect } from "react";

function Toast({ message, type = "error", onClose }) {
    useEffect(() => {
        const timer = setTimeout(() => {
            onClose();
        }, 3000);

        return () => clearTimeout(timer);
    }, [onClose]);

    const bgColor = type === "error" ? "bg-red-500" : "bg-green-500";

    return (
        <div
            className={`fixed top-1/5 left-1/2 transform -translate-x-1/2 -translate-y-1/2 z-[1000]
                        px-6 py-3 rounded shadow-lg text-white text-center ${bgColor}`}
        >
            {message}
        </div>
    );
}

export default Toast;
