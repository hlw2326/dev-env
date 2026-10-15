<template>
	<div class="git-panel">
		<div class="toolbar">
			<div><strong>Git 管理</strong><p class="hint">本地提交与远端推送分开操作</p></div>
			<p-button size="small" severity="secondary" variant="outlined" icon="pi pi-refresh" label="刷新" :loading="loading" :disabled="disabled" @click="refreshStatus" />
		</div>
		<div v-if="loadError !== ''" class="notice error">{{ loadError }}</div>
		<div v-else-if="loading && status === null" class="notice"><i class="pi pi-spin pi-spinner"></i>正在读取 Git 状态…</div>
		<template v-else-if="status !== null">
			<div v-if="status.error !== ''" class="notice error">{{ status.error }}</div>
			<div v-else-if="!status.available" class="notice warning">电脑尚未安装 Git，请安装并加入 PATH 后重启 DevEnv。</div>
			<div v-else-if="status.nested" class="notice warning">项目位于另一个 Git 仓库内。请将仓库根目录添加为项目，避免误操作其他文件。</div>
			<div v-else-if="!status.repository" class="enable-box">
				<i class="pi pi-github"></i>
				<div><strong>此项目尚未启用 Git，是否需要开启？</strong><p class="hint">启用仅创建本地 .git 仓库，不会自动提交或上传文件。</p></div>
				<p-button label="启用 Git" icon="pi pi-plus" :loading="busy === 'initialize'" :disabled="disabled" @click="initializeRepository" />
			</div>
			<template v-else>
				<div class="summary-grid">
					<div class="summary-card"><span>当前分支</span><strong v-if="status.branch !== ''" class="font-mono">{{ status.branch }}</strong><strong v-else>尚无分支</strong></div>
					<div class="summary-card"><span>工作区</span><strong>{{ workspaceLabel }}</strong></div>
					<div class="summary-card"><span>最近提交</span><strong v-if="status.lastCommit !== ''" class="font-mono">{{ status.lastCommit }}</strong><strong v-else>暂无提交</strong><span>{{ status.commitMessage }}</span></div>
					<div class="summary-card"><span>远端跟踪状态（本地记录）</span><template v-if="status.upstream !== ''"><strong>{{ status.upstream }}</strong><span>待推送 {{ status.ahead }} · 落后 {{ status.behind }}</span></template><strong v-else>未关联远端分支</strong></div>
				</div>
				<div v-if="status.changed > 0" class="file-section">
					<strong>待提交变更 · {{ status.changed }} 项</strong>
					<p class="hint">将提交全部变更，包括已暂存、未暂存和未跟踪文件；请先用 .gitignore 排除密码、依赖和构建产物。</p>
					<div class="file-list"><div v-for="file in status.files" :key="file.path" class="file-row"><span class="file-code font-mono">{{ file.status }}</span><span class="file-path font-mono" :title="file.path">{{ file.path }}</span></div></div>
				</div>
				<div class="commit-box">
					<label for="git-commit-message">提交说明</label>
					<p-input-text id="git-commit-message" v-model="message" placeholder="例如：修复登录页面样式" :disabled="disabled" @keyup.enter="commitChanges" />
					<label class="version-option"><p-checkbox v-model="bumpVersion" binary :disabled="!status.versionCanBump || disabled" /><span>提交时自动升级版本号</span></label>
					<span v-if="status.versionCanBump" class="hint font-mono">{{ status.version }} → {{ status.nextVersion }}（{{ status.versionFile }}）</span>
					<span v-else class="hint">{{ status.versionError }}</span>
					<p v-if="bumpVersion" class="hint warning">按自定义规则递增（如 1.0.01）。补丁号带前导零不符合严格 SemVer，npm / Cargo 等工具可能拒绝该格式。</p>
					<div class="toolbar"><span class="hint">本地提交不会自动推送。</span><p-button label="提交到 Git" icon="pi pi-check" :loading="busy === 'commit'" :disabled="!canCommit" @click="commitChanges" /></div>
				</div>
				<CompGitRemotes :status="status" :busy="busy" :disabled="disabled" @save="saveRemote" @remove="removeRemote" @fetch="fetchRemote" @push="pushRemote" />
			</template>
		</template>
	</div>
</template>

<script setup lang="ts">
import { useProjectGit } from '@/composables'
import type { Project } from '@/types'
import CompGitRemotes from './git-remotes.vue'

const props = defineProps<{ project: Project }>()
const emit = defineEmits<{ refresh: [] }>()
const { status, loading, loadError, busy, disabled, message, bumpVersion, canCommit, workspaceLabel,
	refreshStatus, initializeRepository, commitChanges, saveRemote, removeRemote, fetchRemote, pushRemote,
} = useProjectGit(() => props.project, () => emit('refresh'))
</script>

<style scoped lang="scss">
.git-panel, .file-section, .commit-box { display: flex; flex-direction: column; gap: 0.75rem; }
.toolbar, .enable-box { display: flex; align-items: center; justify-content: space-between; gap: 0.75rem; }
.hint, .summary-card > span { margin: 0; font-size: var(--font-xs); color: var(--text-muted); overflow-wrap: anywhere; }
.notice, .enable-box, .summary-card, .file-list { padding: 0.875rem; border: 1px solid var(--border-color); border-radius: var(--h-radius); background: var(--bg-app); }
.notice { display: flex; gap: 0.5rem; }
.error { color: var(--err); }
.warning { color: var(--warn); }
.enable-box > div { flex: 1; }
.enable-box :deep(button) { flex-shrink: 0; }
.summary-grid { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 0.5rem; }
.summary-card { display: flex; flex-direction: column; gap: 0.3rem; min-width: 0; }
.summary-card strong { overflow-wrap: anywhere; }
.file-list { max-height: 10rem; overflow-y: auto; }
.file-row { display: flex; gap: 0.5rem; font-size: var(--font-xs); }
.file-code { width: 2rem; flex: none; color: var(--info); }
.file-path { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.commit-box { border-top: 1px solid var(--border-color); padding-top: 0.875rem; }
.commit-box > label { color: var(--text-secondary); font-size: var(--font-sm); }
.version-option { display: flex; align-items: center; gap: 0.4rem; }
</style>
