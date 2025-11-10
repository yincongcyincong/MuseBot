import React from "react";


const InputField = ({ label, name, value, onChange, placeholder, readOnly = false }) => (
    <div>
        <label className="block text-sm font-medium text-gray-700">{label}</label>
        <input
            type="text"
            name={name}
            value={value || ""}
            onChange={onChange}
            readOnly={readOnly}
            placeholder={placeholder}
            className={`w-full px-3 py-2 border border-gray-300 rounded ${readOnly ? 'bg-gray-100' : 'bg-white'} text-gray-700`}
        />
    </div>
);

export default InputField;