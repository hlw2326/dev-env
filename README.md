# DevEnv - 本地开发环境管理平台

基于 **Go + Wails + Vue 3 + TypeScript + SQLite** 的现代化 Windows 本地开发环境管理工具。

---

## 🛠️ 前置环境

在开发前，请确保系统已配置以下工具（当前系统均已就绪）：

1. **Go** (1.25+ 或更高版本，如 Go 1.27)
2. **Node.js** (v20+，如 v20.13.1) 与 **npm** (10+)
3. **Wails CLI** (`v2.16.0`，已安装至 `GOPATH/bin` 即 `%USERPROFILE%\go\bin`)

> 💡 如果终端提示找不到 `wails` 或 `go` 命令，请确保将 `C:\Program Files\Go\bin` 与 `C:\Users\Administrator\go\bin` 添加至系统的环境变量 `Path` 中（或重新打开终端生效）。

---

## 🚀 启动与开发流程

### 1. 本地热重载开发模式（推荐）

在项目根目录下执行：

```powershell
wails dev
```

该命令会自动：
- 启动前端 Vite 开发服务器（支持 Vue 组件热重载 HMR）
- 启动 Go 后端并开启热编译监听（修改 Go 代码会自动重新编译）
- 弹出原生 Windows 桌面应用窗口进行实时调试
- 也可在浏览器中访问：`http://localhost:34115`

### 2. 仅调试前端界面（可选）

如仅需调整前端样式或交互逻辑：

```powershell
cd frontend
npm run dev
```

浏览器打开 `http://localhost:5173`。

### 3. 运行自动化测试

- **执行后端与 SQLite 仓储测试：**
  ```powershell
  go test -v ./...
  ```
- **执行前端类型检查与打包：**
  ```powershell
  cd frontend
  npm run build
  ```

---

## 📦 生产打包与构建

在项目根目录下执行：

```powershell
wails build
```

打包完成后，生成的可执行文件位于：
```text
build/bin/devenv.exe
```
双击 `build/bin/devenv.exe` 即可独立离线运行。
