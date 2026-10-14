/**
 * 主题模式：浅色 / 深色 / 跟随系统
 */
export type ThemeMode = 'light' | 'dark' | 'system'

/**
 * 主题强调色配置项
 */
export interface ThemeColorItem {
  name: string
  value: string
}

/**
 * 主题状态
 */
export interface ThemeState {
  mode: ThemeMode
  color: string
}
