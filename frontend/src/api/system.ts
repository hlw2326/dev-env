import * as SystemAPI from '../../wailsjs/go/wails/SystemAPI'
import type { Settings, SaveSettingsInput, Result } from '@/types'

export async function loadSettings(): Promise<Settings> {
	return await SystemAPI.Settings()
}

export async function saveSettings(input: SaveSettingsInput): Promise<Result> {
	return await SystemAPI.SaveSettings(input)
}

export async function pickFolder(title: string): Promise<string> {
	return await SystemAPI.PickFolder(title)
}

export async function openURL(url: string): Promise<Result> {
	return await SystemAPI.OpenURL(url)
}

export async function openFolder(path: string): Promise<Result> {
	return await SystemAPI.OpenFolder(path)
}

export async function loadDataFolder(): Promise<string> {
	return await SystemAPI.DataFolder()
}
