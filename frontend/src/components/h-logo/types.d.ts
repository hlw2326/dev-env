// 软件图标元信息
export interface LogoMeta {
	// 简写文本
	short: string
	// 色调标识
	tone: string
	// 图标文件路径
	logo: string
}

// 软件图标组件属性
export interface HLogoProps {
	// 软件类型标识
	kind: string
	// 尺寸规格
	size?: 'small' | 'medium' | 'large'
}
