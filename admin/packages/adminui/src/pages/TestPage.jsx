import React, {useState} from "react";

export default function TestPage() {
    const [show, setShow] = useState(false);
    return (
        <div className="p-6">
            <div className="bg-red-500 text-white p-4 text-xl">
                Tailwind 是否加载成功？
            </div>
        </div>
    );
}
