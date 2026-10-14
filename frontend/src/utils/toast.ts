import ToastEventBus from 'primevue/toasteventbus'
import type { ToastMessageOptions } from 'primevue/toast'

/**
 * 提示级别类型
 */
export type ToastType = 'info' | 'success' | 'warn' | 'error' | 'secondary' | 'contrast'

/**
 * Toast 全局配置
 */
export interface ToastConfig {
  /**
   * 自动关闭时长 (ms)
   */
  duration?: number
  /**
   * 分组标识
   */
  group?: string
  /**
   * 同一时间是否只保留一条最新消息（避免多个堆叠）
   */
  single?: boolean
}

let config: ToastConfig = { duration: 2500, single: true }

/**
 * 设置 Toast 全局配置
 */
export function setupToast(options: ToastConfig = {}) {
  config = { ...config, ...options }
}

/**
 * 清除所有或指定分组提示
 * @param group 分组标识，缺省清除全部
 */
export function clearToast(group?: string) {
  if (group) {
    ToastEventBus.emit('remove-group', group)
  } else {
    ToastEventBus.emit('remove-all-groups')
  }
}

/**
 * 发送一条 Toast 提示
 */
function show(options: ToastMessageOptions) {
  const group = options.group ?? config.group
  if (config.single) {
    clearToast(group)
  }
  ToastEventBus.emit('add', {
    closable: false,
    ...options,
    severity: options.severity ?? 'info',
    life: options.life ?? config.duration,
    group
  })
}

/**
 * 生成快捷级别函数
 */
const level = (severity: ToastType) => (message: string, detail?: string, life?: number) => {
  show({ severity, summary: message, detail, life })
}

/**
 * 全局统一 Toast 调用函数（参考 zbzs/ele 最佳实践）
 *
 * 使用示例：
 * - toast('操作成功', 'success')
 * - toast.success('保存成功', '配置已写入')
 * - toast.error('启动失败')
 * - toast.warn('请先停止服务')
 * - toast.info('正在处理中...')
 * - toast.clear()
 * - toast({ severity: 'success', summary: '成功', detail: '已连接' })
 */
export const toast = Object.assign(
  (messageOrOptions: string | ToastMessageOptions, type: ToastType = 'info') => {
    if (typeof messageOrOptions === 'string') {
      show({ severity: type, summary: messageOrOptions })
    } else {
      show(messageOrOptions)
    }
  },
  {
    info: level('info'),
    success: level('success'),
    warn: level('warn'),
    error: level('error'),
    clear: clearToast
  }
)

export default toast
