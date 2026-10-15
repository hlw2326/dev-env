import type { wails } from '../../wailsjs/go/models'

// 通用操作结果
export type Result = wails.Result

// 配置保存结果
export type ConfigResult = wails.ConfigResult

// 项目数据传输对象
export type Project = wails.ProjectDTO

// 后端注册的项目模板
export type ProjectTemplate = wails.ProjectTemplateDTO

// 创建项目入参
export type ProjectCreateInput = wails.ProjectCreateInput

// 更新项目入参
export type ProjectUpdateInput = wails.ProjectUpdateInput

// 项目插件动作
export type PluginAction = wails.PluginActionDTO

// 项目详情插件
export type ProjectPlugin = wails.ProjectPluginDTO

// 执行项目插件操作入参
export type ProjectPluginActionInput = wails.ProjectPluginActionInput

// 项目 Git 文件状态
export type ProjectGitFile = wails.ProjectGitFileDTO

// 项目 Git 仓库状态
export type ProjectGitStatus = wails.ProjectGitStatusDTO

// 项目 Git 提交参数
export type ProjectGitCommitInput = wails.ProjectGitCommitInput

export type ProjectGitRemote = wails.ProjectGitRemoteDTO

export type DeploymentConfig = wails.DeploymentConfigDTO
export type DeploymentPlan = wails.DeploymentPlanDTO
export type DeploymentTask = wails.DeploymentTaskDTO

// 软件商店卡片项
export type MarketItem = wails.ItemDTO

// 已安装软件版本
export type Software = wails.SoftwareDTO

export type CommandEnvironment = wails.CommandEnvironmentDTO

// 进程资源指标
export type ProcessMetrics = wails.ProcessMetricsDTO

// 项目访问流量指标
export type ProjectTraffic = wails.ProjectTrafficDTO

// 主机资源指标
export type SystemMetrics = wails.SystemMetricsDTO

// 可安装软件版本
export type SoftwareVersion = wails.VersionDTO

// 软件配置表单项
export type ConfigItem = wails.ConfigItem

// 软件配置数据
export type SoftwareConfig = wails.ConfigDTO

// 软件配置保存入参
export type ConfigInput = wails.ConfigInput

// PHP 扩展状态项
export type Extension = wails.ExtensionDTO

// 安装软件入参
export type InstallInput = wails.InstallInput

// Redis 数据库摘要
export type RedisDatabase = wails.RedisDatabaseDTO

// Redis 键摘要
export type RedisKey = wails.RedisKeyDTO

// Redis 键分页结果
export type RedisKeyPage = wails.RedisKeyPageDTO

// Redis 键列表查询参数
export type RedisKeysInput = wails.RedisKeysInput

// Redis 集合项
export type RedisEntry = wails.RedisEntryDTO

// Redis 键详情
export type RedisValue = wails.RedisValueDTO

// Redis 键详情查询参数
export type RedisValueInput = wails.RedisValueInput

// 系统设置数据
export type Settings = wails.SettingsDTO

// 保存系统设置入参
export type SettingsInput = wails.SettingsInput


// 安装进度事件载荷
export interface ProgressEvent {
	// 软件类型
	kind: string
	// 软件版本号
	version: string
	// 进度百分比
	progress: number
	// 进度说明文案
	message: string
}

// 安装完成事件载荷
export interface CompletedEvent {
	// 软件类型
	kind: string
	// 软件版本号
	version: string
	// 是否安装成功
	ok: boolean
	// 失败错误信息
	error: string
}

// 服务状态变更事件载荷
export interface ServiceEvent {
	// 软件记录主键
	id: number
	// 是否正在运行
	running: boolean
	// 进程 PID
	pid: number
	// 状态或错误说明
	message: string
}

// 项目状态变更事件载荷
export interface ProjectEvent {
	// 项目记录主键
	id: number
	// 项目状态码
	state: string
	// 状态或错误说明
	message: string
}

// 项目插件状态变更事件载荷
export interface ProjectPluginEvent {
	id: number
	plugin: string
	action: string
	running: boolean
	ok: boolean
	message: string
}
