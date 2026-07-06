# Changelog

## v0.9.0 — 2026-07-07

- 部署社区中继服务器（阿里云深圳 120.77.255.112:9000）
- P2P 失败自动走中继（中继配对 + 双向数据转发）
- 修复中继清理死循环 bug（清理协程误删待配对条目）
- 修复中继配对超时过短（2min → 5min）
- 新增双端在线玩家实时列表（房间页每 3 秒刷新）
- 新增 Bridge 活跃检测（OnActivity 回调 + 玩家状态同步）
- 新增本地 Yggdrasil 认证服务器（绕过 Mojang 验证）
- 连接码解耦（去掉无效的 127.0.0.1 端点）
- P2P 打洞重试机制（双方每 2 秒互发，最多 120 次）
- 回执码机制（EP1- 前缀，P2P 打洞备用）
- Windows 防火墙自动放行 QUIC 端口
- UPnP 自动端口映射
- 版本号统一管理（main.go / package.json / Settings.tsx）

## v0.8.0 — 2026-07-06

- 架构重构：从 Paper 独立开服 → Open to LAN + P2P 隧道
- 新增「分享我的世界」模式（默认），保留「创建新世界」模式
- 新增 MC LAN 端口自动检测（UDP 组播 + 进程枚举）
- 新增前端模式选择 UI + LAN 开放引导提示
- 自动从 MC 进程提取用户名，不再硬编码
- 修复版本检测（过滤 Fabric/Forge，只识别纯数字版本）
- 修复 Java 版本检测（MC 1.21+ 需要 Java 21）
- 迁移 Paper API 到 fill.papermc.io v3
- 修复 Windows 进程分离（CREATE_BREAKAWAY_FROM_JOB）
- 新增服务端命令 API（/op 等）
- 修复房间退出清理（停服 + 关隧道 + 释放端口）
