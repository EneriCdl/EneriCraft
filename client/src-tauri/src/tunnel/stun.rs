//! STUN 客户端
//!
//! 实现 RFC 5389 STUN Binding Request，
//! 用于获取公网 IP:Port 和判断 NAT 类型。

/// 公共 STUN 服务器列表
pub const PUBLIC_STUN_SERVERS: &[&str] = &[
    "stun.l.google.com:19302",
    "stun1.l.google.com:19302",
    "stun2.l.google.com:19302",
];

/// STUN 响应
#[derive(Debug, Clone)]
pub struct StunResponse {
    /// 公网 IP
    pub public_ip: String,
    /// 公网端口
    pub public_port: u16,
}

/// 发送 STUN Binding Request
///
/// 返回从 STUN 服务器视角看到的公网地址
pub async fn send_binding_request(stun_server: &str) -> Result<StunResponse, String> {
    // TODO: 实现 STUN Binding Request
    // 1. 创建 UDP socket
    // 2. 构造 STUN Binding Request (RFC 5389)
    // 3. 发送到 STUN 服务器
    // 4. 解析 XOR-MAPPED-ADDRESS 响应
    let _ = stun_server;
    Err("未实现".into())
}
