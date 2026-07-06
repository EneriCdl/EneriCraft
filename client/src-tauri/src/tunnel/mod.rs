pub mod nat;
pub mod quic_mgr;
pub mod tcp_bridge;
pub mod stun;
pub mod connect_code;

/// 隧道引擎 - 管理 P2P 连接的生命周期
pub struct TunnelEngine {
    // TODO: 实现隧道引擎
}

impl TunnelEngine {
    pub fn new() -> Self {
        Self {}
    }
}
