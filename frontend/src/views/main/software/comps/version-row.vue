<template>
	<!-- 可安装版本单行条目 -->
	<div class="version-row" :class="{ active: task !== null }">
		<!-- 顶部版本信息与安装按钮 -->
		<div class="row-main">
			<div class="version-info">
				<span class="version-num font-mono">{{ row.version }}</span>
				<span v-if="row.recommended" class="recommend-tag">推荐</span>
				<span v-if="row.note !== ''" class="note-tag">{{ row.note }}</span>
			</div>

			<div class="version-actions">
				<span v-if="installed" class="installed-label">
					<i class="pi pi-check-circle"></i> 已安装
				</span>
				<p-button
					v-else-if="task !== null"
					class="install-cancel-button"
					size="small"
					severity="danger"
					variant="outlined"
					icon="pi pi-times"
					label="取消"
					@click="emit('cancel', row.version)"
				/>
				<p-button
					v-else
					size="small"
					icon="pi pi-download"
					label="安装"
					:disabled="busy"
					@click="emit('install', row.version)"
				/>
			</div>
		</div>

		<!-- 下载与安装实时进度条 -->
		<div v-if="task !== null" class="progress-box">
			<div class="progress-info">
				<span class="progress-text">
					<i class="pi pi-spin pi-spinner"></i>
					{{ task.message }}
				</span>
				<span class="progress-num font-mono">{{ task.progress }}%</span>
			</div>
			<p-progress-bar
				:value="task.progress"
				:show-value="false"
				class="download-bar"
			/>
		</div>
	</div>
</template>

<script setup lang="ts">
import type { SoftwareVersion } from '@/types'

// 接收版本与进度属性
defineProps<{
	row: SoftwareVersion
	installed: boolean
	task: { progress: number; message: string } | null
	busy: boolean
}>()

// 定义安装与取消事件
const emit = defineEmits<{
	install: [version: string]
	cancel: [version: string]
}>()
</script>

<style scoped lang="scss">
/* 版本条目外框 */
.version-row {
	display: flex;
	flex-direction: column;
	gap: 0.5rem;
	padding: 0.75rem 0.875rem;
	background: var(--bg-app);
	border: 1px solid var(--border-color);
	border-radius: var(--h-radius);
	transition: all 0.15s ease;

	/* 下载中激活态 */
	&.active {
		background: var(--bg-card);
		border-color: var(--p-primary-color);
	}
}

/* 主行布局 */
.row-main {
	display: flex;
	align-items: center;
	justify-content: space-between;
	width: 100%;
}

/* 版本号与标签区 */
.version-info {
	display: flex;
	align-items: center;
	gap: 0.5rem;
	flex-wrap: wrap;
}

/* 版本号文字 */
.version-num {
	font-size: var(--font-base);
	color: var(--text-primary);
}

/* 推荐标签 */
.recommend-tag {
	font-size: var(--font-xs);
	color: var(--ok);
	background: var(--ok-bg);
	border: 1px solid var(--ok-border);
	padding: 0.1rem 0.45rem;
	border-radius: var(--h-radius);
}

/* 版本附加说明 */
.note-tag {
	font-size: var(--font-xs);
	color: var(--text-secondary);
}

/* 右侧操作区 */
.version-actions {
	display: flex;
	align-items: center;
	gap: 0.5rem;
}

/* 安装中的取消按钮保持紧凑，避免挤占版本信息空间 */
.install-cancel-button {
	min-width: 4rem;
	height: 2rem;
	padding: 0.3rem 0.5rem;
	font-size: var(--font-xs);
}

/* 已安装状态文字 */
.installed-label {
	display: flex;
	align-items: center;
	gap: 0.35rem;
	font-size: var(--font-xs);
	color: var(--ok);
}

/* 进度条区域 */
.progress-box {
	display: flex;
	flex-direction: column;
	gap: 0.35rem;
	padding-top: 0.5rem;
	border-top: 1px dashed var(--border-color);
}

/* 进度文字行 */
.progress-info {
	display: flex;
	justify-content: space-between;
	align-items: center;
	font-size: var(--font-xs);
	color: var(--text-secondary);
}

/* 进度说明文本 */
.progress-text {
	display: flex;
	align-items: center;
	gap: 0.35rem;
	color: var(--text-primary);
}

/* 进度百分比数值 */
.progress-num {
	color: var(--p-primary-color);
}

/* 进度条高度 */
.download-bar {
	height: 0.45rem;
	border-radius: 999rem;
}
</style>
