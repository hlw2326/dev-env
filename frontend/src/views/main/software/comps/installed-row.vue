<template>
	<!-- 已安装软件版本行 -->
	<div class="installed-row surface-card">
		<!-- 左侧图标与版本状态 -->
		<div class="left">
			<h-logo :kind="item.kind" size="medium" />
			<div class="info">
				<div class="top-line">
					<span class="item-name">{{ item.name }} {{ item.version }}</span>
					<span v-if="item.main" class="main-tag">默认</span>

					<span v-if="item.state === 'installing'" class="status-pill installing">
						<i class="pi pi-spin pi-spinner status-icon"></i>
						<span>安装中 {{ activeProgress.progress }}%</span>
					</span>
					<span v-else-if="item.state === 'failed'" class="status-pill failed">
						<span class="dot"></span>
						<span>安装失败</span>
					</span>
					<span v-else-if="item.service && item.running" class="status-pill running">
						<span class="dot"></span>
						<span>运行中 (PID {{ item.pid }})</span>
					</span>
					<span v-else-if="item.service && !item.running" class="status-pill stopped">
						<span class="dot"></span>
						<span>已停止</span>
					</span>
					<span v-else class="status-pill ok">
						<span class="dot"></span>
						<span>可用</span>
					</span>

					<span v-if="item.daemon && item.port > 0" class="port-badge font-mono">
						端口 {{ item.port }}
					</span>
				</div>

				<div class="sub-line">
					<div v-if="item.state === 'installing'" class="progress-box">
						<div class="progress-info">
							<span class="progress-text">{{ activeProgress.message }}</span>
							<span class="progress-num font-mono">{{ activeProgress.progress }}%</span>
						</div>
						<p-progress-bar :value="activeProgress.progress" :show-value="false" class="progress-bar" />
					</div>
					<span v-else-if="item.state === 'failed'" class="error-text">{{ item.error }}</span>
					<span v-else class="path-text font-mono">{{ item.path }}</span>
				</div>
			</div>
		</div>

		<!-- 右侧版本操作按钮组 -->
		<div class="actions">
			<p-button
				v-if="item.state === 'installing'"
				size="small"
				severity="danger"
				variant="outlined"
				icon="pi pi-times"
				label="取消"
				@click="emit('cancel', item.kind, item.version)"
			/>
			<template v-else-if="item.state === 'failed'">
				<p-button
					size="small"
					icon="pi pi-refresh"
					label="重试"
					@click="emit('retry', item.kind, item.version)"
				/>
				<p-button
					size="small"
					severity="danger"
					variant="outlined"
					icon="pi pi-trash"
					label="清理"
					@click="emit('remove', item)"
				/>
			</template>
			<template v-else>
				<p-button
					v-if="item.daemon && item.running"
					size="small"
					severity="secondary"
					variant="outlined"
					icon="pi pi-stop"
					label="停止"
					@click="emit('stop', item)"
				/>
				<p-button
					v-else-if="item.daemon"
					size="small"
					icon="pi pi-play"
					label="启动"
					@click="emit('start', item)"
				/>
				<p-button
					v-if="!item.main"
					size="small"
					severity="secondary"
					variant="outlined"
					label="设为默认"
					@click="emit('main', item)"
				/>
				<p-button
					size="small"
					severity="secondary"
					variant="outlined"
					icon="pi pi-cog"
					label="设置"
					@click="emit('config', item)"
				/>
				<p-button
					size="small"
					severity="danger"
					variant="outlined"
					icon="pi pi-trash"
					label="卸载"
					@click="confirmRemove($event)"
				/>
			</template>
		</div>
	</div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useConfirm } from 'primevue/useconfirm'
import type { Software } from '@/types'

// 接收软件版本与进度属性
const props = defineProps<{
	item: Software
	progressMap: Record<string, { progress: number; message: string }>
}>()

// 定义版本行操作事件
const emit = defineEmits<{
	start: [item: Software]
	stop: [item: Software]
	main: [item: Software]
	config: [item: Software]
	remove: [item: Software]
	cancel: [kind: string, version: string]
	retry: [kind: string, version: string]
}>()

// 确认弹窗服务
const confirm = useConfirm()

// 当前实时安装进度信息
const activeProgress = computed(() => {
	const key = `${props.item.kind}:${props.item.version}`
	if (props.progressMap[key]) {
		return props.progressMap[key]
	}
	return { progress: props.item.progress, message: '正在安装...' }
})

// 包含独立数据目录的数据库类型集合
const DATA_KINDS = ['mysql', 'mariadb', 'postgresql', 'mongodb']

// 弹出卸载软件确认框
function confirmRemove(event: MouseEvent): void {
	let message = `确认卸载 ${props.item.name} ${props.item.version} 吗？`
	if (DATA_KINDS.includes(props.item.kind)) {
		message = `卸载只删除软件目录，数据库数据目录 dbdata\\${props.item.kind}-${props.item.version} 将完整保留`
	}
	confirm.require({
		target: event.currentTarget as HTMLElement,
		message,
		icon: 'pi pi-exclamation-triangle',
		acceptLabel: '卸载',
		rejectLabel: '取消',
		acceptProps: { severity: 'danger', size: 'small' },
		rejectProps: { severity: 'secondary', variant: 'outlined', size: 'small' },
		accept: () => emit('remove', props.item),
	})
}
</script>

<style scoped lang="scss">
/* 版本行容器 */
.installed-row {
	display: flex;
	align-items: center;
	justify-content: space-between;
	padding: 0.875rem 1.125rem;
	gap: 1rem;
}
/* 左侧信息组合 */
.left {
	display: flex;
	align-items: center;
	gap: 0.875rem;
	min-width: 0;
	flex: 1;
}
/* 文本列容器 */
.info {
	display: flex;
	flex-direction: column;
	gap: 0.3rem;
	min-width: 0;
	flex: 1;
}
/* 顶部标题与状态行 */
.top-line {
	display: flex;
	align-items: center;
	gap: 0.625rem;
	flex-wrap: wrap;
}
/* 软件版本名称 */
.item-name {
	font-size: var(--font-base);
	color: var(--text-primary);
}
/* 默认版本标签 */
.main-tag {
	font-size: var(--font-xs);
	color: var(--info);
	background: var(--info-bg);
	border: 1px solid var(--info-border);
	padding: 0.1rem 0.45rem;
	border-radius: var(--h-radius);
}
/* 状态加载图标 */
.status-icon {
	font-size: var(--icon-xs);
}
/* 端口徽标 */
.port-badge {
	font-size: var(--font-xs);
	color: var(--text-secondary);
	background: var(--bg-app);
	border: 1px solid var(--border-color);
	padding: 0.1rem 0.45rem;
	border-radius: var(--h-radius);
}
/* 次要信息行 */
.sub-line {
	display: flex;
	align-items: center;
	width: 100%;
}
/* 进度条容器 */
.progress-box {
	display: flex;
	flex-direction: column;
	gap: 0.35rem;
	width: 100%;
	max-width: 26rem;
	margin-top: 0.25rem;
}
/* 进度说明与百分比 */
.progress-info {
	display: flex;
	align-items: center;
	justify-content: space-between;
	font-size: var(--font-xs);
	color: var(--text-secondary);
}
/* 进度说明文本 */
.progress-text {
	color: var(--text-secondary);
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
.progress-bar {
	height: 0.45rem;
	border-radius: 999rem;
}
/* 安装路径文本 */
.path-text {
	font-size: var(--font-xs);
	color: var(--text-muted);
}
/* 失败错误消息 */
.error-text {
	font-size: var(--font-xs);
	color: var(--err);
}
/* 右侧操作区 */
.actions {
	display: flex;
	align-items: center;
	gap: 0.45rem;
	flex-shrink: 0;
}
</style>
