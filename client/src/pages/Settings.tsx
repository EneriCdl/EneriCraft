import { useState } from "react";

export default function Settings() {
  const [darkMode, setDarkMode] = useState(true);
  const [nickname, setNickname] = useState("");

  return (
    <div className="max-w-2xl mx-auto">
      <h1 className="text-2xl font-bold mb-6">设置</h1>

      <div className="space-y-6">
        {/* 用户信息 */}
        <div className="card">
          <h3 className="font-bold mb-4">用户信息</h3>
          <div>
            <label className="block text-sm font-medium text-gray-600 dark:text-gray-400 mb-2">
              昵称
            </label>
            <input
              type="text"
              value={nickname}
              onChange={(e) => setNickname(e.target.value)}
              placeholder="输入你的昵称"
              className="input"
            />
          </div>
        </div>

        {/* 外观设置 */}
        <div className="card">
          <h3 className="font-bold mb-4">外观</h3>
          <div className="flex items-center justify-between">
            <span>深色模式</span>
            <button
              onClick={() => setDarkMode(!darkMode)}
              className={`w-12 h-6 rounded-full transition-colors ${
                darkMode ? "bg-primary-500" : "bg-gray-300"
              }`}
            >
              <div className={`w-5 h-5 bg-white rounded-full shadow transition-transform
                              ${darkMode ? "translate-x-6" : "translate-x-0.5"}`} />
            </button>
          </div>
        </div>

        {/* 中继设置 */}
        <div className="card">
          <h3 className="font-bold mb-4">社区中继（可选）</h3>
          <p className="text-sm text-gray-500 mb-3">
            当 P2P 直连失败时，使用社区中继服务器兜底。
            留空则不使用中继。
          </p>
          <input
            type="text"
            placeholder="中继服务器地址（如 relay.example.com:3478）"
            className="input"
          />
        </div>

        {/* 版本信息 */}
        <div className="card">
          <h3 className="font-bold mb-2">关于</h3>
          <p className="text-sm text-gray-500">
            MC 联机器 v0.1.0-alpha<br />
            GPL-3.0 开源协议 · 永久免费<br />
            与 Mojang/Microsoft 无关联
          </p>
        </div>
      </div>
    </div>
  );
}
