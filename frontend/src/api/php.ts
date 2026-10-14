import * as AppAPI from '../../wailsjs/go/wails/AppAPI'
import type { PHPVersion, RegisterPHPVersionInput, PHPExtension, OperationResult } from '../types'

export async function listVersions(): Promise<PHPVersion[]> {
  const items = await AppAPI.ListVersions()
  return (items || []).map((v) => ({
    id: v.id,
    version: v.version,
    architecture: v.architecture,
    installPath: v.installPath,
    status: v.status as PHPVersion['status'],
    isDefault: v.isDefault,
  }))
}

export async function registerVersion(input: RegisterPHPVersionInput): Promise<PHPVersion> {
  const v = await AppAPI.RegisterVersion({
    version: input.version,
    architecture: input.architecture,
    installPath: input.installPath,
  })
  return {
    id: v.id,
    version: v.version,
    architecture: v.architecture,
    installPath: v.installPath,
    status: v.status as PHPVersion['status'],
    isDefault: v.isDefault,
  }
}

export async function setDefaultVersion(versionId: number): Promise<OperationResult> {
  return await AppAPI.SetDefaultVersion(versionId)
}

export async function getExtensions(versionId: number): Promise<PHPExtension[]> {
  const items = await AppAPI.GetExtensions(versionId)
  return (items || []).map((e) => ({
    name: e.name,
    enabled: e.enabled,
    available: e.available,
  }))
}
