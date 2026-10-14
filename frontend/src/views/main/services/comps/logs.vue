<template>
	<!-- 服务控制台实时日志面板 -->
	<div v-if="service" class="log-panel">
		<!-- 日志面板头部 -->
		<div class="log-header">
			<div class="log-title">
				<i class="pi pi-terminal"></i>
				<span>{{ service.name }} - 控制台实时日志</span>
			</div>
			<div class="log-actions">
				<p-button
					label="刷新日志"
					icon="pi pi-refresh"
					severity="secondary"
					text
					size="small"
					@click="emit('refresh')"
				/>
				<p-button
					label="关闭"
					severity="secondary"
					text
					size="small"
					@click="emit('close')"
				/>
			</div>
		</div>

		<!-- 日志控制台输出区 -->
		<div class="log-console font-mono">
			<div v-if="logs.length === 0" class="log-empty">暂无输出日志</div>
			<div v-for="(line, idx) in logs" :key="idx" class="log-row">
				<span class="log-num">{{ idx + 1 }}</span>
				<span class="log-text">{{ line }}</span>
			</div>
		</div>
	</div>
</template>

<script setup lang="ts">
	import type { ServiceInstance } from '@/types'

	// 接收当前服务与日志内容
	defineProps<{
		service: ServiceInstance | null
		logs: string[]
	}>()

	// 定义刷新与关闭事件
	const emit = defineEmits<{
		refresh: []
		close: []
	}>()
</script>

<style scoped lang="scss">
	/* 日志面板容器 */
	.log-panel {
		padding: 16px;
		display: flex;
		flex-direction: column;
		gap: 12px;
		background: var(--bg-card);
		border: 1px solid var(--border-color);
		border-radius: var(--radius-lg);
	}

	/* 面板头部布局 */
	.log-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
	}

	/* 面板标题文字 */
	.log-title {
		display: flex;
		align-items: center;
		gap: 8px;
		font-size: var(--font-base);
		color: var(--text-primary);
	}

	/* 终端输出区 */
	.log-console {
		background: var(--code-bg);
		border-radius: var(--radius-md);
		padding: 14px;
		height: 240px;
		overflow-y: auto;
		color: var(--code-text);
	}

	/* 单行日志 */
	.log-row {
		display: flex;
		gap: 12px;
		font-size: var(--font-sm);
		line-height: 1.5;
	}

	/* 日志行号 */
	.log-num {
		color: var(--text-muted);
		min-width: 24px;
		user-select: none;
	}

	/* 日志文本 */
	.log-text {
		color: var(--code-text);
		word-break: break-all;
	}

	/* 空日志提示 */
	.log-empty {
		text-align: center;
		padding: 70px 0;
		color: var(--text-muted);
		font-size: var(--font-sm);
	}
</style>
