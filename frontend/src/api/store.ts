import type { PackageInfo, PackageVersion } from '../types'
import { listAllRuntimes } from './runtimes'

// 基础软件包目录定义
const CATALOG_BASE: Omit<PackageInfo, 'isInstalled' | 'installedVersions'>[] = [
  {
    key: 'php',
    displayName: 'PHP',
    category: 'language',
    description: '通用开源脚本语言，尤其适用于 Web 开发与服务端编程',
    iconClass: 'pi pi-code',
    brandColor: '#777BB4',
    recommendedVersion: '8.3.35',
    officialWebsite: 'https://www.php.net',
  },
  {
    key: 'nginx',
    displayName: 'Nginx',
    category: 'webserver',
    description: '高性能 HTTP 和反向代理 Web 服务器，兼具 IMAP/POP3 代理功能',
    iconClass: 'pi pi-server',
    brandColor: '#009639',
    recommendedVersion: '1.26.3',
    officialWebsite: 'https://nginx.org',
  },
  {
    key: 'mysql',
    displayName: 'MySQL',
    category: 'database',
    description: '全球广泛使用的开源关系型数据库管理系统 (RDBMS)',
    iconClass: 'pi pi-database',
    brandColor: '#4479A1',
    recommendedVersion: '8.4.11',
    officialWebsite: 'https://www.mysql.com',
  },
  {
    key: 'redis',
    displayName: 'Redis',
    category: 'cache',
    description: '内存级键值数据结构存储系统，可用作数据库、缓存和消息中间件',
    iconClass: 'pi pi-bolt',
    brandColor: '#DC382D',
    recommendedVersion: '8.10.2',
    officialWebsite: 'https://redis.io',
  },
  {
    key: 'postgresql',
    displayName: 'PostgreSQL',
    category: 'database',
    description: '功能强大的企业级开源对象-关系型数据库系统',
    iconClass: 'pi pi-database',
    brandColor: '#336791',
    recommendedVersion: '16.15',
    officialWebsite: 'https://www.postgresql.org',
  },
  {
    key: 'apache',
    displayName: 'Apache HTTP',
    category: 'webserver',
    description: '经典模块化 Web 服务器 (Apache 2.4 / httpd)',
    iconClass: 'pi pi-compass',
    brandColor: '#D22128',
    recommendedVersion: '2.4.69',
    officialWebsite: 'https://httpd.apache.org',
  },
  {
    key: 'nodejs',
    displayName: 'Node.js',
    category: 'language',
    description: '基于 Chrome V8 引擎的高并发异步事件驱动 JavaScript 运行时',
    iconClass: 'pi pi-code',
    brandColor: '#339933',
    recommendedVersion: '24.21.0',
    officialWebsite: 'https://nodejs.org',
  },
]

// 离线基准兜底版本（在网络离线或接口异常时保障可用）
const BASELINE_VERSIONS: Record<string, PackageVersion[]> = {
  php: [
    {
      version: '8.3.35',
      variant: 'nts',
      architecture: 'x64',
      downloadUrl: 'https://windows.php.net/downloads/releases/php-8.3.35-nts-Win32-vs16-x64.zip',
      mirrorUrl: 'https://mirrors.aliyun.com/php/releases/php-8.3.35-nts-Win32-vs16-x64.zip',
      fileSize: 33800000,
      isRecommended: true,
      isLTS: false,
      releaseDate: '2026-09-22',
    },
    {
      version: '8.2.34',
      variant: 'nts',
      architecture: 'x64',
      downloadUrl: 'https://windows.php.net/downloads/releases/php-8.2.34-nts-Win32-vs16-x64.zip',
      mirrorUrl: 'https://mirrors.aliyun.com/php/releases/php-8.2.34-nts-Win32-vs16-x64.zip',
      fileSize: 33400000,
      isRecommended: false,
      isLTS: true,
      releaseDate: '2026-09-22',
    },
    {
      version: '8.1.34',
      variant: 'nts',
      architecture: 'x64',
      downloadUrl: 'https://windows.php.net/downloads/releases/php-8.1.34-nts-Win32-vs16-x64.zip',
      mirrorUrl: 'https://mirrors.aliyun.com/php/releases/php-8.1.34-nts-Win32-vs16-x64.zip',
      fileSize: 30800000,
      isRecommended: false,
      isLTS: true,
      releaseDate: '2025-12-16',
    },
    {
      version: '7.4.33',
      variant: 'nts',
      architecture: 'x64',
      downloadUrl: 'https://windows.php.net/downloads/releases/php-7.4.33-nts-Win32-vc15-x64.zip',
      mirrorUrl: 'https://mirrors.aliyun.com/php/releases/php-7.4.33-nts-Win32-vc15-x64.zip',
      fileSize: 26100000,
      isRecommended: false,
      isLTS: false,
      releaseDate: '2024-07-10',
    },
  ],
  nginx: [
    {
      version: '1.26.3',
      architecture: 'x64',
      downloadUrl: 'https://nginx.org/download/nginx-1.26.3.zip',
      mirrorUrl: 'https://mirrors.aliyun.com/nginx/nginx-1.26.3.zip',
      fileSize: 2100000,
      isRecommended: true,
      isLTS: true,
      releaseDate: '2025-02-10',
    },
    {
      version: '1.24.0',
      architecture: 'x64',
      downloadUrl: 'https://nginx.org/download/nginx-1.24.0.zip',
      mirrorUrl: 'https://mirrors.aliyun.com/nginx/nginx-1.24.0.zip',
      fileSize: 1900000,
      isRecommended: false,
      isLTS: true,
      releaseDate: '2023-04-11',
    },
  ],
  mysql: [
    {
      version: '8.4.11',
      architecture: 'x64',
      downloadUrl: 'https://dev.mysql.com/get/Downloads/MySQL-8.4/mysql-8.4.11-winx64.zip',
      mirrorUrl: 'https://mirrors.aliyun.com/mysql/MySQL-8.4/mysql-8.4.11-winx64.zip',
      fileSize: 245000000,
      isRecommended: true,
      isLTS: true,
      releaseDate: '2026-05-15',
    },
    {
      version: '8.0.48',
      architecture: 'x64',
      downloadUrl: 'https://dev.mysql.com/get/Downloads/MySQL-8.0/mysql-8.0.48-winx64.zip',
      mirrorUrl: 'https://mirrors.aliyun.com/mysql/MySQL-8.0/mysql-8.0.48-winx64.zip',
      fileSize: 232000000,
      isRecommended: false,
      isLTS: true,
      releaseDate: '2026-04-10',
    },
  ],
  redis: [
    {
      version: '8.10.2',
      architecture: 'x64',
      downloadUrl: 'https://github.com/redis-windows/redis-windows/releases/download/8.10.2/Redis-8.10.2-Windows-x64-cygwin-with-Service.zip',
      fileSize: 12500000,
      isRecommended: true,
      isLTS: true,
      releaseDate: '2026-09-18',
    },
    {
      version: '7.2.4',
      architecture: 'x64',
      downloadUrl: 'https://github.com/tporadowski/redis/releases/download/v7.2.4/Redis-x64-7.2.4.zip',
      fileSize: 8500000,
      isRecommended: false,
      isLTS: true,
      releaseDate: '2024-01-10',
    },
  ],
  postgresql: [
    {
      version: '16.15',
      architecture: 'x64',
      downloadUrl: 'https://get.enterprisedb.com/postgresql/postgresql-16.15-windows-x64-binaries.zip',
      fileSize: 310000000,
      isRecommended: true,
      isLTS: true,
      releaseDate: '2026-05-20',
    },
  ],
  apache: [
    {
      version: '2.4.69',
      architecture: 'x64',
      downloadUrl: 'https://dlcdn.apache.org/httpd/binaries/win64/httpd-2.4.69-win64.zip',
      fileSize: 18500000,
      isRecommended: true,
      isLTS: true,
      releaseDate: '2026-06-01',
    },
  ],
  nodejs: [
    {
      version: '24.21.0',
      architecture: 'x64',
      downloadUrl: 'https://nodejs.org/dist/v24.21.0/node-v24.21.0-win-x64.zip',
      mirrorUrl: 'https://npmmirror.com/mirrors/node/v24.21.0/node-v24.21.0-win-x64.zip',
      fileSize: 34000000,
      isRecommended: true,
      isLTS: true,
      releaseDate: '2026-09-07',
    },
    {
      version: '22.18.0',
      architecture: 'x64',
      downloadUrl: 'https://nodejs.org/dist/v22.18.0/node-v22.18.0-win-x64.zip',
      mirrorUrl: 'https://npmmirror.com/mirrors/node/v22.18.0/node-v22.18.0-win-x64.zip',
      fileSize: 32000000,
      isRecommended: false,
      isLTS: true,
      releaseDate: '2025-10-15',
    },
  ],
}

// 缓存管理
const CACHE_PREFIX = 'devenv_store_cache_'
const CACHE_TTL_MS = 30 * 60 * 1000 // 缓存 30 分钟

function getCachedVersions(key: string): PackageVersion[] | null {
  try {
    const raw = localStorage.getItem(CACHE_PREFIX + key)
    if (!raw) return null
    const parsed = JSON.parse(raw)
    if (Date.now() - (parsed.time || 0) < CACHE_TTL_MS && Array.isArray(parsed.data)) {
      return parsed.data
    }
  } catch {
    // ignore
  }
  return null
}

function setCachedVersions(key: string, data: PackageVersion[]) {
  try {
    localStorage.setItem(
      CACHE_PREFIX + key,
      JSON.stringify({ time: Date.now(), data })
    )
  } catch {
    // ignore
  }
}

// 带超时的通用请求封装
async function fetchWithTimeout(url: string, options: RequestInit = {}, timeoutMs = 6000): Promise<Response> {
  const controller = new AbortController()
  const id = setTimeout(() => controller.abort(), timeoutMs)
  try {
    const res = await fetch(url, { ...options, signal: controller.signal })
    clearTimeout(id)
    return res
  } catch (err) {
    clearTimeout(id)
    throw err
  }
}

// 辅助大小转换
function parseSizeToBytes(size: string | number | undefined): number {
  if (!size) return 0
  if (typeof size === 'number') return size
  const s = String(size).toUpperCase().trim()
  if (s.endsWith('MB')) return Math.round(parseFloat(s) * 1024 * 1024)
  if (s.endsWith('KB')) return Math.round(parseFloat(s) * 1024)
  if (s.endsWith('GB')) return Math.round(parseFloat(s) * 1024 * 1024 * 1024)
  return parseFloat(s) || 0
}

// 1. 在线抓取 PHP Windows 官方 releases.json
async function fetchRemotePHPVersions(): Promise<PackageVersion[]> {
  const res = await fetchWithTimeout('https://windows.php.net/downloads/releases/releases.json')
  if (!res.ok) throw new Error(`HTTP ${res.status}`)
  const data = await res.json()
  const list: PackageVersion[] = []

  const sortedBranches = Object.keys(data).sort((a, b) => parseFloat(b) - parseFloat(a))

  for (const branch of sortedBranches) {
    const info = data[branch]
    if (!info || !info.version) continue
    const nts = info['nts-vs17-x64'] || info['nts-vs16-x64'] || info['nts-vc15-x64'] || info['nts-vs16-x86']
    if (!nts || !nts.zip) continue

    const zipPath = nts.zip.path || `php-${info.version}-nts-Win32-vs16-x64.zip`
    const isRecommended = branch === '8.3' || (branch === '8.4' && !sortedBranches.includes('8.3'))
    const isLTS = branch === '8.2' || branch === '8.1'

    list.push({
      version: info.version,
      variant: 'nts',
      architecture: 'x64',
      downloadUrl: `https://windows.php.net/downloads/releases/${zipPath}`,
      mirrorUrl: `https://mirrors.aliyun.com/php/releases/${zipPath}`,
      fileSize: parseSizeToBytes(nts.zip.size),
      isRecommended,
      isLTS,
      releaseDate: nts.mtime ? nts.mtime.split('T')[0] : '',
    })
  }

  return list
}

// 2. 在线抓取 Node.js 官方发布索引
async function fetchRemoteNodeVersions(): Promise<PackageVersion[]> {
  let data: any[] = []
  try {
    const res = await fetchWithTimeout('https://nodejs.org/dist/index.json')
    if (res.ok) data = await res.json()
  } catch {
    // 降级使用 npmmirror 镜像源
    const res = await fetchWithTimeout('https://npmmirror.com/mirrors/node/index.json')
    if (res.ok) data = await res.json()
  }

  if (!Array.isArray(data) || data.length === 0) throw new Error('未能获取 Node.js 列表')

  const list: PackageVersion[] = []
  const seenMajor: Record<string, number> = {}

  for (const item of data) {
    if (!item.files || !item.files.includes('win-x64-zip')) continue
    const ver = item.version.replace(/^v/, '')
    const major = ver.split('.')[0]
    const count = seenMajor[major] || 0

    if (count < 2) {
      list.push({
        version: ver,
        architecture: 'x64',
        downloadUrl: `https://nodejs.org/dist/${item.version}/node-${item.version}-win-x64.zip`,
        mirrorUrl: `https://npmmirror.com/mirrors/node/${item.version}/node-${item.version}-win-x64.zip`,
        fileSize: 34000000,
        isRecommended: !!item.lts,
        isLTS: !!item.lts,
        releaseDate: item.date,
      })
      seenMajor[major] = count + 1
    }
    if (list.length >= 10) break
  }

  return list
}

// 3. 在线抓取 Nginx Windows 官方版本列表
async function fetchRemoteNginxVersions(): Promise<PackageVersion[]> {
  const res = await fetchWithTimeout('https://nginx.org/en/download.html')
  if (!res.ok) throw new Error(`HTTP ${res.status}`)
  const html = await res.text()

  const re = /\/download\/nginx-([0-9\.]+)\.zip/g
  let m
  const vers: string[] = []
  while ((m = re.exec(html)) !== null) {
    if (!vers.includes(m[1])) vers.push(m[1])
  }

  if (vers.length === 0) throw new Error('未在 Nginx 官网找到 Windows 安装包')

  return vers.slice(0, 8).map((v, idx) => ({
    version: v,
    architecture: 'x64',
    downloadUrl: `https://nginx.org/download/nginx-${v}.zip`,
    mirrorUrl: `https://mirrors.aliyun.com/nginx/nginx-${v}.zip`,
    fileSize: 2100000,
    isRecommended: idx === 0 || v.startsWith('1.26.'),
    isLTS: v.startsWith('1.26.') || v.startsWith('1.24.'),
    releaseDate: '最新发布',
  }))
}

// 4. 在线抓取 Redis Windows 版本列表
async function fetchRemoteRedisVersions(): Promise<PackageVersion[]> {
  const res = await fetchWithTimeout('https://api.github.com/repos/redis-windows/redis-windows/releases', {
    headers: { 'User-Agent': 'DevEnv' },
  })
  if (!res.ok) throw new Error(`HTTP ${res.status}`)
  const data = await res.json()
  if (!Array.isArray(data)) throw new Error('获取 Redis 列表异常')

  const list: PackageVersion[] = []
  for (const rel of data.slice(0, 8)) {
    const zipAsset = rel.assets && rel.assets.find((a: any) => a.name && a.name.endsWith('.zip'))
    const ver = rel.tag_name ? rel.tag_name.replace(/^v/, '') : ''
    if (!ver) continue

    list.push({
      version: ver,
      architecture: 'x64',
      downloadUrl: zipAsset?.browser_download_url || `https://github.com/redis-windows/redis-windows/releases/download/${rel.tag_name}/Redis-${ver}-Windows-x64-cygwin-with-Service.zip`,
      fileSize: zipAsset?.size || 12000000,
      isRecommended: list.length === 0,
      isLTS: true,
      releaseDate: rel.published_at ? rel.published_at.split('T')[0] : '',
    })
  }

  return list
}

// 5. 在线抓取 MySQL 版本列表
async function fetchRemoteMySQLVersions(): Promise<PackageVersion[]> {
  const res = await fetchWithTimeout('https://api.github.com/repos/mysql/mysql-server/tags?per_page=100', {
    headers: { 'User-Agent': 'DevEnv' },
  })
  if (!res.ok) throw new Error(`HTTP ${res.status}`)
  const data = await res.json()
  if (!Array.isArray(data)) throw new Error('获取 MySQL 标签异常')

  const vers: string[] = []
  const seen = new Set<string>()

  for (const item of data) {
    const m = item.name.match(/^mysql-(?:cluster-)?([0-9]+\.[0-9]+\.[0-9]+)$/)
    if (m) {
      const v = m[1]
      const majorMinor = v.split('.').slice(0, 2).join('.')
      if (!seen.has(majorMinor) && (v.startsWith('8.') || v.startsWith('9.'))) {
        seen.add(majorMinor)
        vers.push(v)
      }
    }
  }

  if (vers.length === 0) throw new Error('未解析到 MySQL 版本')

  return vers.slice(0, 6).map((v, idx) => ({
    version: v,
    architecture: 'x64',
    downloadUrl: `https://dev.mysql.com/get/Downloads/MySQL-${v.split('.').slice(0, 2).join('.')}/mysql-${v}-winx64.zip`,
    mirrorUrl: `https://mirrors.aliyun.com/mysql/MySQL-${v.split('.').slice(0, 2).join('.')}/mysql-${v}-winx64.zip`,
    fileSize: 245000000,
    isRecommended: v.startsWith('8.4.'),
    isLTS: v.startsWith('8.4.') || v.startsWith('8.0.'),
    releaseDate: idx === 0 ? '最新' : '稳定版',
  }))
}

// 6. 在线抓取 PostgreSQL 版本列表
async function fetchRemotePostgresVersions(): Promise<PackageVersion[]> {
  const res = await fetchWithTimeout('https://api.github.com/repos/postgres/postgres/tags?per_page=100', {
    headers: { 'User-Agent': 'DevEnv' },
  })
  if (!res.ok) throw new Error(`HTTP ${res.status}`)
  const data = await res.json()
  if (!Array.isArray(data)) throw new Error('获取 PostgreSQL 列表异常')

  const list: PackageVersion[] = []
  const seen = new Set<string>()

  for (const item of data) {
    const m = item.name.match(/^REL_([0-9]+)_([0-9]+)$/)
    if (m) {
      const major = m[1]
      const v = `${major}.${m[2]}`
      if (!seen.has(major) && parseInt(major) >= 12 && parseInt(major) <= 18) {
        seen.add(major)
        list.push({
          version: v,
          architecture: 'x64',
          downloadUrl: `https://get.enterprisedb.com/postgresql/postgresql-${v}-windows-x64-binaries.zip`,
          fileSize: 310000000,
          isRecommended: major === '16',
          isLTS: true,
          releaseDate: '长期支持',
        })
      }
    }
  }

  return list
}

// 7. 在线抓取 Apache HTTP Server 版本列表
async function fetchRemoteApacheVersions(): Promise<PackageVersion[]> {
  const res = await fetchWithTimeout('https://dlcdn.apache.org/httpd/')
  if (!res.ok) throw new Error(`HTTP ${res.status}`)
  const html = await res.text()

  const re = /httpd-([0-9\.]+)\.tar\.gz/g
  let m
  const vers: string[] = []
  while ((m = re.exec(html)) !== null) {
    if (!vers.includes(m[1])) vers.push(m[1])
  }

  if (vers.length === 0) throw new Error('未解析到 Apache 版本')

  return vers.map((v) => ({
    version: v,
    architecture: 'x64',
    downloadUrl: `https://dlcdn.apache.org/httpd/binaries/win64/httpd-${v}-win64.zip`,
    fileSize: 18500000,
    isRecommended: true,
    isLTS: true,
    releaseDate: '稳定版',
  }))
}

// 根据包类型动态获取最新版本列表
export async function getPackageVersions(packageKey: string, forceRefresh: boolean = false): Promise<PackageVersion[]> {
  const key = packageKey.toLowerCase()

  // 1. 如果不强制刷新，优先返回有效缓存
  if (!forceRefresh) {
    const cached = getCachedVersions(key)
    if (cached && cached.length > 0) {
      return cached
    }
  }

  // 2. 在线实时抓取
  try {
    let result: PackageVersion[] = []

    switch (key) {
      case 'php':
        result = await fetchRemotePHPVersions()
        break
      case 'nodejs':
        result = await fetchRemoteNodeVersions()
        break
      case 'nginx':
        result = await fetchRemoteNginxVersions()
        break
      case 'redis':
        result = await fetchRemoteRedisVersions()
        break
      case 'mysql':
        result = await fetchRemoteMySQLVersions()
        break
      case 'postgresql':
        result = await fetchRemotePostgresVersions()
        break
      case 'apache':
        result = await fetchRemoteApacheVersions()
        break
      default:
        result = BASELINE_VERSIONS[key] || []
    }

    if (result && result.length > 0) {
      setCachedVersions(key, result)
      return result
    }
  } catch (err) {
    console.warn(`[DevEnv Store] 动态获取 ${packageKey} 远程版本失败，回退到本地基准版本:`, err)
  }

  // 3. 网络异常时兜底回退到基准配置
  return BASELINE_VERSIONS[key] || []
}

// 获取软件商店列表（动态注入推荐版本与安装情况）
export async function listPackages(forceRefresh: boolean = false): Promise<PackageInfo[]> {
  const installedRuntimes = await listAllRuntimes().catch(() => [])

  // 异步获取或基于缓存获取最新推荐版本
  const packagesWithDynamicVersions = await Promise.all(
    CATALOG_BASE.map(async (pkg) => {
      let recVer = pkg.recommendedVersion
      try {
        const versions = await getPackageVersions(pkg.key, forceRefresh)
        if (versions.length > 0) {
          const rec = versions.find((v) => v.isRecommended) || versions[0]
          if (rec) {
            recVer = rec.version
          }
        }
      } catch {
        // keep base
      }

      const installed = installedRuntimes.filter(
        (r) => r.runtimeType.toLowerCase() === pkg.key.toLowerCase() && r.status !== 'removed'
      )

      return {
        ...pkg,
        recommendedVersion: recVer,
        isInstalled: installed.length > 0,
        installedVersions: installed.map((r) => r.version),
      }
    })
  )

  return packagesWithDynamicVersions
}

export interface InstallPackagePayload {
  packageKey: string
  version: string
  variant?: string
  architecture: string
  downloadUrl: string
  installPath: string
  mirror?: string
}

export async function installPackage(payload: InstallPackagePayload): Promise<void> {
  const installPath =
    payload.installPath ||
    `runtimes\\${payload.packageKey}\\${payload.version}${payload.variant ? '-' + payload.variant : ''}-${payload.architecture}`

  const { registerRuntime } = await import('./runtimes')
  await registerRuntime(
    payload.packageKey,
    installPath,
    payload.version,
    payload.architecture,
    payload.variant
  )
}
