// 标签颜色
export type HTagColor =
	| 'default'
	| 'success'
	| 'info'
	| 'warning'
	| 'danger'
	| 'secondary'
	| 'contrast'

// 标签尺寸
export type HTagSize = 'small' | 'medium' | 'large'

// 标签圆角
export type HTagRadius = 'small' | 'large' | 'full'

// 标签组件属性
export interface HTagProps {
	// 标签文本
	value?: string
	// 项目语义颜色
	color?: HTagColor
	// PrimeVue severity 兼容属性，color 未设置时生效
	severity?: 'success' | 'info' | 'warn' | 'danger' | 'secondary' | 'contrast'
	// 标签字号
	size?: HTagSize
	// 标签圆角
	radius?: HTagRadius
	// PrimeIcons 图标
	icon?: string
	// 是否显示状态圆点
	dot?: boolean
}
