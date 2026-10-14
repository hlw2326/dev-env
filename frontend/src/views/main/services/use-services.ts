import { ref, onMounted, onUnmounted } from 'vue'
import { useConfirm } from 'primevue/useconfirm'
import {
	listServices,
	startService,
	stopService,
	restartService,
	getServiceLogs,
	deleteService,
	switchServiceRuntime,
	getServiceRuntimeOptions
} from '@/api/services'
import { listAllRuntimes } from '@/api/runtimes'
import type { ServiceInstance, RuntimeOption } from '@/types'

// 服务实例列表
const services = ref<ServiceInstance[]>([])

// 每个服务实例对应的可切换运行时版本列表
const runtimeOptions = ref<Record<number, RuntimeOption[]>>({})

// 列表加载状态
const loading = ref(false)

// 当前正在操作的服务编号
const busyId = ref<number | null>(null)

// 当前查看日志的服务实例
const curSrv = ref<ServiceInstance | null>(null)

// 控制台日志列表
const logs = ref<string[]>([])

// 加载服务状态列表及各服务可选版本
async function loadServices() {
	loading.value = true
	try {
		// 先触发一次运行时列表查询，确保旧版本 localStorage 数据已迁移至后端
		await listAllRuntimes()
		const list = await listServices()
		services.value = list

		const optMap: Record<number, RuntimeOption[]> = {}
		await Promise.all(
			list.map(async (srv) => {
				try {
					optMap[srv.id] = await getServiceRuntimeOptions(srv.id)
				} catch {
					optMap[srv.id] = []
				}
			})
		)
		runtimeOptions.value = optMap
	} finally {
		loading.value = false
	}
}

// 启动指定服务
async function startSrv(id: number) {
	busyId.value = id
	try {
		await startService(id)
		toast.success('服务启动成功')
		await loadServices()
	} catch (err) {
		toast.error(String(err))
	} finally {
		busyId.value = null
	}
}

// 停止指定服务
async function stopSrv(id: number) {
	busyId.value = id
	try {
		await stopService(id)
		toast.info('服务已停止')
		await loadServices()
	} catch (err) {
		toast.error(String(err))
	} finally {
		busyId.value = null
	}
}

// 重启指定服务
async function restartSrv(id: number) {
	busyId.value = id
	try {
		await restartService(id)
		toast.success('服务重启成功')
		await loadServices()
	} catch (err) {
		toast.error(String(err))
	} finally {
		busyId.value = null
	}
}

// 刷新当前服务日志
async function fetchLogs() {
	if (!curSrv.value) {
		return
	}
	logs.value = await getServiceLogs(curSrv.value.id, 200)
}

// 打开指定服务日志面板
async function openLogs(srv: ServiceInstance) {
	curSrv.value = srv
	await fetchLogs()
}

// 关闭日志面板
function closeLogs() {
	curSrv.value = null
}

// 导出服务管理逻辑
export function useServices() {
	// 气泡确认框实例
	const confirm = useConfirm()

	// 删除指定服务实例
	function removeSrv(event: Event, srv: ServiceInstance) {
		confirm.require({
			target: event.currentTarget as HTMLElement,
			message: `确定要删除服务实例「${srv.name}」吗？`,
			icon: 'pi pi-info-circle',
			rejectProps: {
				label: '取消',
				severity: 'secondary',
				outlined: true,
				size: 'small'
			},
			acceptProps: {
				label: '删除',
				severity: 'danger',
				size: 'small'
			},
			accept: async () => {
				busyId.value = srv.id
				try {
					await deleteService(srv.id)
					if (curSrv.value?.id === srv.id) {
						curSrv.value = null
					}
					toast.info(`服务实例「${srv.name}」已删除`)
					await loadServices()
				} catch (err) {
					toast.error(String(err))
				} finally {
					busyId.value = null
				}
			}
		})
	}

	// 切换服务实例绑定的运行时版本
	async function switchVersion(
		srv: ServiceInstance,
		targetRuntimeId: number,
		targetEl?: HTMLElement | null
	) {
		if (targetRuntimeId === srv.runtimeId) {
			return
		}
		const opts = runtimeOptions.value[srv.id] || []
		const targetOpt = opts.find((o) => o.runtimeId === targetRuntimeId)
		const verLabel = targetOpt ? targetOpt.version : `#${targetRuntimeId}`

		// 已停止状态：直接调用 SwitchRuntime
		if (srv.status !== 'running') {
			busyId.value = srv.id
			try {
				await switchServiceRuntime(srv.id, targetRuntimeId)
				toast.success(`${srv.name} 已切换至版本 ${verLabel}`)
				await loadServices()
			} catch (err) {
				toast.error(String(err))
			} finally {
				busyId.value = null
			}
			return
		}

		// 运行中状态：弹窗确认后执行 Stop -> SwitchRuntime -> Start
		confirm.require({
			target: targetEl || (document.activeElement as HTMLElement),
			message: `将先停止「${srv.name}」服务，切换至版本 ${verLabel} 后重新启动，是否继续？`,
			icon: 'pi pi-exclamation-triangle',
			rejectProps: {
				label: '取消',
				severity: 'secondary',
				outlined: true,
				size: 'small'
			},
			acceptProps: {
				label: '切换并重启',
				severity: 'primary',
				size: 'small'
			},
			accept: async () => {
				busyId.value = srv.id
				try {
					await stopService(srv.id)
					await switchServiceRuntime(srv.id, targetRuntimeId)
					try {
						await startService(srv.id)
						toast.success(`${srv.name} 已切换至版本 ${verLabel} 并重新启动`)
					} catch {
						toast.warn('版本已切换，但新版本启动失败，请查看日志')
					}
					await loadServices()
				} catch (err) {
					toast.error(String(err))
					await loadServices()
				} finally {
					busyId.value = null
				}
			}
		})
	}

	const offFns: Array<() => void> = []

	onMounted(async () => {
		await loadServices()
		try {
			const wailsRt = await import('../../../../wailsjs/runtime/runtime')
			for (const evName of [
				'service:created',
				'service:updated',
				'service:deleted',
				'service:status-changed'
			]) {
				offFns.push(
					wailsRt.EventsOn(evName, () => {
						loadServices()
					})
				)
			}
		} catch {
			// Ignore if not in Wails runtime environment
		}
	})

	onUnmounted(() => {
		while (offFns.length > 0) {
			const fn = offFns.pop()
			if (fn) fn()
		}
	})

	return {
		services,
		runtimeOptions,
		loading,
		busyId,
		curSrv,
		logs,
		loadServices,
		startSrv,
		stopSrv,
		restartSrv,
		removeSrv,
		switchVersion,
		openLogs,
		fetchLogs,
		closeLogs
	}
}


