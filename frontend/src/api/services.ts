import * as AppAPI from '../../wailsjs/go/wails/AppAPI'
import type { ServiceInstance, ServiceStatusInfo, OperationResult, RuntimeType, ServiceStatus } from '../types'

export async function listServices(): Promise<ServiceInstance[]> {
  const items = await AppAPI.ListServices()
  return (items || []).map((s) => ({
    id: s.id,
    name: s.name,
    serviceType: s.serviceType as RuntimeType,
    runtimeVersionId: s.runtimeVersionId,
    port: s.port ?? null,
    configPath: s.configPath ?? null,
    dataPath: s.dataPath ?? null,
    enabled: s.enabled,
    autoStart: s.autoStart,
    status: s.status as ServiceStatus,
    pid: s.pid ?? null,
  }))
}

export async function getService(id: number): Promise<ServiceInstance> {
  const s = await AppAPI.GetService(id)
  return {
    id: s.id,
    name: s.name,
    serviceType: s.serviceType as RuntimeType,
    runtimeVersionId: s.runtimeVersionId,
    port: s.port ?? null,
    configPath: s.configPath ?? null,
    dataPath: s.dataPath ?? null,
    enabled: s.enabled,
    autoStart: s.autoStart,
    status: s.status as ServiceStatus,
    pid: s.pid ?? null,
  }
}

export async function getServiceStatus(id: number): Promise<ServiceStatusInfo> {
  const st = await AppAPI.GetServiceStatus(id)
  return {
    serviceId: st.serviceId,
    status: st.status as ServiceStatus,
    pid: st.pid ?? null,
    startedAt: st.startedAt ?? null,
    lastErrorCode: st.lastErrorCode ?? null,
  }
}

export async function startService(serviceId: number): Promise<OperationResult> {
  return await AppAPI.StartService({ serviceId })
}

export async function stopService(serviceId: number): Promise<OperationResult> {
  return await AppAPI.StopService({ serviceId })
}

export async function restartService(serviceId: number): Promise<OperationResult> {
  return await AppAPI.RestartService({ serviceId })
}

export async function getServiceLogs(serviceId: number, limit = 100): Promise<string[]> {
  const res = await AppAPI.GetServiceLogs({ serviceId, limit })
  return res.lines || []
}
