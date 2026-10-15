import { computed, ref, watch } from 'vue'
import {
	commitProjectGit, fetchProjectGit, getProjectGitStatus, initializeProjectGit,
	pushProjectGit, removeProjectGitRemote, saveProjectGitRemote,
} from '@/api'
import { formatError, toast } from '@/utils'
import type { Project, ProjectGitRemote, ProjectGitStatus } from '@/types'

export function useProjectGit(project: () => Project, changed: () => void) {
	const status = ref<ProjectGitStatus | null>(null)
	const loading = ref(false)
	const loadError = ref('')
	const busy = ref('')
	const message = ref('')
	const bumpVersion = ref(false)
	let request = 0
	const disabled = computed(() => loading.value || busy.value !== '')
	const canCommit = computed(() => {
		const value = status.value
		return !disabled.value && value !== null && value.error === '' && value.repository && !value.nested
			&& message.value.trim() !== '' && (value.changed > 0 || (bumpVersion.value && value.versionCanBump))
	})
	const workspaceLabel = computed(() => {
		if (status.value === null) return ''
		if (status.value.changed === 0) return '工作区干净'
		return `${status.value.changed} 项未提交`
	})

	async function refreshStatus(): Promise<void> {
		const current = ++request
		const id = project().id
		loading.value = true
		loadError.value = ''
		try {
			const result = await getProjectGitStatus(id)
			if (current === request && id === project().id) status.value = result
		} catch (err: unknown) {
			if (current === request) loadError.value = formatError(err)
		} finally {
			if (current === request) loading.value = false
		}
	}

	async function operate(key: string, action: (id: number) => Promise<{ ok: boolean; message: string }>): Promise<boolean> {
		if (disabled.value) return false
		const id = project().id
		busy.value = key
		try {
			const result = await action(id)
			if (!result.ok) { toast.error('Git 操作失败', result.message); return false }
			toast.success('Git 操作完成', result.message)
			return true
		} catch (err: unknown) {
			toast.error('Git 操作失败', formatError(err))
			return false
		} finally {
			await refreshStatus()
			changed()
			busy.value = ''
		}
	}

	async function commitChanges(): Promise<void> {
		if (!canCommit.value) return
		const id = project().id
		const success = await operate('commit', (projectId) => commitProjectGit({ projectId, message: message.value.trim(), bumpVersion: bumpVersion.value }))
		if (success && id === project().id) { message.value = ''; bumpVersion.value = false }
	}
	const initializeRepository = () => operate('initialize', initializeProjectGit)
	const saveRemote = (remote: ProjectGitRemote) => operate('save', (projectId) => saveProjectGitRemote({ projectId, ...remote }))
	const removeRemote = (name: string) => operate(`remove:${name}`, (id) => removeProjectGitRemote(id, name))
	const fetchRemote = (name: string) => operate(`fetch:${name}`, (id) => fetchProjectGit(id, name))
	const pushRemote = (name: string) => operate(`push:${name}`, (id) => pushProjectGit(id, name))

	watch(() => project().id, () => {
		status.value = null
		message.value = ''
		bumpVersion.value = false
		void refreshStatus()
	}, { immediate: true })

	return { status, loading, loadError, busy, disabled, message, bumpVersion, canCommit, workspaceLabel,
		refreshStatus, initializeRepository, commitChanges, saveRemote, removeRemote, fetchRemote, pushRemote }
}
