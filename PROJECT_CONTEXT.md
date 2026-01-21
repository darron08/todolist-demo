# Project Context: Enterprise Todo List Microservice
# Vision: Scalable, Secure, and Maintainable Backend Service

## 1. 项目愿景 (Project Vision)
本项目旨在构建一个**高可扩展、高可维护的待办事项管理微服务**。
该服务将作为标准后端架构的基准实现 (Reference Implementation)，用于展示现代云原生应用的最佳实践。
核心目标是实现一个支持多租户（Multi-tenancy via Auth）、高并发读写，且具备完整可观测性的后端系统。

## 2. 工程标准 (Engineering Standards)
为了确保系统的长期可维护性，所有开发工作必须严格遵循以下原则：

* **Clean Architecture (整洁架构):** 严格分离关注点。外部框架（如 Gin）不应污染核心业务逻辑。
* **SOLID Principles:** 代码设计必须符合 SOLID 原则，特别是单一职责原则 (SRP) 和依赖倒置原则 (DIP)。
* **TDD (测试驱动开发):** 核心业务逻辑必须优先编写测试。禁止提交未经过单元测试覆盖的代码。
* **Consistency (一致性):** 命名规范、错误处理、日志格式必须在全局保持一致。
* **DevOps Ready:** 项目必须包含容器化 (Docker) 及编排 (K8s/Helm) 配置，支持 CI/CD 流水线集成。

## 3. 功能范围 (Scope of Work)

### Phase 1: Core Foundation (MVP)
* **User Management:** 用户注册、登录 (JWT Auth)。
* **Todo Lifecycle:** 创建、查询、更新、删除 (CRUD)。
* **Smart Query:** 支持基于状态 (Status)、截止日期 (Due Date) 的复杂过滤与排序。

### Phase 2: Enhanced Features
* **Metadata:** 优先级 (Priority) 管理、标签 (Tags) 系统。
* **Collaboration:** 团队协作功能，支持 RBAC (基于角色的访问控制)。
* **Real-time:** 基于 WebSocket 的实时协作与动态流。

## 4. 关键约束 (Key Constraints)
* **Performance:** API 响应时间应在 95% 情况下低于 200ms。
* **Security:** 所有敏感数据（密码）必须哈希存储，API 访问必须鉴权。
* **Scalability:** 架构设计需支持水平扩展 (Stateless API)。