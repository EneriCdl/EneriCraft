use serde::{Deserialize, Serialize};
use base64::{Engine as _, engine::general_purpose::STANDARD as BASE64};

/// 连接码版本前缀
const CODE_PREFIX: &str = "MC-CONNECT-v1";

/// 连接码数据结构
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ConnectCode {
    /// 协议版本
    pub v: u8,
    /// 端点列表 [(ip, port)]
    pub ep: Vec<(String, u16)>,
    /// QUIC 公钥 (base64)
    pub pk: String,
    /// MC 游戏版本
    pub mv: String,
    /// MOD 清单 Hash (SHA256)
    pub mh: String,
    /// 生成时间戳
    pub ts: u64,
}

impl ConnectCode {
    /// 生成连接码字符串
    ///
    /// 流程: JSON序列化 → Brotli压缩 → Base64编码 → 添加前缀
    pub fn generate(
        endpoints: Vec<(String, u16)>,
        public_key: &[u8],
        mc_version: &str,
        mod_hash: &str,
    ) -> Result<String, String> {
        let code = ConnectCode {
            v: 1,
            ep: endpoints,
            pk: BASE64.encode(public_key),
            mv: mc_version.to_string(),
            mh: mod_hash.to_string(),
            ts: std::time::SystemTime::now()
                .duration_since(std::time::UNIX_EPOCH)
                .unwrap()
                .as_secs(),
        };

        // JSON 序列化
        let json = serde_json::to_vec(&code)
            .map_err(|e| format!("序列化失败: {}", e))?;

        // Brotli 压缩
        let mut compressed = Vec::new();
        {
            let mut writer = brotli::CompressorWriter::new(&mut compressed, 4096, 11, 22);
            std::io::Write::write_all(&mut writer, &json)
                .map_err(|e| format!("压缩失败: {}", e))?;
        }

        // Base64 编码
        let encoded = BASE64.encode(&compressed);

        Ok(format!("{}-{}", CODE_PREFIX, encoded))
    }

    /// 解析连接码字符串
    ///
    /// 流程: 去除前缀 → Base64解码 → Brotli解压 → JSON反序列化
    pub fn parse(code: &str) -> Result<Self, String> {
        let stripped = code
            .strip_prefix(&format!("{}-", CODE_PREFIX))
            .ok_or_else(|| format!("无效的连接码格式，应以 {} 开头", CODE_PREFIX))?;

        // Base64 解码
        let compressed = BASE64
            .decode(stripped)
            .map_err(|e| format!("Base64解码失败: {}", e))?;

        // Brotli 解压
        let mut decompressed = Vec::new();
        {
            let mut reader = brotli::DecompressorReader::new(
                &compressed[..],
                4096,
            );
            std::io::Read::read_to_end(&mut reader, &mut decompressed)
                .map_err(|e| format!("解压失败: {}", e))?;
        }

        // JSON 反序列化
        let code: ConnectCode = serde_json::from_slice(&decompressed)
            .map_err(|e| format!("解析失败: {}", e))?;

        // 验证时间戳（2小时有效期）
        let now = std::time::SystemTime::now()
            .duration_since(std::time::UNIX_EPOCH)
            .unwrap()
            .as_secs();
        if now - code.ts > 7200 {
            return Err("连接码已过期（超过2小时）".into());
        }

        Ok(code)
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_generate_and_parse() {
        let code = ConnectCode::generate(
            vec![
                ("123.45.67.89".into(), 54321),
                ("192.168.1.5".into(), 54321),
            ],
            b"test-public-key",
            "1.21",
            "abc123",
        )
        .unwrap();

        assert!(code.starts_with(CODE_PREFIX));

        let parsed = ConnectCode::parse(&code).unwrap();
        assert_eq!(parsed.mv, "1.21");
        assert_eq!(parsed.ep.len(), 2);
    }
}
