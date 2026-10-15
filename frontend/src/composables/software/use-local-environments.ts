import { computed, reactive, ref } from 'vue'
import { scanEnvironments, pickFile } from '@/api'
import { formatError } from '@/utils'
import type { CommandEnvironment } from '@/types'

const visible = ref(false)
const selectedKind = ref('')
const loading = ref(false)
const activeKind = ref('')
const allScanned = ref(false)
const outdated = reactive(new Set<string>())
const errors = reactive<Record<string, string>>({})
const reports = reactive<Record<string, CommandEnvironment>>({})

function describeError(cause: unknown): string {
	const message = formatError(cause)
	if (message.includes('ScanEnvironments') && message.includes('not a function')) {
		return `当前运行的 DevEnv 尚未加载本机环境接口，请从托盘退出后启动最新版本。\n${message}`
	}
	return message
}

// 仅在用户点击查找或选择程序时执行，不随商店加载、刷新或安装事件启动。
async function scanExisting(kind: string, executable = ''): Promise<void> {
	if (loading.value) return
	loading.value = true
	activeKind.value = kind
	errors[kind] = ''
	for (const key of Object.keys(reports)) {
		if (kind === '' || kind === key) outdated.add(key)
	}
	try {
		const result = await scanEnvironments(kind, executable)
		if (kind === '' && executable === '') {
			for (const key of Object.keys(reports)) delete reports[key]
			outdated.clear()
		}
		for (const item of result) {
			const previous = reports[item.kind]
			if (executable !== '' && previous !== undefined) {
				const added = new Set(item.entries.map((entry) => entry.path.toLowerCase()))
				item.entries = [...previous.entries.filter((entry) => !added.has(entry.path.toLowerCase())), ...item.entries]
				item.conflict = previous.conflict
				item.warnings = previous.warnings
			}
			reports[item.kind] = item
			outdated.delete(item.kind)
		}
		if (kind === '') allScanned.value = true
	} catch (cause: unknown) {
		errors[kind] = describeError(cause)
	} finally {
		loading.value = false
		activeKind.value = ''
	}
}

async function selectExecutable(kind: string): Promise<void> {
	if (kind === '' || loading.value) return
	errors[kind] = ''
	try {
		const path = await pickFile('选择已有环境的程序文件，例如 node.exe、php.exe 或 mvn.cmd')
		if (path !== '') await scanExisting(kind, path)
	} catch (cause: unknown) {
		errors[kind] = describeError(cause)
	}
}

function openEnvironments(kind = ''): void {
	selectedKind.value = kind
	visible.value = true
}

export function invalidateEnvironments(): void {
	for (const kind of Object.keys(reports)) outdated.add(kind)
}

export function externalVersions(kind: string): string[] {
	const report = reports[kind]
	if (outdated.has(kind) || report === undefined) return []
	return [...new Set(report.entries.filter((entry) => {
		if (entry.managed || entry.error !== '' || entry.role === '客户端') return false
		if (kind === 'java' && entry.command !== 'javac') return false
		if (kind === 'dotnet' && entry.role !== 'SDK') return false
		return true
	}).flatMap((entry) => entry.versions))]
}

export function hasExternalVersion(kind: string, version: string): boolean {
	return externalVersions(kind).includes(version)
}

export function useLocalEnvironments() {
	const scanned = computed(() => {
		if (selectedKind.value === '') return allScanned.value
		return reports[selectedKind.value] !== undefined
	})
	const stale = computed(() => {
		if (selectedKind.value === '') return outdated.size > 0
		return outdated.has(selectedKind.value)
	})
	const displayedReports = computed(() => Object.values(reports).filter((item) => {
		if (selectedKind.value !== '' && item.kind !== selectedKind.value) return false
		return item.entries.some((entry) => !entry.managed) || item.error !== '' || item.warnings.length > 0
	}))
	const detectedCount = computed(() => Object.values(reports).filter((item) => item.entries.some((entry) => !entry.managed && entry.versions.length > 0)).length)
	const managerError = computed(() => {
		const value = errors[selectedKind.value]
		if (value === undefined) return ''
		return value
	})
	function findReport(kind: string): CommandEnvironment | undefined { return reports[kind] }
	function findError(kind: string): string {
		const error = errors[kind]
		if (error === undefined) return ''
		return error
	}
	return { visible, selectedKind, loading, activeKind, scanned, stale, displayedReports, detectedCount, managerError, openEnvironments, scanExisting, selectExecutable, findReport, findError }
}
