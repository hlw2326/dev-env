import { ref, onMounted } from 'vue'
import { useConfirm } from 'primevue/useconfirm'
import { listTasks, cancelTask } from '@/api/tasks'
import type { Task } from '@/types'

// 任务列表状态
const tasks = ref<Task[]>([])

// 列表加载状态
const loading = ref(false)

// 加载所有任务
async function loadTasks() {
	loading.value = true
	try {
		tasks.value = await listTasks()
	} finally {
		loading.value = false
	}
}

// 导出任务逻辑
export function useTasks() {
	// 气泡确认框实例
	const confirm = useConfirm()

	// 取消指定任务
	function stopTask(event: Event, id: number) {
		confirm.require({
			target: event.currentTarget as HTMLElement,
			message: `确定要取消任务 #${id} 吗？`,
			icon: 'pi pi-exclamation-triangle',
			rejectProps: {
				label: '取消',
				severity: 'secondary',
				outlined: true,
				size: 'small'
			},
			acceptProps: {
				label: '取消任务',
				severity: 'danger',
				size: 'small'
			},
			accept: async () => {
				try {
					await cancelTask(id)
					toast.info('任务已请求取消')
					await loadTasks()
				} catch (err) {
					toast.error(String(err))
				}
			}
		})
	}

	onMounted(loadTasks)
	return {
		tasks,
		loading,
		loadTasks,
		stopTask
	}
}
