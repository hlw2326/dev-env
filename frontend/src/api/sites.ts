import * as SiteAPI from '../../wailsjs/go/wails/SiteAPI'
import type { Site, CreateSiteInput, UpdateSiteInput, Result } from '@/types'

export async function listSites(): Promise<Site[]> {
	return await SiteAPI.List()
}

export async function getSite(id: number): Promise<Site> {
	return await SiteAPI.Get(id)
}

export async function createSite(input: CreateSiteInput): Promise<Site> {
	return await SiteAPI.Create(input)
}

export async function updateSite(input: UpdateSiteInput): Promise<Site> {
	return await SiteAPI.Update(input)
}

export async function findRun(root: string, kind: string): Promise<string> {
	return await SiteAPI.FindRun(root, kind)
}

export async function startSite(id: number): Promise<Result> {
	return await SiteAPI.Start(id)
}

export async function stopSite(id: number): Promise<Result> {
	return await SiteAPI.Stop(id)
}

export async function restartSite(id: number): Promise<Result> {
	return await SiteAPI.Restart(id)
}

export async function deleteSite(id: number): Promise<Result> {
	return await SiteAPI.Delete(id)
}

export async function loadSiteLogs(id: number, kind: string, limit: number): Promise<string[]> {
	return await SiteAPI.Logs(id, kind, limit)
}
