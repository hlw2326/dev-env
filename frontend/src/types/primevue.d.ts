declare module 'primevue/toasteventbus' {
	// Toast 事件总线实例接口
	interface ToastBus {
		// 绑定事件监听
		on(type: string, handler: (...args: unknown[]) => void): void
		// 解绑事件监听
		off(type: string, handler: (...args: unknown[]) => void): void
		// 触发事件通知
		emit(type: string, ...args: unknown[]): void
	}

	// 全局 Toast 事件总线
	const ToastEventBus: ToastBus
	export default ToastEventBus
}
