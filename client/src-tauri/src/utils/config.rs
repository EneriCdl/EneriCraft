use serde::{Deserialize, Serialize};
use std::path::PathBuf;
use std::sync::Mutex;

/// 应用配置
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct AppConfig {
    /// 用户昵称
    pub nickname: String,
    /// 明/暗主题
    pub dark_mode: bool,
    /// MC 数据目录
    pub mc_data_dir: PathBuf,
    /// MOD 缓存目录
    pub mod_cache_dir: PathBuf,
    /// Java 路径（留空则自动检测）
    pub java_path: Option<String>,
    /// 分配给 MC 服务端的内存 (MB)
    pub server_memory_mb: u32,
    /// 社区中继服务器地址（可选）
    pub relay_server: Option<String>,
    /// 信令服务器地址（可选）
    pub signaling_server: Option<String>,
    /// 是否接受社区中继
    pub allow_community_relay: bool,
}

impl Default for AppConfig {
    fn default() -> Self {
        Self {
            nickname: "MC玩家".into(),
            dark_mode: true,
            mc_data_dir: dirs_data_dir().join("mc-connector").join("servers"),
            mod_cache_dir: dirs_data_dir().join("mc-connector").join("mods"),
            java_path: None,
            server_memory_mb: 2048,
            relay_server: None,
            signaling_server: None,
            allow_community_relay: true,
        }
    }
}

impl AppConfig {
    /// 加载配置
    pub fn load() -> Result<Self, String> {
        let config_path = config_path()?;
        if config_path.exists() {
            let data = std::fs::read_to_string(&config_path)
                .map_err(|e| format!("读取配置失败: {}", e))?;
            serde_json::from_str(&data)
                .map_err(|e| format!("解析配置失败: {}", e))
        } else {
            let config = AppConfig::default();
            config.save()?;
            Ok(config)
        }
    }

    /// 保存配置
    pub fn save(&self) -> Result<(), String> {
        let config_path = config_path()?;
        if let Some(parent) = config_path.parent() {
            std::fs::create_dir_all(parent)
                .map_err(|e| format!("创建配置目录失败: {}", e))?;
        }
        let data = serde_json::to_string_pretty(self)
            .map_err(|e| format!("序列化配置失败: {}", e))?;
        std::fs::write(&config_path, data)
            .map_err(|e| format!("写入配置失败: {}", e))
    }
}

fn config_path() -> Result<PathBuf, String> {
    let base = dirs_data_dir();
    Ok(base.join("mc-connector").join("config.json"))
}

fn dirs_data_dir() -> PathBuf {
    // Windows: %APPDATA%
    #[cfg(target_os = "windows")]
    {
        let appdata = std::env::var("APPDATA").unwrap_or_else(|_| ".".into());
        PathBuf::from(appdata)
    }
    #[cfg(not(target_os = "windows"))]
    {
        let home = std::env::var("HOME").unwrap_or_else(|_| ".".into());
        PathBuf::from(home).join(".local").join("share")
    }
}
