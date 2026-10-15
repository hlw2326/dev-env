export interface HlwDrawerProps {
	title: string
	width?: string
	position?: 'left' | 'right'
	modal?: boolean
	closable?: boolean
	maskClose?: boolean
	closeOnEscape?: boolean
	padding?: boolean
	footer?: boolean
	okText?: string
	okIcon?: string
	cancelText?: string
	loading?: boolean
}

export type HlwDrawerEmits = {
	ok: []
	cancel: []
	open: []
	close: []
}

export interface HlwDrawerExpose {
	open: () => void
	close: () => void
}
