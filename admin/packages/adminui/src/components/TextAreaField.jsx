import React from "react";

const TextareaField = ({ label, name, value, onChange, placeholder }) => (
    <div>
        <label className="block text-sm font-medium text-gray-700">{label}</label>
        <textarea
            name={name}
            value={value || ""}
            onChange={onChange}
            placeholder={placeholder}
            rows="3"
            className="w-full px-3 py-2 border border-gray-300 rounded bg-white text-gray-700"
        />
    </div>
);

export default TextareaField;