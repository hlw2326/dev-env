import type { Software } from '@/types'

// 服务行组件属性
export interface HServiceProps {
	// 软件服务信息
	item: Software
	// 当前服务是否正在执行操作
	busy: boolean
	// 是否显示服务设置入口
	configurable?: boolean
}
