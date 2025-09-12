import React, { useEffect, useState } from "react";
import Toast from "../components/Toast";
import {useTranslation} from "react-i18next";

function ConfigForm({ botId }) {
    const [configData, setConfigData] = useState(null);
    const [loading, setLoading] = useState(true);
    const [saveStatusMap, setSaveStatusMap] = useState({});
    const [toast, setToast] = useState({ show: false, message: "", type: "error" });

    const { t } = useTranslation();

    const showToast = (message, type = "error") => {
        setToast({ show: true, message, type });
    };

    useEffect(() => {
        fetchConfig();
    }, []);

    const fetchConfig = async () => {
        try {
            const res = await fetch(`/bot/conf/get?id=${botId}`);
            const data = await res.json();
            if (data.code !== 0) {
                showToast(data.message || "Failed to fetch config");
                return;
            }
            setConfigData(data.data);
        } catch (err) {
            showToast("Request error: " + err.message);
        } finally {
            setLoading(false);
        }
    };

    const handleChange = (section, key, value) => {
        setConfigData((prev) => ({
            ...prev,
            [section]: {
                ...prev[section],
                [key]: value,
            },
        }));
    };

    const handleSaveSingle = async (section, key) => {
        const value = configData[section][key];
        const statusKey = `${section}.${key}`;
        setSaveStatusMap((prev) => ({ ...prev, [statusKey]: "Saving..." }));

        try {
            const res = await fetch(`/bot/conf/update?id=${botId}`, {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({ type: section, key, value }),
            });

            const data = await res.json();
            if (data.code !== 0) {
                setSaveStatusMap((prev) => ({ ...prev, [statusKey]: "❌ Failed" }));
                showToast(data.message || "Update failed");
                return;
            }

            setSaveStatusMap((prev) => ({ ...prev, [statusKey]: "✔ Saved" }));
            showToast("Saved successfully", "success");
        } catch (err) {
            setSaveStatusMap((prev) => ({ ...prev, [statusKey]: "❌ Failed" }));
            showToast("Request error: " + err.message);
        }

        setTimeout(() => {
            setSaveStatusMap((prev) => {
                const copy = { ...prev };
                delete copy[statusKey];
                return copy;
            });
        }, 3000);
    };

    if (loading) return <div className="p-4 text-gray-600">Loading config...</div>;
    if (!configData) return <div className="p-4 text-gray-600">No config data</div>;

    return (
        <div className="p-5 max-h-[70vh] overflow-y-auto border border-gray-300 rounded-lg bg-white relative">
            {toast.show && (
                <Toast
                    message={toast.message}
                    type={toast.type}
                    onClose={() => setToast({ ...toast, show: false })}
                />
            )}

            <h2 className="text-xl font-semibold mb-6">System Configurations</h2>

            {Object.entries(configData).map(([sectionName, sectionValues]) => (
                <div key={sectionName} className="mb-10">
                    <h3 className="text-lg font-bold mb-4 border-b border-gray-200 pb-1">
                        {sectionName.toUpperCase()}
                    </h3>

                    <div className="space-y-4">
                        {Object.entries(sectionValues).map(([key, value]) => {
                            const statusKey = `${sectionName}.${key}`;
                            const statusText = saveStatusMap[statusKey] || "";

                            return (
                                <div key={key} className="flex items-center space-x-4">
                                    <label className="w-48 font-semibold text-gray-700 break-words">{key}</label>

                                    <input
                                        type="text"
                                        value={value ?? ""}
                                        onChange={(e) => handleChange(sectionName, key, e.target.value)}
                                        className="flex-1 px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-400 focus:border-transparent text-sm"
                                    />

                                    <button
                                        onClick={() => handleSaveSingle(sectionName, key)}
                                        className="px-3 py-1 bg-blue-600 hover:bg-blue-700 text-white rounded-md text-sm transition"
                                    >
                                        {t("save")}
                                    </button>

                                    <span className="text-gray-500 text-sm select-none">{statusText}</span>
                                </div>
                            );
                        })}
                    </div>
                </div>
            ))}
        </div>
    );
}

export default ConfigForm;
