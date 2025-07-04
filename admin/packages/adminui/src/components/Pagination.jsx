// src/components/Pagination.jsx
import React from "react";

export default function Pagination({ page, pageSize, total, onPageChange }) {
    const totalPages = Math.ceil(total / pageSize);

    const pages = Array.from({ length: totalPages }, (_, i) => i + 1);

    return (
        <div className="flex justify-center mt-6 space-x-2">
            {pages.map((p) => (
                <button
                    key={p}
                    onClick={() => onPageChange(p)}
                    className={`px-3 py-1 rounded border ${
                        p === page ? "bg-blue-600 text-white" : "bg-white text-gray-800"
                    }`}
                >
                    {p}
                </button>
            ))}
        </div>
    );
}
