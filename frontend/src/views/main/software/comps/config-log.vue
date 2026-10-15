<template>
	<!-- 软件运行日志查看器 -->
	<div class="config-log">
		<!-- 顶部日志工具栏 -->
		<div class="log-bar">
			<span class="log-hint">显示最近 200 行运行输出</span>
			<p-button
				size="small"
				severity="secondary"
				variant="outlined"
				icon="pi pi-refresh"
				label="刷新日志"
				@click="reloadLogs"
			/>
		</div>

		<!-- 日志内容输出区 -->
		<div class="log-box font-mono">
			<div v-if="logLines.length === 0" class="empty-log">暂无运行日志</div>
			<div v-for="(line, index) in logLines" :key="index" class="log-line">
				{{ line }}
			</div>
		</div>
	</div>
</template>

<script setup lang="ts">
import { watch } from 'vue'
import { useConfig } from '@/composables'

// 接收软件标识属性
const props = defineProps<{
	softwareId: number
}>()

// 日志状态与读取方法
const { logLines, fetchLogs } = useConfig()

// 刷新运行日志
async function reloadLogs(): Promise<void> {
	await fetchLogs(props.softwareId)
}

// 监听软件标识变化加载日志
watch(
	() => props.softwareId,
	() => {
		reloadLogs()
	},
	{ immediate: true },
)
</script>

<style scoped lang="scss">
/* 日志查看容器 */
.config-log {
	display: flex;
	flex-direction: column;
	gap: 0.75rem;
}

/* 顶部工具栏 */
.log-bar {
	display: flex;
	align-items: center;
	justify-content: space-between;
}

/* 日志说明文本 */
.log-hint {
	font-size: var(--font-xs);
	color: var(--text-muted);
}

/* 日志黑框容器 */
.log-box {
	background: var(--code-bg);
	color: var(--code-text);
	border-radius: var(--h-radius);
	padding: 0.875rem;
	height: 18rem;
	overflow-y: auto;
	font-size: var(--font-xs);
	display: flex;
	flex-direction: column;
	gap: 0.25rem;
}

/* 空日志提示 */
.empty-log {
	color: var(--text-muted);
}

/* 单行日志文本 */
.log-line {
	white-space: pre-wrap;
	word-break: break-all;
}
</style>
