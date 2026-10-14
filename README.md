# DevEnv - 本地开发环境管理平台

基于 **Go + Wails + Vue 3 + TypeScript + SQLite** 的 Windows 本地开发环境与站点管理工具。

---

## 核心特性

1. **内置 HTTP 本地网关**：监听 `127.0.0.1:<port>`（默认 80），按 `Host` 头分发请求：
   - **静态站点**：直接提供目录文件与索引服务，内置路径穿越与敏感隐藏文件防护。
   - **PHP 站点**：直接调用已安装 PHP 的 `php-cgi.exe` 处理请求，支持单入口框架回退（Laravel `public`、ThinkPHP `public` 等自动识别）。
   - **Node 站点**：由进程管理器拉起项目进程并注入 `PORT` 与所选 Node 版本的 `PATH`，网关自动反向代理（含 WebSocket 升级）。
2. **软件商店与真实进程管理**：
   - 支持在线获取并安装 **PHP (NTS x64)**、**Node.js (x64)**、**MySQL**、**Redis**。
   - SHA-256 完整性校验、Zip-Slip 防护、真实子进程生命周期管理与结构化配置抽屉。
3. **SQLite 严格迁移与防漂移备份**：
   - 启动时对已执行迁移做 SHA-256 漂移校验，迁移前通过 `VACUUM INTO` 自动备份（保留最近 5 份）。

---

## 目录约定

- **应用数据目录**：`%LOCALAPPDATA%\DevEnv\`
  - `data\devenv.db`：SQLite 数据库
  - `data\backups\`：迁移与重置自动备份
  - `software\<kind>\<version>\`：已安装软件目录
  - `mysql-data\<version>\`：MySQL 数据目录
  - `logs\sites\`、`logs\services\`：站点与服务运行日志
- **默认站点父目录**：`%USERPROFILE%\DevEnv\www\`（可在「设置」页修改）

---

## 开发与构建

### 1. 运行后端测试

```powershell
go test ./...
```

### 2. 前端类型检查与构建

```powershell
cd frontend
npm run build
```

### 3. 本地开发与打包

```powershell
# 开发模式
wails dev

# 生产构建
wails build
```
