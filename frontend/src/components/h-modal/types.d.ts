// 弹窗组件属性
export interface HModalProps {
	// 弹窗标题
	title?: string
	// 弹窗宽度
	width?: string
	// 是否模态遮罩
	modal?: boolean
	// 是否允许关闭
	closable?: boolean
	// 是否点击遮罩关闭
	maskClose?: boolean
	// 是否显示底部栏
	footer?: boolean
	// 确定按钮文本
	okText?: string
	// 确定按钮图标
	okIcon?: string
	// 取消按钮文本
	cancelText?: string
	// 加载中状态
	loading?: boolean
}

// 弹窗组件事件
export type HModalEmits = {
	(event: 'ok'): void
	(event: 'cancel'): void
}

// 弹窗实例暴露方法
export interface HModalExpose {
	// 打开弹窗
	open: () => void
	// 关闭弹窗
	close: () => void
}
