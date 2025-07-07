import React from "react";

export default function Pagination({ page, pageSize, total, onPageChange }) {
    const totalPages = Math.ceil(total / pageSize);

    const getPagesToShow = () => {
        const pages = new Set();

        if (totalPages <= 7) {
            for (let i = 1; i <= totalPages; i++) {
                pages.add(i);
            }
        } else {
            pages.add(1); // first page

            if (page > 3) {
                pages.add("...");
            }

            for (let i = page - 1; i <= page + 1; i++) {
                if (i > 1 && i < totalPages) {
                    pages.add(i);
                }
            }

            if (page < totalPages - 2) {
                pages.add("...");
            }

            pages.add(totalPages); // last page
        }

        return Array.from(pages);
    };

    const pagesToShow = getPagesToShow();

    return (
        <div className="flex justify-center mt-6 space-x-2">
            {pagesToShow.map((p, index) =>
                p === "..." ? (
                    <span key={`ellipsis-${index}`} className="px-3 py-1 text-gray-500">
                        ...
                    </span>
                ) : (
                    <button
                        key={`page-${p}`}
                        onClick={() => onPageChange(p)}
                        className={`px-3 py-1 rounded border ${
                            p === page ? "bg-blue-600 text-white" : "bg-white text-gray-800"
                        }`}
                    >
                        {p}
                    </button>
                )
            )}
        </div>
    );
}
