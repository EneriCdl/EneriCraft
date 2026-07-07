<h1 align="center">
  <img src="https://img.shields.io/badge/version-v0.9.0-brightgreen" alt="v0.9.0">
  <img src="https://img.shields.io/badge/license-GPL--3.0-blue" alt="GPL-3.0">
  <img src="https://img.shields.io/badge/status-dev-yellow" alt="开发中">
</h1>

<p align="center">
  <img src="https://raw.githubusercontent.com/EneriCdl/EneriCraft/main/docs/logo.svg" width="120" onerror="this.style.display='none'">
</p>

# EneriCraft

> 🎮 让 Minecraft 联机像发微信一样简单 —— 连接码交换，P2P + 中继混合连接，全国可通。

**永久免费 · 全开源 · 社区驱动 · 无充值**

---

## 💡 这是什么

EneriCraft 是一个**零门槛 MC 联机工具**。不需要懂端口转发、NAT 穿透、内网映射——打开 MC，打开 EneriCraft，把连接码发给朋友，完事。

```
你:  MC → 打开存档 → EneriCraft 创建房间 → 复制连接码 → 发微信
👥
朋友: 粘贴连接码 → EneriCraft 加入 → MC 连 127.0.0.1:25566 → 进入你的世界
```

---

## 🏗️ 工作原理

```
┌─────────────┐          ┌─────────────────┐          ┌─────────────┐
│  你的 MC     │  TCP     │   EneriCraft     │  QUIC    │   EneriCraft │  TCP    │ 朋友的 MC  │
│ (局域网开放)  │◄────────│    主机端        │◄────────►│   客户端     │◄────────│ 连25566    │
│  :56580     │          │  Bridge→LAN      │  P2P优先  │  TCP→Bridge │          │            │
└─────────────┘          └───────┬──────────┘          └─────────────┘
                                 │ P2P 失败时
                                 ▼
                      ┌──────────────────┐
                      │   中继服务器      │
                      │ 120.77.255.112   │
                      │   (纯转发)        │
                      └──────────────────┘
```

**P2P 优先，中继兜底。** 同一 WiFi 或一方有公网 IP → P2P 直连。都在大内网 → 自动切中继。

---

## 🚀 快速开始

### 房主

1. 打开 MC → 进入存档 → **Esc → 对局域网开放**
2. 双击 `EneriCraft.exe` → 选择「分享我的世界」→ 创建房间
3. 复制连接码（`EC1-` 开头）→ 发微信 / QQ 给朋友

### 朋友

1. 双击 `EneriCraft.exe`
2. 粘贴连接码 → 点击「加入」
3. 打开 MC → 多人游戏 → 直接连接 `127.0.0.1:25566`
4. 进入房主的世界 🎉

> **注意：** 双方需要安装相同版本的 Minecraft。建议使用 HMCL 启动器的离线模式以避免 Mojang 验证问题。

---

## 🛠️ 技术栈

| 层 | 技术 |
|---|---|
| 后端 | Go 1.25 + quic-go v0.60 |
| 前端 | React 18 + TypeScript + TailwindCSS |
| 隧道 | QUIC + STUN (RFC 5389) + TCP Bridge |
| 中继 | 自建 QUIC Relay（8.6MB，社区可自托管） |
| 打包 | 单 exe 文件（13MB，前端内嵌） |

---

## 📊 开发进度

### 已完成

- [x] MC 版本自动检测（PCL2/HMCL 进程 + 文件扫描）
- [x] LAN 端口自动检测（UDP 组播 + 进程枚举）
- [x] 连接码生成/解析（EC1- + EP1-，zlib + Base64）
- [x] QUIC P2P 隧道 + TCP ↔ QUIC 双向桥接
- [x] 社区中继服务器 `120.77.255.112:9000`
- [x] 中继配对 + 数据流转发
- [x] 在线玩家实时列表
- [x] 本地 Yggdrasil 认证服务器（绕过 Mojang 验证）
- [x] Windows 防火墙自动放行
- [x] UPnP 端口自动映射
- [x] Paper 独立开服模式（可选）
- [x] 房间退出自动清理

### 待完成

- [ ] MC "无效的会话" 兼容性完善
- [ ] MOD 列表检测与同步
- [ ] 多中继节点负载均衡
- [ ] 中文 / English 双语

---

## 🧑‍💻 开发

```bash
git clone https://github.com/EneriCdl/EneriCraft.git
cd EneriCraft

# 前端
cd frontend
npm install
npx vite build

# 后端（编译为 13MB 单文件）
cd ..
go build -o EneriCraft.exe .

# 中继服务器（Linux）
GOOS=linux GOARCH=amd64 go build -o relay-server-linux ./cmd/relay/
```

---

## 🖥️ 部署中继

任何人都可以部署自己的中继节点，帮助社区扩大覆盖：

```bash
# 在任意有公网 IP 的服务器上
./relay-server 9000

# 在 EneriCraft 设置中填写: <你的IP>:9000
```

---

## 📄 许可

[GNU General Public License v3.0](LICENSE)

本项目与 Mojang AB / Microsoft 无任何关联。
