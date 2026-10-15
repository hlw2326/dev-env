import * as ProjectAPI from '../../wailsjs/go/wails/ProjectAPI'

// 获取项目列表
export const listProjects = ProjectAPI.List

// 获取单个项目详情
export const getProject = ProjectAPI.Get

// 创建项目
export const createProject = ProjectAPI.Create

// 更新项目配置
export const updateProject = ProjectAPI.Update

// 自动探测运行目录
export const findRun = ProjectAPI.FindRun

// 读取小程序项目可用的 package.json 启动命令
export const findCommands = ProjectAPI.FindCommands

// 获取后端注册的项目模板
export const listProjectTemplates = ProjectAPI.Templates

// 启动项目
export const startProject = ProjectAPI.Start

// 停止项目
export const stopProject = ProjectAPI.Stop

// 重启项目
export const restartProject = ProjectAPI.Restart

// 删除项目记录
export const deleteProject = ProjectAPI.Delete

// 读取项目日志
export const readLogs = ProjectAPI.Logs

// 获取项目可用插件
export const listProjectPlugins = ProjectAPI.Plugins

// 执行项目插件操作
export const runProjectPlugin = ProjectAPI.RunPlugin

// 打开项目生成的 Nginx 配置文件
export const openProjectNginxConfig = ProjectAPI.OpenNginxConfig

// 读取项目生成的完整 Nginx 配置
export const readProjectNginxConfig = ProjectAPI.ReadNginxConfig

// 保存项目生成的 Nginx 配置
export const writeProjectNginxConfig = ProjectAPI.WriteNginxConfig

// 获取全部项目 Git 状态
export const listProjectGitStatuses = ProjectAPI.GitStatuses

// 获取单个项目 Git 状态
export const getProjectGitStatus = ProjectAPI.GitStatus

// 初始化项目 Git 仓库
export const initializeProjectGit = ProjectAPI.InitializeGit

// 提交项目 Git 变更
export const commitProjectGit = ProjectAPI.CommitGit

export const saveProjectGitRemote = ProjectAPI.SaveGitRemote
export const removeProjectGitRemote = ProjectAPI.RemoveGitRemote
export const fetchProjectGit = ProjectAPI.FetchGit
export const pushProjectGit = ProjectAPI.PushGit
