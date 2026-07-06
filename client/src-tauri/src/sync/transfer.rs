//! P2P MOD 文件传输
//!
//! 通过已建立的 QUIC 隧道传输 MOD 文件。
//! - 分块传输（16KB chunk）
//! - 每个 chunk SHA256 校验
//! - 断点续传

/// 传输配置
#[derive(Debug, Clone)]
pub struct TransferConfig {
    /// 分块大小（字节）
    pub chunk_size: usize,
    /// 超时时间（秒）
    pub timeout_secs: u64,
}

impl Default for TransferConfig {
    fn default() -> Self {
        Self {
            chunk_size: 16 * 1024, // 16KB
            timeout_secs: 60,
        }
    }
}

/// 传输进度
#[derive(Debug, Clone)]
pub struct TransferProgress {
    pub file_name: String,
    pub total_size: u64,
    pub transferred: u64,
    pub speed_bytes_per_sec: f64,
}

/// P2P 文件传输器
pub struct P2PTransfer {
    config: TransferConfig,
}

impl P2PTransfer {
    pub fn new(config: TransferConfig) -> Self {
        Self { config }
    }

    /// 发送文件
    pub async fn send_file(
        &self,
        file_path: &str,
        progress_callback: impl Fn(TransferProgress),
    ) -> Result<(), String> {
        // TODO: 实现 P2P 文件发送
        // 1. 读取文件
        // 2. 分块
        // 3. 通过 QUIC Stream 发送每个 chunk
        // 4. 每个 chunk 附带 SHA256
        // 5. 对端确认收到后继续下一个 chunk
        let _ = file_path;
        let _ = progress_callback;
        Err("未实现".into())
    }

    /// 接收文件
    pub async fn receive_file(
        &self,
        save_path: &str,
        expected_size: u64,
        expected_hash: &str,
        progress_callback: impl Fn(TransferProgress),
    ) -> Result<(), String> {
        // TODO: 实现 P2P 文件接收
        let _ = save_path;
        let _ = expected_size;
        let _ = expected_hash;
        let _ = progress_callback;
        Err("未实现".into())
    }
}
