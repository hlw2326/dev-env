import type { PackageInfo, PackageVersion } from '../types'
import { listAllRuntimes } from './runtimes'

// 静态软件包目录定义（符合设计规范第 2.1 节与第 3.2 节）
const CATALOG: Omit<PackageInfo, 'isInstalled' | 'installedVersions'>[] = [
  {
    key: 'php',
    displayName: 'PHP',
    category: 'language',
    description: '通用开源脚本语言，尤其适用于 Web 开发与服务端编程',
    iconClass: 'pi pi-code',
    brandColor: '#777BB4',
    recommendedVersion: '8.3.12',
    officialWebsite: 'https://www.php.net',
  },
  {
    key: 'nginx',
    displayName: 'Nginx',
    category: 'webserver',
    description: '高性能 HTTP 和反向代理 Web 服务器，兼具 IMAP/POP3 代理功能',
    iconClass: 'pi pi-server',
    brandColor: '#009639',
    recommendedVersion: '1.26.0',
    officialWebsite: 'https://nginx.org',
  },
  {
    key: 'mysql',
    displayName: 'MySQL',
    category: 'database',
    description: '全球广泛使用的开源关系型数据库管理系统 (RDBMS)',
    iconClass: 'pi pi-database',
    brandColor: '#4479A1',
    recommendedVersion: '8.4.0',
    officialWebsite: 'https://www.mysql.com',
  },
  {
    key: 'redis',
    displayName: 'Redis',
    category: 'cache',
    description: '内存级键值数据结构存储系统，可用作数据库、缓存和消息中间件',
    iconClass: 'pi pi-bolt',
    brandColor: '#DC382D',
    recommendedVersion: '7.2.4',
    officialWebsite: 'https://redis.io',
  },
  {
    key: 'postgresql',
    displayName: 'PostgreSQL',
    category: 'database',
    description: '功能强大的企业级开源对象-关系型数据库系统',
    iconClass: 'pi pi-database',
    brandColor: '#336791',
    recommendedVersion: '16.2',
    officialWebsite: 'https://www.postgresql.org',
  },
  {
    key: 'apache',
    displayName: 'Apache HTTP',
    category: 'webserver',
    description: '经典模块化 Web 服务器 (Apache 2.4 / httpd)',
    iconClass: 'pi pi-compass',
    brandColor: '#D22128',
    recommendedVersion: '2.4.58',
    officialWebsite: 'https://httpd.apache.org',
  },
  {
    key: 'nodejs',
    displayName: 'Node.js',
    category: 'language',
    description: '基于 Chrome V8 引擎的高并发异步事件驱动 JavaScript 运行时',
    iconClass: 'pi pi-code',
    brandColor: '#339933',
    recommendedVersion: '20.12.2',
    officialWebsite: 'https://nodejs.org',
  },
]

// 各软件包可用版本定义
const PACKAGE_VERSIONS: Record<string, PackageVersion[]> = {
  php: [
    {
      version: '8.3.12',
      variant: 'nts',
      architecture: 'x64',
      downloadUrl: 'https://windows.php.net/downloads/releases/php-8.3.12-nts-Win32-vs16-x64.zip',
      mirrorUrl: 'https://mirrors.aliyun.com/php/releases/php-8.3.12-nts-Win32-vs16-x64.zip',
      fileSize: 31800000,
      isRecommended: true,
      isLTS: false,
      releaseDate: '2024-09-26',
    },
    {
      version: '8.2.25',
      variant: 'nts',
      architecture: 'x64',
      downloadUrl: 'https://windows.php.net/downloads/releases/php-8.2.25-nts-Win32-vs16-x64.zip',
      mirrorUrl: 'https://mirrors.aliyun.com/php/releases/php-8.2.25-nts-Win32-vs16-x64.zip',
      fileSize: 31200000,
      isRecommended: false,
      isLTS: true,
      releaseDate: '2024-09-26',
    },
    {
      version: '8.1.30',
      variant: 'nts',
      architecture: 'x64',
      downloadUrl: 'https://windows.php.net/downloads/releases/php-8.1.30-nts-Win32-vs16-x64.zip',
      mirrorUrl: 'https://mirrors.aliyun.com/php/releases/php-8.1.30-nts-Win32-vs16-x64.zip',
      fileSize: 30500000,
      isRecommended: false,
      isLTS: true,
      releaseDate: '2024-09-26',
    },
    {
      version: '7.4.33',
      variant: 'nts',
      architecture: 'x64',
      downloadUrl: 'https://windows.php.net/downloads/releases/archives/php-7.4.33-nts-Win32-vc15-x64.zip',
      mirrorUrl: 'https://mirrors.aliyun.com/php/releases/php-7.4.33-nts-Win32-vc15-x64.zip',
      fileSize: 25400000,
      isRecommended: false,
      isLTS: false,
      releaseDate: '2022-11-02',
    },
  ],
  nginx: [
    {
      version: '1.26.0',
      architecture: 'x64',
      downloadUrl: 'https://nginx.org/download/nginx-1.26.0.zip',
      mirrorUrl: 'https://mirrors.aliyun.com/nginx/nginx-1.26.0.zip',
      fileSize: 2100000,
      isRecommended: true,
      isLTS: true,
      releaseDate: '2024-04-23',
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
      version: '8.4.0',
      architecture: 'x64',
      downloadUrl: 'https://dev.mysql.com/get/Downloads/MySQL-8.4/mysql-8.4.0-winx64.zip',
      mirrorUrl: 'https://mirrors.aliyun.com/mysql/MySQL-8.4/mysql-8.4.0-winx64.zip',
      fileSize: 245000000,
      isRecommended: true,
      isLTS: true,
      releaseDate: '2024-04-30',
    },
    {
      version: '8.0.36',
      architecture: 'x64',
      downloadUrl: 'https://dev.mysql.com/get/Downloads/MySQL-8.0/mysql-8.0.36-winx64.zip',
      mirrorUrl: 'https://mirrors.aliyun.com/mysql/MySQL-8.0/mysql-8.0.36-winx64.zip',
      fileSize: 232000000,
      isRecommended: false,
      isLTS: true,
      releaseDate: '2024-01-16',
    },
  ],
  redis: [
    {
      version: '7.2.4',
      architecture: 'x64',
      downloadUrl: 'https://github.com/tporadowski/redis/releases/download/v7.2.4/Redis-x64-7.2.4.zip',
      mirrorUrl: 'https://mirrors.aliyun.com/redis/Redis-x64-7.2.4.zip',
      fileSize: 8500000,
      isRecommended: true,
      isLTS: true,
      releaseDate: '2024-01-10',
    },
  ],
  postgresql: [
    {
      version: '16.2',
      architecture: 'x64',
      downloadUrl: 'https://get.enterprisedb.com/postgresql/postgresql-16.2-1-windows-x64-binaries.zip',
      mirrorUrl: 'https://mirrors.aliyun.com/postgresql/postgresql-16.2-1-windows-x64-binaries.zip',
      fileSize: 310000000,
      isRecommended: true,
      isLTS: true,
      releaseDate: '2024-02-08',
    },
  ],
  apache: [
    {
      version: '2.4.58',
      architecture: 'x64',
      downloadUrl: 'https://www.apachelounge.com/download/VS17/binaries/httpd-2.4.58-231018-win64-VS17.zip',
      fileSize: 18000000,
      isRecommended: true,
      isLTS: true,
      releaseDate: '2023-10-19',
    },
  ],
  nodejs: [
    {
      version: '20.12.2',
      architecture: 'x64',
      downloadUrl: 'https://nodejs.org/dist/v20.12.2/node-v20.12.2-win-x64.zip',
      mirrorUrl: 'https://npmmirror.com/mirrors/node/v20.12.2/node-v20.12.2-win-x64.zip',
      fileSize: 34000000,
      isRecommended: true,
      isLTS: true,
      releaseDate: '2024-04-10',
    },
    {
      version: '18.20.2',
      architecture: 'x64',
      downloadUrl: 'https://nodejs.org/dist/v18.20.2/node-v18.20.2-win-x64.zip',
      mirrorUrl: 'https://npmmirror.com/mirrors/node/v18.20.2/node-v18.20.2-win-x64.zip',
      fileSize: 31000000,
      isRecommended: false,
      isLTS: true,
      releaseDate: '2024-04-10',
    },
  ],
}

export async function listPackages(): Promise<PackageInfo[]> {
  const installedRuntimes = await listAllRuntimes().catch(() => [])

  return CATALOG.map((pkg) => {
    const installed = installedRuntimes.filter(
      (r) => r.runtimeType.toLowerCase() === pkg.key.toLowerCase() && r.status !== 'removed'
    )
    return {
      ...pkg,
      isInstalled: installed.length > 0,
      installedVersions: installed.map((r) => r.version),
    }
  })
}

export async function getPackageVersions(packageKey: string): Promise<PackageVersion[]> {
  const key = packageKey.toLowerCase()
  return PACKAGE_VERSIONS[key] || []
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
  // 模拟并向运行环境注册，或调用 Wails 对应接口
  const installPath =
    payload.installPath ||
    `runtimes\\${payload.packageKey}\\${payload.version}${payload.variant ? '-' + payload.variant : ''}-${payload.architecture}`

  // 调用内部存储/环境持久化注册
  const { registerRuntime } = await import('./runtimes')
  await registerRuntime(
    payload.packageKey,
    installPath,
    payload.version,
    payload.architecture,
    payload.variant
  )
}
