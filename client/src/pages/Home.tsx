import { useState } from "react";
import { HiPlus, HiArrowRightOnRectangle, HiClipboard, HiQrCode } from "react-icons/hi2";
import toast from "react-hot-toast";

export default function Home() {
  const [roomCode, setRoomCode] = useState("");
  const [showCreateModal, setShowCreateModal] = useState(false);

  const handleCreateRoom = () => {
    // TODO: 调用 Tauri IPC create_room
    toast.success("房间创建成功！");
    setShowCreateModal(false);
  };

  const handleJoinRoom = () => {
    if (!roomCode || roomCode.length < 4) {
      toast.error("请输入有效的房间码");
      return;
    }
    // TODO: 调用 Tauri IPC join_room
    toast.success("正在加入房间...");
  };

  return (
    <div className="max-w-2xl mx-auto pt-20">
      {/* 标题 */}
      <div className="text-center mb-12 animate-fade-in">
        <h1 className="text-4xl font-bold mb-3">
          MC <span className="text-primary-500">联机器</span>
        </h1>
        <p className="text-gray-500 dark:text-gray-400">
          零服务器成本 · P2P 直连 · 一键联机
        </p>
      </div>

      {/* 主操作区 */}
      <div className="grid grid-cols-2 gap-6">
        {/* 创建房间 */}
        <button
          onClick={() => setShowCreateModal(true)}
          className="card hover:shadow-md transition-all duration-300 group
                     flex flex-col items-center gap-4 py-12 cursor-pointer
                     hover:border-primary-200 dark:hover:border-primary-800"
        >
          <div className="w-20 h-20 rounded-2xl bg-primary-50 dark:bg-primary-950
                          flex items-center justify-center
                          group-hover:bg-primary-100 dark:group-hover:bg-primary-900
                          transition-colors">
            <HiPlus className="w-10 h-10 text-primary-500" />
          </div>
          <div className="text-center">
            <h3 className="text-xl font-bold mb-1">创建房间</h3>
            <p className="text-sm text-gray-500 dark:text-gray-400">
              一键生成连接码，发给朋友即可加入
            </p>
          </div>
        </button>

        {/* 加入房间 */}
        <div className="card flex flex-col items-center gap-4 py-12">
          <div className="w-20 h-20 rounded-2xl bg-green-50 dark:bg-green-950
                          flex items-center justify-center">
            <HiArrowRightOnRectangle className="w-10 h-10 text-green-500" />
          </div>
          <div className="text-center w-full px-4">
            <h3 className="text-xl font-bold mb-3">加入房间</h3>
            <input
              type="text"
              value={roomCode}
              onChange={(e) => setRoomCode(e.target.value.toUpperCase())}
              placeholder="输入连接码或房间码"
              maxLength={32}
              className="input text-center text-lg tracking-widest"
            />
            <button
              onClick={handleJoinRoom}
              disabled={!roomCode}
              className="btn-primary w-full mt-3"
            >
              加入
            </button>
          </div>
        </div>
      </div>

      {/* 创建房间弹窗 */}
      {showCreateModal && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50"
             onClick={() => setShowCreateModal(false)}>
          <div className="card max-w-md w-full mx-4 animate-slide-up"
               onClick={e => e.stopPropagation()}>
            <h2 className="text-2xl font-bold mb-6">创建房间</h2>

            {/* 游戏模式选择 */}
            <div className="space-y-4 mb-6">
              <label className="block text-sm font-medium text-gray-600 dark:text-gray-400">
                游戏模式
              </label>
              <div className="grid grid-cols-2 gap-3">
                {["生存", "创造", "冒险", "小游戏"].map(mode => (
                  <button key={mode} className="btn-secondary py-3 text-sm">
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
                placeholder="给房间起个名字（可选）"
                className="input"
              />
            </div>

            <button onClick={handleCreateRoom} className="btn-primary w-full">
              创建
            </button>
          </div>
        </div>
      )}

      {/* 底部提示 */}
      <p className="text-center text-xs text-gray-400 mt-8">
        连接码通过 QQ/微信 发送给朋友即可联机 · 无需服务器 · 完全免费
      </p>
    </div>
  );
}
