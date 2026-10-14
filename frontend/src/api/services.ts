import * as ServiceAPI from '../../wailsjs/go/wails/ServiceAPI'
import type {
  ServiceInstance,
  ServiceStatusInfo,
  OperationResult,
  RuntimeType,
  ServiceStatus,
  RuntimeOption,
} from '../types'

export async function listServices(): Promise<ServiceInstance[]> {
  const items = await ServiceAPI.List()
  return items.map((s) => ({
    id: s.id,
    name: s.name,
    serviceType: s.serviceType as RuntimeType,
    runtimeId: s.runtimeId,
    runtimeVersion: s.runtimeVersion,
    runtimePath: s.runtimePath,
    port: s.port,
    configPath: s.configPath,
    dataPath: s.dataPath,
    enabled: s.enabled,
    autoStart: s.autoStart,
    status: s.status as ServiceStatus,
    pid: s.pid,
  }))
}

export async function getService(id: number): Promise<ServiceInstance> {
  const s = await ServiceAPI.Get(id)
  return {
    id: s.id,
    name: s.name,
    serviceType: s.serviceType as RuntimeType,
    runtimeId: s.runtimeId,
    runtimeVersion: s.runtimeVersion,
    runtimePath: s.runtimePath,
    port: s.port,
    configPath: s.configPath,
    dataPath: s.dataPath,
    enabled: s.enabled,
    autoStart: s.autoStart,
    status: s.status as ServiceStatus,
    pid: s.pid,
  }
}

export async function getServiceStatus(id: number): Promise<ServiceStatusInfo> {
  const st = await ServiceAPI.Status(id)
  return {
    serviceId: st.serviceId,
    status: st.status as ServiceStatus,
    pid: st.pid,
    startedAt: st.startedAt,
    errorCode: st.errorCode,
  }
}

export async function startService(serviceId: number): Promise<OperationResult> {
  return await ServiceAPI.Start({ serviceId })
}

export async function stopService(serviceId: number): Promise<OperationResult> {
  return await ServiceAPI.Stop({ serviceId })
}

export async function restartService(serviceId: number): Promise<OperationResult> {
  return await ServiceAPI.Restart({ serviceId })
}

export async function getServiceLogs(serviceId: number, limit = 100): Promise<string[]> {
  const res = await ServiceAPI.Logs({ serviceId, limit })
  return res.lines
}

export async function deleteService(serviceId: number): Promise<OperationResult> {
  return await ServiceAPI.Delete(serviceId)
}

export async function switchServiceRuntime(
  serviceId: number,
  runtimeId: number
): Promise<OperationResult> {
  return await ServiceAPI.SwitchRuntime({ serviceId, runtimeId })
}

export async function getServiceRuntimeOptions(serviceId: number): Promise<RuntimeOption[]> {
  const items = await ServiceAPI.RuntimeOptions(serviceId)
  return items.map((opt) => ({
    runtimeId: opt.id,
    version: opt.version,
    architecture: opt.architecture,
    installPath: opt.installPath,
    status: 'installed',
    isCurrent: false,
  }))
}


