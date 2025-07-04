import React, {useState} from "react";
import Modal from "../components/Modal";

export default function TestPage() {
    const [show, setShow] = useState(false);
    return (
        <div className="p-6">
            <button onClick={() => setShow(true)} className="px-4 py-2 bg-blue-600 text-white rounded">
                Open Modal
            </button>
            <Modal visible={show} title="Test Modal" onClose={() => setShow(false)}>
                <p>This is a test modal content.</p>
            </Modal>
        </div>
    );
}
