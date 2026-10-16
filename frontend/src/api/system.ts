import * as SystemAPI from '../../wailsjs/go/wails/SystemAPI'

// 读取系统设置
export const loadSettings = SystemAPI.Settings

// 读取主机资源状态
export const loadSystemMetrics = SystemAPI.Metrics

// 保存系统设置
export const saveSettings = SystemAPI.SaveSettings

export const loadApplicationStartup = SystemAPI.StartupEnabled
export const setApplicationStartup = SystemAPI.SetStartupEnabled

// 打开系统文件夹选择对话框
export const pickFolder = SystemAPI.PickFolder

export const pickFile = SystemAPI.PickFile

// 使用系统默认浏览器打开链接
export const openURL = SystemAPI.OpenURL

// 在资源管理器中打开目录
export const openFolder = SystemAPI.OpenFolder

// 打开 Nginx 主配置文件
export const openNginxConfig = SystemAPI.OpenNginxConfig

// 获取应用数据根目录路径
export const loadFolder = SystemAPI.DataFolder
