import * as SoftwareAPI from '../../wailsjs/go/wails/SoftwareAPI'

// 获取软件商店列表
export const loadMarket = SoftwareAPI.Market

// 获取已安装软件列表
export const listSoftware = SoftwareAPI.List

// 获取软件可安装版本列表
export const listVersions = SoftwareAPI.Versions

// 用户主动查询本机环境；空类型表示全部，程序路径用于手动指定入口
export const scanEnvironments = SoftwareAPI.ScanEnvironments

// 读取软件常用配置项
export const loadConfig = SoftwareAPI.Config

// 保存软件常用配置项
export const saveConfig = SoftwareAPI.SaveConfig

// 读取软件配置文件全文
export const readConfig = SoftwareAPI.ReadConfig

// 写入软件配置文件全文
export const writeConfig = SoftwareAPI.WriteConfig

// 获取 PHP 扩展列表
export const listExtensions = SoftwareAPI.Extensions

// 切换 PHP 扩展启用状态
export const setExtension = SoftwareAPI.SetExtension

// 打开软件所在目录
export const openSoftware = SoftwareAPI.OpenFolder

// 安装指定版本软件
export const installSoftware = SoftwareAPI.Install

// 取消正在进行的安装任务
export const cancelInstall = SoftwareAPI.Cancel

// 卸载已安装软件版本
export const removeSoftware = SoftwareAPI.Remove

// 设为该类型默认软件版本
export const setMain = SoftwareAPI.SetMain

// 启动守护进程服务
export const startSoftware = SoftwareAPI.Start

// 停止守护进程服务
export const stopSoftware = SoftwareAPI.Stop

// 异步重启守护进程服务
export const restartSoftware = SoftwareAPI.Restart

// 设置守护进程随应用启动
export const setAutostart = SoftwareAPI.SetAutostart

// 读取软件运行日志
export const listLogs = SoftwareAPI.Logs

// 读取 Redis 数据库列表
export const listRedisDatabases = SoftwareAPI.RedisDatabases

// 分页读取 Redis 键列表
export const listRedisKeys = SoftwareAPI.RedisKeys

// 读取 Redis 键内容
export const loadRedisValue = SoftwareAPI.RedisValue
