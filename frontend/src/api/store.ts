import type { PackageInfo, PackageVersion, InstallResult } from '../types'
import { listAllRuntimes } from './runtimes'

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

export async function getPackageVersions(packageKey: string): Promise<PackageVersion[]> {
  const key = packageKey.toLowerCase()
  const versions = BASELINE_VERSIONS[key]
  if (!versions) {
    return []
  }
  return versions
}

export async function listPackages(): Promise<PackageInfo[]> {
  const runtimes = await listAllRuntimes()

  return Promise.all(
    CATALOG_BASE.map(async (pkg) => {
      let recVer = pkg.recommendedVersion
      const versions = await getPackageVersions(pkg.key)
      if (versions.length > 0) {
        const rec = versions.find((v) => v.isRecommended)
        if (rec) {
          recVer = rec.version
        } else {
          recVer = versions[0].version
        }
      }

      const installed = runtimes.filter(
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

export async function installPackage(payload: InstallPackagePayload): Promise<InstallResult> {
  const RuntimeAPI = await import('../../wailsjs/go/wails/RuntimeAPI')
  const res = await RuntimeAPI.Install({
    runtimeType: payload.packageKey.toLowerCase(),
    version: payload.version,
    variant: payload.variant || '',
    architecture: payload.architecture || 'x64',
    downloadUrl: payload.downloadUrl,
    installPath: payload.installPath || '',
    mirror: payload.mirror || '',
  })
  return {
    runtimeId: res.runtimeId,
    serviceId: res.serviceId,
    serviceCreated: res.serviceCreated,
  }
}


