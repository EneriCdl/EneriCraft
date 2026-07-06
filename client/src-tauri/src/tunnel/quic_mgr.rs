//! QUIC 连接管理器
//!
//! 基于 quinn crate 管理 QUIC 连接的生命周期：
//! - 生成自签名证书（仅用于加密，不用于身份验证）
//! - 建立/接受 QUIC 连接
//! - 多路复用 Stream
//! - 连接状态监控

use quinn::{Endpoint, Connection};
use std::sync::Arc;

/// QUIC 连接管理器
pub struct QuicManager {
    endpoint: Option<Endpoint>,
    connection: Option<Connection>,
}

impl QuicManager {
    /// 创建新的 QUIC 管理器（客户端模式）
    pub fn new_client() -> Result<Self, String> {
        // TODO: 创建客户端 Endpoint
        Ok(Self {
            endpoint: None,
            connection: None,
        })
    }

    /// 创建新的 QUIC 管理器（服务端模式）
    pub fn new_server(bind_addr: &str) -> Result<Self, String> {
        // TODO: 创建服务端 Endpoint，绑定到 bind_addr
        Ok(Self {
            endpoint: None,
            connection: None,
        })
    }

    /// 主动连接到远端
    pub async fn connect(&mut self, addr: &str) -> Result<(), String> {
        // TODO: 建立 QUIC 连接
        let _ = addr;
        Ok(())
    }

    /// 打开双向 Stream（用于承载 Minecraft TCP 流量）
    pub async fn open_stream(&self) -> Result<quinn::SendStream, String> {
        // TODO: 打开新的双向 Stream
        Err("未实现".into())
    }

    /// 关闭连接
    pub async fn close(&mut self) {
        // TODO: 优雅关闭连接
    }
}

/// 生成自签名证书
///
/// 仅用于 QUIC 加密传输，不作为身份验证凭据。
/// 身份验证通过连接码中的公钥 Hash 完成。
pub fn generate_self_signed_cert() -> Result<(rustls::Certificate, rustls::PrivateKey), String> {
    // TODO: 使用 rcgen 生成自签名证书
    Err("未实现".into())
}
