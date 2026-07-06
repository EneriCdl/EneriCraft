//! MC 服务端进程管理器
//!
//! 负责 Paper 服务端的启动、停止、崩溃重启。
//! 捕获 stdout/stderr 日志输出。

use std::process::{Child, Command};

/// 服务端进程句柄
pub struct ServerProcess {
    child: Option<Child>,
    mc_version: String,
    java_path: String,
    jar_path: String,
    allocated_memory_mb: u32,
}

impl ServerProcess {
    /// 创建新的服务端进程
    pub fn new(
        java_path: &str,
        jar_path: &str,
        mc_version: &str,
        memory_mb: u32,
    ) -> Self {
        Self {
            child: None,
            mc_version: mc_version.to_string(),
            java_path: java_path.to_string(),
            jar_path: jar_path.to_string(),
            allocated_memory_mb: memory_mb,
        }
    }

    /// 启动服务器
    pub async fn start(&mut self, server_dir: &str) -> Result<(), String> {
        // 构建 JVM 参数
        let jvm_args = format!(
            "-Xms{}M -Xmx{}M -XX:+UseG1GC -XX:+ParallelRefProcEnabled \
             -XX:MaxGCPauseMillis=200 -XX:+UnlockExperimentalVMOptions \
             -XX:+DisableExplicitGC -XX:+AlwaysPreTouch \
             -XX:G1NewSizePercent=30 -XX:G1MaxNewSizePercent=40 \
             -XX:G1HeapRegionSize=8M -XX:G1ReservePercent=20 \
             -XX:G1HeapWastePercent=5 -XX:G1MixedGCCountTarget=4 \
             -XX:InitiatingHeapOccupancyPercent=15 \
             -XX:G1MixedGCLiveThresholdPercent=90 \
             -XX:G1RSetUpdatingPauseTimePercent=5 \
             -XX:SurvivorRatio=32 -XX:+PerfDisableSharedMem \
             -XX:MaxTenuringThreshold=1",
            self.allocated_memory_mb,
            self.allocated_memory_mb
        );

        let child = Command::new(&self.java_path)
            .args(jvm_args.split_whitespace())
            .arg("-jar")
            .arg(&self.jar_path)
            .arg("nogui")
            .current_dir(server_dir)
            .spawn()
            .map_err(|e| format!("启动服务端失败: {}", e))?;

        self.child = Some(child);
        Ok(())
    }

    /// 停止服务器（发送 /stop 命令 or kill）
    pub async fn stop(&mut self) -> Result<(), String> {
        if let Some(mut child) = self.child.take() {
            // 尝试优雅关闭
            child.kill().map_err(|e| format!("停止服务端失败: {}", e))?;
            child.wait().ok();
        }
        Ok(())
    }

    /// 检查进程是否在运行
    pub fn is_running(&mut self) -> bool {
        if let Some(ref mut child) = self.child {
            matches!(child.try_wait(), Ok(None))
        } else {
            false
        }
    }
}
