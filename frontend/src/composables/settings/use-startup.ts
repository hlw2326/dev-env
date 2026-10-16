import { ref } from 'vue'
import { loadApplicationStartup, setApplicationStartup } from '@/api'
import { formatError, toast } from '@/utils'

const enabled = ref(false)
const ready = ref(false)
const busy = ref(false)
const error = ref('')

export function useStartup() {
	async function load(): Promise<void> {
		if (busy.value) return
		busy.value = true
		error.value = ''
		try {
			enabled.value = await loadApplicationStartup()
			ready.value = true
		} catch (err: unknown) {
			ready.value = false
			error.value = formatError(err)
		} finally { busy.value = false }
	}

	async function change(value: boolean): Promise<void> {
		if (busy.value || !ready.value) return
		busy.value = true
		try {
			const result = await setApplicationStartup(value)
			if (!result.ok) { toast.error('更新开机启动失败', result.message); return }
			enabled.value = await loadApplicationStartup()
			toast.success('已更新', result.message)
		} catch (err: unknown) {
			toast.error('更新开机启动失败', formatError(err))
		} finally { busy.value = false }
	}
	return { enabled, ready, busy, error, load, change }
}
