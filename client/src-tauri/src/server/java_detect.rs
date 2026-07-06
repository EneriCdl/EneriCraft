//! Java 运行时检测
//!
//! 扫描常见的 Java 安装路径，检测版本。
//! Minecraft 1.18+ 需要 Java 17+
//! Minecraft 1.17 需要 Java 16
//! Minecraft 1.16.5 及以下需要 Java 8/11

/// Java 安装信息
#[derive(Debug, Clone)]
pub struct JavaInstallation {
    pub path: String,
    pub version: String,
    pub major_version: u32,
    pub is_64bit: bool,
}

/// 检测系统上已安装的 Java
pub fn detect_java() -> Vec<JavaInstallation> {
    let mut installations = Vec::new();

    // Windows 常见路径
    let search_paths = [
        "C:\\Program Files\\Java",
        "C:\\Program Files (x86)\\Java",
        "C:\\Program Files\\Eclipse Adoptium",
        "C:\\Program Files\\Amazon Corretto",
    ];

    for path in &search_paths {
        if let Ok(entries) = std::fs::read_dir(path) {
            for entry in entries.flatten() {
                let java_path = entry.path().join("bin").join("java.exe");
                if java_path.exists() {
                    // TODO: 运行 java -version 获取版本
                    installations.push(JavaInstallation {
                        path: java_path.to_string_lossy().to_string(),
                        version: "unknown".into(),
                        major_version: 0,
                        is_64bit: true,
                    });
                }
            }
        }
    }

    installations
}

/// 查找最适合目标 MC 版本的 Java
pub fn find_best_java(mc_version: &str) -> Option<JavaInstallation> {
    let required = required_java_version(mc_version);
    let installations = detect_java();

    installations
        .into_iter()
        .filter(|j| j.major_version >= required)
        .max_by_key(|j| j.major_version)
}

/// MC 版本 → 最低 Java 大版本
fn required_java_version(mc_version: &str) -> u32 {
    // 解析主版本号
    let parts: Vec<&str> = mc_version.split('.').collect();
    if parts.len() < 2 {
        return 17;
    }

    let major: u32 = parts[0].parse().unwrap_or(1);
    let minor: u32 = parts[1].parse().unwrap_or(21);

    match (major, minor) {
        (1, m) if m <= 16 => 8,   // 1.16.5 及以下 → Java 8
        (1, 17) => 16,             // 1.17 → Java 16
        _ => 17,                   // 1.18+ → Java 17+
    }
}
