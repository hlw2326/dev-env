# DevEnv - 本地开发环境管理平台

基于 **Go + Wails + Vue 3 + TypeScript + SQLite** 的 Windows 本地开发环境与项目管理工具。

---

## 核心特性

1. **Nginx 统一本地网关**：原生支持 Nginx 作为高性能项目网关，动态生成 `vhosts/*.conf` 虚拟主机配置并热重载：
   - **静态项目**：直接托管目录并配置默认索引与回退规则。
   - **PHP 项目**：原生对接 `php-cgi.exe` FastCGI 常驻进程池，高效处理请求并支持单入口框架伪静态。
   - **Node 项目**：由进程管理器拉起项目进程并注入 `PORT` 与 Node 运行时环境变量，Nginx 自动反向代理并支持 WebSocket 升级。
2. **软件商店与真实进程生命周期**：
   - 支持在线获取并安装 **Nginx**、**Apache**、**PHP (NTS x64)**、**Node.js (x64)**、**MySQL**、**Redis**。
   - **Windows Job Object 防孤儿进程**：通过内核作业对象绑定所有子进程生命周期，随主程序退出彻底清理。
   - **端口进程自动认领与防死锁**：应用重启时自动探测并认领已在后台运行的自身服务（避免误判未启动），再次启动时自动清理同软件遗留孤儿进程防止端口占用死锁。
   - **系统开机自启**：写入当前用户注册表 `Run`，支持无 UAC 弹窗并带 `--minimized` 静默启动。
3. **SQLite 严格迁移与防漂移备份**：
   - 启动时对已执行迁移做 SHA-256 漂移校验，迁移前通过 `VACUUM INTO` 自动备份（保留最近 5 份）。

---

## 目录约定

- **应用数据目录**：`%LOCALAPPDATA%\DevEnv\`
  - `data\devenv.db`：SQLite 数据库
  - `data\backups\`：迁移与重置自动备份
  - `software\<kind>\<version>\`：已安装软件目录
  - `mysql-data\<version>\`：MySQL 数据目录
  - `logs\projects\`、`logs\services\`：项目与服务运行日志
- **默认项目父目录**：`%USERPROFILE%\DevEnv\www\`（可在「设置」页修改）

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
