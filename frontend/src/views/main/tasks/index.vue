<template>
	<!-- 任务中心页面容器 -->
	<h-page title="任务与后台作业中心" desc="查看环境安装、备份、恢复及异步处理任务的进度与状态">
		<!-- 页面右上角刷新按钮 -->
		<template #extra>
			<p-button
				label="刷新"
				icon="pi pi-refresh"
				severity="secondary"
				outlined
				:loading="loading"
				@click="loadTasks"
			/>
		</template>

		<!-- 任务列表卡片 -->
		<h-card>
			<!-- 空状态提示 -->
			<div v-if="tasks.length === 0" class="empty-state">
				<span v-if="loading">加载任务中...</span>
				<span v-else>当前没有正在运行或历史记录的任务</span>
			</div>

			<!-- 任务列表展示 -->
			<div v-else class="task-list">
				<task-item
					v-for="t in tasks"
					:key="t.id"
					:item="t"
					@cancel="stopTask"
				/>
			</div>
		</h-card>
	</h-page>
</template>

<script setup lang="ts">
	import TaskItem from './comps/item.vue'
	import { useTasks } from './use-tasks'

	// 解构任务列表状态与方法
	const {
		tasks,
		loading,
		loadTasks,
		stopTask
	} = useTasks()
</script>

<style scoped lang="scss">
	/* 任务列表垂直排列 */
	.task-list {
		display: flex;
		flex-direction: column;
		gap: 14px;
	}

	/* 空状态提示样式 */
	.empty-state {
		text-align: center;
		padding: 40px;
		color: var(--text-muted);
	}
</style>
