// components/ConfirmModal.jsx
import React from "react";

function ConfirmModal({ visible, title = "Confirm", message, onCancel, onConfirm }) {
    if (!visible) return null;

    return (
        <div className="fixed inset-0 bg-black bg-opacity-30 flex items-center justify-center z-50">
            <div className="bg-white p-6 rounded shadow-lg max-w-sm w-full">
                <h3 className="text-lg font-semibold text-gray-800 mb-4">{title}</h3>
                <p className="text-gray-700 mb-4">{message}</p>
                <div className="flex justify-end space-x-4">
                    <button onClick={onCancel} className="px-4 py-2 text-gray-600 hover:text-gray-800">
                        Cancel
                    </button>
                    <button onClick={onConfirm} className="px-4 py-2 bg-red-600 text-white rounded hover:bg-red-700">
                        Delete
                    </button>
                </div>
            </div>
        </div>
    );
}

export default ConfirmModal;
