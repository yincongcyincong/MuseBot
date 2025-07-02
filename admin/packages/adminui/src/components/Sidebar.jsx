import React from "react";
import { Link, useLocation } from "react-router-dom";

function Sidebar() {
    const location = useLocation();

    const links = [
        { path: "/dashboard", label: "Dashboard" },
        { path: "/users", label: "Users" },
        { path: "/bot", label: "bot" },
    ];

    return (
        <div style={{ width: "200px", background: "#f4f4f4", padding: "20px" }}>
            {links.map((link) => (
                <div key={link.path} style={{ marginBottom: "10px" }}>
                    <Link
                        to={link.path}
                        style={{
                            color: location.pathname === link.path ? "blue" : "black",
                            textDecoration: "none",
                            fontWeight: location.pathname === link.path ? "bold" : "normal",
                        }}
                    >
                        {link.label}
                    </Link>
                </div>
            ))}
        </div>
    );
}

export default Sidebar;
