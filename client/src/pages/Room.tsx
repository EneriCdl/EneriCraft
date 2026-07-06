import { HiClipboard, HiQrCode, HiUserGroup, HiServer, HiCog } from "react-icons/hi2";
import toast from "react-hot-toast";

export default function Room() {
  // TODO: 从 Tauri 获取房间状态
  const mockConnected = false;
  const mockRoomCode = "ABC123";

  return (
    <div className="max-w-3xl mx-auto">
      <h1 className="text-2xl font-bold mb-6">房间管理</h1>

      {!mockConnected ? (
        <div className="card text-center py-16">
          <HiUserGroup className="w-16 h-16 mx-auto text-gray-300 dark:text-gray-700 mb-4" />
          <p className="text-gray-500 dark:text-gray-400">
            暂未加入房间，请先在首页创建或加入房间
          </p>
        </div>
      ) : (
        <div className="space-y-6">
          {/* 房间码展示 */}
          <div className="card text-center">
            <p className="text-sm text-gray-500 mb-3">房间码</p>
            <div className="room-code mb-4">{mockRoomCode}</div>
            <div className="flex justify-center gap-3">
              <button
                onClick={() => toast.success("已复制连接码")}
                className="btn-secondary flex items-center gap-2"
              >
                <HiClipboard className="w-4 h-4" /> 复制连接码
              </button>
              <button className="btn-secondary flex items-center gap-2">
                <HiQrCode className="w-4 h-4" /> 二维码
              </button>
            </div>
            <p className="text-xs text-gray-400 mt-4">
              连接类型：P2P 直连 · 延迟 12ms
            </p>
          </div>

          {/* 玩家列表 */}
          <div className="card">
            <h3 className="font-bold mb-3 flex items-center gap-2">
              <HiUserGroup className="w-5 h-5" /> 在线玩家 (2)
            </h3>
            <div className="space-y-2">
              {["Steve (房主)", "Alex"].map((name, i) => (
                <div key={i} className="flex items-center gap-3 py-2">
                  <div className="w-8 h-8 rounded-full bg-primary-100 dark:bg-primary-900
                                  flex items-center justify-center text-sm font-bold text-primary-600">
                    {name[0]}
                  </div>
                  <span>{name}</span>
                  {i === 0 && (
                    <span className="text-xs bg-yellow-100 text-yellow-700 px-2 py-0.5 rounded-full">
                      房主
                    </span>
                  )}
                </div>
              ))}
            </div>
          </div>

          {/* 服务器控制 */}
          <div className="card">
            <h3 className="font-bold mb-3 flex items-center gap-2">
              <HiServer className="w-5 h-5" /> 服务器状态
            </h3>
            <div className="grid grid-cols-3 gap-4 text-center">
              <div>
                <p className="text-2xl font-bold text-green-500">20.0</p>
                <p className="text-xs text-gray-500">TPS</p>
              </div>
              <div>
                <p className="text-2xl font-bold">45%</p>
                <p className="text-xs text-gray-500">内存</p>
              </div>
              <div>
                <p className="text-2xl font-bold">0:32</p>
                <p className="text-xs text-gray-500">运行时间</p>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
