# Technical Specification: Todo Service
# Version: 1.0.0
# Stack: Go / Gin / MySQL / Redis

## 1. 技术栈选型 (Technology Stack)

* **Language:** Go (Golang) 1.24.0
* **Web Framework:** Gin (github.com/gin-gonic/gin) - 追求高性能与轻量级。
* **Database:** MySQL 8.0 - 使用 InnoDB 引擎，强一致性事务支持。
* **ORM/Data Access:** GORM v2 (需严格隔离 Model 与 DTO)。
* **Cache (Optional):** Redis - 用于热点数据缓存及分布式 Session 管理。
* **Configuration:** Viper - 支持多环境配置 (Env/YAML)。
* **Documentation:** Swagger / OpenAPI 3.0 (自动生成)。
* **Containerization:** Docker & Kubernetes (Helm Charts).

## 2. 数据模型设计 (Schema Design)

> **Instruction:** 所有数据库表必须包含 `created_at`, `updated_at`, `deleted_at` (Soft Delete)。

### 2.1 Users Table
* `id`: Primary Key (BigInt/UUID)
* `username`: Unique, Index
* `password_hash`: String (BCrypt)
* `role`: Enum (Admin, User)

### 2.2 Todos Table
* `id`: Primary Key
* `user_id`: Foreign Key (Indexed)
* `title`: String (Not Null)
* `description`: Text
* `due_date`: DateTime (Indexed for range queries)
* `status`: Enum (Not Started, In Progress, Completed) - Indexed
* `priority`: Enum (Low, Medium, High)

### 2.3 Tags Table (Many-to-Many)
* `id`: Primary Key
* `name`: String (Unique)

## 3. API 接口规范 (API Specification)

遵循 RESTful 最佳实践。所有 API 必须返回统一的 JSON 结构。

### Global Response Wrapper
```json
{
  "code": 200,
  "message": "Success",
  "data": { ... },
  "request_id": "c7a9-..."
}