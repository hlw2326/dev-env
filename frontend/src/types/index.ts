export interface Result {
	ok: boolean
	message: string
}

export interface SaveConfigResult {
	ok: boolean
	message: string
	restart: boolean
}

export interface Site {
	id: number
	name: string
	kind: string
	hosts: string[]
	root: string
	run: string
	path: string
	software?: number
	version: string
	command: string
	port: number
	autostart: boolean
	state: string
	url: string
	need: string
	created: string
	updated: string
}

export interface CreateSiteInput {
	kind: string
	name: string
	host: string
	root: string
	run: string
	command: string
}

export interface UpdateSiteInput {
	id: number
	hosts: string[]
	root: string
	run: string
	software?: number
	command: string
	autostart: boolean
}

export interface MarketItem {
	kind: string
	name: string
	description: string
	category: string
	count: number
	recommended: string
}

export interface Software {
	id: number
	kind: string
	version: string
	path: string
	state: string
	progress: number
	error: string
	main: boolean
	port: number
	autostart: boolean
	running: boolean
	pid: number
}

export interface SoftwareVersion {
	version: string
	recommended: boolean
	installed: boolean
	note: string
}

export interface ConfigItem {
	key: string
	label: string
	value: string
	description: string
}

export interface SoftwareConfig {
	items: ConfigItem[]
	file: string
}

export interface SaveConfigInput {
	id: number
	items: Record<string, string>
}

export interface Extension {
	name: string
	enabled: boolean
}

export interface InstallInput {
	kind: string
	version: string
}

export interface Settings {
	theme: string
	root: string
	port: number
	hosts: boolean
	listening: boolean
	error: string
}

export interface SaveSettingsInput {
	theme: string
	root: string
	port: number
	hosts: boolean
}

export interface TaskProgressEvent {
	kind: string
	version: string
	progress: number
	message: string
}

export interface TaskCompletedEvent {
	kind: string
	version: string
	ok: boolean
	error: string
}

export interface ServiceStatusEvent {
	id: number
	running: boolean
	pid: number
	message: string
}

export interface SiteStatusEvent {
	id: number
	state: string
	message: string
}
