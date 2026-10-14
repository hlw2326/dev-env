import * as AppAPI from '../../wailsjs/go/wails/AppAPI'
import type { AppSettings } from '../types'

export async function getSettings(): Promise<AppSettings> {
  const s = await AppAPI.GetSettings()
  return {
    language: s.language,
    theme: s.theme as AppSettings['theme'],
    runtimeRootPath: s.runtimeRootPath,
    defaultProjectPath: s.defaultProjectPath,
    autoStartServices: s.autoStartServices,
  }
}

export async function updateSettings(settings: AppSettings): Promise<AppSettings> {
  const s = await AppAPI.UpdateSettings(settings)
  return {
    language: s.language,
    theme: s.theme as AppSettings['theme'],
    runtimeRootPath: s.runtimeRootPath,
    defaultProjectPath: s.defaultProjectPath,
    autoStartServices: s.autoStartServices,
  }
}

export async function resetSettings(): Promise<AppSettings> {
  const s = await AppAPI.ResetSettings()
  return {
    language: s.language,
    theme: s.theme as AppSettings['theme'],
    runtimeRootPath: s.runtimeRootPath,
    defaultProjectPath: s.defaultProjectPath,
    autoStartServices: s.autoStartServices,
  }
}
