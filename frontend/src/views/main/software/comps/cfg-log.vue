<template>
	<div class="cfg-log">
		<div class="log-toolbar">
			<span class="sec-title">运行与错误日志</span>
			<p-button
				size="small"
				severity="secondary"
				variant="outlined"
				icon="pi pi-refresh"
				label="刷新"
				@click="emit('refresh')"
			/>
		</div>
		<div class="log-box font-mono">
			<div v-if="lines.length === 0" class="empty-log">暂无日志输出</div>
			<div v-for="(line, idx) in lines" :key="idx" class="log-line">{{ line }}</div>
		</div>
	</div>
</template>

<script setup lang="ts">
defineProps<{
	lines: string[]
}>()

const emit = defineEmits<{
	refresh: []
}>()
</script>

<style scoped lang="scss">
.cfg-log {
	display: flex;
	flex-direction: column;
	gap: 0.625rem;
}

.log-toolbar {
	display: flex;
	align-items: center;
	justify-content: space-between;
}

.sec-title {
	font-size: var(--font-base);
	color: var(--text-primary);
}

.log-box {
	background: var(--code-bg);
	color: var(--code-text);
	border-radius: var(--radius-md);
	padding: 0.875rem;
	height: 20rem;
	overflow-y: auto;
	font-size: var(--font-xs);
	display: flex;
	flex-direction: column;
	gap: 0.25rem;
}

.empty-log {
	color: var(--text-muted);
}

.log-line {
	white-space: pre-wrap;
	word-break: break-all;
}
</style>
