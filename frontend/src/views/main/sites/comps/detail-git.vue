<template>
	<!-- 详情页签2：Git 仓库状态与提交管理 -->
	<div v-if="curSite" class="detail-pane">
		<!-- 远程仓库与当前分支头部卡片 -->
		<div class="git-head">
			<div class="meta-col">
				<div class="git-line">
					<span class="label">远程仓库：</span>
					<span class="font-mono">{{ remoteLabel }}</span>
				</div>
				<div class="git-line">
					<span class="label">当前分支：</span>
					<p-tag severity="info" :value="branchText" />
					<span class="label">最后提交：</span>
					<span class="font-mono">{{ commitText }}</span>
					<span v-if="gitStatus?.message" class="text-muted">({{ gitStatus.message }})</span>
				</div>
			</div>
			<div class="action-col">
				<p-button label="刷新状态" icon="pi pi-refresh" size="small" severity="secondary" outlined :loading="gitLoading" @click="loadGit(curSite.id)" />
				<p-button label="拉取更新 (Pull)" icon="pi pi-cloud-download" size="small" severity="secondary" outlined :loading="gitLoading" @click="pullRepo" />
				<p-button label="推送到远程 (Push)" icon="pi pi-cloud-upload" size="small" :loading="gitLoading" @click="pushRepo" />
			</div>
		</div>

		<!-- 工作区变更与提交表单双列网格 -->
		<div class="git-grid">
			<div class="info-card">
				<div class="sec-title">
					<span>工作区变更状态</span>
					<p-tag :severity="cleanTone" :value="cleanText" size="small" />
				</div>
				<div v-if="gitFiles.length === 0" class="empty-small">当前工作区没有已修改或未跟踪的文件</div>
				<div v-else class="files-list font-mono">
					<label v-for="f in gitFiles" :key="f.file" class="file-row">
						<input v-model="selectedFiles" type="checkbox" :value="f.file" />
						<span class="file-tag" :class="f.style">{{ f.tag }}</span>
						<span class="file-path">{{ f.file }}</span>
					</label>
				</div>
			</div>

			<div class="info-card">
				<div class="sec-title">提交更改 (Commit)</div>
				<div class="form-field">
					<label>提交说明 (Commit Message)</label>
					<textarea v-model="commitMsg" rows="4" class="code-area" placeholder="请输入本次代码变更说明，如：Update styles and fix bug"></textarea>
				</div>
				<div class="commit-btns">
					<p-button label="提交更改" icon="pi pi-check" size="small" :disabled="!commitMsg.trim()" :loading="gitLoading" @click="commitRepo(false)" />
					<p-button label="提交并推送" icon="pi pi-send" size="small" severity="success" :disabled="!commitMsg.trim()" :loading="gitLoading" @click="commitRepo(true)" />
				</div>
			</div>
		</div>
	</div>
</template>

<script setup lang="ts">
	import { computed } from 'vue'
	import { useDetail } from '../use-detail'

	// 解构 Git 状态与操作方法
	const {
		curSite, gitStatus, gitLoading, selectedFiles, commitMsg,
		loadGit, pullRepo, commitRepo, pushRepo
	} = useDetail()

	// 远程仓库地址文案
	const remoteLabel = computed(() => {
		if (curSite.value && curSite.value.remote) {
			return curSite.value.remote
		}
		return '本地仓库 / 未绑定远程地址'
	})

	// 当前分支名称
	const branchText = computed(() => {
		if (gitStatus.value && gitStatus.value.branch) {
			return gitStatus.value.branch
		}
		if (curSite.value && curSite.value.branch) {
			return curSite.value.branch
		}
		return 'main'
	})

	// 最后提交短哈希
	const commitText = computed(() => {
		if (gitStatus.value && gitStatus.value.commit) {
			return gitStatus.value.commit
		}
		if (curSite.value && curSite.value.commit) {
			return curSite.value.commit
		}
		return '-'
	})

	// 工作区干净状态颜色与文案
	const cleanTone = computed(() => {
		if (gitStatus.value && gitStatus.value.clean) {
			return 'success'
		}
		return 'warn'
	})
	const cleanText = computed(() => {
		if (gitStatus.value && gitStatus.value.clean) {
			return '工作区干净'
		}
		return '存在未提交更改'
	})

	// 合并已修改与未跟踪文件列表
	const gitFiles = computed(() => {
		if (!gitStatus.value) {
			return []
		}
		const list: Array<{ tag: string; style: string; file: string }> = []
		for (const m of gitStatus.value.modified) {
			list.push({ tag: 'M', style: 'modified', file: m })
		}
		for (const u of gitStatus.value.untracked) {
			list.push({ tag: 'U', style: 'untracked', file: u })
		}
		return list
	})
</script>

<style scoped lang="scss">
	/* 页签主体容器与头部 */
	.detail-pane { display: flex; flex-direction: column; gap: 14px; }
	.git-head {
		display: flex; justify-content: space-between; align-items: center;
		padding: 12px 14px; background: var(--bg-app); border: 1px solid var(--border-color); border-radius: var(--radius-md);
	}
	.meta-col { display: flex; flex-direction: column; gap: 6px; font-size: var(--font-xs); }
	.git-line { display: flex; align-items: center; gap: 8px; }
	.label { color: var(--text-muted); }
	.action-col, .commit-btns { display: flex; gap: 8px; }

	/* 双列工作区网格与卡片 */
	.git-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 14px; }
	.info-card {
		background: var(--bg-app); border: 1px solid var(--border-color);
		border-radius: var(--radius-md); padding: 14px; display: flex; flex-direction: column; gap: 10px;
	}
	.sec-title { display: flex; justify-content: space-between; align-items: center; font-size: var(--font-sm); font-weight: 500; color: var(--text-primary); }

	/* 变更文件列表 */
	.files-list { display: flex; flex-direction: column; gap: 6px; max-height: 180px; overflow-y: auto; }
	.file-row { display: flex; align-items: center; gap: 8px; font-size: var(--font-xs); cursor: pointer; }
	.file-tag {
		padding: 1px 5px; border-radius: var(--radius-sm); font-size: var(--font-xs);
		&.modified { background: var(--warn-bg); color: var(--warn); }
		&.untracked { background: var(--ok-bg); color: var(--ok); }
	}
	.empty-small { padding: 30px 0; text-align: center; color: var(--text-muted); font-size: var(--font-xs); }
	.code-area {
		width: 100%; padding: 8px 10px; background: var(--bg-card);
		border: 1px solid var(--border-color); border-radius: var(--radius-sm);
		color: var(--text-primary); font-size: var(--font-xs); resize: vertical;
	}
	.text-muted { color: var(--text-muted); }
</style>
