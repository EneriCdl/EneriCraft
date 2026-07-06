//! Paper 服务端 JAR 下载器
//!
//! 从 PaperMC API 获取指定 MC 版本的 Paper 服务端 JAR

/// 下载 Paper 服务端 JAR
///
/// 流程：
/// 1. 查询 PaperMC API 获取最新 Build
/// 2. 检查本地缓存（避免重复下载）
/// 3. 下载 JAR 文件（支持断点续传）
/// 4. SHA256 校验
pub async fn download_paper(
    mc_version: &str,
    cache_dir: &str,
) -> Result<String, String> {
    // TODO: 实现 Paper 下载
    // API: https://api.papermc.io/v2/projects/paper/versions/{version}
    let _ = mc_version;
    let _ = cache_dir;
    Err("未实现".into())
}

/// 检查本地缓存
pub fn check_cache(mc_version: &str, cache_dir: &str) -> Option<String> {
    let jar_path = std::path::Path::new(cache_dir)
        .join("paper")
        .join(format!("paper-{}.jar", mc_version));

    if jar_path.exists() {
        Some(jar_path.to_string_lossy().to_string())
    } else {
        None
    }
}
