import * as SoftwareAPI from '../../wailsjs/go/wails/SoftwareAPI'
import type {
	MarketItem,
	Software,
	SoftwareVersion,
	SoftwareConfig,
	SaveConfigInput,
	SaveConfigResult,
	Extension,
	InstallInput,
	Result,
} from '@/types'

export async function loadMarket(): Promise<MarketItem[]> {
	return await SoftwareAPI.Market()
}

export async function listSoftware(): Promise<Software[]> {
	return await SoftwareAPI.List()
}

export async function listVersions(kind: string): Promise<SoftwareVersion[]> {
	return await SoftwareAPI.Versions(kind)
}

export async function loadConfig(id: number): Promise<SoftwareConfig> {
	return await SoftwareAPI.Config(id)
}

export async function saveConfig(input: SaveConfigInput): Promise<SaveConfigResult> {
	return await SoftwareAPI.SaveConfig(input)
}

export async function readConfig(id: number): Promise<string> {
	return await SoftwareAPI.ReadConfig(id)
}

export async function writeConfig(id: number, text: string): Promise<Result> {
	return await SoftwareAPI.WriteConfig(id, text)
}

export async function listExtensions(id: number): Promise<Extension[]> {
	return await SoftwareAPI.Extensions(id)
}

export async function setExtension(id: number, name: string, on: boolean): Promise<Result> {
	return await SoftwareAPI.SetExtension(id, name, on)
}

export async function openSoftwareFolder(id: number): Promise<Result> {
	return await SoftwareAPI.OpenFolder(id)
}

export async function installSoftware(input: InstallInput): Promise<Result> {
	return await SoftwareAPI.Install(input)
}

export async function cancelInstall(kind: string, version: string): Promise<Result> {
	return await SoftwareAPI.Cancel(kind, version)
}

export async function removeSoftware(id: number): Promise<Result> {
	return await SoftwareAPI.Remove(id)
}

export async function setMainSoftware(id: number): Promise<Result> {
	return await SoftwareAPI.SetMain(id)
}

export async function startSoftware(id: number): Promise<Result> {
	return await SoftwareAPI.Start(id)
}

export async function stopSoftware(id: number): Promise<Result> {
	return await SoftwareAPI.Stop(id)
}

export async function setSoftwareAutostart(id: number, on: boolean): Promise<Result> {
	return await SoftwareAPI.SetAutostart(id, on)
}

export async function loadSoftwareLogs(id: number, limit: number): Promise<string[]> {
	return await SoftwareAPI.Logs(id, limit)
}
