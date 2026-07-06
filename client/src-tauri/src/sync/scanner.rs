//! MOD 文件扫描器
//!
//! 递归扫描 mods/ 目录，计算每个文件的 SHA256 Hash

use serde::{Deserialize, Serialize};
use sha2::{Sha256, Digest};
use std::path::Path;

/// MOD 文件信息
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ModFile {
    /// 相对路径 (如 "mods/fabric-api-0.92.0.jar")
    pub path: String,
    /// 文件大小（字节）
    pub size: u64,
    /// SHA256 Hash
    pub sha256: String,
}

/// MOD 目录清单
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ModManifest {
    pub mc_version: String,
    pub mods: Vec<ModFile>,
}

/// 扫描 MOD 目录
pub fn scan_mods_dir(mods_dir: &Path) -> Result<ModManifest, String> {
    let mut mods = Vec::new();

    if !mods_dir.exists() {
        return Ok(ModManifest {
            mc_version: "unknown".into(),
            mods,
        });
    }

    for entry in walkdir::WalkDir::new(mods_dir)
        .follow_links(false)
        .into_iter()
        .filter_map(|e| e.ok())
        .filter(|e| e.file_type().is_file())
    {
        let path = entry.path();
        let ext = path.extension()
            .and_then(|e| e.to_str())
            .unwrap_or("");

        // 只扫描 .jar 文件
        if ext != "jar" {
            continue;
        }

        let metadata = std::fs::metadata(path).map_err(|e| e.to_string())?;
        let hash = compute_file_hash(path)?;

        mods.push(ModFile {
            path: path.file_name()
                .unwrap()
                .to_string_lossy()
                .to_string(),
            size: metadata.len(),
            sha256: hash,
        });
    }

    Ok(ModManifest {
        mc_version: "unknown".into(),
        mods,
    })
}

/// 计算文件 SHA256
fn compute_file_hash(path: &Path) -> Result<String, String> {
    let data = std::fs::read(path).map_err(|e| e.to_string())?;
    let mut hasher = Sha256::new();
    hasher.update(&data);
    let hash = hasher.finalize();
    Ok(format!("{:x}", hash))
}
