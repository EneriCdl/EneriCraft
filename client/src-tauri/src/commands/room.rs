use serde::{Deserialize, Serialize};
use tauri::State;
use crate::tunnel::TunnelEngine;
use crate::utils::config::AppConfig;

/// 房间状态
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct RoomStatus {
    pub connected: bool,
    pub room_code: Option<String>,
    pub connect_code: Option<String>,
    pub connection_type: String,
    pub players: Vec<String>,
    pub mc_version: String,
}

/// 创建房间
#[tauri::command]
pub async fn create_room(
    config: State<'_, AppConfig>,
    game_mode: Option<String>,
    room_name: Option<String>,
) -> Result<RoomStatus, String> {
    tracing::info!("创建房间: mode={:?}, name={:?}", game_mode, room_name);

    // 1. 启动 MC 服务端
    // let server = crate::server::ServerManager::start(config).await?;

    // 2. 探测 NAT 类型
    // let nat_type = crate::tunnel::nat::detect_nat_type().await?;

    // 3. 生成连接码
    // let connect_code = crate::tunnel::connect_code::generate(...)?;

    // 4. 返回房间状态
    Ok(RoomStatus {
        connected: false,
        room_code: None,
        connect_code: None,
        connection_type: "unknown".into(),
        players: vec![],
        mc_version: "1.21".into(),
    })
}

/// 加入房间
#[tauri::command]
pub async fn join_room(code: String) -> Result<RoomStatus, String> {
    tracing::info!("加入房间: code={}", code);

    // 1. 解析连接码或房间码
    // let endpoints = if code.starts_with("MC-CONNECT-") {
    //     crate::tunnel::connect_code::parse(&code)?
    // } else {
    //     // 房间码，需要信令服务器查询（可选功能）
    //     return Err("需要信令服务器支持，请使用连接码".into());
    // };

    // 2. 建立 P2P 连接
    // let tunnel = crate::tunnel::TunnelEngine::connect(endpoints).await?;

    // 3. 同步 MOD
    // crate::sync::sync_mods(&tunnel).await?;

    Ok(RoomStatus {
        connected: true,
        room_code: Some(code),
        connect_code: None,
        connection_type: "p2p".into(),
        players: vec!["玩家1".into()],
        mc_version: "1.21".into(),
    })
}

/// 离开房间
#[tauri::command]
pub async fn leave_room() -> Result<(), String> {
    tracing::info!("离开房间");
    // crate::tunnel::TunnelEngine::disconnect().await?;
    // crate::server::ServerManager::stop().await?;
    Ok(())
}

/// 获取房间状态
#[tauri::command]
pub async fn get_room_status() -> Result<RoomStatus, String> {
    Ok(RoomStatus {
        connected: false,
        room_code: None,
        connect_code: None,
        connection_type: "none".into(),
        players: vec![],
        mc_version: "1.21".into(),
    })
}

/// 生成连接码
#[tauri::command]
pub async fn generate_connect_code() -> Result<String, String> {
    // TODO: 实现连接码生成
    // let endpoints = get_local_endpoints().await?;
    // let nat_type = detect_nat_type().await?;
    // let code = ConnectCode::generate(endpoints, nat_type, mc_version, mod_hash)?;
    Ok("MC-CONNECT-v1-示例连接码".into())
}

/// 解析连接码
#[tauri::command]
pub async fn parse_connect_code(code: String) -> Result<serde_json::Value, String> {
    // TODO: 实现连接码解析
    tracing::info!("解析连接码: {}", code);
    Ok(serde_json::json!({
        "version": "1.21",
        "endpoints": []
    }))
}
