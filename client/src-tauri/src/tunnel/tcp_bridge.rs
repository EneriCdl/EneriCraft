//! TCP ↔ QUIC 桥接
//!
//! 核心功能：
//! 1. 在本地监听 TCP 端口（如 25565）
//! 2. 将接入的 TCP 连接数据包装成 QUIC Stream 发送
//! 3. 对端从 QUIC Stream 读取数据，还原为 TCP 连接到 MC 服务端
//!
//! 这是实现"透明联机"的关键模块。

use tokio::net::{TcpListener, TcpStream};

/// TCP Bridge 方向
pub enum BridgeDirection {
    /// 主机端：监听本地 TCP → 转发到 QUIC Stream
    Host,
    /// 客户端：从 QUIC Stream 读取 → 转发到本地 MC
    Client,
}

/// TCP Bridge
pub struct TcpBridge {
    direction: BridgeDirection,
    local_port: u16,
    listener: Option<TcpListener>,
}

impl TcpBridge {
    /// 创建主机端 Bridge
    ///
    /// 监听本地端口，将 Minecraft 客户端连接转发到 QUIC 隧道
    pub async fn new_host(local_port: u16) -> Result<Self, String> {
        let listener = TcpListener::bind(format!("127.0.0.1:{}", local_port))
            .await
            .map_err(|e| format!("无法绑定端口 {}: {}", local_port, e))?;

        Ok(Self {
            direction: BridgeDirection::Host,
            local_port,
            listener: Some(listener),
        })
    }

    /// 创建客户端 Bridge
    ///
    /// 接受 QUIC Stream 数据，转发到本地 Minecraft 客户端
    pub fn new_client(remote_port: u16) -> Self {
        Self {
            direction: BridgeDirection::Client,
            local_port: remote_port,
            listener: None,
        }
    }

    /// 获取本地监听端口
    pub fn local_port(&self) -> u16 {
        self.local_port
    }

    /// 启动桥接
    pub async fn start(&mut self) -> Result<(), String> {
        // TODO: 实现 TCP ↔ QUIC 数据转发
        // Host 模式：
        //   loop { accept TCP → spawn task → read TCP → write QUIC Stream }
        // Client 模式：
        //   read QUIC Stream → connect to localhost:25565 → write TCP
        Ok(())
    }

    /// 停止桥接
    pub async fn stop(&mut self) {
        // TODO: 关闭监听器和所有活跃连接
    }
}
