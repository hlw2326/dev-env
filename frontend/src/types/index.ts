export type ID = number

export type RuntimeType =
  | 'php'
  | 'nginx'
  | 'mysql'
  | 'postgresql'
  | 'redis'
  | 'apache'

export type RuntimeStatus =
  | 'available'
  | 'installing'
  | 'installed'
  | 'failed'
  | 'removing'

export type ServiceStatus =
  | 'stopped'
  | 'starting'
  | 'running'
  | 'stopping'
  | 'failed'

export type TaskStatus =
  | 'queued'
  | 'running'
  | 'cancelling'
  | 'succeeded'
  | 'failed'
  | 'cancelled'

export interface OperationResult {
  success: boolean
  message?: string
}

export interface Project {
  id: ID
  name: string
  path: string
  phpVersionId: ID | null
  status: 'ready' | 'missing' | 'invalid'
  createdAt: string
  updatedAt: string
}

export interface CreateProjectInput {
  name: string
  path: string
  phpVersionId?: ID | null
}

export interface UpdateProjectInput {
  id: ID
  name?: string
  phpVersionId?: ID | null
}

export interface ProjectPathValidation {
  valid: boolean
  normalizedPath: string
  exists: boolean
  writable: boolean
  errors: string[]
}

export interface PHPVersion {
  id: ID
  version: string
  architecture: string
  installPath: string
  status: RuntimeStatus
  isDefault: boolean
}

export interface RegisterPHPVersionInput {
  version: string
  architecture: string
  installPath: string
}

export interface PHPExtension {
  name: string
  enabled: boolean
  available: boolean
}

export interface ServiceInstance {
  id: ID
  name: string
  serviceType: RuntimeType
  runtimeVersionId: ID
  port: number | null
  configPath: string | null
  dataPath: string | null
  enabled: boolean
  autoStart: boolean
  status: ServiceStatus
  pid: number | null
}

export interface ServiceStatusInfo {
  serviceId: ID
  status: ServiceStatus
  pid: number | null
  startedAt: string | null
  lastErrorCode: string | null
}

export interface VirtualHost {
  id: ID
  projectId: ID
  domain: string
  documentRoot: string
  webServerType: 'nginx' | 'apache'
  port: number
  enabled: boolean
  configPath: string | null
}

export interface CreateVirtualHostInput {
  projectId: ID
  domain: string
  documentRoot: string
  webServerType: 'nginx' | 'apache'
  port: number
}

export interface ValidationResult {
  valid: boolean
  errors: string[]
  warnings: string[]
}

export interface Task {
  id: ID
  taskType: string
  status: TaskStatus
  progress: number
  message: string
  errorCode: string | null
  errorMessage: string | null
  createdAt: string
  updatedAt: string
  finishedAt: string | null
}

export interface AppSettings {
  language: string
  theme: 'light' | 'dark' | 'system'
  runtimeRootPath: string
  defaultProjectPath: string
  autoStartServices: boolean
}
