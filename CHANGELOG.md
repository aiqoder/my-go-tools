# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [v0.1.0] - 2026-02-16

### Added

- **gin-static-server**: 基于 Gin 的高性能静态文件服务器
  - 内存缓存（LRO 淘汰策略）
  - Gzip/Zstd 压缩
  - ETag/Last-Modified 条件请求
  - SPA 路由回退支持
  - 目录遍历防护

- **oauth2**: OAuth2 授权码模式登录后端
  - 授权码模式完整支持
  - 令牌换取和刷新
  - 用户信息获取

- **uf**: 用户反馈服务 API 客户端
  - 简洁易用的 API 设计
  - 自定义 HTTP 客户端支持

### Documentation

- 完善的各模块 README 文档
- 项目根目录 README 补全

[Unreleased]: https://github.com/aiqoder/my-go-tools/compare/v0.1.0...HEAD
[v0.1.0]: https://github.com/aiqoder/my-go-tools/releases/tag/v0.1.0
