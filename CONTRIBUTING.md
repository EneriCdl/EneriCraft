# 贡献指南 (Contributing Guide)

感谢你对 MC 联机器的关注！我们欢迎任何形式的贡献。

## 行为准则

- 尊重所有贡献者，保持友好和专业
- 对新手保持耐心，乐于帮助
- 接受建设性批评，专注于项目改进

## 如何贡献

### 报告 Bug

1. 在 GitHub Issues 中搜索是否已有相同问题
2. 使用 Bug Report 模板
3. 提供详细的：操作系统版本、MC 版本、操作步骤、截图/日志

### 功能建议

1. 在 Discussions 中先讨论想法
2. 描述你希望解决的问题，而非具体实现
3. 说明该功能对普通用户的价值

### 提交代码

1. Fork 本仓库
2. 创建功能分支：`git checkout -b feature/your-feature`
3. 遵循代码规范（见下文）
4. 确保所有测试通过
5. 提交 PR 到 `dev` 分支
6. PR 描述中关联相关 Issue

### 代码规范

- **Rust**: 遵循 `rustfmt` 和 `clippy` 默认规则
- **TypeScript**: 使用 Prettier + ESLint
- **Go**: 遵循 `gofmt` 标准格式
- 提交信息格式：`[模块] 简短描述`

## 开发环境搭建

详见 [docs/architecture.md](docs/architecture.md)

## 项目结构

```
mc-connector/
├── client/        # Tauri 桌面客户端
├── server/        # Go 服务端
├── protocol/      # 通信协议定义
├── plugins/       # 官方插件
└── docs/          # 文档
```

## 联系方式

- GitHub Discussions: 技术讨论 & 功能建议
- QQ 群: （待创建）
- B站: （待创建）
