import * as SiteAPI from '../../wailsjs/go/wails/SiteAPI'

// 获取站点列表
export const listSites = SiteAPI.List

// 获取单个站点详情
export const getSite = SiteAPI.Get

// 创建站点
export const createSite = SiteAPI.Create

// 更新站点配置
export const updateSite = SiteAPI.Update

// 自动探测运行目录
export const findRun = SiteAPI.FindRun

// 启动站点
export const startSite = SiteAPI.Start

// 停止站点
export const stopSite = SiteAPI.Stop

// 重启站点
export const restartSite = SiteAPI.Restart

// 删除站点记录
export const deleteSite = SiteAPI.Delete

// 读取站点日志
export const readLogs = SiteAPI.Logs
