<template>
	<!-- 最近项目条目行 -->
	<div class="project-row" @click="$router.push('/projects')">
			<!-- 左侧项目名称与路径 -->
		<div class="info">
			<div class="title-line">
				<span class="name">{{ project.name }}</span>
				<span class="kind font-mono">{{ project.kind.toUpperCase() }}</span>
			</div>
			<span class="path font-mono" :title="hostText">
				{{ hostText }} · {{ project.root }}
			</span>
		</div>

		<!-- 右侧状态标签 -->
		<h-tag
			v-if="project.state === 'starting'"
			color="info"
			value="启动中"
			dot
			radius="full"
		/>
		<h-tag
			v-else-if="project.state === 'running'"
			color="success"
			value="运行中"
			dot
			radius="full"
		/>
		<h-tag
			v-else-if="project.need !== ''"
			color="warning"
			value="缺少环境"
		/>
		<h-tag
			v-else-if="project.state === 'error'"
			color="danger"
			value="异常"
		/>
		<h-tag
			v-else
			color="default"
			value="已停止"
		/>
	</div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { Project } from '@/types'

// 接收项目属性
const props = defineProps<{
	project: Project
}>()

// 域名拼接文本
const hostText = computed(() => {
	if (props.project.hosts.length === 0) {
		return '本地调试'
	}
	return props.project.hosts.join(', ')
})
</script>

<style scoped lang="scss">
/* 项目行容器 */
.project-row {
	display: flex;
	justify-content: space-between;
	align-items: center;
	padding: 0.75rem 0.875rem;
	background: var(--bg-app);
	border: 1px solid var(--border-color);
	border-radius: var(--h-radius);
	cursor: pointer;
	transition: all 0.15s ease;

	/* 悬停态 */
	&:hover {
		background: var(--bg-card);
		border-color: var(--border-hover);
	}

	/* 左侧文字区 */
	.info {
		display: flex;
		flex-direction: column;
		gap: 0.2rem;
		min-width: 0;
	}

	/* 项目名与类型行 */
	.title-line {
		display: flex;
		align-items: center;
		gap: 0.5rem;
	}

	/* 项目名称 */
	.name {
		font-size: var(--font-base);
		color: var(--text-primary);
	}

	/* 项目类型标识 */
	.kind {
		font-size: var(--font-xs);
		color: var(--text-muted);
	}

	/* 域名与目录路径 */
	.path {
		font-size: var(--font-xs);
		color: var(--text-muted);
		max-width: 20rem;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}
}
</style>
