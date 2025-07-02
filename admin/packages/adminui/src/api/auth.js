export async function login(username, password) {
    try {
        const response = await fetch("/user/login", {
            method: "POST",
            headers: {
                "Content-Type": "application/json"
            },
            credentials: "include", // 关键！带上 cookie
            body: JSON.stringify({ username, password })
        });

        const result = await response.json();

        if (response.ok && result.code === 0) {
            // 登录成功，可根据需要保存其他信息
            return result.data; // 比如返回用户信息或登录成功标识
        } else {
            throw new Error(result.message || "登录失败");
        }
    } catch (err) {
        throw new Error(err.message || "网络错误");
    }
}
