# 微信公众号写作助手 - 开发计划

## 概述
基于 Flutter 的跨平台微信公众号写作助手，采用 Clean Architecture + Feature-First 架构，提供热点追踪、AI 写作辅助、文章管理、数据分析和一键发布功能。

## 任务分解

### Task 1: 核心基础设施搭建
- **ID**: task-1
- **type**: default
- **描述**: 搭建项目核心基础设施，包括目录结构、状态管理、数据库层、网络层、安全存储和测试基础设施
- **文件范围**:
  - `lib/app/` - 应用入口和全局配置
  - `lib/core/network/` - Dio 网络层封装（拦截器、错误处理、重试机制）
  - `lib/core/database/` - Drift 数据库配置（表定义、DAO、迁移脚本）
  - `lib/core/storage/` - flutter_secure_storage 安全存储封装
  - `lib/core/error/` - 统一错误模型和异常处理
  - `lib/core/utils/` - 工具类（日期格式化、验证器、扩展方法）
  - `lib/shared/providers/` - Riverpod 全局 Provider 配置
  - `lib/shared/theme/` - 主题配置和样式常量
  - `lib/shared/widgets/` - 通用 UI 组件（按钮、输入框、加载指示器）
  - `test/core/**` - 核心模块单元测试
  - `test/shared/**` - 共享模块单元测试
- **依赖关系**: None
- **测试命令**: `flutter test test/core test/shared --coverage --coverage-path=coverage/task-1.lcov.info`
- **测试重点**:
  - Dio 拦截器正确处理请求/响应/错误
  - 网络超时和重试机制验证
  - Drift 数据库初始化和表创建
  - 数据库迁移脚本执行正确性
  - 安全存储的读写和加密验证
  - 错误模型序列化和反序列化
  - 工具类函数边界值测试（空值、特殊字符、极限值）
  - Riverpod Provider 依赖注入正确性

### Task 2: 热点扫描模块
- **ID**: task-2
- **type**: default
- **描述**: 实现多源热点话题扫描功能，包括 HotTopicSource 抽象层、多平台数据源适配器（微博、知乎、百度、36氪）、缓存策略、Repository 和 UseCase
- **文件范围**:
  - `lib/features/hot_topics/domain/entities/` - 热点实体模型（HotTopic、TopicSource）
  - `lib/features/hot_topics/domain/repositories/` - 热点仓库接口
  - `lib/features/hot_topics/domain/usecases/` - 业务用例（获取热点、刷新热点、搜索热点）
  - `lib/features/hot_topics/data/datasources/` - 数据源实现（远程 API、本地缓存）
  - `lib/features/hot_topics/data/models/` - 数据模型和 JSON 映射
  - `lib/features/hot_topics/data/repositories/` - 仓库实现（缓存优先策略）
  - `lib/features/hot_topics/presentation/providers/` - Riverpod 状态管理
  - `lib/features/hot_topics/presentation/pages/` - 热点列表页面
  - `lib/features/hot_topics/presentation/widgets/` - 热点卡片、筛选器组件
  - `test/features/hot_topics/**` - 完整测试覆盖
- **依赖关系**: depends on task-1
- **测试命令**: `flutter test test/features/hot_topics --coverage --coverage-path=coverage/task-2.lcov.info`
- **测试重点**:
  - 多平台 API 解析正确性（微博、知乎、百度、36氪）
  - 网络异常时降级到缓存数据
  - 缓存过期策略验证（TTL 机制）
  - 热点数据去重和排序算法
  - Repository 层缓存优先逻辑
  - UseCase 业务规则验证（刷新间隔限制、并发请求控制）
  - 热点搜索和过滤功能
  - UI 状态管理（加载、成功、错误、空状态）
  - 下拉刷新和分页加载
  - Widget 渲染性能测试（大列表场景）

### Task 3: AI 写作模块
- **ID**: task-3
- **type**: default
- **描述**: 实现 AI 写作辅助功能，包括多 AI 提供商抽象层（OpenAI、Claude、Gemini）、标题生成、内容扩写、文章改写、流式输出处理、提示词模板管理和结果持久化
- **文件范围**:
  - `lib/features/ai_writing/domain/entities/` - AI 实体（AIProvider、GenerationRequest、GenerationResult）
  - `lib/features/ai_writing/domain/repositories/` - AI 服务仓库接口
  - `lib/features/ai_writing/domain/usecases/` - 业务用例（生成标题、扩写内容、改写文章）
  - `lib/features/ai_writing/data/datasources/` - AI 提供商适配器（OpenAI、Claude、Gemini）
  - `lib/features/ai_writing/data/models/` - 请求/响应模型
  - `lib/features/ai_writing/data/repositories/` - 仓库实现（提供商路由、流式处理）
  - `lib/features/ai_writing/presentation/providers/` - Riverpod 状态管理（流式状态）
  - `lib/features/ai_writing/presentation/pages/` - AI 写作页面
  - `lib/features/ai_writing/presentation/widgets/` - 提示词输入、结果展示、流式动画组件
  - `lib/core/services/prompt_template_service.dart` - 提示词模板管理
  - `test/features/ai_writing/**` - 完整测试覆盖
- **依赖关系**: depends on task-1
- **测试命令**: `flutter test test/features/ai_writing --coverage --coverage-path=coverage/task-3.lcov.info`
- **测试重点**:
  - 多 AI 提供商统一接口适配正确性
  - OpenAI/Claude/Gemini API 调用和响应解析
  - 流式响应处理和状态更新
  - API 密钥验证和错误处理（401、429、500）
  - 提示词模板变量替换逻辑
  - 生成结果保存到数据库
  - 超时和重试机制（指数退避）
  - 并发请求限制和队列管理
  - UseCase 业务规则（字数限制、敏感词过滤）
  - UI 流式输出动画效果
  - 多轮对话历史管理
  - 结果复制和应用到文章编辑器

### Task 4: 文章管理模块
- **ID**: task-4
- **type**: ui
- **描述**: 实现文章全生命周期管理，包括 CRUD 操作、富文本/Markdown 编辑器、自动保存草稿、版本历史、搜索过滤、分类标签、发布状态管理和文章列表 UI
- **文件范围**:
  - `lib/features/articles/domain/entities/` - 文章实体（Article、Draft、Version）
  - `lib/features/articles/domain/repositories/` - 文章仓库接口
  - `lib/features/articles/domain/usecases/` - 业务用例（创建、更新、删除、搜索、发布）
  - `lib/features/articles/data/datasources/` - 本地数据源（Drift DAO）
  - `lib/features/articles/data/models/` - 数据模型
  - `lib/features/articles/data/repositories/` - 仓库实现（自动保存、版本控制）
  - `lib/features/articles/presentation/providers/` - Riverpod 状态管理
  - `lib/features/articles/presentation/pages/` - 文章列表、编辑器、详情页面
  - `lib/features/articles/presentation/widgets/` - 富文本编辑器、Markdown 编辑器、工具栏、预览组件
  - `lib/shared/widgets/rich_text_editor/` - 通用富文本编辑器组件
  - `test/features/articles/**` - 完整测试覆盖
- **依赖关系**: depends on task-1
- **测试命令**: `flutter test test/features/articles --coverage --coverage-path=coverage/task-4.lcov.info`
- **测试重点**:
  - 文章 CRUD 操作数据库事务正确性
  - 草稿自动保存机制（防抖、增量保存）
  - 版本历史记录和回滚功能
  - 文章搜索算法（全文搜索、标签匹配）
  - 分类和标签管理
  - 发布状态流转（草稿→待发布→已发布）
  - 富文本编辑器功能（加粗、斜体、列表、链接、图片）
  - Markdown 语法解析和渲染
  - 编辑器性能测试（大文档场景）
  - 图片上传和本地缓存
  - 实时预览切换
  - 文章列表分页和虚拟滚动
  - 搜索和过滤 UI 交互
  - 批量操作（删除、导出）
  - 空状态和错误提示
  - 无障碍访问支持（语义化标签、键盘导航）

### Task 5: 数据仪表盘与微信发布模块
- **ID**: task-5
- **type**: ui
- **描述**: 实现数据分析仪表盘和微信公众号一键发布功能，包括指标统计（总文章数、命中率、平均阅读量、质量评分）、图表展示、微信 API 集成（Access Token、素材上传、草稿箱、发布）、发布预览、封面设置和发布记录管理
- **文件范围**:
  - `lib/features/analytics/domain/entities/` - 分析实体（Metrics、Trend、Report）
  - `lib/features/analytics/domain/repositories/` - 分析仓库接口
  - `lib/features/analytics/domain/usecases/` - 业务用例（获取指标、生成报告、趋势分析）
  - `lib/features/analytics/data/datasources/` - 数据源（本地聚合、微信 API）
  - `lib/features/analytics/data/repositories/` - 仓库实现
  - `lib/features/analytics/presentation/providers/` - Riverpod 状态管理
  - `lib/features/analytics/presentation/pages/` - 仪表盘页面
  - `lib/features/analytics/presentation/widgets/` - 图表组件（折线图、柱状图、饼图）
  - `lib/features/wechat/domain/entities/` - 微信实体（WeChatConfig、Material、Draft、PublishRecord）
  - `lib/features/wechat/domain/repositories/` - 微信仓库接口
  - `lib/features/wechat/domain/usecases/` - 业务用例（上传素材、创建草稿、发布文章、重试发布）
  - `lib/features/wechat/data/datasources/` - 微信 API 数据源
  - `lib/features/wechat/data/models/` - 微信 API 模型
  - `lib/features/wechat/data/repositories/` - 仓库实现（Token 管理、错误重试）
  - `lib/features/wechat/presentation/providers/` - Riverpod 状态管理
  - `lib/features/wechat/presentation/pages/` - 发布页面、发布记录页面
  - `lib/features/wechat/presentation/widgets/` - 发布预览、封面编辑、摘要编辑组件
  - `lib/shared/widgets/charts/` - 通用图表组件库
  - `test/features/analytics/**` - 完整测试覆盖
  - `test/features/wechat/**` - 完整测试覆盖
- **依赖关系**: depends on task-1, task-4
- **测试命令**: `flutter test test/features/analytics test/features/wechat --coverage --coverage-path=coverage/task-5.lcov.info`
- **测试重点**:
  - 数据指标聚合计算正确性（总数、平均值、百分比）
  - 时间序列数据查询和趋势分析
  - 图表数据转换和渲染
  - 图表交互（缩放、工具提示、数据点选择）
  - 时间范围选择和数据刷新
  - 数据导出功能（CSV、JSON）
  - 响应式布局适配（桌面、平板、手机）
  - 微信 Access Token 获取和自动刷新
  - Token 过期处理和重试机制
  - 图片素材上传（格式验证、大小限制、压缩）
  - 草稿箱文章创建和更新
  - 文章发布流程（预检查、发布、回调）
  - 微信 API 错误码处理（40001、40014、45009 等）
  - 发布预览渲染（HTML 转微信格式）
  - 封面图片裁剪和尺寸适配
  - 摘要和作者信息编辑
  - 发布进度提示和状态更新
  - 发布记录查询和重试功能
  - 网络异常和超时处理

## 验收标准
- [ ] 核心基础设施完整，Riverpod、Drift、Dio 配置正确
- [ ] 热点扫描支持至少 3 个平台（微博、知乎、百度），缓存策略有效
- [ ] AI 写作支持至少 2 个提供商（OpenAI、Claude），流式输出正常
- [ ] 文章管理支持富文本和 Markdown 编辑，自动保存草稿功能正常
- [ ] 数据仪表盘展示核心指标，图表渲染正确
- [ ] 微信发布功能正常，支持素材上传、草稿创建和一键发布
- [ ] 所有单元测试通过，代码覆盖率 ≥90%
- [ ] Widget 测试覆盖核心 UI 组件，交互逻辑正确
- [ ] 集成测试覆盖关键用户流程（热点→AI 写作→文章编辑→发布）
- [ ] 应用在 Windows/macOS/Linux/Android 平台正常运行
- [ ] 性能指标达标：列表滚动 60fps，编辑器输入延迟 <50ms，启动时间 <3s

## 技术要点

### 架构设计
- **Clean Architecture**: 严格三层分离（domain/data/presentation），依赖倒置原则
- **Feature-First**: 按功能模块组织代码，每个 feature 独立完整
- **依赖注入**: 使用 Riverpod 2.x Provider 作为 DI 容器，支持依赖覆盖和测试

### 状态管理
- **Riverpod 2.x**: 使用 StateNotifierProvider 管理复杂状态，FutureProvider 处理异步数据
- **流式状态**: AI 写作使用 StreamProvider 处理流式输出
- **状态持久化**: 关键状态使用 StateNotifierProvider + Drift 持久化

### 数据库设计
- **Drift (SQLite)**: 类型安全的查询构建器，支持 DAO 模式
- **表设计**: articles、hot_topics、ai_generations、wechat_publish_records、analytics_data
- **索引优化**: 为搜索字段（title、content）和时间字段（created_at）建立索引
- **迁移策略**: 使用 Drift 迁移脚本管理数据库版本升级

### 网络层设计
- **Dio 5.x**: 统一网络请求封装，支持拦截器链
- **拦截器**: 日志拦截器、认证拦截器、错误处理拦截器、重试拦截器
- **超时配置**: 连接超时 15s，接收超时 30s，流式请求超时 120s
- **重试策略**: 指数退避算法，最多重试 3 次，仅对幂等请求重试

### 安全性
- **敏感信息存储**: 使用 flutter_secure_storage 存储 API 密钥、微信凭证
- **加密传输**: 所有 API 请求使用 HTTPS
- **Token 管理**: Access Token 自动刷新，过期前 5 分钟主动更新
- **输入验证**: 前端验证 + 后端验证，防止 XSS 和 SQL 注入

### 性能优化
- **列表优化**: 使用 ListView.builder 和虚拟滚动，支持分页加载
- **图片优化**: 使用 cached_network_image，本地缓存和内存缓存
- **数据库优化**: 批量插入使用事务，查询使用索引，避免 N+1 查询
- **编辑器优化**: 使用防抖（debounce）减少自动保存频率，增量更新 DOM

### 测试策略
- **单元测试**: 覆盖所有 domain 和 data 层逻辑，使用 mockito 模拟依赖
- **Widget 测试**: 覆盖核心 UI 组件，使用 ProviderScope 模拟状态
- **集成测试**: 使用 integration_test 包，覆盖关键用户流程
- **覆盖率要求**: 每个 task 独立覆盖率 ≥90%，整体覆盖率 ≥90%

### 错误处理
- **统一错误模型**: NetworkError、DatabaseError、ValidationError、BusinessError
- **错误传播**: 使用 Either<Failure, Success> 模式传播错误
- **用户提示**: 友好的错误提示，提供重试和反馈入口
- **日志记录**: 使用 logger 包记录错误堆栈，支持远程日志上报

### 微信 API 集成
- **API 版本**: 微信公众平台 API v2
- **认证方式**: AppID + AppSecret 获取 Access Token
- **Token 刷新**: 有效期 7200s，提前 300s 刷新
- **错误码处理**: 40001（Token 过期）、40014（AppID 无效）、45009（接口调用超限）
- **素材上传**: 支持图片（jpg/png，<2MB）、视频（mp4，<10MB）

### AI 提供商集成
- **OpenAI**: GPT-4/GPT-3.5-turbo，支持流式输出
- **Claude**: Claude 3 Opus/Sonnet，支持长文本处理
- **Gemini**: Gemini Pro，支持多模态输入
- **统一接口**: AIProvider 抽象层，支持动态切换提供商
- **提示词模板**: 预定义模板（标题生成、内容扩写、改写优化），支持变量替换

### 国际化与本地化
- **当前版本**: 仅支持中文（zh_CN）
- **预留支持**: 使用 flutter_localizations，预留 i18n 文件结构
- **日期格式**: 使用 intl 包格式化日期和数字

### 平台适配
- **桌面平台**: Windows/macOS/Linux，支持窗口大小调整和键盘快捷键
- **移动平台**: Android，支持触摸手势和底部导航
- **响应式布局**: 使用 LayoutBuilder 和 MediaQuery 适配不同屏幕尺寸
