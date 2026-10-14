import * as AppAPI from '../../wailsjs/go/wails/AppAPI'
import type { VirtualHost, CreateVirtualHostInput, ValidationResult, OperationResult } from '../types'

export async function listVirtualHosts(): Promise<VirtualHost[]> {
  const items = await AppAPI.ListVirtualHosts()
  return (items || []).map((h) => ({
    id: h.id,
    projectId: h.projectId,
    domain: h.domain,
    documentRoot: h.documentRoot,
    webServerType: h.webServerType as 'nginx' | 'apache',
    port: h.port,
    enabled: h.enabled,
    configPath: h.configPath ?? null,
  }))
}

export async function getVirtualHost(id: number): Promise<VirtualHost> {
  const h = await AppAPI.GetVirtualHost(id)
  return {
    id: h.id,
    projectId: h.projectId,
    domain: h.domain,
    documentRoot: h.documentRoot,
    webServerType: h.webServerType as 'nginx' | 'apache',
    port: h.port,
    enabled: h.enabled,
    configPath: h.configPath ?? null,
  }
}

export async function createVirtualHost(input: CreateVirtualHostInput): Promise<VirtualHost> {
  const h = await AppAPI.CreateVirtualHost(input)
  return {
    id: h.id,
    projectId: h.projectId,
    domain: h.domain,
    documentRoot: h.documentRoot,
    webServerType: h.webServerType as 'nginx' | 'apache',
    port: h.port,
    enabled: h.enabled,
    configPath: h.configPath ?? null,
  }
}

export async function updateVirtualHost(host: VirtualHost): Promise<VirtualHost> {
  const h = await AppAPI.UpdateVirtualHost({
    id: host.id,
    projectId: host.projectId,
    domain: host.domain,
    documentRoot: host.documentRoot,
    webServerType: host.webServerType,
    port: host.port,
    enabled: host.enabled,
    configPath: host.configPath ?? undefined,
  })
  return {
    id: h.id,
    projectId: h.projectId,
    domain: h.domain,
    documentRoot: h.documentRoot,
    webServerType: h.webServerType as 'nginx' | 'apache',
    port: h.port,
    enabled: h.enabled,
    configPath: h.configPath ?? null,
  }
}

export async function deleteVirtualHost(id: number): Promise<OperationResult> {
  return await AppAPI.DeleteVirtualHost(id)
}

export async function enableVirtualHost(id: number): Promise<OperationResult> {
  return await AppAPI.EnableVirtualHost(id)
}

export async function disableVirtualHost(id: number): Promise<OperationResult> {
  return await AppAPI.DisableVirtualHost(id)
}

export async function validateVirtualHost(input: CreateVirtualHostInput): Promise<ValidationResult> {
  return await AppAPI.ValidateVirtualHost(input)
}
