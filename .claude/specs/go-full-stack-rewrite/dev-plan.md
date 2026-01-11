# Go 全栈重写 - 开发计划

## 概述
使用 Go + Fyne 框架将现有 Flutter 应用重写为纯 Go 跨平台实现，支持桌面端（Windows/macOS/Linux）和移动端，采用 Clean Architecture + Feature-First 模式。

## 任务分解

### Task 1: Go Monorepo 骨架 + 平台工具 + CI
- **ID**: go-task-1
- **type**: default
- **Description**: 搭建 Go 项目基础架构，包括 monorepo 目录结构、平台工具库（HTTP 客户端、时钟抽象、日志）、SQLite 数据库配置、依赖注入框架、CI/CD 流水线配置
- **File Scope**:
  - `go.mod`, `go.sum`
  - `cmd/app/main.go`
  - `internal/platform/**` (http_client.go, clock.go, logger.go, database.go)
  - `internal/di/**` (container.go, wire.go)
  - `.github/workflows/go-ci.yml`
  - `Makefile`
  - `scripts/**` (build.sh, test.sh)
- **Dependencies**: None
- **Test Command**:
  ```bash
  go test ./internal/platform/... -coverprofile=coverage.out -covermode=atomic
  go tool cover -func=coverage.out | grep total | awk '{print $3}'
  ```
- **Test Focus**:
  - HTTP 客户端的重试机制和超时处理
  - 时钟抽象的 mock 能力（用于时间相关测试）
  - SQLite 连接池的并发安全性
  - 依赖注入容器的生命周期管理
  - 日志输出格式和级别控制

### Task 2: 热点话题模块
- **ID**: go-task-2
- **type**: default
- **Description**: 实现热点话题功能模块，包括领域模型、用例层（获取热榜、刷新、缓存）、数据层（API 适配器、本地缓存）、单元测试和集成测试
- **File Scope**:
  - `internal/features/hot_topics/domain/**` (topic.go, repository.go)
  - `internal/features/hot_topics/usecase/**` (fetch_topics.go, refresh_topics.go)
  - `internal/features/hot_topics/data/**` (api_client.go, cache_repository.go, sqlite_repository.go)
  - `internal/features/hot_topics/data/models/**` (topic_dto.go)
  - `tests/hot_topics/**` (usecase_test.go, repository_test.go, integration_test.go)
- **Dependencies**: depends on go-task-1
- **Test Command**:
  ```bash
  go test ./internal/features/hot_topics/... ./tests/hot_topics/... -coverprofile=coverage_hot_topics.out -covermode=atomic -v
  go tool cover -func=coverage_hot_topics.out | grep total | awk '{print $3}'
  ```
- **Test Focus**:
  - 热榜数据获取的成功和失败场景（网络错误、API 限流）
  - 缓存过期策略和刷新逻辑
  - 多平台热榜数据的聚合和排序
  - 并发请求的安全性
  - SQLite 持久化的 CRUD 操作
  - Mock HTTP 客户端和时钟的集成测试

### Task 3: 文章管理模块
- **ID**: go-task-3
- **type**: default
- **Description**: 实现文章管理功能模块，包括文章 CRUD、草稿保存、版本历史、标签分类、搜索过滤、数据持久化
- **File Scope**:
  - `internal/features/articles/domain/**` (article.go, tag.go, repository.go)
  - `internal/features/articles/usecase/**` (create_article.go, update_article.go, delete_article.go, list_articles.go, search_articles.go)
  - `internal/features/articles/data/**` (sqlite_repository.go, search_index.go)
  - `internal/features/articles/data/models/**` (article_dto.go, tag_dto.go)
  - `tests/articles/**` (usecase_test.go, repository_test.go, search_test.go)
- **Dependencies**: depends on go-task-1
- **Test Command**:
  ```bash
  go test ./internal/features/articles/... ./tests/articles/... -coverprofile=coverage_articles.out -covermode=atomic -v
  go tool cover -func=coverage_articles.out | grep total | awk '{print $3}'
  ```
- **Test Focus**:
  - 文章创建、更新、删除的完整流程
  - 草稿自动保存和恢复机制
  - 版本历史的存储和回滚
  - 标签的增删改查和关联关系
  - 全文搜索的准确性和性能
  - 并发写入的事务隔离
  - 边界条件（空标题、超长内容、特殊字符）

### Task 4: AI 写作模块
- **ID**: go-task-4
- **type**: default
- **Description**: 实现 AI 写作辅助功能，包括内容生成、续写、改写、摘要提取、与文章模块的集成、流式响应处理
- **File Scope**:
  - `internal/features/ai_writing/domain/**` (prompt.go, generation.go, repository.go)
  - `internal/features/ai_writing/usecase/**` (generate_content.go, rewrite_content.go, summarize.go)
  - `internal/features/ai_writing/data/**` (ai_client.go, prompt_repository.go)
  - `internal/features/ai_writing/data/models/**` (prompt_dto.go, generation_dto.go)
  - `tests/ai_writing/**` (usecase_test.go, ai_client_test.go, integration_test.go)
- **Dependencies**: depends on go-task-3
- **Test Command**:
  ```bash
  go test ./internal/features/ai_writing/... ./tests/ai_writing/... -coverprofile=coverage_ai_writing.out -covermode=atomic -v
  go tool cover -func=coverage_ai_writing.out | grep total | awk '{print $3}'
  ```
- **Test Focus**:
  - AI API 调用的成功和失败场景（超时、限流、无效响应）
  - 流式响应的解析和错误处理
  - Prompt 模板的参数替换和验证
  - 生成内容与文章的关联保存
  - 并发生成请求的队列管理
  - Mock AI 客户端的单元测试
  - 与文章模块的集成测试（生成后保存为草稿）

### Task 5: Fyne GUI 集成
- **ID**: go-task-5
- **type**: ui
- **Description**: 使用 Fyne 框架实现跨平台 GUI，包括主窗口布局、四个核心模块的 UI 界面、导航路由、数据绑定、主题适配、移动端适配
- **File Scope**:
  - `internal/ui/**` (app.go, theme.go, navigation.go)
  - `internal/ui/screens/hot_topics/**` (hot_topics_screen.go, topic_card.go)
  - `internal/ui/screens/articles/**` (articles_list_screen.go, article_editor_screen.go)
  - `internal/ui/screens/ai_writing/**` (ai_writing_screen.go, generation_panel.go)
  - `internal/ui/screens/analytics/**` (analytics_screen.go, charts.go)
  - `internal/ui/widgets/**` (custom_button.go, loading_indicator.go)
  - `internal/ui/bindings/**` (topic_binding.go, article_binding.go)
  - `tests/ui/**` (screen_test.go, widget_test.go, navigation_test.go)
- **Dependencies**: depends on go-task-2, go-task-3, go-task-4
- **Test Command**:
  ```bash
  go test ./internal/ui/... ./tests/ui/... -coverprofile=coverage_ui.out -covermode=atomic -v
  go tool cover -func=coverage_ui.out | grep total | awk '{print $3}'
  ```
- **Test Focus**:
  - 各屏幕的初始化和渲染逻辑
  - 数据绑定的双向同步
  - 导航路由的正确性（前进、后退、深度链接）
  - 主题切换的视觉一致性
  - 移动端布局的响应式适配
  - 自定义组件的交互行为
  - 错误状态的 UI 反馈（加载失败、网络错误）
  - 长列表的虚拟滚动性能

## 验收标准
- [ ] 所有五个任务的单元测试通过
- [ ] 每个模块的代码覆盖率 ≥90%
- [ ] CI/CD 流水线成功构建 Windows、macOS、Linux 三个平台的可执行文件
- [ ] 热点话题模块能成功获取并缓存多平台热榜数据
- [ ] 文章管理模块支持完整的 CRUD、搜索、标签功能
- [ ] AI 写作模块能调用 AI API 并将生成内容保存为文章草稿
- [ ] Fyne GUI 在桌面端正常运行，四个核心模块界面可导航和交互
- [ ] 移动端布局适配完成（至少在模拟器中验证）
- [ ] SQLite 数据库迁移脚本可正常执行
- [ ] 所有依赖注入关系正确配置，应用可正常启动

## 技术要点
- **架构模式**: Clean Architecture 三层分离（domain/usecase/data），每个 feature 独立模块
- **GUI 框架**: Fyne v2.x，支持桌面端和移动端跨平台编译
- **数据库**: SQLite（使用 `modernc.org/sqlite` 纯 Go 驱动，无 CGO 依赖）
- **依赖注入**: 使用 `google/wire` 或手动构造，确保可测试性
- **测试策略**:
  - 单元测试使用 mock 接口（HTTP 客户端、时钟、数据库）
  - 集成测试使用内存 SQLite 数据库
  - UI 测试使用 Fyne 的 test 包进行无头测试
- **并发安全**: 所有共享状态使用 mutex 或 channel 保护
- **错误处理**: 使用 `errors.Is/As` 进行错误类型判断，避免字符串比较
- **构建约束**:
  - 桌面端使用 `go build` 直接编译
  - 移动端使用 `fyne package -os android/ios`
- **CI/CD**: GitHub Actions 矩阵构建，缓存 Go 模块和构建产物
- **性能考虑**:
  - 热榜数据使用 TTL 缓存减少 API 调用
  - 文章列表使用分页加载
  - AI 生成使用流式响应提升用户体验
