import {
	WindowMinimise,
	WindowToggleMaximise,
	WindowIsMaximised,
	Quit,
} from '../../wailsjs/runtime/runtime'

// 最小化窗口
export const minWin = WindowMinimise

// 切换最大化窗口
export const toggleMaxWin = WindowToggleMaximise

// 查询窗口是否最大化
export const isMaxWin = WindowIsMaximised

// 退出应用
export const quitWin = Quit
