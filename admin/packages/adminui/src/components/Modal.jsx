// src/components/Modal.jsx
import React from "react";

export default function Modal({ visible, title, children, onClose }) {
    if (!visible) return null;

    return (
        <div
            style={{
                position: "fixed",
                top: 0,
                left: 0,
                width: "100%",
                height: "100%",
                background: "rgba(0,0,0,0.3)",
                display: "flex",
                alignItems: "center",
                justifyContent: "center",
                zIndex: 999,
            }}
        >
            <div
                style={{
                    background: "#fff",
                    padding: "20px",
                    borderRadius: "8px",
                    minWidth: "300px",
                    maxWidth: "90%",
                    boxShadow: "0 4px 8px rgba(0,0,0,0.2)",
                }}
            >
                <div style={{ display: "flex", justifyContent: "space-between", marginBottom: "10px" }}>
                    <h3>{title}</h3>
                    <button onClick={onClose} style={{ fontSize: "16px" }}>✕</button>
                </div>
                {children}
            </div>
        </div>
    );
}
