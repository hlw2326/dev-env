import { computed, onUnmounted, ref, watch } from 'vue'
import { useConfirm } from 'primevue/useconfirm'
import { cancelDeployment, getDeploymentConfig, getDeploymentHistory, pickFile, previewDeployment, saveDeploymentConfig, startDeployment, testDeployment, openURL } from '@/api'
import { formatError, toast } from '@/utils'
import type { DeploymentConfig, DeploymentPlan, DeploymentTask, Project } from '@/types'

export function useDeployment(project: () => Project) {
	const config = ref<DeploymentConfig | null>(null)
	const history = ref<DeploymentTask[]>([])
	const plan = ref<DeploymentPlan | null>(null)
	const password = ref('')
	const passphrase = ref('')
	const busy = ref('')
	const error = ref('')
	const saved = ref('')
	const confirm = useConfirm()
	let generation = 0
	let disposed = false
	let timer: ReturnType<typeof setTimeout> | undefined
	const running = computed(() => history.value.some((task) => task.state === 'running'))
	const disabled = computed(() => busy.value !== '' || running.value)
	const dirty = computed(() => JSON.stringify(config.value) !== saved.value)

	function stopPolling(): void { if (timer !== undefined) clearTimeout(timer); timer = undefined }
	async function refreshHistory(): Promise<void> {
		stopPolling()
		const id = project().id
		try {
			const tasks = await getDeploymentHistory(id)
			if (disposed || id !== project().id) return
			history.value = tasks
			if (tasks.some((task) => task.state === 'running')) timer = setTimeout(() => void refreshHistory(), 1200)
		} catch (err: unknown) { toast.error('读取发布记录失败', formatError(err)) }
	}
	async function load(): Promise<void> {
		const current = ++generation
		stopPolling(); config.value = null; plan.value = null; history.value = []; password.value = ''; passphrase.value = ''
		busy.value = 'load'; error.value = ''
		try {
			const value = await getDeploymentConfig(project().id)
			if (current !== generation || disposed) return
			config.value = value; saved.value = JSON.stringify(value)
			await refreshHistory()
		} catch (err: unknown) { if (current === generation) error.value = formatError(err) }
		finally { if (current === generation) busy.value = '' }
	}
	async function save(): Promise<void> {
		if (disabled.value || config.value === null) return
		busy.value = 'save'
		const snapshot = JSON.parse(JSON.stringify(config.value)) as DeploymentConfig
		try {
			const result = await saveDeploymentConfig(snapshot)
			if (!result.ok) { toast.error('保存失败', result.message); return }
			saved.value = JSON.stringify(snapshot); plan.value = null; toast.success('配置已保存', '不会自动同步或上传')
		} catch (err: unknown) { toast.error('保存失败', formatError(err)) }
		finally { busy.value = '' }
	}
	async function preview(): Promise<void> {
		if (disabled.value || dirty.value) return
		busy.value = 'preview'; plan.value = null
		const id = project().id
		try { const value = await previewDeployment(id); if (id === project().id && !disposed) plan.value = value }
		catch (err: unknown) { toast.error('生成预览失败', formatError(err)) }
		finally { busy.value = '' }
	}
	async function test(): Promise<void> {
		if (disabled.value || dirty.value) return
		busy.value = 'test'
		try {
			const result = await testDeployment({ projectId: project().id, token: '', password: password.value, passphrase: passphrase.value })
			if (!result.ok) { toast.error('检查失败', result.message); return }; toast.success('检查完成', result.message)
		} catch (err: unknown) { toast.error('检查失败', formatError(err)) }
		finally { busy.value = ''; password.value = ''; passphrase.value = '' }
	}
	function confirmStart(event: MouseEvent): void {
		if (disabled.value || dirty.value || plan.value === null) return
		const token = plan.value.token
		const message = `将 ${plan.value.files.length} 个文件发布到 ${plan.value.target}？服务器同名文件将被覆盖，不删除额外文件、不自动回滚；微信仅上传开发版本。`
		confirm.require({ target: event.currentTarget as HTMLElement, message, icon: 'pi pi-exclamation-triangle', acceptLabel: '确认手动发布', rejectLabel: '取消', acceptProps: { severity: 'warn', size: 'small' }, rejectProps: { severity: 'secondary', variant: 'outlined', size: 'small' }, accept: () => void start(token) })
	}
	async function start(token: string): Promise<void> {
		if (disabled.value || dirty.value || plan.value === null || plan.value.token !== token) return
		busy.value = 'start'
		try {
			const result = await startDeployment({ projectId: project().id, token, password: password.value, passphrase: passphrase.value })
			if (!result.ok) { toast.error('启动发布失败', result.message); return }
			plan.value = null; toast.success('手动发布已开始'); await refreshHistory()
		} catch (err: unknown) { toast.error('启动发布失败', formatError(err)) }
		finally { busy.value = ''; password.value = ''; passphrase.value = '' }
	}
	async function cancel(): Promise<void> {
		if (busy.value !== '') return
		busy.value = 'cancel'
		try { const result = await cancelDeployment(project().id); if (!result.ok) toast.error('取消失败', result.message); else toast.warn('已请求取消', result.message); await refreshHistory() }
		catch (err: unknown) { toast.error('取消失败', formatError(err)) }
		finally { busy.value = '' }
	}
	async function pick(field: 'privateKey'): Promise<void> {
		if (disabled.value || config.value === null) return
		try {
			const value = await pickFile('选择 SSH 私钥文件')
			if (value !== '' && config.value !== null) config.value[field] = value
		} catch (err: unknown) { toast.error('选择文件失败', formatError(err)) }
	}
	async function openWeChat(): Promise<void> {
		try { const result = await openURL('https://mp.weixin.qq.com/'); if (!result.ok) toast.error('打开失败', result.message) }
		catch (err: unknown) { toast.error('打开失败', formatError(err)) }
	}
	watch(() => project().id, load, { immediate: true })
	watch(config, () => { plan.value = null }, { deep: true })
	onUnmounted(() => { disposed = true; generation++; stopPolling(); password.value = ''; passphrase.value = '' })
	return { config, history, plan, password, passphrase, busy, error, running, disabled, dirty, load, save, preview, test, confirmStart, cancel, pick, refreshHistory, openWeChat }
}
