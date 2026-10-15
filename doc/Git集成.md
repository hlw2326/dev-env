# Git 集成

产品目标:项目 Git 提交、创建标签、推送到多个远程。

不做:分支操作、拉取、合并、变基、回退、冲突解决、子模块。

## 一、配置

`system_project` 表新增 `git TEXT` 字段,JSON:

```json
{
  "enabled": true,
  "remotes": [
    {
      "name": "github",
      "url": "https://github.com/user/repo.git",
      "username": "user",
      "token": "encrypted",
      "main": true
    }
  ],
  "auto": false,
  "template": "update: ${files} files"
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `enabled` | bool | 是 | 启用 Git |
| `remotes` | array | 否 | 远程仓库组 |
| `remotes[].name` | string | 是 | 远程名称,唯一 |
| `remotes[].url` | string | 是 | HTTPS URL |
| `remotes[].username` | string | 否 | 用户名 |
| `remotes[].token` | string | 否 | 访问令牌,加密存储 |
| `remotes[].main` | bool | 是 | 默认推送目标,只能有一个 |
| `auto` | bool | 否 | 保存时自动提交,默认 `false` |
| `template` | string | 否 | 提交消息模板,支持 `${files}` `${date}` `${time}` |

- 只在项目详情"Git"页签编辑。
- `token` AES-256-GCM 加密,复用 `sync.key`。
- 不支持 SSH,只支持 HTTPS。
- `remotes` 空时禁用推送,只能本地提交。

## 二、操作

### 初始化

项目详情"Git"页签【初始化仓库】,执行 `git init`。

### 提交

【提交更改】:

1. 检查未初始化返回 `GIT_NOT_INITIALIZED`。
2. 发 `git:commit-progress`,`stage=staging`。
3. 执行 `git add .`。
4. 发 `git:commit-progress`,`stage=committing`。
5. 执行 `git commit -m "<message>"`。
6. 读提交哈希:`git rev-parse HEAD`。
7. 发 `git:operation-completed`,返回哈希。

### 创建标签

【创建标签】:

1. 输入标签名,校验格式 `^v?\d+\.\d+\.\d+$`。
2. 输入标签消息(可选)。
3. 执行 `git tag -a <tag> -m "<message>"`。
4. 发 `git:operation-completed`。

### 推送

【推送】:

1. 读 `git`,检查 `remotes`。
2. 找目标 `remote`,解密 `token`。
3. 构造认证 URL:`https://<username>:<token>@github.com/user/repo.git`。
4. 发 `git:push-progress`。
5. 执行 `git push <url> <refspec>`。
6. 发 `git:operation-completed`。

推送模式:

- 推送到主远程:推送当前分支
- 推送到所有远程:遍历 `remotes`,逐个推送
- 推送标签到主远程:推送指定标签
- 推送标签到所有远程:遍历推送标签

### 自动提交

`auto=true` 时监听 `root` 文件变化:

- 防抖 2 秒。
- 用 `template` 生成消息,替换 `${files}` `${date}` `${time}`。
- 后台提交,失败 toast。

## 三、GitAPI

| 方法 | 入参 | 返回 |
|------|------|------|
| `GetStatus(projectID)` | | `StatusDTO` |
| `Init(projectID)` | | `Result` |
| `Commit(projectID, message)` | | `Result`,返回哈希 |
| `CreateTag(projectID, tag, message)` | | `Result` |
| `Push(projectID, remote, ref)` | `ref` 为 `main` 或 `v1.0.0` | `Result` |
| `PushAll(projectID, ref)` | 推送到所有远程 | `Result`,每个远程的结果 |
| `SaveConfig(projectID, config)` | `{enabled, remotes[], auto, template}` | `Result` |
| `GetConfig(projectID)` | | `ConfigDTO`,`token` 返回掩码 |
| `TestRemote(projectID, name)` | | `Result` |

`StatusDTO`:

```json
{
  "initialized": true,
  "branch": "main",
  "uncommitted": 12,
  "untracked": 3,
  "modified": 8,
  "deleted": 1,
  "last": {
    "hash": "a3f5c2d",
    "message": "update: 5 files",
    "time": "2026-10-15T10:30:00Z"
  },
  "remotes": ["github", "gitlab"],
  "main": "github"
}
```

`ConfigDTO`:

```json
{
  "enabled": true,
  "remotes": [
    {
      "name": "github",
      "url": "https://github.com/user/repo.git",
      "username": "user",
      "main": true
    }
  ],
  "auto": false,
  "template": "update: ${files} files"
}
```

## 四、事件

`git:commit-progress`:

```json
{
  "project": 1,
  "stage": "staging",
  "message": "正在暂存..."
}
```

`stage`:`staging` | `committing` | `done`

`git:push-progress`:

```json
{
  "project": 1,
  "remote": "github",
  "stage": "pushing",
  "message": "正在推送..."
}
```

`git:operation-completed`:

```json
{
  "project": 1,
  "operation": "commit",
  "ok": true,
  "result": "a3f5c2d",
  "error": ""
}
```

`operation`:`init` | `commit` | `tag` | `push`

## 五、后端

分层:

```
internal/
├── domain/git/       配置模型,Token 加解密
├── ports/git.go      GitClient 接口
├── app/git.go        GitService,编排操作
├── infra/git/
│   ├── client.go    封装 git 命令
│   └── auth.go      HTTPS 认证
└── transport/wails/git.go  GitAPI 绑定
```

依赖:

- 调用系统 `git.exe`,不用 `go-git`。
- 检查:`git --version`,找不到返回 `GIT_NOT_INSTALLED`。

`GitClient` 接口:

```go
type GitClient interface {
    Init(dir string) error
    Status(dir string) (*Status, error)
    Commit(dir, message string) (string, error)
    Tag(dir, tag, message string) error
    Push(dir, remote, ref string, auth *Auth) error
    LastCommit(dir string) (*Commit, error)
}
```

Token 加密:

- 复用 `sync.key`。
- `SaveConfig` 收明文,加密存库。
- `Push` 解密,拼 URL:`https://<username>:<token>@<host>/<path>`。

命令注入防护:

- 提交消息、标签名、标签消息做检查,拒绝 `&&` `||` `;` `|`。
- 认证 URL 不写入 `.git/config`,只在命令行传递。

错误码(`shared/errors`):

- `GIT_NOT_INSTALLED`
- `GIT_NOT_INITIALIZED`
- `GIT_COMMIT_FAILED`
- `GIT_TAG_FAILED`
- `GIT_PUSH_FAILED`
- `GIT_AUTH_FAILED`
- `GIT_REMOTE_NOT_FOUND`

## 六、数据库

迁移,改 `000001_init.sql`:

```sql
ALTER TABLE system_project ADD COLUMN git TEXT DEFAULT '';
```

改表同步全链路(见 [数据库.md](rules/数据库.md) 第六节)。

## 七、安全

1. Token 加密,不在日志、事件、错误里输出。
2. 推送时认证 URL 只在命令行传递,不写 `.git/config`。
3. 检查 `.git` 目录权限,拒绝超出 `root` 的 `.git`。
4. 提交消息、标签名注入检查。
5. 推送超时 5 分钟。

## 八、限制

1. 只支持 HTTPS,不支持 SSH。
2. 不支持分支操作,默认在当前分支工作。
3. 不支持交互式操作。
4. 推送冲突返回错误,不自动解决。
5. 远程仓库必须已存在。

## 九、测试

- `domain/git`:Token 加解密、模板变量替换。
- `infra/git`:临时目录 + 本地 Git 服务器(gitea Docker):初始化、提交、标签、HTTPS 推送、认证失败、推送冲突。
- `app/git`:mock `GitClient`,测试编排、事件。
- 手动:真实 GitHub/GitLab,大仓库,断网,自动提交。

## 自查

```bash
go build ./... && go test ./...
cd frontend && npx vue-tsc --noEmit
rg 'token.*string' internal/infra/sqlite --glob '*.go' | rg -v 'encrypt'
rg '_ = .*(Init|Commit|Push|Tag)' internal/app internal/infra/git
rg 'exec.Command.*\+' internal/infra/git
```
