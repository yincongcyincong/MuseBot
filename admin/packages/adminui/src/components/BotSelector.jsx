import React, { useEffect, useRef, useState } from "react";

function BotSelector({ value, onChange }) {
    const [bots, setBots] = useState([]);
    const [filteredBots, setFilteredBots] = useState([]);
    const [searchText, setSearchText] = useState("");
    const [dropdownOpen, setDropdownOpen] = useState(false);
    const wrapperRef = useRef(null);

    useEffect(() => {
        fetchBots();
    }, []);

    useEffect(() => {
        const lower = searchText.toLowerCase();
        setFilteredBots(bots.filter(bot => bot.address.toLowerCase().includes(lower)));
    }, [searchText, bots]);

    useEffect(() => {
        function handleClickOutside(event) {
            if (wrapperRef.current && !wrapperRef.current.contains(event.target)) {
                setDropdownOpen(false);
            }
        }

        document.addEventListener("mousedown", handleClickOutside);
        return () => document.removeEventListener("mousedown", handleClickOutside);
    }, []);

    const fetchBots = async () => {
        try {
            const res = await fetch("/bot/online");
            const data = await res.json();
            if (data.data && data.data.length > 0) {
                setBots(data.data);
                setFilteredBots(data.data);
                const defaultBot = data.data[0];
                setSearchText(defaultBot.address);
                onChange(defaultBot); // 选中第一个 bot
            }
        } catch (err) {
            console.error("Failed to fetch bots:", err);
        }
    };

    const handleSelectBot = (bot) => {
        setSearchText(bot.address);
        setDropdownOpen(false);
        onChange(bot);
    };

    return (
        <div className="relative" ref={wrapperRef}>
            <label className="block font-medium text-gray-700 mb-1">Select Bot:</label>
            <input
                type="text"
                value={searchText}
                onChange={e => {
                    setSearchText(e.target.value);
                    setDropdownOpen(true);
                }}
                onFocus={() => setDropdownOpen(true)}
                placeholder="Search and select bot"
                className="w-full px-4 py-2 border border-gray-300 rounded shadow-sm focus:outline-none focus:ring focus:border-blue-400"
            />
            {dropdownOpen && (
                <ul className="absolute z-10 mt-1 w-full max-h-48 overflow-auto bg-white border border-gray-300 rounded shadow-lg">
                    {filteredBots.length > 0 ? (
                        filteredBots.map(bot => (
                            <li
                                key={bot.id}
                                onClick={() => handleSelectBot(bot)}
                                className="px-4 py-2 cursor-pointer hover:bg-blue-100"
                            >
                                {bot.address}
                            </li>
                        ))
                    ) : (
                        <li className="px-4 py-2 text-gray-500">No bots found</li>
                    )}
                </ul>
            )}
        </div>
    );
}

export default BotSelector;
