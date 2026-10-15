import type { Software } from '@/types'

export interface HServiceConfigProps {
	item: Software | null
	busy: boolean
}

export type HServiceConfigEmits = {
	start: [item: Software]
	stop: [item: Software]
	main: [item: Software]
	autostart: [item: Software, on: boolean]
	folder: [id: number]
}
