/** @type {import('tailwindcss').Config} */
module.exports = {
    // 配置 Tailwind 扫描你的 HTML 和 JSX/TSX 文件
    content: [
        "./index.html", // Vite 项目的入口 HTML 文件
        "./src/**/*.{js,jsx,ts,tsx}", // 你的 React 组件文件
    ],
    theme: {
        extend: {},
    },
    plugins: [],
}