import { ref } from 'vue'
import { listRedisDatabases, listRedisKeys, loadRedisValue } from '@/api'
import { formatError, toast } from '@/utils'
import type { RedisDatabase, RedisKey, RedisKeyPage, RedisValue, Software } from '@/types'

const REDIS_PAGE_LIMIT = 100
const DEFAULT_PATTERN = '*'

const visible = ref(false)
const activeSoftware = ref<Software | null>(null)
const loadingDatabases = ref(false)
const loadingKeys = ref(false)
const loadingValue = ref(false)
const databases = ref<RedisDatabase[]>([])
const selectedDatabase = ref(0)
const pattern = ref(DEFAULT_PATTERN)
const keys = ref<RedisKey[]>([])
const cursors = ref<number[]>([0])
const pageIndex = ref(0)
const nextCursor = ref(0)
const selectedKey = ref<RedisKey | null>(null)
const value = ref<RedisValue | null>(null)

export function useRedisData() {
	function open(item: Software): void {
		activeSoftware.value = item
		visible.value = true
		void loadDatabases()
	}

	async function loadDatabases(): Promise<void> {
		if (activeSoftware.value === null || loadingDatabases.value) {
			return
		}
		loadingDatabases.value = true
		try {
			const items = await listRedisDatabases(activeSoftware.value.id)
			databases.value = items
			const selected = items.find((item) => item.index === selectedDatabase.value)
			if (selected === undefined && items.length > 0) {
				selectedDatabase.value = items[0].index
			}
			await loadKeys(true)
		} catch (cause: unknown) {
			toast.error('读取 Redis 数据库失败', formatError(cause))
		} finally {
			loadingDatabases.value = false
		}
	}

	async function loadKeys(reset: boolean): Promise<void> {
		if (activeSoftware.value === null || loadingKeys.value) {
			return
		}
		if (reset) {
			cursors.value = [0]
			pageIndex.value = 0
			nextCursor.value = 0
			keys.value = []
			selectedKey.value = null
			value.value = null
		}
		const cursor = cursors.value[pageIndex.value]
		if (cursor === undefined) {
			toast.error('读取 Redis 键失败', '分页游标不存在')
			return
		}
		loadingKeys.value = true
		try {
			const result: RedisKeyPage = await listRedisKeys({
				id: activeSoftware.value.id,
				database: selectedDatabase.value,
				pattern: pattern.value,
				cursor,
				limit: REDIS_PAGE_LIMIT,
			})
			keys.value = result.keys
			nextCursor.value = result.cursor
		} catch (cause: unknown) {
			toast.error('读取 Redis 键失败', formatError(cause))
		} finally {
			loadingKeys.value = false
		}
	}

	function selectDatabase(database: number): void {
		selectedDatabase.value = database
		void loadKeys(true)
	}

	function searchKeys(): void {
		void loadKeys(true)
	}

	function previousKeys(): void {
		if (pageIndex.value === 0 || loadingKeys.value) {
			return
		}
		pageIndex.value -= 1
		void loadKeys(false)
	}

	function nextKeys(): void {
		if (nextCursor.value === 0 || loadingKeys.value) {
			return
		}
		const next = nextCursor.value
		cursors.value = [...cursors.value.slice(0, pageIndex.value + 1), next]
		pageIndex.value += 1
		void loadKeys(false)
	}

	async function loadValue(item: RedisKey): Promise<void> {
		if (activeSoftware.value === null || loadingValue.value) {
			return
		}
		selectedKey.value = item
		loadingValue.value = true
		try {
			value.value = await loadRedisValue({
				id: activeSoftware.value.id,
				database: selectedDatabase.value,
				key: item.name,
				cursor: 0,
				limit: REDIS_PAGE_LIMIT,
			})
		} catch (cause: unknown) {
			value.value = null
			toast.error('读取 Redis 键内容失败', formatError(cause))
		} finally {
			loadingValue.value = false
		}
	}

	return {
		visible,
		activeSoftware,
		loadingDatabases,
		loadingKeys,
		loadingValue,
		databases,
		selectedDatabase,
		pattern,
		keys,
		pageIndex,
		nextCursor,
		selectedKey,
		value,
		open,
		loadDatabases,
		selectDatabase,
		searchKeys,
		previousKeys,
		nextKeys,
		loadValue,
	}
}
