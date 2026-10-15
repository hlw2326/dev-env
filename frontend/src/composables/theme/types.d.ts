// 主题模式类型
export type ThemeMode = 'light' | 'dark' | 'system'

// 主题强调色选项
export interface ThemeColor {
	// 颜色名称
	name: string
	// 颜色色值
	value: string
}

// 主题状态结构
export interface ThemeState {
	// 色彩模式
	mode: ThemeMode
	// 强调色值
	color: string
}
