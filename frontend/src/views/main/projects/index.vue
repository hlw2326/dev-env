<template>
	<!-- 项目管理页面 -->
	<h-page title="项目管理" description="管理本地开发项目" padding>
		<!-- 页面右上角创建按钮 -->
		<template #extra>
			<p-button icon="pi pi-plus" label="创建项目" @click="openCreate" />
		</template>

		<!-- 顶部筛选与搜索工具栏 -->
		<comp-toolbar :templates="projectTemplates" />

		<!-- 空状态提示 -->
		<div v-if="filteredProjects.length === 0" class="surface-card">
			<h-empty
				icon="pi pi-globe"
				title="还没有创建任何项目"
				description="选择项目类型并填写项目目录，即可创建本地开发项目"
			>
				<p-button icon="pi pi-plus" label="创建第一个项目" @click="openCreate" />
			</h-empty>
		</div>

		<!-- 项目列表区 -->
		<div v-else class="project-list">
			<comp-row
				v-for="item in filteredProjects"
				:key="item.id"
				:project="item"
				:templates="projectTemplates"
				:git-status="gitStatuses[item.id]"
				:installing-kind="installingKind"
				@toggle="toggleProject"
				@open="openProject"
				@folder="browseFolder"
				@detail="openDetail"
				@git="openGit"
				@delete="removeProject"
				@install="quickInstall"
			/>
		</div>

		<!-- 创建项目弹窗 -->
		<comp-create
			v-model:visible="createVisible"
			:default-root="defaultRoot"
			:initial-kind="kindFilter"
			:templates="projectTemplates"
			@submit="onCreate"
		/>

		<!-- 项目详情抽屉 -->
		<project-detail
			v-model:visible="detailVisible"
			:project="activeProject"
			:initial-tab="detailInitialTab"
			:software-list="softwareList"
			:templates="projectTemplates"
			:plugins="projectPlugins"
			:plugin-busy="pluginBusy"
			@save="submitUpdate"
			@plugin="runPlugin"
			@refresh-git="refreshGitStatuses"
		/>
	</h-page>
</template>

<script setup lang="ts">
import { onMounted, onUnmounted } from 'vue'
import { useProjects } from '@/composables'
import CompToolbar from './comps/toolbar.vue'
import CompRow from './comps/row.vue'
import CompCreate from './comps/create.vue'
import ProjectDetail from './detail.vue'
import type { ProjectCreateInput } from '@/types'

// 项目管理状态与操作方法
const {
	filteredProjects,
	softwareList,
	projectTemplates,
	gitStatuses,
	defaultRoot,
	kindFilter,
	createVisible,
	activeProject,
	detailVisible,
	detailInitialTab,
	installingKind,
	projectPlugins,
	pluginBusy,
	refreshAll,
	refreshGitStatuses,
	submitCreate,
	submitUpdate,
	toggleProject,
	removeProject,
	quickInstall,
	openProject,
	browseFolder,
	openDetail,
	openGit,
	runPlugin,
	bindEvents,
} = useProjects()

// 事件解绑函数引用
let stopEvents: (() => void) | null = null
let refreshTimer: number | null = null

function openCreate(): void {
	createVisible.value = true
}

// 处理创建项目表单提交
async function onCreate(input: ProjectCreateInput): Promise<void> {
	const ok = await submitCreate(input)
	if (ok) {
		createVisible.value = false
	}
}

// 挂载时刷新项目并监听事件
onMounted(() => {
	refreshAll()
	stopEvents = bindEvents()
	refreshTimer = window.setInterval(() => {
		void refreshAll(true)
	}, 5000)
})

// 卸载时清理事件监听
onUnmounted(() => {
	if (stopEvents) {
		stopEvents()
		stopEvents = null
	}
	if (refreshTimer !== null) {
		window.clearInterval(refreshTimer)
		refreshTimer = null
	}
})
</script>

<style scoped lang="scss">
/* 项目列表容器 */
.project-list {
	display: flex;
	flex-direction: column;
	gap: 0.625rem;
}
</style>
