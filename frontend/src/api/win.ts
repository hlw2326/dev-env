import {
	WindowMinimise,
	WindowToggleMaximise,
	WindowIsMaximised,
	WindowHide,
	EventsOn,
} from '../../wailsjs/runtime/runtime'

// 最小化窗口
export const minWin = WindowMinimise

// 切换最大化窗口
export const toggleWin = WindowToggleMaximise

// 查询窗口是否最大化
export const checkWin = WindowIsMaximised

// 隐藏窗口并保留托盘进程
export const hideWin = WindowHide

// 监听后端事件
export const listenEvent = EventsOn
