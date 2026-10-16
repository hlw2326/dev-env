import { ref } from 'vue'
import { listSoftware } from '@/api'
import { formatError } from '@/utils'
import type { Software } from '@/types'

const items = ref<Software[]>([])
const ready = ref(false)
const checking = ref(true)
const error = ref('')
let pending: Promise<Software[]> | null = null

// 启动、首页和服务管理共用一次服务检测，未知状态不能当成已停止。
export function useServiceStatus() {
	async function refresh(silent = false): Promise<Software[]> {
		if (pending !== null) return pending
		if (!silent || !ready.value) checking.value = true
		pending = (async () => {
			try {
				const result = await listSoftware()
				items.value = result
				ready.value = true
				error.value = ''
				return result
			} catch (failure: unknown) {
				error.value = formatError(failure)
				throw failure
			} finally {
				checking.value = false
				pending = null
			}
		})()
		return pending
	}
	return { items, ready, checking, error, refresh }
}
