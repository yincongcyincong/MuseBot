import React, { useEffect, useRef, useState } from "react";
import { useTranslation } from "react-i18next";

// 定义用于 localStorage 存储 botId 的键
const CACHE_BOT_ID_KEY = "lastSelectedBotId";

function BotSelector({ value, onChange }) {
    const [bots, setBots] = useState([]);
    const [filteredBots, setFilteredBots] = useState([]);
    const [searchText, setSearchText] = useState("");
    const [selectedBot, setSelectedBot] = useState(null);
    const [dropdownOpen, setDropdownOpen] = useState(false);
    const wrapperRef = useRef(null);

    const { t } = useTranslation();

    useEffect(() => {
        fetchBots();
    }, []);

    useEffect(() => {
        const lower = searchText.toLowerCase();
        if (lower === "") {
            setFilteredBots(bots); // 空搜索显示全部
        } else {
            setFilteredBots(
                bots.filter(
                    (bot) =>
                        bot.name.toLowerCase().includes(lower) ||
                        bot.address.toLowerCase().includes(lower)
                )
            );
        }
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
                const fetchedBots = data.data;
                setBots(fetchedBots);
                setFilteredBots(fetchedBots);

                // --- ⭐️ 核心修改逻辑：优先 value，value 为空时才走缓存 ---
                let defaultBot = null;

                // 1. 优先使用传入的 value prop (如果 value 是 bot object 且存在于列表中)
                if (value && value.id) {
                    const botFromValue = fetchedBots.find(bot => bot.id === value.id);
                    if (botFromValue) {
                        defaultBot = botFromValue;
                    }
                }

                // 2. 只有当 value 为空或无效时，才检查缓存
                if (!defaultBot) {
                    const cachedBotId = localStorage.getItem(CACHE_BOT_ID_KEY);
                    if (cachedBotId) {
                        const botFromCache = fetchedBots.find(bot => String(bot.id) === cachedBotId);
                        if (botFromCache) {
                            defaultBot = botFromCache; // 缓存命中
                        } else {
                            // 缓存失效或 bot 已下线，清理缓存
                            localStorage.removeItem(CACHE_BOT_ID_KEY);
                        }
                    }
                }

                // 3. 最后的 fallback：列表中的第一个 bot
                if (!defaultBot) {
                    defaultBot = fetchedBots[0];
                }
                // --- ⭐️ 结束核心修改逻辑 ---

                setSelectedBot(defaultBot);
                setSearchText("");
                onChange(defaultBot); // 触发父组件的 onChange
            } else {
                // 如果没有获取到 bot 列表，清除缓存
                localStorage.removeItem(CACHE_BOT_ID_KEY);
            }
        } catch (err) {
            console.error("Failed to fetch bots:", err);
            // 失败时也清除缓存
            localStorage.removeItem(CACHE_BOT_ID_KEY);
        }
    };

    const handleSelectBot = (bot) => {
        // --- 存储 botId 到缓存 (保持) ---
        localStorage.setItem(CACHE_BOT_ID_KEY, String(bot.id));
        // --- 结束存储 ---

        setSelectedBot(bot);
        setSearchText("");
        setDropdownOpen(false);
        onChange(bot); // 触发父组件的 onChange
    };

    return (
        <div className="relative" ref={wrapperRef}>
            <label className="block font-medium text-gray-700 mb-1">{t("bot_choose")}:</label>
            <input
                type="text"
                // 使用 selectedBot 来显示当前值
                value={searchText || (selectedBot?.name || selectedBot?.address || "")}
                onChange={(e) => {
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
                        filteredBots.map((bot) => (
                            <li
                                key={bot.id}
                                onClick={() => handleSelectBot(bot)}
                                className="px-4 py-2 cursor-pointer hover:bg-blue-100"
                            >
                                {bot.name || bot.address}
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