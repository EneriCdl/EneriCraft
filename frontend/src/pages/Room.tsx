import { useState, useEffect } from "react";
import { HiClipboard, HiCheck, HiUserGroup, HiServer } from "react-icons/hi2";
import toast from "react-hot-toast";
import * as api from "../api/client";

interface Player {
  name: string;
  online: boolean;
}

export default function Room() {
  const [copied, setCopied] = useState(false);
  const [connectCode, setConnectCode] = useState("");
  const [connected, setConnected] = useState(false);
  const [natType, setNatType] = useState("");
  const [loading, setLoading] = useState(true);
  const [players, setPlayers] = useState<Player[]>([]);

  useEffect(() => {
    let cancelled = false;
    const poll = async () => {
      try {
        const status = await api.getRoomStatus();
        if (cancelled) return;
        setConnected(status.connected);
        setConnectCode(status.connect_code || "");
        setNatType(status.nat_type || "");
        // 玩家列表
        const plist = (status as Record<string, unknown>).players as Player[] | undefined;
        if (plist) setPlayers(plist);
      } catch { /* no room */ }
      finally { if (!cancelled) setLoading(false); }
    };
    poll();
    const timer = setInterval(poll, 3000);
    return () => { cancelled = true; clearInterval(timer); };
  }, []);

  const handleCopy = async () => {
    if (!connectCode) return;
    try {
      await navigator.clipboard.writeText(connectCode);
      setCopied(true);
      toast.success("连接码已复制！");
      setTimeout(() => setCopied(false), 2000);
    } catch { toast.error("复制失败"); }
  };

  const handleLeave = async () => {
    try {
      await api.leaveRoom();
      toast.success("已离开房间");
      setConnected(false); setConnectCode(""); setPlayers([]);
    } catch (err) { toast.error("离开失败: " + (err as Error).message); }
  };

  if (loading) {
    return <div className="max-w-3xl mx-auto pt-20 text-center"><div className="animate-pulse text-gray-400">加载中...</div></div>;
  }

  if (!connected) {
    return (
      <div className="max-w-3xl mx-auto pt-20">
        <div className="card text-center py-16">
          <HiUserGroup className="w-16 h-16 mx-auto text-gray-300 dark:text-gray-700 mb-4" />
          <p className="text-gray-500 dark:text-gray-400 mb-4">暂未加入任何房间</p>
          <a href="/" className="btn-primary bg-emerald-500 inline-block">去创建或加入房间</a>
        </div>
      </div>
    );
  }

  return (
    <div className="max-w-3xl mx-auto">
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-2xl font-bold dark:text-white">房间管理</h1>
        <button onClick={handleLeave} className="text-sm text-red-500 hover:text-red-600 font-medium">离开房间</button>
      </div>

      <div className="space-y-5">
        {connectCode && (
          <div className="card text-center">
            <p className="text-xs text-gray-500 dark:text-gray-400 uppercase tracking-wider mb-3">连接码</p>
            <div className="bg-gray-50 dark:bg-gray-800 rounded-xl p-4 mb-4 font-mono text-xs text-gray-700 dark:text-gray-300 break-all select-all text-left">{connectCode}</div>
            <button onClick={handleCopy} className={`flex items-center gap-2 px-5 py-2.5 rounded-lg font-medium text-sm transition-all mx-auto ${copied ? "bg-emerald-100 dark:bg-emerald-900 text-emerald-700" : "bg-gray-100 dark:bg-gray-800 text-gray-700 dark:text-gray-300 hover:bg-gray-200"}`}>
              {copied ? <><HiCheck className="w-4 h-4" /> 已复制</> : <><HiClipboard className="w-4 h-4" /> 复制连接码</>}
            </button>
            <p className="text-xs text-gray-400 dark:text-gray-500 mt-4 font-mono">P2P + 中继 · {natType || "运行中"}</p>
          </div>
        )}

        {/* 在线玩家 */}
        <div className="card">
          <h3 className="font-bold mb-4 flex items-center gap-2 dark:text-white"><HiUserGroup className="w-5 h-5 text-emerald-500" />在线玩家 ({players.length})</h3>
          <div className="space-y-2">
            {players.length === 0 && <p className="text-sm text-gray-400">等待玩家加入...</p>}
            {players.map((p, i) => (
              <div key={i} className="flex items-center gap-3 py-2.5 px-3 rounded-lg hover:bg-gray-50 dark:hover:bg-gray-800 transition-colors">
                <div className={`w-2.5 h-2.5 rounded-full ${p.online ? "bg-emerald-500" : "bg-gray-300 dark:bg-gray-600"}`} />
                <span className="font-medium dark:text-white">{p.name}</span>
                <span className={`text-xs px-2 py-0.5 rounded-full font-medium ${p.online ? "bg-emerald-100 dark:bg-emerald-900 text-emerald-700" : "bg-gray-100 dark:bg-gray-800 text-gray-500"}`}>
                  {p.online ? "在线" : "等待中"}
                </span>
              </div>
            ))}
          </div>
        </div>

        {/* 服务器状态 */}
        <div className="card">
          <h3 className="font-bold mb-4 flex items-center gap-2 dark:text-white"><HiServer className="w-5 h-5 text-emerald-500" />连接信息</h3>
          <div className="text-sm text-gray-500 dark:text-gray-400 space-y-1">
            <p>连接类型: P2P + 中继</p>
            <p>中继地址: 120.77.255.112:9000</p>
            <p>朋友连接地址: 127.0.0.1:25566</p>
          </div>
        </div>
      </div>
    </div>
  );
}
