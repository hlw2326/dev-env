<template>
	<!-- 网站管理页面 -->
	<h-page title="网站" description="管理本地 PHP、Node.js 与静态网站，所有站点共用本地网关">
		<!-- 页面右上角创建按钮 -->
		<template #extra>
			<p-button icon="pi pi-plus" label="创建网站" @click="createVisible = true" />
		</template>

		<!-- 顶部筛选与搜索工具栏 -->
		<comp-toolbar />

		<!-- 空状态提示 -->
		<div v-if="filteredSites.length === 0" class="surface-card">
			<h-empty
				icon="pi pi-globe"
				title="还没有创建任何网站"
				description="选择站点类型、填写名称和域名即可一键创建本地网站"
			>
				<p-button icon="pi pi-plus" label="创建第一个网站" @click="createVisible = true" />
			</h-empty>
		</div>

		<!-- 网站列表区 -->
		<div v-else class="site-list">
			<comp-row
				v-for="item in filteredSites"
				:key="item.id"
				:site="item"
				:installing-kind="installingKind"
				@toggle="toggleSite"
				@open="openSite"
				@folder="browseFolder"
				@detail="openDetail"
				@delete="removeSite"
				@install="quickInstall"
			/>
		</div>

		<!-- 创建网站弹窗 -->
		<comp-create
			v-model:visible="createVisible"
			:default-root="defaultRoot"
			@submit="onCreate"
		/>

		<!-- 站点详情弹窗 -->
		<comp-detail
			v-model:visible="detailVisible"
			:site="activeSite"
			:software-list="softwareList"
			@save="submitUpdate"
		/>
	</h-page>
</template>

<script setup lang="ts">
import { onMounted, onUnmounted } from 'vue'
import { useSites } from './use-sites'
import CompToolbar from './comps/toolbar.vue'
import CompRow from './comps/row.vue'
import CompCreate from './comps/create.vue'
import CompDetail from './comps/detail.vue'
import type { CreateInput } from '@/types'

// 网站管理状态与操作方法
const {
	filteredSites,
	softwareList,
	defaultRoot,
	createVisible,
	activeSite,
	detailVisible,
	installingKind,
	refreshAll,
	submitCreate,
	submitUpdate,
	toggleSite,
	removeSite,
	quickInstall,
	openSite,
	browseFolder,
	openDetail,
	bindEvents,
} = useSites()

// 事件解绑函数引用
let stopEvents: (() => void) | null = null

// 处理创建网站表单提交
async function onCreate(input: CreateInput): Promise<void> {
	const ok = await submitCreate(input)
	if (ok) {
		createVisible.value = false
	}
}

// 挂载时刷新站点并监听事件
onMounted(() => {
	refreshAll()
	stopEvents = bindEvents()
})

// 卸载时清理事件监听
onUnmounted(() => {
	if (stopEvents) {
		stopEvents()
		stopEvents = null
	}
})
</script>

<style scoped lang="scss">
/* 网站列表容器 */
.site-list {
	display: flex;
	flex-direction: column;
	gap: 0.625rem;
}
</style>
