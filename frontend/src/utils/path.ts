// 清理Windows路径尾部斜杠
export function trimSlash(value: string): string {
	return value.replace(/[\\/]+$/, '')
}

// 拼接项目根路径与子名称
export function joinPath(root: string, name: string): string {
	const base = trimSlash(root.trim())
	const segment = name.trim().replace(/^[\\/]+/, '')
	if (!base) {
		return segment
	}
	if (!segment) {
		return base
	}
	return `${base}\\${segment}`
}

// 拼接根目录与相对运行目录
export function resolvePath(root: string, run: string): string {
	const base = trimSlash(root.trim())
	const sub = run.trim().replace(/^[\\/]+/, '').replace(/\//g, '\\')
	if (!sub) {
		return base
	}
	return `${base}\\${sub}`
}

// 提取错误提示文本
export function formatError(err: unknown): string {
	if (err instanceof Error) {
		return err.message
	}
	return String(err)
}
