import React from "react";

export default function Dashboard() {
    return (
        <div className="max-w-3xl mx-auto p-8 bg-white rounded-lg shadow-md mt-10">
            <h1 className="text-4xl font-extrabold mb-6 text-indigo-600 flex items-center gap-3">
                🤖 Telegram DeepSeek Bot
            </h1>

            <p className="text-gray-700 mb-6 leading-relaxed">
                This repository provides a <strong>Telegram bot</strong> built with <strong>Golang</strong> that integrates with <strong>LLM API</strong> to provide AI-powered responses.
                The bot supports <strong>OpenAI</strong>, <strong>DeepSeek</strong>, <strong>Gemini</strong>, <strong>OpenRouter</strong> LLMs, making interactions feel more natural and dynamic.
                <br />
                <a
                    href="https://github.com/yincongcyincong/telegram-deepseek-bot"
                    target="_blank"
                    rel="noopener noreferrer"
                    className="text-indigo-600 font-semibold underline hover:text-indigo-800"
                >
                    GitHub
                </a>
                <br />
                <span className="text-indigo-600 font-semibold cursor-pointer underline mt-2 inline-block">
                <a
                    href="https://github.com/yincongcyincong/telegram-deepseek-bot/blob/main/README_ZH.md"
                    target="_blank"
                    rel="noopener noreferrer"
                    className="text-indigo-600 font-semibold underline hover:text-indigo-800"
                >
                 中文文档
                </a>
        </span>
            </p>

            <h2 className="text-2xl font-semibold mb-4">🚀 Features</h2>
            <ul className="list-disc list-inside space-y-3 text-gray-700">
                <li>
                    <span className="mr-2">🤖</span>
                    <strong>AI Responses:</strong> Uses DeepSeek API for chatbot replies.
                </li>
                <li>
                    <span className="mr-2">⏳</span>
                    <strong>Streaming Output:</strong> Sends responses in real-time to improve user experience.
                </li>
                <li>
                    <span className="mr-2">🏗</span>
                    <strong>Easy Deployment:</strong> Run locally or deploy to a cloud server.
                </li>
                <li>
                    <span className="mr-2">👀</span>
                    <strong>Identify Image:</strong> Use images to communicate with DeepSeek.
                </li>
                <li>
                    <span className="mr-2">🎺</span>
                    <strong>Support Voice:</strong> Use voice to communicate with DeepSeek.
                </li>
                <li>
                    <span className="mr-2">🐂</span>
                    <strong>Function Call:</strong> Transform MCP protocol to function call.
                </li>
                <li>
                    <span className="mr-2">🌊</span>
                    <strong>RAG:</strong> Support RAG to fill contextc.
                </li>
                <li>
                    <span className="mr-2">⛰️</span>
                    <strong>OpenRouter:</strong> Support OpenRouter with more than 400 LLMs.
                </li>
            </ul>
        </div>
    );
}
