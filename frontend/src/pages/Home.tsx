import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { HiPlus, HiArrowRightOnRectangle, HiClipboard, HiCheck } from "react-icons/hi2";
import toast from "react-hot-toast";
import * as api from "../api/client";
import BorderGlow from "../components/borderglow/BorderGlow";

const GAME_MODES = ["生存", "创造", "冒险", "小游戏"];

export default function Home() {
  const navigate = useNavigate();
  const [roomCode, setRoomCode] = useState("");
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [createMode, setCreateMode] = useState<"lan" | "paper">("lan");
  const [selectedMode, setSelectedMode] = useState("生存");
  const [roomName, setRoomName] = useState("");
  const [creating, setCreating] = useState(false);
  const [createStep, setCreateStep] = useState("");
  const [joining, setJoining] = useState(false);
  const [createdCode, setCreatedCode] = useState("");
  const [createdVersion, setCreatedVersion] = useState("");
  const [copied, setCopied] = useState(false);
  const [detectedVersion, setDetectedVersion] = useState("");
  const [fromProcess, setFromProcess] = useState(false);
  const [needOpenLAN, setNeedOpenLAN] = useState(false);
  const [punchCode, setPunchCode] = useState("");
  const [punchInstruction, setPunchInstruction] = useState("");
  const [hostPunchInput, setHostPunchInput] = useState("");
  const [punching, setPunching] = useState(false);

  const openCreateModal = async () => {
    setShowCreateModal(true);
    try {
      const v = await api.detectVersions();
      setDetectedVersion(v.latest || "未检测到");
      setFromProcess(v.from_process || false);
    } catch {
      setDetectedVersion("检测失败");
      setFromProcess(false);
    }
  };

  const handleCreateRoom = async () => {
    setCreating(true);
    setNeedOpenLAN(false);
    setCreateStep("检测 MC 版本...");
    let shouldStop = false;
    try {
      const result = await api.createRoom({
        game_mode: selectedMode.toLowerCase(),
        room_name: roomName || undefined,
        mc_version: detectedVersion,
        mode: createMode,
      });
      if (result.need_open_lan) {
        setNeedOpenLAN(true);
        shouldStop = true;
      } else {
        setCreatedCode(result.connect_code || "");
        setCreatedVersion(result.mc_version || "");
        setCreateStep("");
        toast.success("房间创建成功！");
      }
    } catch (err) {
      toast.error("创建失败: " + (err as Error).message);
    } finally {
      if (!shouldStop) setCreating(false);
    }
  };

  const handleJoinRoom = async () => {
    if (!roomCode || roomCode.length < 4) { toast.error("请输入有效的连接码"); return; }
    setJoining(true); setPunchCode("");
    try {
      const result = await api.joinRoom(roomCode);
      if (result.punch_required) {
        setPunchCode(result.punch_code || "");
        setPunchInstruction(result.instruction || "请把回执码发给房主");
        toast.error("P2P 直连失败，需要房主配合打洞");
      } else {
        toast.success("加入成功！"); navigate("/room");
      }
    } catch (err) {
      toast.error("加入失败: " + (err as Error).message);
    } finally { setJoining(false); }
  };

  const handleHostPunch = async () => {
    if (!hostPunchInput) return;
    setPunching(true);
    try {
      const res = await fetch("/api/room/punch", { method: "POST", headers: { "Content-Type": "application/json" }, body: JSON.stringify({ punch_code: hostPunchInput }) });
      const data = await res.json();
      if (res.ok) { toast.success("打洞中，请等待..."); setHostPunchInput(""); }
      else { toast.error((data as { error: string }).error || "失败"); }
    } catch { toast.error("打洞请求失败"); }
    finally { setPunching(false); }
  };

  const handleCopyCode = async () => {
    try {
      await navigator.clipboard.writeText(createdCode);
      setCopied(true);
      toast.success("连接码已复制！");
      setTimeout(() => setCopied(false), 2000);
    } catch {
      toast.error("复制失败");
    }
  };

  return (
      <div className="relative max-w-2xl mx-auto pt-16">
        {/* 标题 */}
        <div className="text-center mb-10">
        <h1 className="text-5xl font-bold mb-3 tracking-tight">
          <span className="bg-gradient-to-r from-emerald-400 to-cyan-500 bg-clip-text text-transparent">
            Eneri
          </span>
          <span className="text-gray-800 dark:text-white">Craft</span>
        </h1>
        <p className="text-gray-500 dark:text-gray-400 text-sm">
          零服务器 · P2P 直连 · 连接码交换 · 永久免费
        </p>
      </div>

      {/* 主操作区 */}
      <div className="grid grid-cols-2 gap-5">
        {/* 创建房间 */}
        <BorderGlow
          glowColor="160 80 60"
          backgroundColor="rgb(255 255 255 / 0.03)"
          borderRadius={24}
          glowRadius={30}
          glowIntensity={1.2}
          coneSpread={20}
          fillOpacity={0.3}
        >
          <button
            onClick={openCreateModal}
            className="flex flex-col items-center gap-4 py-14 cursor-pointer w-full
                       bg-transparent border-0 outline-none"
          >
            <div
              className="w-20 h-20 rounded-2xl bg-emerald-50 dark:bg-emerald-950
                            flex items-center justify-center
                            group-hover:bg-emerald-100 dark:group-hover:bg-emerald-900
                            transition-colors"
            >
              <HiPlus className="w-10 h-10 text-emerald-500" />
            </div>
            <div className="text-center">
              <h3 className="text-xl font-bold mb-1 dark:text-white">创建房间</h3>
              <p className="text-sm text-gray-500 dark:text-gray-400">
                一键生成连接码，发给朋友即可加入
              </p>
            </div>
          </button>
        </BorderGlow>

        {/* 加入房间 */}
        <BorderGlow
          colors={["#06b6d4", "#10b981", "#6366f1"]}
          glowColor="190 80 60"
          backgroundColor="rgb(255 255 255 / 0.03)"
          borderRadius={24}
          glowRadius={30}
          glowIntensity={1.2}
          coneSpread={20}
          fillOpacity={0.3}
        >
          <div className="flex flex-col items-center gap-4 py-14 w-full">
            <div
              className="w-20 h-20 rounded-2xl bg-cyan-50 dark:bg-cyan-950
                            flex items-center justify-center"
            >
              <HiArrowRightOnRectangle className="w-10 h-10 text-cyan-500" />
            </div>
            <div className="text-center w-full px-4">
              <h3 className="text-xl font-bold mb-3 dark:text-white">加入房间</h3>
              <input
                type="text"
                value={roomCode}
                onChange={(e) => setRoomCode(e.target.value)}
                placeholder="粘贴连接码"
                maxLength={500}
                className="input text-center text-sm tracking-wide font-mono"
              />
              <button
                onClick={handleJoinRoom}
                disabled={!roomCode || joining}
                className="btn-primary w-full mt-3"
              >
                {joining ? "连接中..." : "加入"}
              </button>
            </div>
          </div>
        </BorderGlow>
      </div>

      {/* 回执码提示（客户端连不上时显示） */}
      {punchCode && (
        <div className="mt-4 p-4 rounded-xl bg-amber-50 dark:bg-amber-950 border border-amber-200 dark:border-amber-800">
          <h4 className="font-bold text-amber-700 dark:text-amber-300 mb-2">P2P 直连失败，需要房主帮助穿透</h4>
          <p className="text-sm text-amber-600 dark:text-amber-400 mb-2">{punchInstruction}</p>
          <div className="bg-white dark:bg-gray-900 rounded-lg p-3 font-mono text-xs break-all select-all mb-2">{punchCode}</div>
          <button onClick={() => { navigator.clipboard.writeText(punchCode); toast.success("回执码已复制"); }} className="px-3 py-1.5 bg-amber-500 text-white rounded-lg text-sm font-medium">复制回执码</button>
        </div>
      )}

      {/* 房主打洞输入 */}
      <div className="mt-4 p-4 rounded-xl bg-gray-50 dark:bg-gray-800/50">
        <h4 className="font-bold text-sm text-gray-600 dark:text-gray-400 mb-2">收到朋友的「回执码」？粘贴到这里</h4>
        <div className="flex gap-2">
          <input type="text" value={hostPunchInput} onChange={(e) => setHostPunchInput(e.target.value)} placeholder="EP1-..." className="input flex-1 text-sm font-mono" />
          <button onClick={handleHostPunch} disabled={!hostPunchInput || punching} className="btn-primary text-sm whitespace-nowrap">{punching ? "打洞中..." : "打洞连接"}</button>
        </div>
      </div>

      {/* 创建房间弹窗 */}
      {showCreateModal && (
        <div
          className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 animate-fade-in"
          onClick={() => setShowCreateModal(false)}
        >
          <div
            className="card max-w-md w-full mx-4 animate-slide-up shadow-xl"
            onClick={(e) => e.stopPropagation()}
          >
            <h2 className="text-2xl font-bold mb-6 dark:text-white">创建房间</h2>

            {/* 创建模式选择 */}
            <div className="mb-5">
              <label className="block text-sm font-medium text-gray-600 dark:text-gray-400 mb-3">
                房间类型
              </label>
              <div className="grid grid-cols-2 gap-2">
                <button
                  onClick={() => setCreateMode("lan")}
                  className={`py-3 px-4 text-sm rounded-lg font-medium transition-all duration-200 text-left
                    ${createMode === "lan"
                      ? "bg-emerald-500 text-white shadow-md shadow-emerald-500/25"
                      : "bg-gray-100 dark:bg-gray-800 text-gray-600 dark:text-gray-400 hover:bg-gray-200 dark:hover:bg-gray-700"
                    }`}
                >
                  <div className="font-bold">分享我的世界</div>
                  <div className="text-xs opacity-75">使用你的存档，MOD兼容</div>
                </button>
                <button
                  onClick={() => setCreateMode("paper")}
                  className={`py-3 px-4 text-sm rounded-lg font-medium transition-all duration-200 text-left
                    ${createMode === "paper"
                      ? "bg-emerald-500 text-white shadow-md shadow-emerald-500/25"
                      : "bg-gray-100 dark:bg-gray-800 text-gray-600 dark:text-gray-400 hover:bg-gray-200 dark:hover:bg-gray-700"
                    }`}
                >
                  <div className="font-bold">创建新世界</div>
                  <div className="text-xs opacity-75">独立服务端，新存档</div>
                </button>
              </div>
            </div>

            {/* LAN 模式提示 */}
            {createMode === "lan" && !needOpenLAN && (
              <div className="mb-5 p-3 rounded-lg bg-blue-50 dark:bg-blue-950 text-sm">
                <span className="text-blue-700 dark:text-blue-300">
                  请在 MC 中按 Esc → 对局域网开放
                </span>
              </div>
            )}

            {/* 未检测到 LAN 提示 */}
            {needOpenLAN && (
              <div className="mb-5 p-4 rounded-lg bg-amber-50 dark:bg-amber-950">
                <h4 className="font-bold text-amber-700 dark:text-amber-300 mb-2">
                  请先在 MC 中对局域网开放
                </h4>
                <ol className="text-sm text-amber-600 dark:text-amber-400 list-decimal ml-4 space-y-1">
                  <li>切换到 Minecraft 窗口</li>
                  <li>按 Esc 打开菜单</li>
                  <li>点击「对局域网开放」</li>
                  <li>回到这里，重新点击创建房间</li>
                </ol>
                <button
                  onClick={() => { setNeedOpenLAN(false); handleCreateRoom(); }}
                  className="mt-3 px-4 py-2 bg-amber-500 text-white rounded-lg text-sm font-medium hover:bg-amber-600"
                >
                  我已开放，重新检测
                </button>
              </div>
            )}

            {/* 游戏模式 */}
            <div className="mb-5">
              <label className="block text-sm font-medium text-gray-600 dark:text-gray-400 mb-3">
                游戏模式
              </label>
              <div className="grid grid-cols-4 gap-2">
                {GAME_MODES.map((mode) => (
                  <button
                    key={mode}
                    onClick={() => setSelectedMode(mode)}
                    className={`py-2.5 text-sm rounded-lg font-medium transition-all duration-200
                      ${selectedMode === mode
                        ? "bg-emerald-500 text-white shadow-md shadow-emerald-500/25"
                        : "bg-gray-100 dark:bg-gray-800 text-gray-600 dark:text-gray-400 hover:bg-gray-200 dark:hover:bg-gray-700"
                      }`}
                  >
                    {mode}
                  </button>
                ))}
              </div>
            </div>

            {/* 房间名 */}
            <div className="mb-6">
              <label className="block text-sm font-medium text-gray-600 dark:text-gray-400 mb-2">
                房间名称
              </label>
              <input
                type="text"
                value={roomName}
                onChange={(e) => setRoomName(e.target.value)}
                placeholder="给房间起个名字（可选）"
                className="input"
              />
            </div>

            {/* 版本检测 */}
            {fromProcess ? (
              <div className="mb-5 p-3 rounded-lg bg-emerald-50 dark:bg-emerald-950 text-sm">
                <span className="text-emerald-700 dark:text-emerald-300 font-medium">
                  🟢 正在运行 MC {detectedVersion}
                </span>
              </div>
            ) : (
              <div className="mb-5 p-3 rounded-lg bg-amber-50 dark:bg-amber-950 text-sm">
                <span className="text-amber-700 dark:text-amber-300">
                  ⚠ 未检测到运行中的 Minecraft
                </span>
                <span className="text-gray-500 dark:text-gray-400 ml-2 text-xs">
                  {detectedVersion !== "未检测到" ? `将使用已安装版本 ${detectedVersion}` : "请先打开游戏再创建房间"}
                </span>
              </div>
            )}

            <div className="mb-6 p-3 rounded-lg bg-gray-50 dark:bg-gray-800 text-sm text-gray-500 dark:text-gray-400">
              将创建 <strong className="text-gray-700 dark:text-gray-300">{selectedMode}</strong> 模式房间
              {fromProcess ? <> · 运行版本 <strong className="text-gray-700 dark:text-gray-300">{detectedVersion}</strong></> : <> · 使用 <strong className="text-gray-700 dark:text-gray-300">{detectedVersion}</strong></>}
              {roomName && <> — "{roomName}"</>}
            </div>

            {creating && (
              <div className="mb-4 p-3 rounded-lg bg-cyan-50 dark:bg-cyan-950 text-sm text-cyan-700 dark:text-cyan-300 animate-pulse">
                {createStep || "创建中..."}
              </div>
            )}

            <button
              onClick={handleCreateRoom}
              disabled={creating}
              className="btn-primary w-full bg-emerald-500 hover:bg-emerald-600 shadow-emerald-500/25"
            >
              {creating ? "创建中..." : "创建房间"}
            </button>
          </div>
        </div>
      )}

      {/* 连接码结果弹窗 */}
      {createdCode && (
        <div
          className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 animate-fade-in"
          onClick={() => setCreatedCode("")}
        >
          <div
            className="card max-w-lg w-full mx-4 animate-slide-up shadow-xl text-center"
            onClick={(e) => e.stopPropagation()}
          >
            <div className="w-16 h-16 rounded-full bg-emerald-100 dark:bg-emerald-900
                            flex items-center justify-center mx-auto mb-4">
              <HiCheck className="w-8 h-8 text-emerald-500" />
            </div>
            <h2 className="text-xl font-bold mb-2 dark:text-white">房间已创建！</h2>
            <p className="text-xs text-emerald-600 dark:text-emerald-400 mb-2">
              MC {createdVersion} · {selectedMode} 模式 · 服务器已启动
            </p>
            <p className="text-sm text-gray-500 dark:text-gray-400 mb-4">
              把连接码发给朋友，对方打开 EneriCraft 粘贴即可加入
            </p>
            <div className="bg-gray-50 dark:bg-gray-800 rounded-xl p-4 mb-4 font-mono text-xs
                            text-gray-700 dark:text-gray-300 break-all select-all text-left">
              {createdCode}
            </div>
            <div className="flex gap-3 justify-center">
              <button
                onClick={handleCopyCode}
                className={`flex items-center gap-2 px-5 py-2.5 rounded-lg font-medium text-sm transition-all
                  ${copied
                    ? "bg-emerald-100 dark:bg-emerald-900 text-emerald-700"
                    : "bg-emerald-500 text-white hover:bg-emerald-600"
                  }`}
              >
                {copied ? <><HiCheck className="w-4 h-4" /> 已复制</>
                        : <><HiClipboard className="w-4 h-4" /> 复制连接码</>}
              </button>
              <button
                onClick={() => navigate("/room")}
                className="btn-secondary"
              >
                进入房间
              </button>
            </div>
          </div>
        </div>
      )}

      {/* 底部提示 */}
      <p className="text-center text-xs text-gray-400 dark:text-gray-500 mt-10">
        连接码通过 QQ / 微信 发给朋友 · P2P 直连 · 无需服务器
      </p>
    </div>
  );
}
