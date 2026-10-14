declare module 'primevue/toasteventbus' {
  interface ToastEventBusInstance {
    on(type: string, handler: (...args: any[]) => void): void
    off(type: string, handler: (...args: any[]) => void): void
    emit(type: string, ...args: any[]): void
  }

  const ToastEventBus: ToastEventBusInstance
  export default ToastEventBus
}
