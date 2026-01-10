# GitHub 部署指南

## 🚀 快速开始

### 方案 1: GitHub Pages（推荐 - 免费）

**优点**:
- ✅ 完全免费
- ✅ 自动 HTTPS
- ✅ 全球 CDN 加速
- ✅ 自动部署

**步骤**:

#### 1. 创建 GitHub 仓库

```bash
cd "E:\公众号写作助手"

# 初始化 Git
git init
git add .
git commit -m "Initial commit: WeChat Writing Assistant"

# 创建 GitHub 仓库后
git remote add origin https://github.com/YOUR_USERNAME/wechat_writing_assistant.git
git branch -M main
git push -u origin main
```

#### 2. 启用 GitHub Pages

1. 进入仓库 Settings → Pages
2. Source 选择 "GitHub Actions"
3. 保存

#### 3. 自动部署

推送代码后，GitHub Actions 会自动：
- ✅ 运行测试
- ✅ 构建 Web 应用
- ✅ 部署到 GitHub Pages

**访问地址**: `https://YOUR_USERNAME.github.io/wechat_writing_assistant/`

---

### 方案 2: Vercel（推荐 - 免费 + 更快）

**优点**:
- ✅ 免费
- ✅ 更快的全球 CDN
- ✅ 自动预览部署
- ✅ 自定义域名

**步骤**:

1. 访问 [vercel.com](https://vercel.com)
2. 使用 GitHub 账号登录
3. 点击 "Import Project"
4. 选择你的仓库
5. 配置构建设置：
   ```
   Framework Preset: Other
   Build Command: flutter build web --release
   Output Directory: build/web
   Install Command: flutter pub get
   ```
6. 点击 "Deploy"

**访问地址**: `https://your-project.vercel.app`

---

### 方案 3: Netlify（免费）

**步骤**:

1. 访问 [netlify.com](https://netlify.com)
2. 连接 GitHub 仓库
3. 配置构建：
   ```
   Build command: flutter build web --release
   Publish directory: build/web
   ```
4. 部署

**访问地址**: `https://your-project.netlify.app`

---

### 方案 4: Firebase Hosting（免费额度）

**步骤**:

```bash
# 安装 Firebase CLI
npm install -g firebase-tools

# 登录
firebase login

# 初始化项目
firebase init hosting

# 配置
# Public directory: build/web
# Single-page app: Yes
# GitHub integration: Yes

# 部署
flutter build web --release
firebase deploy
```

**访问地址**: `https://your-project.web.app`

---

## 📋 GitHub Actions 工作流

已创建 3 个工作流：

### 1. `.github/workflows/deploy.yml` - 自动部署
**触发**: 推送到 main 分支
**功能**:
- 运行测试
- 构建 Web 应用
- 部署到 GitHub Pages

### 2. `.github/workflows/test.yml` - 持续测试
**触发**: 推送到任何分支、Pull Request
**功能**:
- 代码分析
- 运行测试
- 检查覆盖率（≥90%）
- 上传到 Codecov

### 3. `.github/workflows/build-all.yml` - 多平台构建
**触发**: 创建 tag（如 `v1.0.0`）
**功能**:
- 构建 Web、Windows、Linux、Android
- 创建 GitHub Release
- 上传所有构建产物

---

## 🔐 环境变量配置

### GitHub Secrets

在仓库 Settings → Secrets and variables → Actions 中添加：

```
OPENAI_API_KEY=your_key
CLAUDE_API_KEY=your_key
GEMINI_API_KEY=your_key
WECHAT_APP_ID=your_app_id
WECHAT_APP_SECRET=your_secret
```

### 在代码中使用

修改 `.github/workflows/deploy.yml`：

```yaml
- name: Build web
  env:
    OPENAI_API_KEY: ${{ secrets.OPENAI_API_KEY }}
    CLAUDE_API_KEY: ${{ secrets.CLAUDE_API_KEY }}
    GEMINI_API_KEY: ${{ secrets.GEMINI_API_KEY }}
    WECHAT_APP_ID: ${{ secrets.WECHAT_APP_ID }}
    WECHAT_APP_SECRET: ${{ secrets.WECHAT_APP_SECRET }}
  run: flutter build web --release --dart-define=OPENAI_API_KEY=$OPENAI_API_KEY
```

---

## 🎯 部署流程

### 日常开发

```bash
# 1. 开发功能
git checkout -b feature/new-feature
# ... 编写代码 ...

# 2. 提交代码
git add .
git commit -m "feat: add new feature"
git push origin feature/new-feature

# 3. 创建 Pull Request
# GitHub Actions 会自动运行测试

# 4. 合并到 main
# 自动部署到 GitHub Pages
```

### 发布新版本

```bash
# 1. 更新版本号
# 编辑 pubspec.yaml: version: 1.0.0+1

# 2. 创建 tag
git tag -a v1.0.0 -m "Release v1.0.0"
git push or0

# 3. GitHub Actions 自动构建所有平台
# 4. 创建 GitHub Release 并上传产物
```

---

## 📊 监控和分析

### 1. GitHub Actions 状态

查看构建状态：
- 仓库首页会显示 Actions 徽章
- Actions 标签页查看详细日志

### 2. 代码覆盖率

集成 Codecov：
1. 访问 [codps://codecov.io)
2. 使用 GitHub 登录
3. 启用你的仓库
4. 获取徽章添加到 README

### 3. 性能监控

添加 Google Analytics：

```html
<!-- web/index.html -->
<script async src="https://www.googletagmanager.com/gtag/js?id=GA_MEASUREMENT_ID"></script>
<script>
  window.dataLayer = window.dataLayer || [];
  function gtag(){dataLayer.push(arguments);}
  gtag('js', new Date());
  gtag('config', 'GA_MEASUREMENT_ID');
</script>
```

---

## 🐛 故障排除

### 问题 1: GitHub Actions 构建失败

**检查**:
1. Actions 标签页查看错误日志
2. 确认 Flutter 版本兼容性
3. 检查依赖是否正确

**解决**:
```yaml
# 使用稳定版本
- uses: subosito/flutter-action@v2
  with:
    flutter-version: '3.24.0'  # 而不是 3.38.6
```

### 问题 2: GitHub Pages 404

**原因**: base-href 配置错误

**解决**:
```yaml
# 确保 base-href 正确
- run: flutter build web --release --base-href "/wechat_writing_assistant/"
```

### 问题 3: API 密钥未生效

**检查**:
1. Secrets 是否正确配置
2. 工作流是否正确引用
3. 代码是否正确读取

---

## 📦 自定义域名

### GitHub Pages

1. 购买域名（如 `example.com`）
2. 添加 DNS 记录：
   ```
   Type: CNAME
   Name: www
   Value: YOUR_USERNAME.github.io
   ```
3. 在仓库 Settings → Pages → Custom domain 输入域名
4. 启用 "Enforce HTTPS"

### Vercel/Netlify

1. 在平台设置中添加自定义域名
2. 按照提示配置 DNS
3. 自动获得 SSL 证书

---

## 🚀 性能优化

### 1. 启用缓存

```yaml
# .github/workflows/deploy.yml
- uses: subosito/flutter-action@v2
  with:
    cache: true  # 启用缓存
```

### 2. 并行构建

```yaml
jobs:
  test:
    strategy:
      matrix:
        os: [ubuntu-latest, windows-latest, macos-latest]
    runs-on: ${{ matrix.os }}
```

### 3. 条件部署

```yaml
# 只在 main 分支部署
if: github.ref == 'refs/heads/main'
```

---

## 📈 最佳实践

### 1. 分支策略

```
main (生产环境)
  ↑
develop (开发环境)
  ↑
feature/* (功能分支)
```

### 2. 提交规范

使用 Conventional Commits：
```
feat: 新功能
fix: 修复 bug
docs: 文档更新
style: 代码格式
refactor: 重构
test: 测试
chore: 构建/工具
```

### 3. 版本管理

遵循语义化版本：
```
v1.0.0 (主版本.次版本.修订号)
```

---

## 🎊 完整部署检查清单

### 准备阶段
- [ ] 创建 GitHub 仓库
- [ ] 配置 GitHub Actions 工作流
- [ ] 添加 Secrets（API 密钥）
- [ ] 更新 README.md

### 首次部署
- [ ] 推送代码到 GitHub
- [ ] 启用 GitHub Pages
- [ ] 验证自动部署成功
- [ ] 测试部署的应用

### 持续集成
- [ ] 每次提交自动运行测试
- [ ] Pull Request 自动检查
- [ ] 代码覆盖率监控
- [ ] 自动部署到生产环境

### 发布管理
- [ ] 创建 Release tag
- [ ] 自动构建多平台
- [ ] 生成 Release Notes
- [ ] 通知用户更新

---

## 📞 支持

**文档**:
- GitHub Actions: https://docs.github.com/actions
- Flutter Web: https://docs.flutter.dev/platform-integration/web
- GitHub Pages: https://pages.github.com

**社区**:
- GitHub Discussions
- Stack Overflow
- Flutter Discord

---

**最后更新**: 2026-01-10
**推荐方案**: GitHub Pages（免费） 或 Vercel（更快）
