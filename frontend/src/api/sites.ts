import * as SiteAPI from '../../wailsjs/go/wails/SiteAPI'
import * as GitAPI from '../../wailsjs/go/wails/GitAPI'
import type { Site, CreateSiteInput, UpdateSiteInput, GitStatus, OperationResult } from '../types'

export async function listSites(): Promise<Site[]> {
  const list = await SiteAPI.List()
  return list as unknown as Site[]
}

export async function getSite(id: number): Promise<Site> {
  const res = await SiteAPI.Get(id)
  return res as unknown as Site
}

export async function createSite(input: CreateSiteInput): Promise<Site> {
  const res = await SiteAPI.Create(input)
  return res as unknown as Site
}

export async function updateSite(input: UpdateSiteInput): Promise<Site> {
  const res = await SiteAPI.Update(input)
  return res as unknown as Site
}

export async function startSite(id: number): Promise<OperationResult> {
  return await SiteAPI.Start(id)
}

export async function stopSite(id: number): Promise<OperationResult> {
  return await SiteAPI.Stop(id)
}

export async function restartSite(id: number): Promise<OperationResult> {
  return await SiteAPI.Restart(id)
}

export async function deleteSite(id: number, purge = false): Promise<OperationResult> {
  return await SiteAPI.Delete({ id, purge })
}

export async function getSiteConf(id: number): Promise<string> {
  return await SiteAPI.Conf(id)
}

export async function renderSiteConf(id: number): Promise<string> {
  return await SiteAPI.RenderConf(id)
}

export async function saveSiteConf(siteId: number, content: string, summary: string): Promise<OperationResult> {
  return await SiteAPI.SaveConf({ siteId, content, summary })
}

export async function checkSiteConf(content: string): Promise<{ valid: boolean; message: string }> {
  return await SiteAPI.CheckConf(content)
}

export async function getSiteLogs(siteId: number, kind: 'access' | 'error' = 'access', limit = 100): Promise<string[]> {
  const res = await SiteAPI.Logs({ siteId, kind, limit })
  return res.lines
}

export async function getGitStatus(siteId: number): Promise<GitStatus> {
  return await GitAPI.Status(siteId)
}

export async function pullGit(siteId: number): Promise<OperationResult> {
  return await GitAPI.Pull(siteId)
}

export async function commitGit(siteId: number, message: string, files: string[] = []): Promise<OperationResult> {
  return await GitAPI.Commit({ siteId, message, files })
}

export async function pushGit(siteId: number): Promise<OperationResult> {
  return await GitAPI.Push(siteId)
}
