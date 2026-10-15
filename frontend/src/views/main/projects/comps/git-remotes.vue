<template>
	<section class="remote-section">
		<div class="toolbar"><strong>远端仓库</strong><p-button size="small" severity="secondary" variant="outlined" icon="pi pi-plus" label="添加远端" :disabled="disabled" @click="add" /></div>
		<p class="hint">新增地址不会创建云端仓库，请先在 GitHub / GitLab 等平台创建仓库。认证使用本机 Git 凭据或 SSH 密钥，不在地址中填写密码或令牌。</p>
		<p v-if="status.remotes.length === 0" class="hint">尚未配置远端，本地提交不会上传到任何服务器。</p>
		<div v-for="remote in status.remotes" :key="remote.name" class="remote-card">
			<strong>{{ remote.name }}</strong>
			<div class="address"><span>拉取地址</span><code>{{ remote.url }}</code></div>
			<div class="address"><span>推送地址</span><code>{{ remote.pushUrl }}</code></div>
			<div class="actions">
				<p-button size="small" severity="secondary" variant="outlined" label="编辑" icon="pi pi-pencil" :disabled="disabled" @click="edit(remote)" />
				<p-button size="small" severity="secondary" variant="outlined" label="获取远端状态" icon="pi pi-download" :disabled="disabled" :loading="busy === `fetch:${remote.name}`" @click="emit('fetch', remote.name)" />
				<p-button size="small" label="推送当前分支" icon="pi pi-upload" :disabled="disabled || status.lastCommit === ''" :loading="busy === `push:${remote.name}`" @click="confirmPush($event, remote)" />
				<p-button size="small" severity="danger" variant="outlined" label="移除" icon="pi pi-trash" :disabled="disabled" :loading="busy === `remove:${remote.name}`" @click="confirmRemove($event, remote)" />
			</div>
		</div>
		<form v-if="editing" class="remote-form" @submit.prevent="save">
			<label for="git-remote-name">远端名称</label>
			<p-input-text id="git-remote-name" v-model="name" placeholder="origin" :disabled="disabled || existing" />
			<label for="git-remote-url">拉取地址</label>
			<p-input-text id="git-remote-url" v-model="url" placeholder="https://host/group/repository.git 或 git@host:group/repository.git" :disabled="disabled" />
			<label for="git-remote-push-url">推送地址（留空时与拉取地址相同）</label>
			<p-input-text id="git-remote-push-url" v-model="pushUrl" :disabled="disabled" />
			<p v-if="formError !== ''" class="error">{{ formError }}</p>
			<div class="actions"><p-button size="small" type="submit" label="保存地址" :disabled="disabled || name.trim() === '' || url.trim() === ''" :loading="busy === 'save'" /><p-button size="small" severity="secondary" variant="outlined" label="关闭编辑" :disabled="disabled" @click="editing = false" /></div>
		</form>
		<p class="hint">“获取远端状态”只更新远端记录，不会合并工作区。推送不使用强制覆盖；若远端领先，请先处理分支差异。</p>
	</section>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useConfirm } from 'primevue/useconfirm'
import type { ProjectGitRemote, ProjectGitStatus } from '@/types'
const props = withDefaults(defineProps<{ status: ProjectGitStatus; busy: string; disabled?: boolean }>(), { disabled: false })
const emit = defineEmits<{ save: [remote: ProjectGitRemote]; remove: [name: string]; fetch: [name: string]; push: [name: string] }>()
const confirm = useConfirm()
const editing = ref(false)
const existing = ref(false)
const name = ref('origin')
const url = ref('')
const pushUrl = ref('')
const formError = ref('')
function add(): void { name.value = 'origin'; url.value = ''; pushUrl.value = ''; existing.value = false; editing.value = true; formError.value = '' }
function edit(remote: ProjectGitRemote): void { name.value = remote.name; url.value = remote.url; pushUrl.value = remote.pushUrl; existing.value = true; editing.value = true; formError.value = '' }
function save(): void {
	if (props.disabled) return
	if (!existing.value && props.status.remotes.some((remote) => remote.name === name.value.trim())) { formError.value = '此名称已存在，请点击对应远端的“编辑”。'; return }
	formError.value = ''
	emit('save', { name: name.value.trim(), url: url.value.trim(), pushUrl: pushUrl.value.trim() })
}
function confirmRemove(event: MouseEvent, remote: ProjectGitRemote): void {
	confirm.require({ target: event.currentTarget as HTMLElement, message: `移除 ${remote.name} 的本地关联及跟踪引用？不会删除云端仓库。`, icon: 'pi pi-exclamation-triangle', acceptLabel: '移除', rejectLabel: '取消', acceptProps: { severity: 'danger', size: 'small' }, rejectProps: { severity: 'secondary', variant: 'outlined', size: 'small' }, accept: () => emit('remove', remote.name) })
}
function confirmPush(event: MouseEvent, remote: ProjectGitRemote): void {
	confirm.require({ target: event.currentTarget as HTMLElement, message: `将当前分支 ${props.status.branch} 的本地提交上传到 ${remote.name}（${remote.pushUrl}）？`, icon: 'pi pi-upload', acceptLabel: '确认推送', rejectLabel: '取消', acceptProps: { size: 'small' }, rejectProps: { severity: 'secondary', variant: 'outlined', size: 'small' }, accept: () => emit('push', remote.name) })
}
</script>

<style scoped lang="scss">
.remote-section, .remote-card, .remote-form { display: flex; flex-direction: column; gap: 0.6rem; }
.remote-section { padding-top: 0.875rem; border-top: 1px solid var(--border-color); }
.toolbar, .actions { display: flex; align-items: center; gap: 0.5rem; flex-wrap: wrap; }
.toolbar { justify-content: space-between; }
.hint { margin: 0; color: var(--text-muted); font-size: var(--font-xs); }
.remote-card, .remote-form { padding: 0.875rem; border: 1px solid var(--border-color); border-radius: var(--h-radius); background: var(--bg-app); }
.address { display: flex; gap: 0.75rem; font-size: var(--font-xs); }
.address > span { flex: none; color: var(--text-muted); }
.address code { min-width: 0; overflow-wrap: anywhere; color: var(--text-secondary); }
.remote-form > label { color: var(--text-secondary); font-size: var(--font-sm); }
.error { color: var(--err); font-size: var(--font-xs); }
</style>
