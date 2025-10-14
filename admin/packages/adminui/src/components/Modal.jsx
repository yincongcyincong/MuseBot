import React from "react";

export default function Modal({ visible, title, children, onClose }) {
    if (!visible) return null;

    const handleBackgroundClick = (e) => {
        if (e.target === e.currentTarget) {
            onClose();
        }
    };

    return (
        <div
            className="fixed inset-0 bg-black/30 flex items-center justify-center z-[999]"
            onClick={handleBackgroundClick}
        >
            <div
                className="bg-white p-6 rounded-lg shadow-lg max-w-[50%] w-full mx-4"
                onClick={(e) => e.stopPropagation()}
            >
                <div className="flex justify-between items-center mb-4">
                    <h3 className="text-lg font-semibold">{title}</h3>
                    <button
                        onClick={onClose}
                        className="text-gray-500 hover:text-gray-700 text-xl leading-none"
                    >
                        ✕
                    </button>
                </div>
                {children}
            </div>
        </div>
    );
}
