/**
 * 弹窗组件参数
 */
export interface HModalProps {
  /**
   * 弹窗标题
   */
  title?: string
  /**
   * 弹窗宽度
   */
  width?: string | number
  /**
   * 模态遮罩
   */
  modal?: boolean
  /**
   * 允许关闭
   */
  closable?: boolean
  /**
   * 退出按键关闭
   */
  closeOnEscape?: boolean
  /**
   * 遮罩点击关闭
   */
  dismissableMask?: boolean
  /**
   * 显示底部栏
   */
  footer?: boolean
  /**
   * 确定按钮文本
   */
  okText?: string
  /**
   * 确定按钮图标
   */
  okIcon?: string
  /**
   * 取消按钮文本
   */
  cancelText?: string
  /**
   * 加载状态
   */
  loading?: boolean
}

/**
 * 弹窗事件
 */
export type HModalEmits = {
  (event: 'ok'): void
  (event: 'cancel'): void
}

/**
 * 弹窗实例方法
 */
export interface HModalInst {
  open: () => void
  close: () => void
}
