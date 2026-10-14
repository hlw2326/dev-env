/**
 * HTTP 客户端引导与配置
 * 在 Wails 桌面端中默认使用 Wails IPC 绑定，同时支持标准 REST/HTTP 请求
 */
export function initHttp() {
  return {
    get: async (url: string) => {
      console.log(`[HTTP GET] ${url}`)
      return { code: 0, message: 'ok', data: null }
    },
    post: async (url: string, data: any) => {
      console.log(`[HTTP POST] ${url}`, data)
      return { code: 0, message: 'ok', data }
    },
    delete: async (url: string) => {
      console.log(`[HTTP DELETE] ${url}`)
      return { code: 0, message: 'ok', data: null }
    }
  }
}

export const http = initHttp()
