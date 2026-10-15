# DevEnv

DevEnv 是 Windows 本地开发环境与项目管理平台，使用 Go、Wails、Vue 3、TypeScript、PrimeVue 和 SQLite 构建。它将运行环境、Web 网关、项目启动、Git、手动发布、服务状态和系统资源监控集中到一个桌面应用中。

## 功能概览

### 项目管理

- 支持静态、PHP、Node.js、Python、Go、Rust、Java、Electron 和小程序项目。
- Web 项目通过本地 Nginx 网关访问；Electron 和小程序用于本地调试，不经过 Nginx。
- PHP 项目支持选择 `.env`、`.env.dev`、`.env.prod` 等项目根目录内的环境文件，并将变量注入当前 PHP 请求。
- 项目详情支持运行目录、运行环境、启动命令、自动启动、访问日志、Nginx 高级配置和项目插件。
- 根据项目文件自动提供 Composer、Node 包管理、Python 依赖、Go 模块、Cargo 和 Java 构建等插件。

### 服务管理

- 软件商城支持 PHP、Node.js、Go、Python、Composer、OpenJDK、.NET、Deno、Bun、Nginx、Apache、Caddy、MySQL、MariaDB、PostgreSQL、MongoDB、Redis、Memcached、SQLite、Gradle 和 Maven。
- Nginx、数据库和缓存服务可以独立启动、停止、查看日志和修改配置。
- Windows 下的数据库、缓存等守护服务使用独立的 Windows 服务运行，不依赖 DevEnv 进程；电脑重启后可继续运行，重新打开 DevEnv 会自动读取真实进程和端口状态。
- 支持服务开机自启、进程状态识别、端口检查、启动超时诊断和服务日志。

### 全局命令

- 安装、切换默认版本、卸载和启动 DevEnv 时，同步默认版本的命令入口到 `%LOCALAPPDATA%\dev-env\bin`，并写入当前用户 PATH。
- 新打开的 CMD 可以在任意目录执行 `composer` 等已绑定的命令；Composer 使用 DevEnv 默认 PHP。已有终端需要重新打开，IDE 内的终端需要重启 IDE 后再打开。
- 系统 PATH 中已有的同名工具仍可能优先匹配，可用 `where composer`、`where php` 检查实际入口。
- 商店的“本机环境”打开独立抽屉；点击“扫描已有环境”查询全部或指定类型，读取 Windows 已保存的 PATH、适用的安装变量和原生 Windows 服务登记。选择软件类型后，也可通过“选择程序”识别未配置 PATH 的便携安装。
- 商店加载、刷新、打开安装窗口和安装完成时均不自动扫描；安装窗口提供“查找本机已有版本”按钮。确认结果才展示“本机已有”徽章；安装、卸载或切换默认版本后将结果标为待更新，由用户手动重查。
- 全部查询通过一次后端调用完成，并发、输出量和时长均受限制。单项错误保留在抽屉详情；接口不可用等公共错误只提示一次，不在每张卡片显示“检测失败”。开发调试时，新增 Go 接口后需要使用包含该接口的最新程序，单独热更新前端不能更新正在运行的旧后端。
- 外部环境只读识别，不计入“DevEnv 已安装”，不自动导入、改 PATH、覆盖、卸载或启停。相同版本标记“系统同版本入口”，可以另行安装到 DevEnv；项目仍使用明确绑定的托管运行环境。
- 区分数据库客户端与服务端、Java 运行环境与 JDK 编译器、.NET 运行时与 SDK；原生服务状态来自 Windows，不用端口或客户端命令伪造。Windows 应用执行别名不作为安装成功，检测失败与未发现分别显示。
- 不全盘扫描；Docker / WSL 内部环境或 NSSM 等第三方服务包装器不在自动查询范围内，便携环境可手动选择程序。检测不到不等于未安装；旧终端、当前目录同名程序和版本管理器临时环境可能产生不同结果。已有服务可能占用端口，需确认后再启动同类服务。
- 开发环境中，如果已安装软件但命令入口缺失，在本仓库根目录执行 `go run ./cmd/sync-commands` 修复。该命令只读软件记录并刷新命令与用户 PATH，不启动服务或重新安装软件。

### 本地网关与项目运行

- Nginx 监听本机回环地址，项目域名和端口由项目配置生成。
- 静态项目由 Nginx 直接托管；PHP 项目使用 `php-cgi`；Node.js、Python、Go、Rust、Java 项目由进程管理器启动后反向代理。
- 项目进程使用 Job Object 管理，DevEnv 退出时会清理由 DevEnv 启动的项目进程；独立服务不受此生命周期影响。
- Nginx 配置会在检查前创建必要的日志、虚拟主机和临时目录；高级配置只允许写入当前项目的 `server` 区块。

### Git 集成

- 项目列表和详情实时显示 Git 是否可用、仓库状态、分支、未提交变更和远端状态。
- 支持用户主动初始化仓库、提交、管理多个远端、获取远端状态和推送当前分支。
- 提交时可选择自动递增版本号，支持 `package.json`、`composer.json`、`pyproject.toml`、`Cargo.toml`、Maven/Gradle 配置和 `VERSION` 文件。
- Git 不会自动初始化、自动提交或自动推送。

详细规则见 [Git 集成说明](doc/Git集成.md)。

### 手动发布

- PHP 和 Node.js 项目支持通过 SSH/SFTP 手动同步到服务器。
- 服务器认证支持连接密码或 SSH 私钥，主机指纹必须由用户确认。
- 小程序支持通过微信开发者工具 CLI 上传编译后的微信代码，可绑定多个 AppID。
- 发布必须经过保存配置、环境检查、生成预览和用户确认，不会因项目启动、Git 提交或文件变化自动发布。
- 发布任务支持进度、取消、历史记录和真实错误反馈。

详细规则见 [同步说明](doc/同步.md)。

### 监控与托盘

- 环境概览显示主机 CPU、内存和项目访问流量。
- 服务管理显示服务 CPU、内存、读写 I/O 和运行时长。
- 项目显示请求数、响应流量、错误数、平均响应时间和最后访问时间。
- 关闭窗口或按 Alt+F4 会隐藏到系统托盘，服务和项目不会因此被主动停止。
- 托盘菜单支持显示窗口、隐藏窗口和退出 DevEnv；只有选择“退出 DevEnv”才会执行完整关闭流程。

## 卸载

- 卸载 DevEnv 时会停止并删除 DevEnv 注册的 Windows 服务，结束数据目录内运行的服务进程，并移除 `%LOCALAPPDATA%\dev-env\bin` 的用户 PATH 项。
- 卸载器会询问是否删除 `%LOCALAPPDATA%\dev-env` 的全部数据；选择“是”会删除已安装软件、数据库、配置、日志和缓存，且无法恢复；选择“否”仅卸载程序，保留数据供下次安装继续使用。
- 静默卸载默认保留数据目录。

## 数据目录

默认数据根目录为 `%LOCALAPPDATA%\dev-env`，项目默认父目录为 `%USERPROFILE%\DevEnv\www`，均可在应用设置中修改。

```text
%LOCALAPPDATA%\dev-env\
├─ data\devenv.db                 SQLite 数据库
├─ data\backups\                 迁移和重置前的数据库备份
├─ data\nginx\nginx.conf         当前生成的 Nginx 主配置
├─ data\nginx\vhosts\           项目虚拟主机配置
├─ data\services\                Windows 服务定义文件
├─ software\<kind>\<version>\    已安装软件
├─ dbdata\<kind>-<version>\       数据库服务数据目录
├─ cache\                           下载缓存
└─ logs\                            项目、Nginx 和软件服务日志
```

软件卸载不会自动删除数据库服务的数据目录。请在确认不再需要数据后，再手动清理对应的 `dbdata` 目录。

## 开发环境

建议使用：

- Windows x64
- Go 1.26+
- Node.js 20.19+ 或 22.12+
- npm
- Wails CLI v2
- WebView2 Runtime

安装依赖后，在项目根目录执行：

```powershell
# 后端测试
go test ./...

# 后端构建
go build ./...

# 前端类型检查与生产构建
Set-Location frontend
npx vue-tsc --noEmit
npm run build
Set-Location ..

# Wails 开发模式
wails dev

# Wails 生产构建
wails build
```

修改 Wails 导出的 API 后，先执行 `wails generate module`，再运行前端类型检查和构建；不要直接修改 `frontend/wailsjs` 下的生成文件。

## 数据库结构漂移

开发阶段如果启动提示数据库结构与迁移脚本不一致，可以使用一次性重置流程：

```powershell
$env:DEVENV_DB_RESET = '1'
wails dev
Remove-Item Env:DEVENV_DB_RESET
```

重置前程序会在 `data\backups` 创建备份，并重新执行当前迁移。该操作只用于开发环境，请勿在需要保留数据的环境中使用。

## 项目规则

开发约定集中在 [doc/rules](doc/rules/README.md)，其中《总则》优先级最高。数据库、项目、软件、前端、后端、样式、Git 和同步功能分别有对应说明。
