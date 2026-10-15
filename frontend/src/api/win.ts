import {
	WindowMinimise,
	WindowToggleMaximise,
	WindowIsMaximised,
	Quit,
	EventsOn,
} from '../../wailsjs/runtime/runtime'

// 最小化窗口
export const minWin = WindowMinimise

// 切换最大化窗口
export const toggleWin = WindowToggleMaximise

// 查询窗口是否最大化
export const checkWin = WindowIsMaximised

// 退出应用
export const quitWin = Quit

// 监听后端事件
export const listenEvent = EventsOn
