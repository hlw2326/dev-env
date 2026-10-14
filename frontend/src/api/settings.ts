import * as SettingAPI from '../../wailsjs/go/wails/SettingAPI'
import type { AppSettings } from '../types'

export async function getSettings(): Promise<AppSettings> {
  const s = await SettingAPI.Get()
  return {
    language: s.language,
    theme: s.theme as AppSettings['theme'],
    runtimeRootPath: s.runtimeRootPath,
    defaultProjectPath: s.defaultProjectPath,
    autoStartServices: s.autoStartServices,
  }
}

export async function updateSettings(settings: AppSettings): Promise<AppSettings> {
  const s = await SettingAPI.Update(settings)
  return {
    language: s.language,
    theme: s.theme as AppSettings['theme'],
    runtimeRootPath: s.runtimeRootPath,
    defaultProjectPath: s.defaultProjectPath,
    autoStartServices: s.autoStartServices,
  }
}

export async function resetSettings(): Promise<AppSettings> {
  const s = await SettingAPI.Reset()
  return {
    language: s.language,
    theme: s.theme as AppSettings['theme'],
    runtimeRootPath: s.runtimeRootPath,
    defaultProjectPath: s.defaultProjectPath,
    autoStartServices: s.autoStartServices,
  }
}

export async function pickDir(title = '选择目录'): Promise<string> {
  return await SettingAPI.SelectDir(title)
}

export async function pickFile(title = '选择文件'): Promise<string> {
  return await SettingAPI.SelectFile(title)
}
