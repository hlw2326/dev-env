export type ID = number

export type RuntimeType =
  | 'php'
  | 'nginx'
  | 'mysql'
  | 'postgresql'
  | 'redis'
  | 'apache'
  | 'nodejs'

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

export interface PHPVersion {
  id: ID
  runtimeType?: string
  version: string
  architecture: string
  installPath: string
  status: RuntimeStatus
  isDefault: boolean
  siteCount?: number
  serviceCount?: number
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
  runtimeId: ID
  runtimeVersion?: string
  runtimePath?: string
  port?: number
  configPath?: string
  dataPath?: string
  enabled: boolean
  autoStart: boolean
  status: ServiceStatus
  pid?: number
}

export interface RuntimeOption {
  runtimeId: ID
  version: string
  architecture: string
  installPath: string
  status: string
  isCurrent: boolean
}

export interface InstallResult {
  runtimeId: ID
  serviceId?: ID
  serviceCreated: boolean
}

export interface ServiceStatusInfo {
  serviceId: ID
  status: ServiceStatus
  pid?: number
  startedAt?: string
  errorCode?: string
}

export interface Task {
  id: ID
  taskType: string
  status: TaskStatus
  progress: number
  message: string
  errorCode?: string
  errorMessage?: string
  createdAt: string
  updatedAt: string
  finishedAt?: string
}

export interface AppSettings {
  language: string
  theme: 'light' | 'dark' | 'system'
  runtimeRootPath: string
  defaultProjectPath: string
  autoStartServices: boolean
}

export interface PackageInfo {
  key: string
  displayName: string
  category: 'language' | 'webserver' | 'database' | 'cache'
  description: string
  iconClass: string
  brandColor: string
  isInstalled: boolean
  installedVersions: string[]
  recommendedVersion?: string
  officialWebsite?: string
}

export interface PackageVersion {
  version: string
  variant?: string
  architecture: string
  downloadUrl: string
  mirrorUrl?: string
  fileSize: number
  isRecommended: boolean
  isLTS: boolean
  releaseDate?: string
}

export interface RuntimeVersion {
  id: number
  runtimeType: string
  version: string
  variant?: string
  architecture: string
  installPath: string
  status: 'available' | 'installing' | 'uninitialized' | 'error' | 'in_use' | 'installed' | 'removed'
  metadataJson?: string
  createdAt?: string
  updatedAt?: string
  isDefault?: boolean
  siteCount?: number
  serviceCount?: number
}

export type SiteKind = 'static' | 'php' | 'nodejs' | 'go' | 'python'

export type SiteStatus = 'stopped' | 'running' | 'error'

export interface SiteSnapshot {
  version: string
  timestamp: string
  summary: string
  content: string
}

export interface SiteMeta {
  entry: string
  cmd: string
  autoStart: boolean
  env: Record<string, string>
  custom: string
  accessLog: string
  errorLog: string
  history: SiteSnapshot[]
}

export interface Site {
  id: ID
  name: string
  kind: SiteKind
  domains: string[]
  root: string
  runtimeId?: ID
  serviceId?: ID
  git: boolean
  remote: string
  branch: string
  commit: string
  pulledAt: string
  conf: string
  port: number
  tls: boolean
  cert: string
  key: string
  issuer: string
  proxy: boolean
  upstream?: number
  meta: SiteMeta
  status: SiteStatus
  createdAt: string
  updatedAt: string
}

export interface CreateSiteInput {
  name: string
  kind: SiteKind
  domains: string[]
  root: string
  runtimeId?: ID
  serviceId?: ID
  autoService: boolean
  git: boolean
  remote: string
  branch: string
  port: number
  tls: boolean
  issuer: string
  cert: string
  key: string
  upstream?: number
  entry: string
  cmd: string
  autoStart: boolean
  startNow: boolean
  custom: string
  env?: Record<string, string>
}

export interface UpdateSiteInput {
  id: ID
  name?: string
  domains?: string[]
  root?: string
  runtimeId?: ID
  serviceId?: ID
  port?: number
  tls?: boolean
  issuer?: string
  cert?: string
  key?: string
  upstream?: number
  entry?: string
  cmd?: string
  custom?: string
  autoStart?: boolean
}

export interface GitStatus {
  branch: string
  commit: string
  message: string
  clean: boolean
  ahead: number
  behind: number
  modified: string[]
  untracked: string[]
  checkedAt: string
}
