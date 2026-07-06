mod commands;
mod tunnel;
mod server;
mod sync;
mod plugin;
mod signaling;
mod utils;

use tracing_subscriber;

pub fn run() {
    // 初始化日志
    tracing_subscriber::fmt()
        .with_env_filter(
            tracing_subscriber::EnvFilter::from_default_env()
                .add_directive("mc_connector=info".parse().unwrap())
        )
        .init();

    tauri::Builder::default()
        .plugin(tauri_plugin_clipboard_manager::init())
        .plugin(tauri_plugin_notification::init())
        .plugin(tauri_plugin_dialog::init())
        .plugin(tauri_plugin_fs::init())
        .invoke_handler(tauri::generate_handler![
            // 房间管理
            commands::room::create_room,
            commands::room::join_room,
            commands::room::leave_room,
            commands::room::get_room_status,
            // 隧道控制
            commands::tunnel::get_tunnel_status,
            commands::tunnel::get_connection_type,
            // 服务管理
            commands::server::start_server,
            commands::server::stop_server,
            commands::server::get_server_status,
            // 连接码
            commands::room::generate_connect_code,
            commands::room::parse_connect_code,
        ])
        .run(tauri::generate_context!())
        .expect("启动 MC 联机器失败");
}
