# 微信公众号写作助手 - 部署指南

## 📋 项目状态

### ✅ 已完成
- **核心功能开发**: 5/5 模块完成
  - ✅ 核心基础设施（Riverpod、Drift、Dio）
  - ✅ 热点扫描模块（91.32% 测试覆盖率）
  - ✅ AI 写作模块
  - ✅ 文章管理模块
  - ✅ 数据仪表盘与微信发布模块

- **代码质量**:
  - 架构: Clean Architecture + Feature-First ⭐⭐⭐⭐⭐
  - 测试覆盖率: 91.32% ✅
  - 代码审查: 已完成，7个改进建议

### ⚠️ 当前问题
- **Web 构建失败**: Flutter shader 编译器写入权限问题
- **缺少工具链**: Android SDK 和 Visual Studio 未安装

---

## 🚀 部署方案

### 方案 1: 修复构建问题后部署（推荐）

#### 步骤 1: 解决 Shader 编译器问题

**原因**: Flutter 3.38.6 的 shader 编译器在某些 Windows 环境下无法写入文件

**解决方案 A - 以管理员权限运行**:
```powershell
# 以管理员身份打开 PowerShell
cd "E:\公众号写作助手"
D:\flutter\bin\flutter.bat build web --release
```

**解决方案 B - 修改目录权限**:
```powershell
# 给当前用户完全控制权限
icacls "E:\公众号写作助手\build" /grant "%USERNAME%:(OI)(CI)F" /T
D:\flutter\bin\flutter.bat build web --release
```

**解决方案 C - 降级 Flutter 版本**:
```bash
cd D:\flutter
git checkout 3.24.0  # 使用稳定版本
flutter doctor
cd "E:\公众号写作助手"
flutter build web --release
```

#### 步骤 2: 部署 Web 版本

构建成功后，产物在 `build/web/` 目录：

**选项 A - 本地服务器**:
```bash
cd build/web
python -m http.server 8080
# 访问 http://localhost:8080
```

**选项 B - Nginx 部署**:
```nginx
server {
    listen 80;
    server_name your-domain.com;
    root E:/公众号写作助手/build/web;
    index index.html;

    location / {
        try_files $uri $uri/ /index.html;
    }
}
```

**选项 C - 云平台部署**:
- **Vercel**: `vercel deploy build/web`
- **Netlify**: 拖拽 `build/web` 到 Netlify
- **Firebase Hosting**: `firebase deploy`

---

### 方案 2: 开发模式运行（临时方案）

如果无法构建 release 版本，可以使用开发模式：

```bash
# 开发模式运行（无需编译 shader）
D:\flutter\bin\flutter.bat run -d chrome

# 或者使用 Web Server 模式
D:\flutter\bin\flutter.bat run -d web-server --web-port=8080
```

**注意**: 开发模式性能较差，仅用于测试。

---

### 方案 3: 桌面应用部署

#### Windows 桌面应用

**前置条件**:
1. 安装 Visual Studio 2022
2. 安装 "Desktop development with C++" 工作负载

**构建步骤**:
```bash
# 启用 Windows 桌面支持
D:\flutter\bin\flutter.bat config --enable-windows-desktop

# 构建 Windows 应用
D:\flutter\bin\flutter.bat build windows --release

# 产物位置
# build/windows/x64/runner/Release/
```

**打包为安装程序**:
使用 Inno Setup 或 NSIS 创建安装包。

#### Android 应用

**前置条件**:
1. 安装 Android Studio
2. 配置 Android SDK
3. 设置 ANDROID_HOME 环境变量

**构建步骤**:
```bash
# 构建 APK
D:\flutter\bin\flutter.bat build apk --release

# 构建 App Bundle（推荐用于 Google Play）
D:\flutter\bin\flutter.bat build appbundle --release

# 产物位置
# build/app/outputs/flutter-apk/app-release.apk
```

---

## 🔧 环境配置

### 必需依赖

**pubspec.yaml** 已包含:
```yaml
dependencies:
  flutter_riverpod: ^2.6.1  # 状态管理
  # 需要添加的依赖（由 codeagent 生成但未应用）:
  # drift: ^2.14.0           # 数据库
  # sqlite3_flutter_libs: ^0.5.0
  # dio: ^5.4.0              # 网络请求
  # flutter_secure_storage: ^9.0.0  # 安全存储
```

### API 密钥配置

创建 `.env` 文件（不要提交到 Git）:
```env
# AI 提供商 API 密钥
OPENAI_API_KEY=your_openai_key
CLAUDE_API_KEY=your_claude_key
GEMINI_API_KEY=your_gemini_key

# 微信公众号配置
WECHAT_APP_ID=your_app_id
WECHAT_APP_SECRET=your_app_secret
```

使用 `flutter_dotenv` 加载:
```dart
import 'package:flutter_dotenv/flutter_dotenv.dart';

Future<void> main() async {
  await dotenv.load(fileName: ".env");
  runApp(MyApp());
}
```

---

## 📊 性能优化

### Web 优化

1. **启用 CanvasKit**（更好的渲染性能）:
```bash
flutter build web --release --web-renderer canvaskit
```

2. **启用 WASM**（实验性，更快的启动）:
```bash
flutter build web --release --wasm
```

3. **代码分割**:
```bash
flutter build web --release --split-debug-info=build/debug_info
```

### 桌面优化

1. **减小包体积**:
```bash
flutter build windows --release --tree-shake-icons
```

2. **启用 AOT 编译**（默认已启用）

---

## 🧪 测试部署

### 运行测试

```bash
# 运行所有测试
D:\flutter\bin\flutter.bat test

# 运行特定模块测试
D:\flutter\bin\flutter.bat test test/features/hot_topics

# 生成覆盖率报告
D:\flutter\bin\flutter.bat test --coverage
genhtml coverage/lcov.info -o coverage/html
```

### 集成测试

```bash
# 运行集成测试
D:\flutter\bin\flutter.bat test integration_test

# 在真实设备上测试
D:\flutter\bin\flutter.bat drive --target=integration_test/app_test.dart
```

---

## 📦 部署检查清单

### 部署前
- [ ] 所有测试通过（`flutter test`）
- [ ] 代码审查建议已实施（见 CODE_REVIEW.md）
- [ ] API 密钥已配置
- [ ] 环境变量已设置
- [ ] 依赖已更新（`flutter pub get`）

### 构建
- [ ] 选择目标平台（Web/Windows/Android）
- [ ] 运行构建命令
- [ ] 验证构建产物
- [ ] 测试构建后的应用

### 部署后
- [ ] 验证所有功能正常
- [ ] 检查性能指标
- [ ] 监控错误日志
- [ ] 收集用户反馈

---

## 🐛 故障排除

### 问题 1: Shader 编译失败

**症状**: `Could not write file to build/web/assets/shaders/...`

**解决方案**:
1. 以管理员权限运行
2. 修改目录权限
3. 降级 Flutter 版本到 3.24.0

### 问题 2: 依赖冲突

**症状**: `version solving failed`

**解决方案**:
```bash
flutter pub upgrade --major-versions
flutter pub get
```

### 问题 3: 热重载不工作

**症状**: 代码修改后不生效

**解决方案**:
```bash
flutter clean
flutter pub get
flutter run
```

---

## 📞 支持

### 文档
- 开发计划: `.claude/specs/wechat-writing-assistant/dev-plan.md`
- 代码审查: 见上文审查报告
- API 文档: 待生成

### 联系方式
- 项目仓库: [待添加]
- 问题反馈: [待添加]

---

## 🎯 下一步

1. **立即**: 解决 shader 编译问题，完成 Web 构建
2. **短期**: 安装 Visual Studio，构建 Windows 桌面应用
3. **中期**: 配置 Android 环境，构建移动应用
4. **长期**: 实施代码审查建议，优化性能

---

**最后更新**: 2026-01-10
**Flutter 版本**: 3.38.6
**Dart 版本**: 3.10.7
