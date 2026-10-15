export function formatBytes(value: number): string {
	if (value < 1024) {
		return `${value} B`
	}
	const units = ['KB', 'MB', 'GB', 'TB']
	let amount = value
	let unit = 'B'
	for (const candidate of units) {
		amount /= 1024
		unit = candidate
		if (amount < 1024 || candidate === units[units.length - 1]) {
			break
		}
	}
	return `${amount.toFixed(amount >= 10 ? 0 : 1)} ${unit}`
}

export function formatDuration(seconds: number): string {
	if (seconds < 60) {
		return `${Math.max(0, Math.floor(seconds))}秒`
	}
	const minutes = Math.floor(seconds / 60)
	if (minutes < 60) {
		return `${minutes}分`
	}
	const hours = Math.floor(minutes / 60)
	if (hours < 24) {
		return `${hours}小时${minutes % 60}分`
	}
	return `${Math.floor(hours / 24)}天${hours % 24}小时`
}
