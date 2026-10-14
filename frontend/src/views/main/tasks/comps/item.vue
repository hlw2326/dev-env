<template>
	<!-- 单条任务卡片 -->
	<div class="task-box">
		<!-- 任务头部信息与操作 -->
		<div class="task-top">
			<div class="task-left">
				<span class="task-type font-mono">{{ item.taskType }}</span>
				<span class="task-id font-mono">#{{ item.id }}</span>
				<span class="task-msg">{{ item.message }}</span>
			</div>

			<div class="task-right">
				<p-tag
					:severity="statusTone"
					:value="statusText"
				/>
				<p-button
					v-if="item.status === 'running' || item.status === 'queued'"
					label="取消任务"
					severity="danger"
					size="small"
					text
					@click="emit('cancel', $event, item.id)"
				/>
			</div>
		</div>

		<!-- 任务执行进度条 -->
		<p-progress-bar
			:value="item.progress"
			:showValue="false"
			style="height: 6px;"
		/>

		<!-- 任务创建时间与百分比 -->
		<div class="task-meta">
			<span class="text-sm text-muted">创建时间: {{ timeText }}</span>
			<span class="text-sm font-mono">{{ item.progress }}%</span>
		</div>

		<!-- 任务异常错误框 -->
		<div v-if="item.errorMessage" class="task-err font-mono">
			错误: [{{ item.errorCode }}] {{ item.errorMessage }}
		</div>
	</div>
</template>

<script setup lang="ts">
	import { computed } from 'vue'
	import type { Task } from '@/types'

	// 接收任务数据
	const props = defineProps<{
		item: Task
	}>()

	// 定义取消事件
	const emit = defineEmits<{
		cancel: [event: Event, id: number]
	}>()

	// 任务状态文案映射
	const statusMap: Record<string, string> = {
		queued: '排队中',
		running: '执行中',
		cancelling: '取消中',
		succeeded: '已完成',
		failed: '失败',
		cancelled: '已取消'
	}

	// 任务状态颜色映射
	const toneMap: Record<string, string> = {
		succeeded: 'success',
		running: 'info',
		failed: 'danger',
		cancelling: 'warn',
		cancelled: 'warn'
	}

	// 当前状态展示文案
	const statusText = computed(() => {
		const hit = statusMap[props.item.status]
		if (hit) {
			return hit
		}
		return props.item.status
	})

	// 当前状态标签样式
	const statusTone = computed(() => {
		const hit = toneMap[props.item.status]
		if (hit) {
			return hit
		}
		return 'secondary'
	})

	// 格式化时间文本
	const timeText = computed(() => {
		if (!props.item.createdAt) {
			return ''
		}
		return new Date(props.item.createdAt).toLocaleString()
	})
</script>

<style scoped lang="scss">
	/* 任务单项容器 */
	.task-box {
		background: var(--bg-app);
		border: 1px solid var(--border-color);
		border-radius: var(--radius-md);
		padding: 16px;
		display: flex;
		flex-direction: column;
		gap: 10px;
	}

	/* 顶部信息行 */
	.task-top {
		display: flex;
		justify-content: space-between;
		align-items: center;
	}

	/* 左侧标题区 */
	.task-left {
		display: flex;
		align-items: center;
		gap: 10px;
	}

	/* 任务类型标签 */
	.task-type {
		background: var(--info-bg);
		color: var(--info);
		padding: 2px 6px;
		border-radius: var(--radius-sm);
		font-size: var(--font-xs);
		border: 1px solid var(--info-border);
	}

	/* 任务编号 */
	.task-id {
		color: var(--text-muted);
		font-size: var(--font-sm);
	}

	/* 任务描述 */
	.task-msg {
		font-size: var(--font-base);
		color: var(--text-primary);
	}

	/* 右侧状态与操作区 */
	.task-right {
		display: flex;
		align-items: center;
		gap: 10px;
	}

	/* 底部元信息行 */
	.task-meta {
		display: flex;
		justify-content: space-between;
		align-items: center;
	}

	/* 错误信息框 */
	.task-err {
		background: var(--err-bg);
		border: 1px solid var(--err-border);
		color: var(--err);
		padding: 8px 12px;
		border-radius: var(--radius-sm);
		font-size: var(--font-xs);
	}

	/* 小号字体 */
	.text-sm {
		font-size: var(--font-sm);
	}

	/* 弱化文字颜色 */
	.text-muted {
		color: var(--text-muted);
	}
</style>
