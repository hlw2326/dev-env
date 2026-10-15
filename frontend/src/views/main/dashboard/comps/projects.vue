<template>
	<!-- 最近项目卡片 -->
	<h-card title="最近项目">
		<!-- 卡片右上角跳转按钮 -->
		<template #extra>
			<p-button
				label="进入项目管理"
				icon="pi pi-arrow-right"
				icon-pos="right"
				size="small"
				@click="$router.push('/projects')"
			/>
		</template>

		<!-- 空状态提示 -->
		<h-empty
			v-if="projects.length === 0"
			icon="pi pi-globe"
			description="尚未创建任何项目"
		/>

		<!-- 最近项目列表 -->
		<div v-else class="list">
			<comp-project-row
				v-for="project in recentProjects"
				:key="project.id"
				:project="project"
			/>
		</div>
	</h-card>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useDashboard } from '@/composables'
import CompProjectRow from './project-row.vue'

// 概览项目数据
const { projects } = useDashboard()

// 最近五个项目
const recentProjects = computed(() => {
	return projects.value.slice(0, 5)
})
</script>

<style scoped lang="scss">
/* 项目列表容器 */
.list {
	display: flex;
	flex-direction: column;
	gap: 0.5rem;
}
</style>
