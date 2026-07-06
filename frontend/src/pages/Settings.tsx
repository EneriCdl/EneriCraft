import { useState, useEffect, useCallback } from "react";
import { useTheme } from "../App";
import * as api from "../api/client";
import toast from "react-hot-toast";

export default function Settings() {
  const { theme, toggle } = useTheme();
  const [nickname, setNickname] = useState("");
  const [relayServer, setRelayServer] = useState("");
  const [saved, setSaved] = useState(true);

  // 加载配置
  useEffect(() => {
    api.getConfig().then((cfg) => {
      setNickname((cfg.nickname as string) || "");
      setRelayServer((cfg.relay_server as string) || "");
    }).catch(() => {});
  }, []);

  // 自动保存（输入后 1 秒防抖）
  const save = useCallback(
    debounce(async (nick: string, relay: string) => {
      try {
        await api.saveConfig({ nickname: nick, relay_server: relay });
        setSaved(true);
        toast.success("已保存");
      } catch {
        toast.error("保存失败");
      }
    }, 800),
    []
  );

  const handleNicknameChange = (val: string) => {
    setNickname(val);
    setSaved(false);
    save(val, relayServer);
  };

  const handleRelayChange = (val: string) => {
    setRelayServer(val);
    setSaved(false);
    save(nickname, val);
  };

  return (
    <div className="max-w-2xl mx-auto">
      <h1 className="text-2xl font-bold mb-6 dark:text-white">设置</h1>
      <div className="space-y-5">
        {/* 用户信息 */}
        <div className="card">
          <h3 className="font-bold mb-4 dark:text-white">用户信息</h3>
          <div className="flex items-center gap-3">
            <input
              type="text"
              value={nickname}
              onChange={(e) => handleNicknameChange(e.target.value)}
              placeholder="输入你的昵称"
              className="input flex-1"
            />
            <span className={`text-xs min-w-12 ${saved ? "text-emerald-500" : "text-amber-500"}`}>
              {saved ? "✓ 已保存" : "保存中..."}
            </span>
          </div>
        </div>

        {/* 外观 */}
        <div className="card">
          <h3 className="font-bold mb-4 dark:text-white">外观</h3>
          <div className="flex items-center justify-between">
            <span className="dark:text-gray-300">
              {theme === "dark" ? "深色模式" : "浅色模式"}
            </span>
            <button
              onClick={toggle}
              className={`w-12 h-6 rounded-full transition-colors duration-200 ${
                theme === "dark" ? "bg-emerald-500" : "bg-gray-300"
              }`}
            >
              <div
                className={`w-5 h-5 bg-white rounded-full shadow transition-transform duration-200 ${
                  theme === "dark" ? "translate-x-6" : "translate-x-0.5"
                }`}
              />
            </button>
          </div>
        </div>

        {/* 社区中继 */}
        <div className="card">
          <h3 className="font-bold mb-4 dark:text-white">社区中继（可选）</h3>
          <p className="text-sm text-gray-500 dark:text-gray-400 mb-3">
            当 P2P 直连失败时，使用社区中继服务器兜底。留空则不使用。
          </p>
          <input
            type="text"
            value={relayServer}
            onChange={(e) => handleRelayChange(e.target.value)}
            placeholder="中继服务器地址"
            className="input"
          />
        </div>

        {/* 关于 */}
        <div className="card">
          <h3 className="font-bold mb-2 dark:text-white">关于</h3>
          <p className="text-sm text-gray-500 dark:text-gray-400">
            EneriCraft v0.9.0<br />
            GPL-3.0 · 永久免费 · 开源<br />
            与 Mojang/Microsoft 无关联
          </p>
        </div>
      </div>
    </div>
  );
}

// 防抖工具函数
// eslint-disable-next-line @typescript-eslint/no-explicit-any
function debounce(fn: (...args: any[]) => void, ms: number) {
  let timer: ReturnType<typeof setTimeout>;
  return (...args: any[]) => {
    clearTimeout(timer);
    timer = setTimeout(() => fn(...args), ms);
  };
}
