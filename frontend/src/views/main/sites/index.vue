<template>
	<!-- 网站管理主页面 -->
	<h-page title="网站管理" desc="整合项目管理与虚拟主机，统一管理静态、PHP、Node.js、Go、Python 网站及 Git 与 Nginx 配置">
		<!-- 页面右上角快捷操作按钮组 -->
		<template #extra>
			<p-button
				label="刷新"
				icon="pi pi-refresh"
				severity="secondary"
				outlined
				:loading="loading"
				@click="loadAll"
			/>
			<p-button
				label="批量启动"
				icon="pi pi-play"
				severity="secondary"
				outlined
				:disabled="sites.length === 0"
				@click="batchRun('start')"
			/>
			<p-button
				label="批量停止"
				icon="pi pi-stop"
				severity="secondary"
				outlined
				:disabled="sites.length === 0"
				@click="batchRun('stop')"
			/>
			<p-button
				label="添加网站"
				icon="pi pi-plus"
				@click="openWizard"
			/>
		</template>

		<!-- 顶部搜索、目录选择与类型筛选栏 -->
		<sites-toolbar />

		<!-- 空状态引导卡片 -->
		<sites-empty
			v-if="filteredSites.length === 0 && !loading"
			@create="openWizard"
		/>

		<!-- 网站卡片列表 -->
		<div v-else class="sites-list">
			<site-card
				v-for="site in filteredSites"
				:key="site.id"
				:item="site"
			/>
		</div>

		<!-- 4步添加网站向导弹窗 -->
		<site-wizard v-model:visible="showWizard" />

		<!-- 网站详情综合管理弹窗 -->
		<site-detail v-model:visible="showDetail" />
	</h-page>
</template>

<script setup lang="ts">
	import { onMounted } from 'vue'
	import SitesToolbar from './comps/toolbar.vue'
	import SitesEmpty from './comps/empty.vue'
	import SiteCard from './comps/card.vue'
	import SiteWizard from './comps/wizard.vue'
	import SiteDetail from './comps/detail.vue'
	import { useSites } from './use-sites'
	import { useWizard } from './use-wizard'
	import { useDetail } from './use-detail'

	// 解构网站列表状态与方法
	const {
		sites,
		loading,
		filteredSites,
		loadAll,
		batchRun
	} = useSites()

	// 解构添加网站向导状态
	const { showWizard, openWizard } = useWizard()

	// 解构网站详情弹窗状态
	const { showDetail } = useDetail()

	// 页面挂载时加载网站与环境列表
	onMounted(loadAll)
</script>

<style scoped lang="scss">
	/* 网站卡片列表垂直排列 */
	.sites-list {
		display: flex;
		flex-direction: column;
		gap: 14px;
	}
</style>
