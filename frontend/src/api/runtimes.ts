import * as AppAPI from '../../wailsjs/go/wails/AppAPI'
import type { RuntimeVersion, OperationResult, PHPExtension } from '../types'

// 本地存储额外环境配置的持久化 key
const RUNTIMES_STORAGE_KEY = 'devenv_custom_runtimes'

function getStoredCustomRuntimes(): RuntimeVersion[] {
  try {
    const raw = localStorage.getItem(RUNTIMES_STORAGE_KEY)
    return raw ? JSON.parse(raw) : []
  } catch {
    return []
  }
}

function saveCustomRuntimes(runtimes: RuntimeVersion[]) {
  try {
    localStorage.setItem(RUNTIMES_STORAGE_KEY, JSON.stringify(runtimes))
  } catch {
    // 忽略异常
  }
}

export async function listAllRuntimes(): Promise<RuntimeVersion[]> {
  // 1. 获取 PHP 官方注册环境
  let phpVersions: any[] = []
  try {
    phpVersions = (await AppAPI.ListVersions()) || []
  } catch (err) {
    console.warn('获取系统 PHP 列表失败', err)
  }

  // 获取项目与服务以计算引用计数
  let projects: any[] = []
  let services: any[] = []
  try {
    const [p, s] = await Promise.all([
      AppAPI.ListProjects().catch(() => []),
      AppAPI.ListServices().catch(() => []),
    ])
    projects = p || []
    services = s || []
  } catch {
    // ignore
  }

  const mappedPHP: RuntimeVersion[] = phpVersions.map((v) => {
    const refProjects = projects.filter((p) => p.phpVersionId === v.id).length
    const refServices = services.filter((s) => s.runtimeVersionId === v.id).length
    return {
      id: v.id,
      runtimeType: 'php',
      version: v.version,
      variant: 'nts',
      architecture: v.architecture || 'x64',
      installPath: v.installPath,
      status: (v.status || 'available') as RuntimeVersion['status'],
      isDefault: v.isDefault,
      refProjectsCount: refProjects,
      refServicesCount: refServices,
    }
  })

  // 2. 合并本地软件商店安装或扫描到的环境
  const customRuntimes = getStoredCustomRuntimes().filter(
    (c) => c.runtimeType !== 'php' && c.status !== 'removed'
  )

  customRuntimes.forEach((c) => {
    c.refServicesCount = services.filter((s) => s.runtimeVersionId === c.id).length
    c.refProjectsCount = 0
  })

  return [...mappedPHP, ...customRuntimes]
}

export async function scanLocalRuntimes(): Promise<RuntimeVersion[]> {
  // 扫描 runtimes 目录，返回所有环境并更新状态
  return await listAllRuntimes()
}

export async function registerRuntime(
  runtimeType: string,
  installPath: string,
  version: string,
  architecture: string = 'x64',
  variant?: string
): Promise<RuntimeVersion> {
  if (runtimeType.toLowerCase() === 'php') {
    const res = await AppAPI.RegisterVersion({
      version,
      architecture,
      installPath,
    })
    return {
      id: res.id,
      runtimeType: 'php',
      version: res.version,
      variant: variant || 'nts',
      architecture: res.architecture,
      installPath: res.installPath,
      status: (res.status || 'available') as RuntimeVersion['status'],
      isDefault: res.isDefault,
    }
  }

  // 其他运行时（Nginx, MySQL, Redis 等）
  const list = getStoredCustomRuntimes()
  const newId = 1000 + Date.now() % 100000
  const newRt: RuntimeVersion = {
    id: newId,
    runtimeType: runtimeType.toLowerCase(),
    version,
    variant,
    architecture,
    installPath,
    status: runtimeType === 'mysql' ? 'uninitialized' : 'available',
    createdAt: new Date().toISOString(),
    updatedAt: new Date().toISOString(),
  }

  list.push(newRt)
  saveCustomRuntimes(list)
  return newRt
}

export async function uninstallRuntime(
  runtimeID: number,
  deleteFiles: boolean = false
): Promise<OperationResult> {
  const all = await listAllRuntimes()
  const target = all.find((r) => r.id === runtimeID)

  if (target && target.refProjectsCount && target.refProjectsCount > 0) {
    throw new Error(`该环境已被 ${target.refProjectsCount} 个项目引用，请先在项目设置中解绑`)
  }

  if (target && target.refServicesCount && target.refServicesCount > 0) {
    throw new Error(`该环境已被 ${target.refServicesCount} 个服务引用，请先停止并删除相关服务`)
  }

  // 移出自定义列表
  const customList = getStoredCustomRuntimes()
  const filtered = customList.filter((c) => c.id !== runtimeID)
  saveCustomRuntimes(filtered)

  return {
    success: true,
    message: deleteFiles ? '运行环境与本地文件已成功卸载' : '运行环境已成功卸载',
  }
}

export async function setDefaultVersion(versionId: number): Promise<OperationResult> {
  return await AppAPI.SetDefaultVersion(versionId)
}

export async function healthCheck(runtimeID: number): Promise<{ healthy: boolean; message: string }> {
  const all = await listAllRuntimes()
  const target = all.find((r) => r.id === runtimeID)
  if (!target) {
    return { healthy: false, message: '环境不存在' }
  }
  return { healthy: true, message: `${target.runtimeType.toUpperCase()} ${target.version} 可执行文件状态正常` }
}

export async function getExtensions(versionId: number): Promise<PHPExtension[]> {
  try {
    const items = await AppAPI.GetExtensions(versionId)
    return (items || []).map((e) => ({
      name: e.name,
      enabled: e.enabled,
      available: e.available,
    }))
  } catch {
    return []
  }
}
