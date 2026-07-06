use serde::{Deserialize, Serialize};

/// 服务器状态
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ServerStatus {
    pub running: bool,
    pub version: String,
    pub tps: f64,
    pub memory_usage: f64,
    pub uptime_secs: u64,
    pub players_online: u32,
    pub max_players: u32,
}

/// 启动 MC 服务端
#[tauri::command]
pub async fn start_server(
    mc_version: Option<String>,
    game_mode: Option<String>,
) -> Result<ServerStatus, String> {
    tracing::info!("启动服务端: version={:?}, mode={:?}", mc_version, game_mode);

    // TODO: 实现服务端启动
    // 1. 检测 Java 运行时
    // 2. 下载/缓存 Paper 服务端 JAR
    // 3. 生成 server.properties
    // 4. 启动子进程
    // 5. 等待就绪

    Ok(ServerStatus {
        running: true,
        version: mc_version.unwrap_or_else(|| "1.21".into()),
        tps: 20.0,
        memory_usage: 45.0,
        uptime_secs: 0,
        players_online: 0,
        max_players: 8,
    })
}

/// 停止 MC 服务端
#[tauri::command]
pub async fn stop_server() -> Result<(), String> {
    tracing::info!("停止服务端");
    // TODO: 发送 /stop 命令，等待进程退出
    Ok(())
}

/// 获取服务器状态
#[tauri::command]
pub async fn get_server_status() -> Result<ServerStatus, String> {
    Ok(ServerStatus {
        running: false,
        version: "1.21".into(),
        tps: 0.0,
        memory_usage: 0.0,
        uptime_secs: 0,
        players_online: 0,
        max_players: 8,
    })
}
