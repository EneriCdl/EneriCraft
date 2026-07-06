//! MOD 差异对比
//!
//! 对比主机和客户端的 MOD 清单，输出需要同步的文件列表。

use super::scanner::ModManifest;

/// 差异分析结果
#[derive(Debug)]
pub struct DiffResult {
    /// 客户端需要下载的文件
    pub to_download: Vec<String>,
    /// 客户端需要删除的旧版本文件
    pub to_remove: Vec<String>,
    /// 版本冲突（同名不同版本）
    pub conflicts: Vec<String>,
}

/// 对比主机和客户端 MOD 清单
pub fn diff(host: &ModManifest, client: &ModManifest) -> DiffResult {
    let mut to_download = Vec::new();
    let mut to_remove = Vec::new();
    let mut conflicts = Vec::new();

    // 客户端缺少的 MOD
    for host_mod in &host.mods {
        let client_mod = client.mods.iter()
            .find(|m| m.path == host_mod.path);

        match client_mod {
            Some(cm) if cm.sha256 != host_mod.sha256 => {
                // 同名但 Hash 不同 → 版本不一致
                to_download.push(host_mod.path.clone());
                conflicts.push(format!(
                    "{} (主机: {}, 客户端: {})",
                    host_mod.path,
                    &host_mod.sha256[..8],
                    &cm.sha256[..8]
                ));
            }
            Some(_) => {
                // 文件一致，跳过
            }
            None => {
                // 客户端缺少此文件
                to_download.push(host_mod.path.clone());
            }
        }
    }

    // 客户端多余的 MOD（主机没有的）
    for client_mod in &client.mods {
        if !host.mods.iter().any(|m| m.path == client_mod.path) {
            to_remove.push(client_mod.path.clone());
        }
    }

    DiffResult {
        to_download,
        to_remove,
        conflicts,
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use super::super::scanner::ModFile;

    #[test]
    fn test_no_diff() {
        let host = ModManifest {
            mc_version: "1.21".into(),
            mods: vec![ModFile {
                path: "mod.jar".into(),
                size: 100,
                sha256: "abc".into(),
            }],
        };
        let client = host.clone();
        let result = diff(&host, &client);
        assert!(result.to_download.is_empty());
        assert!(result.to_remove.is_empty());
        assert!(result.conflicts.is_empty());
    }

    #[test]
    fn test_missing_mod() {
        let host = ModManifest {
            mc_version: "1.21".into(),
            mods: vec![ModFile {
                path: "mod.jar".into(),
                size: 100,
                sha256: "abc".into(),
            }],
        };
        let client = ModManifest {
            mc_version: "1.21".into(),
            mods: vec![],
        };
        let result = diff(&host, &client);
        assert_eq!(result.to_download.len(), 1);
    }
}
