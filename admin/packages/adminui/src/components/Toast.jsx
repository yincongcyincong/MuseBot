import React, { useEffect } from "react";

function Toast({ message, type = "error", onClose }) {
    useEffect(() => {
        const timer = setTimeout(() => {
            onClose();
        }, 3000); // 自动消失时间

        return () => clearTimeout(timer);
    }, [onClose]);

    const bgColor = type === "error" ? "bg-red-500" : "bg-green-500";

    return (
        <div className={`fixed top-5 right-5 z-50 px-4 py-2 rounded shadow-lg text-white ${bgColor}`}>
            {message}
        </div>
    );
}

export default Toast;
