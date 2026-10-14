<template>
	<!-- 软件商店页面容器 -->
	<h-page title="软件商店" desc="发现并在线获取 PHP、Nginx、MySQL、Redis 等主流开发环境最新版本">
		<!-- 分类筛选、搜索与同步工具栏 -->
		<store-toolbar
			v-model:category="curCat"
			v-model:keyword="keyword"
			:syncing="syncing"
			@sync="syncPackages"
		/>

		<!-- 空状态提示 -->
		<div v-if="filtered.length === 0" class="empty-state">
			<i class="pi pi-inbox empty-icon"></i>
			<p class="empty-text">未找到与 "{{ keyword }}" 匹配的软件包</p>
		</div>

		<!-- 软件包卡片网格 -->
		<div v-else class="pkg-grid">
			<store-card
				v-for="pkg in filtered"
				:key="pkg.key"
				:item="pkg"
				@select="openDetail"
			/>
		</div>

		<!-- 软件版本选择与安装弹窗 -->
		<store-install
			v-model:visible="showModal"
			v-model:curVer="curVer"
			v-model:curMirror="curMirror"
			v-model:customDir="customDir"
			:pkg="curPkg"
			:versions="versions"
			:loadingVer="loadingVer"
			:installing="installing"
			:progress="progress"
			:checkInstalled="isVerInstalled"
			@refresh="refreshVer"
			@pick-dir="pickInstallDir"
			@install="startInstall"
		/>
	</h-page>
</template>

<script setup lang="ts">
	import StoreToolbar from './comps/toolbar.vue'
	import StoreCard from './comps/card.vue'
	import StoreInstall from './comps/install.vue'
	import { useStore } from './use-store'

	// 解构软件商店状态与操作方法
	const {
		keyword,
		curCat,
		filtered,
		syncing,
		showModal,
		curPkg,
		versions,
		curVer,
		curMirror,
		customDir,
		loadingVer,
		installing,
		progress,
		isVerInstalled,
		pickInstallDir,
		syncPackages,
		openDetail,
		refreshVer,
		startInstall
	} = useStore()
</script>

<style scoped lang="scss">
	/* 软件包网格布局 */
	.pkg-grid {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
		gap: var(--h-padding);
	}

	/* 空状态提示 */
	.empty-state {
		text-align: center;
		padding: 60px 20px;
		color: var(--text-muted);
	}

	/* 空状态图标 */
	.empty-icon {
		font-size: 2.5rem;
		margin-bottom: 10px;
	}
</style>
