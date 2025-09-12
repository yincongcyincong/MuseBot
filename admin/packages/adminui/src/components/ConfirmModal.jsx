import React, { useState } from "react";
import {useTranslation} from "react-i18next";

function ConfirmModal({ visible, title = "Confirm", message, onCancel, onConfirm }) {
    const [loading, setLoading] = useState(false);

    const { t } = useTranslation();

    if (!visible) return null;

    const handleConfirm = async () => {
        setLoading(true);
        try {
            await onConfirm();
        } finally {
            setLoading(false);
        }
    };

    return (
        <div className="fixed inset-0 bg-black bg-opacity-30 flex items-center justify-center z-50">
            <div className="bg-white p-6 rounded shadow-lg max-w-sm w-full">
                <h3 className="text-lg font-semibold text-gray-800 mb-4">{title}</h3>
                <p className="text-gray-700 mb-4">{message}</p>
                <div className="flex justify-end space-x-4">
                    <button
                        onClick={onCancel}
                        disabled={loading}
                        className="px-4 py-2 text-gray-600 hover:text-gray-800 disabled:opacity-50"
                    >
                        {t("cancel")}
                    </button>
                    <button
                        onClick={handleConfirm}
                        disabled={loading}
                        className="px-4 py-2 bg-red-600 text-white rounded hover:bg-red-700 disabled:opacity-50 flex items-center justify-center min-w-[90px]"
                    >
                        {loading ? (
                            <div className="flex items-center space-x-2">
                                <svg
                                    className="animate-spin h-4 w-4 text-white"
                                    xmlns="http://www.w3.org/2000/svg"
                                    fill="none"
                                    viewBox="0 0 24 24"
                                >
                                    <circle
                                        className="opacity-25"
                                        cx="12"
                                        cy="12"
                                        r="10"
                                        stroke="currentColor"
                                        strokeWidth="4"
                                    />
                                    <path
                                        className="opacity-75"
                                        fill="currentColor"
                                        d="M4 12a8 8 0 018-8v4a4 4 0 00-4 4H4z"
                                    />
                                </svg>
                                <span>Loading</span>
                            </div>
                        ) : (
                            t("save")
                        )}
                    </button>
                </div>
            </div>
        </div>
    );
}

export default ConfirmModal;
