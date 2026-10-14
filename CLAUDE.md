# DevEnv 开发与 AI 编程约束规范

本项目是一个基于 **Go + Wails + Vue 3 + TypeScript + SQLite** 的现代本地开发环境管理工具（支持 Windows 10/11 64 位）。

## 1. 核心架构约束

1. **分层架构**：
   - `main.go` / `app.go`：桌面应用入口与生命周期管理。
   - `internal/transport/wails/`：Wails 绑定方法、DTO 转换、事件发布。禁止在此层直接写复杂 SQL、进程操作或文件安装。
   - `internal/app/`：应用服务层（编排业务用例，如项目管理、PHP版本注册、服务启停、任务调度）。
   - `internal/domain/`：领域业务层（模型定义、业务规则、领域错误）。禁止依赖 Wails、Vue 或具体数据库驱动。
   - `internal/ports/`：端口契约（存储库、进程管理器、文件系统、下载器、端口检查器、事件发布器等接口）。
   - `internal/infra/`：基础设施实现（SQLite 与迁移、Windows 进程管理、文件系统、网络端口检查等）。
   - `frontend/`：Vue 3 + Vite + TypeScript 现代化前端。统一通过 `frontend/src/api/` 封装后端调用，禁止在页面组件中散落底层绑定调用。
2. **依赖方向**：
   - 基础设施层不能反向依赖应用服务层。
   - 业务领域层禁止依赖外部实现和 UI 框架。
   - 杜绝循环依赖与大杂烩模块（禁止创建无明确业务职责的 common/utils）。
3. **数据隔离与安全**：
   - DevEnv 内部管理数据库采用 `%LOCALAPPDATA%/DevEnv/data/devenv.db`（SQLite），与用户项目使用的 MySQL/PostgreSQL 数据目录严格隔离。
   - 删除项目记录仅删除元数据，严禁静默递归删除用户项目物理目录。
   - 卸载环境版本前必须检查项目和服务引用。

## 2. 数据库与表结构规范（用户最新定制）

1. **主键与标识符**：统一采用**数字自增 ID**（Go 中为 `int64`，TypeScript 中为 `number`，SQLite 中为 `INTEGER PRIMARY KEY AUTOINCREMENT`），外键均为 `INTEGER`。
2. **统一表名前缀**：所有数据库表必须统一加上 `system_` 前缀：
   - `system_projects`
   - `system_runtime_versions`
   - `system_service_instances`
   - `system_virtual_hosts`
   - `system_tasks`
   - `system_app_settings`
   - `system_schema_migrations`
3. **时间格式**：统一存储和传输 ISO 8601 / RFC 3339 格式（例如 `2026-10-14T10:30:00Z`）。
4. **迁移规范**：所有数据库表结构变更必须编写编号迁移脚本（`internal/infra/sqlite/migrations/`），支持事务执行与版本记录。
5. **SQLite 选项**：开启外键检查 (`PRAGMA foreign_keys = ON;`)、WAL 模式 (`PRAGMA journal_mode = WAL;`) 与合理 busy_timeout。

## 3. 错误与事件规范

1. **错误代码**：错误必须映射为标准化错误代码（如 `INVALID_ARGUMENT`, `PORT_ALREADY_IN_USE`, `PHP_VERSION_IN_USE`, `PATH_INVALID` 等）。
2. **事件通知**：长时间任务及服务状态变更统一使用 `task:progress`, `task:completed`, `service:status-changed` 等标准事件。
