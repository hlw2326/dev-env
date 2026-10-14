import * as RuntimeAPI from '../../wailsjs/go/wails/RuntimeAPI'
import type { RuntimeVersion, OperationResult, PHPExtension } from '../types'

const LEGACY_RUNTIMES_STORAGE_KEY = 'devenv_custom_runtimes'
let legacyMigrated = false

async function migrateLegacyLocalStorageOnce() {
  if (legacyMigrated) return
  legacyMigrated = true
  try {
    const raw = localStorage.getItem(LEGACY_RUNTIMES_STORAGE_KEY)
    if (!raw) return
    const items: RuntimeVersion[] = JSON.parse(raw)
    for (const item of items) {
      if (!item || item.status === 'removed' || !item.runtimeType || !item.version) continue
      try {
        await RuntimeAPI.Register({
          runtimeType: item.runtimeType.toLowerCase(),
          version: item.version,
          variant: item.variant || '',
          architecture: item.architecture || 'x64',
          installPath: item.installPath || `runtimes\\${item.runtimeType.toLowerCase()}\\${item.version}-${item.architecture || 'x64'}`,
        })
      } catch {
        // Ignore duplicate or already migrated entries
      }
    }
    localStorage.removeItem(LEGACY_RUNTIMES_STORAGE_KEY)
  } catch {
    // Ignore parse errors
  }
}

export async function listAllRuntimes(): Promise<RuntimeVersion[]> {
  await migrateLegacyLocalStorageOnce()
  const runtimes = await RuntimeAPI.List()

  return runtimes.map((v) => ({
    id: v.id,
    runtimeType: (v.runtimeType || 'php').toLowerCase(),
    version: v.version,
    variant: v.variant || ((v.runtimeType || 'php').toLowerCase() === 'php' ? 'nts' : undefined),
    architecture: v.architecture,
    installPath: v.installPath,
    status: v.status as RuntimeVersion['status'],
    isDefault: v.isDefault,
    siteCount: v.siteCount ?? 0,
    serviceCount: v.serviceCount ?? 0,
  }))
}

export async function scanLocalRuntimes(): Promise<RuntimeVersion[]> {
  return await listAllRuntimes()
}

export async function registerRuntime(
  runtimeType: string,
  installPath: string,
  version: string,
  architecture = 'x64',
  variant = 'nts'
): Promise<RuntimeVersion> {
  const rtType = runtimeType.toLowerCase()
  const res = await RuntimeAPI.Register({
    runtimeType: rtType,
    version,
    variant: rtType === 'php' ? variant : '',
    architecture,
    installPath,
  })
  return {
    id: res.id,
    runtimeType: (res.runtimeType || rtType).toLowerCase(),
    version: res.version,
    variant: res.variant || (rtType === 'php' ? variant : undefined),
    architecture: res.architecture,
    installPath: res.installPath,
    status: res.status as RuntimeVersion['status'],
    isDefault: res.isDefault,
    siteCount: res.siteCount ?? 0,
    serviceCount: res.serviceCount ?? 0,
  }
}

export async function uninstallRuntime(
  runtimeId: number,
  deleteFiles = false
): Promise<OperationResult> {
  return await RuntimeAPI.Delete(runtimeId, deleteFiles)
}

export async function setDefaultVersion(versionId: number): Promise<OperationResult> {
  return await RuntimeAPI.SetDefault(versionId)
}

export async function healthCheck(runtimeId: number): Promise<{ healthy: boolean; message: string }> {
  const all = await listAllRuntimes()
  const target = all.find((r) => r.id === runtimeId)
  if (!target) {
    return { healthy: false, message: '环境不存在' }
  }
  return { healthy: true, message: `${target.runtimeType.toUpperCase()} ${target.version} 可执行文件状态正常` }
}

export async function getExtensions(versionId: number): Promise<PHPExtension[]> {
  return await RuntimeAPI.Extensions(versionId)
}

