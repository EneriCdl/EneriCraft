use serde::{Deserialize, Serialize};

/// 隧道状态
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct TunnelStatus {
    pub connected: bool,
    pub connection_type: String,
    pub latency_ms: u32,
    pub bytes_sent: u64,
    pub bytes_received: u64,
}

/// 获取隧道状态
#[tauri::command]
pub async fn get_tunnel_status() -> Result<TunnelStatus, String> {
    Ok(TunnelStatus {
        connected: false,
        connection_type: "none".into(),
        latency_ms: 0,
        bytes_sent: 0,
        bytes_received: 0,
    })
}

/// 获取连接类型
#[tauri::command]
pub async fn get_connection_type() -> Result<String, String> {
    Ok("p2p".into())
}
