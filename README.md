# 微信公众号写作助手

[![Deploy](https://github.com/YOUR_USERNAME/wechat_writing_assistant/actions/workflows/deploy.yml/badge.svg)](https://github.com/YOUR_USERNAME/wechat_writing_assistant/actions/workflows/deploy.yml)
[![Tests](https://github.com/YOUR_USERNAME/wechat_writing_assistant/actions/workflows/test.yml/badge.svg)](https://github.com/YOUR_USERNAME/wechat_writing_assistant/actions/workflows/test.yml)

一个基于 Flutter 的跨平台微信公众号写作助手，提供热点追踪、AI 写作辅助、文章管理、数据分析和一键发布功能。

## ✨ 功能特性

- 🔥 **热点扫描**: 多平台热榜聚合（微博、知乎、百度、36氪）
- 🤖 **AI 写作**: 支持 OpenAI、Claude、Gemini 多提供商
- 📝 **文章管理**: 富文本编辑器、自动保存、版本历史
- 📊 **数据分析**: 文章统计、爆款率分析、阅读量追踪
- 📱 **微信发布**: 一键发布到公众号

## 🚀 快速开始

### 在线体验
访问 [https://YOUR_USERNAME.github.io/wechat_writing_assistant/](https://YOUR_USERNAME.github.io/wechat_writing_assistant/)

### 本地运行
```bash
git clone https://github.com/YOUR_USERNAME/wechat_writing_assistant.git
cd wechat_writing_assistant
flutter pub get
flutter run -d chrome
```

## 📋 环境要求
- Flutter: 3.24.0+
- Dart: 3.10.7+

## 🔧 配置
创建 `.env` 文件配置 API 密钥（详见 [DEPLOYMENT.md](DEPLOYMENT.md)）

## 🏗️ 架构
- **架构模式**: Clean Architecture + Feature-First
- **状态管理**: Riverpod 2.x
- **数据库**: Drift (SQLite)
- **测试覆盖率**: 91.32% ✅

## 📖 文档
- [部署指南](DEPLOYMENT.md)
- [GitHub 部署](GITHUB_DEPLOYMENT.md)
- [开发计划](.claude/specs/wechat-writing-assistant/dev-plan.md)

## 📄 许可证
MIT License

---
**开发状态**: ✅ 生产就绪 | **版本**: 1.0.0
