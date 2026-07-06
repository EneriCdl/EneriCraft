//! server.properties 动态生成

use serde::{Deserialize, Serialize};

/// 服务器配置（映射 MC server.properties）
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ServerConfig {
    pub server_port: u16,
    pub gamemode: String,
    pub difficulty: String,
    pub max_players: u32,
    pub level_name: String,
    pub level_seed: Option<String>,
    pub online_mode: bool,
    pub pvp: bool,
    pub allow_nether: bool,
    pub allow_flight: bool,
    pub spawn_animals: bool,
    pub spawn_monsters: bool,
    pub spawn_npcs: bool,
    pub view_distance: u32,
    pub simulation_distance: u32,
}

impl Default for ServerConfig {
    fn default() -> Self {
        Self {
            server_port: 25565,
            gamemode: "survival".into(),
            difficulty: "normal".into(),
            max_players: 8,
            level_name: "world".into(),
            level_seed: None,
            online_mode: false, // 联机器使用自己的验证
            pvp: true,
            allow_nether: true,
            allow_flight: false,
            spawn_animals: true,
            spawn_monsters: true,
            spawn_npcs: true,
            view_distance: 12,
            simulation_distance: 8,
        }
    }
}

impl ServerConfig {
    /// 生成 server.properties 文件内容
    pub fn to_properties_string(&self) -> String {
        let mut props = Vec::new();

        props.push(format!("server-port={}", self.server_port));
        props.push(format!("gamemode={}", self.gamemode));
        props.push(format!("difficulty={}", self.difficulty));
        props.push(format!("max-players={}", self.max_players));
        props.push(format!("level-name={}", self.level_name));
        if let Some(ref seed) = self.level_seed {
            props.push(format!("level-seed={}", seed));
        }
        props.push(format!("online-mode={}", self.online_mode));
        props.push(format!("pvp={}", self.pvp));
        props.push(format!("allow-nether={}", self.allow_nether));
        props.push(format!("allow-flight={}", self.allow_flight));
        props.push(format!("spawn-animals={}", self.spawn_animals));
        props.push(format!("spawn-monsters={}", self.spawn_monsters));
        props.push(format!("spawn-npcs={}", self.spawn_npcs));
        props.push(format!("view-distance={}", self.view_distance));
        props.push(format!("simulation-distance={}", self.simulation_distance));

        // 隐藏设置：禁止公网访问（安全）
        props.push("server-ip=127.0.0.1".into());
        props.push("enable-query=false".into());
        props.push("enable-rcon=false".into());

        props.join("\n")
    }
}
