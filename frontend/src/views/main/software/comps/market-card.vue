<template>
	<!-- 软件商店单个软件卡片 -->
	<div class="market-card surface-card">
		<!-- 卡片顶部图标与描述 -->
		<div class="card-top">
			<h-logo :kind="item.kind" size="large" />
			<div class="title-col">
				<div class="name-line">
					<span class="software-name">{{ item.name }}</span>
					<span class="category-badge">{{ item.category }}</span>
					<span v-if="item.count > 0" class="installed-badge">
						已安装 {{ item.count }} 个版本
					</span>
				</div>
				<p class="software-summary">{{ item.description }}</p>
			</div>
		</div>

		<!-- 安装中实时进度区 -->
		<div v-if="activeTask !== null" class="progress-area">
			<div class="progress-info">
				<span class="progress-text">
					<i class="pi pi-spin pi-spinner"></i>
					{{ activeTask.version }} · {{ activeTask.message }}
				</span>
				<span class="progress-num font-mono">{{ activeTask.progress }}%</span>
			</div>
			<p-progress-bar :value="activeTask.progress" :show-value="false" class="bar" />
			<div class="progress-actions">
				<p-button
					size="small"
					severity="danger"
					variant="outlined"
					icon="pi pi-times"
					label="取消"
					@click="emit('cancel', item.kind, activeTask.version)"
				/>
			</div>
		</div>

		<!-- 安装失败原因与重试区 -->
		<div v-else-if="failedTask !== null" class="failed-area">
			<span class="failed-text">{{ failedTask.version }} 安装失败：{{ failedTask.error }}</span>
			<div class="button-group">
				<p-button
					size="small"
					icon="pi pi-refresh"
					label="重试"
					@click="emit('retry', item.kind, failedTask.version)"
				/>
				<p-button
					size="small"
					severity="secondary"
					variant="outlined"
					icon="pi pi-download"
					label="选择版本"
					@click="emit('install', item)"
				/>
			</div>
		</div>

		<!-- 卡片底部推荐版本与操作按钮 -->
		<div v-else class="card-footer">
			<span v-if="item.recommended !== ''" class="recommend-hint font-mono">
				推荐版本：{{ item.recommended }}
			</span>
			<span v-else class="recommend-hint">点击选择版本安装</span>

			<div class="button-group">
				<p-button
					v-if="item.count > 0"
					size="small"
					severity="secondary"
					variant="outlined"
					icon="pi pi-cog"
					label="设置"
					@click="emit('config', item.kind)"
				/>
				<p-button
					v-if="item.recommended !== ''"
					size="small"
					icon="pi pi-download"
					:label="`安装 ${item.recommended}`"
					@click="emit('install', item)"
				/>
				<p-button
					v-else
					size="small"
					icon="pi pi-download"
					label="选择版本安装"
					@click="emit('install', item)"
				/>
			</div>
		</div>
	</div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { MarketItem, Software } from '@/types'

// 接收商店卡片属性
const props = defineProps<{
	item: MarketItem
	installedList: Software[]
	progressMap: Record<string, { progress: number; message: string }>
}>()

// 定义卡片操作事件
const emit = defineEmits<{
	install: [item: MarketItem]
	cancel: [kind: string, version: string]
	retry: [kind: string, version: string]
	config: [kind: string]
}>()

// 当前正在进行的安装任务
const activeTask = computed(() => {
	for (const entry of props.installedList) {
		if (entry.kind === props.item.kind && entry.state === 'installing') {
			const key = `${entry.kind}:${entry.version}`
			const live = props.progressMap[key]
			if (live) {
				return { version: entry.version, progress: live.progress, message: live.message }
			}
			return { version: entry.version, progress: entry.progress, message: '安装中...' }
		}
	}
	for (const [key, value] of Object.entries(props.progressMap)) {
		if (key.startsWith(`${props.item.kind}:`)) {
			const version = key.slice(props.item.kind.length + 1)
			return { version, progress: value.progress, message: value.message }
		}
	}
	return null
})

// 最近一次安装失败的版本记录
const failedTask = computed(() => {
	const found = props.installedList.find(
		(entry) => entry.kind === props.item.kind && entry.state === 'failed',
	)
	if (!found) {
		return null
	}
	return { version: found.version, error: found.error }
})
</script>

<style scoped lang="scss">
/* 软件卡片外框 */
.market-card {
	display: flex;
	flex-direction: column;
	justify-content: space-between;
	padding: 1.125rem;
	gap: 1rem;
}

/* 顶部图标与文字区 */
.card-top {
	display: flex;
	align-items: flex-start;
	gap: 0.875rem;
}

/* 标题与简介列 */
.title-col {
	display: flex;
	flex-direction: column;
	gap: 0.35rem;
	flex: 1;
}

/* 名称与徽标行 */
.name-line {
	display: flex;
	align-items: center;
	gap: 0.5rem;
	flex-wrap: wrap;
}

/* 软件主名称 */
.software-name {
	font-size: var(--font-lg);
	color: var(--text-primary);
}

/* 分类徽标 */
.category-badge {
	font-size: var(--font-xs);
	color: var(--text-secondary);
	background: var(--bg-app);
	border: 1px solid var(--border-color);
	padding: 0.1rem 0.45rem;
	border-radius: var(--h-radius);
}

/* 已安装数量徽标 */
.installed-badge {
	font-size: var(--font-xs);
	color: var(--ok);
	background: var(--ok-bg);
	border: 1px solid var(--ok-border);
	padding: 0.1rem 0.45rem;
	border-radius: var(--h-radius);
}

/* 软件说明文字 */
.software-summary {
	font-size: var(--font-sm);
	color: var(--text-secondary);
}

/* 进度条区域 */
.progress-area {
	display: flex;
	flex-direction: column;
	gap: 0.5rem;
	padding-top: 0.5rem;
	border-top: 1px dashed var(--border-color);
}

/* 进度文字与百分比行 */
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
	overflow: hidden;
	text-overflow: ellipsis;
	white-space: nowrap;
}

/* 进度百分比数值 */
.progress-num {
	color: var(--p-primary-color);
	margin-left: 0.5rem;
}

/* 进度条高度 */
.bar {
	height: 0.45rem;
	border-radius: 999rem;
}

/* 进度取消按钮区 */
.progress-actions {
	display: flex;
	justify-content: flex-end;
}

/* 失败提示与重试操作栏 */
.failed-area,
.card-footer {
	display: flex;
	align-items: center;
	justify-content: space-between;
	padding-top: 0.75rem;
	border-top: 1px dashed var(--border-color);
	gap: 0.5rem;
}

/* 失败原因说明 */
.failed-text {
	font-size: var(--font-xs);
	color: var(--err);
	overflow: hidden;
	text-overflow: ellipsis;
	white-space: nowrap;
}

/* 推荐版本提示 */
.recommend-hint {
	font-size: var(--font-xs);
	color: var(--text-muted);
}

/* 底部按钮组 */
.button-group {
	display: flex;
	align-items: center;
	gap: 0.5rem;
	flex-shrink: 0;
}
</style>
