use serde::{Deserialize, Serialize};

/// NAT 类型
#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum NatType {
    /// 完全锥形 NAT - 最容易打洞
    FullCone,
    /// 受限锥形 NAT
    RestrictedCone,
    /// 端口受限锥形 NAT
    PortRestrictedCone,
    /// 对称 NAT - 最难打洞
    Symmetric,
    /// 未知
    Unknown,
}

/// 探测 NAT 类型
///
/// 通过向 STUN 服务器发送多个请求，
/// 根据响应判断 NAT 类型。
///
/// 算法基于 RFC 5780。
pub async fn detect_nat_type() -> Result<NatType, String> {
    // TODO: 实现 NAT 探测
    // 1. 发送 Binding Request 到主 STUN 服务器
    // 2. 请求从不同 IP:Port 响应 (CHANGE-REQUEST)
    // 3. 根据响应判断 NAT 类型
    Ok(NatType::Unknown)
}

/// 获取本机公网地址
pub async fn get_public_addr(stun_server: &str) -> Result<(String, u16), String> {
    // TODO: 实现 STUN 地址获取
    Err("未实现".into())
}

/// 获取本机内网地址
pub fn get_local_addrs() -> Vec<(String, u16)> {
    // TODO: 枚举本地网络接口
    vec![]
}
