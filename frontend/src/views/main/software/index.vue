<template>
	<!-- 软件商店页面 -->
	<h-page title="软件商店" description="一键安装、多版本并存、按需启停与可视化配置本地开发环境服务">
		<!-- 顶部页签与搜索栏 -->
		<comp-toolbar />

		<!-- 版本清单加载失败提示 -->
		<div v-if="marketError !== ''" class="error-banner surface-card">
			<span class="err-text">{{ marketError }}</span>
			<p-button size="small" icon="pi pi-refresh" label="重试" @click="refreshAll" />
		</div>

		<!-- 全部软件商店双列网格 -->
		<div v-if="tab === 'market'" class="market-grid">
			<comp-market-card
				v-for="item in filteredMarket"
				:key="item.kind"
				:item="item"
				:installed-list="installedList"
				:progress-map="progressMap"
				@install="openInstall"
				@cancel="stopInstall"
				@retry="startInstall"
				@config="showConfig"
			/>
		</div>

		<!-- 已安装软件分组列表 -->
		<comp-installed-groups v-else />

		<!-- 选择版本安装弹窗 -->
		<comp-install-modal
			v-model:visible="modalVisible"
			:item="marketItem"
			:versions="versionOptions"
			:progress-map="progressMap"
			:installed-list="installedList"
			:loading="versionsLoading"
			:error="versionsError"
			@install="startInstall"
			@cancel="stopInstall"
			@retry="fetchVersions"
		/>

		<!-- 软件配置弹窗 -->
		<comp-config-drawer
			v-model:visible="configVisible"
			:item="activeSoftware"
			:busy="activeSoftware !== null && isServiceBusy(activeSoftware.id)"
			@start="startItem"
			@stop="stopItem"
			@main="markMain"
			@autostart="toggleAutostart"
			@folder="browseDir"
		/>
	</h-page>
</template>

<script setup lang="ts">
import { onMounted, onUnmounted } from 'vue'
import { useSoftware } from '@/composables'
import CompToolbar from './comps/toolbar.vue'
import CompMarketCard from './comps/market-card.vue'
import CompInstalledGroups from './comps/installed-groups.vue'
import CompInstallModal from './comps/install-modal.vue'
import CompConfigDrawer from './comps/config-drawer.vue'

// 软件商店状态与操作方法
const {
	tab,
	filteredMarket,
	installedList,
	marketError,
	progressMap,
	modalVisible,
	marketItem,
	versionOptions,
	versionsLoading,
	versionsError,
	configVisible,
	activeSoftware,
	refreshAll,
	openInstall,
	fetchVersions,
	startInstall,
	stopInstall,
	markMain,
	startItem,
	stopItem,
	isServiceBusy,
	toggleAutostart,
	browseDir,
	showConfig,
	bindEvents,
} = useSoftware()

// 事件解绑函数引用
let stopEvents: (() => void) | null = null

// 挂载时刷新软件商店并监听事件
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
/* 错误提示横幅 */
.error-banner {
	display: flex;
	align-items: center;
	justify-content: space-between;
	padding: 0.75rem 1rem;
	background: var(--err-bg);
	border-color: var(--err-border);

	/* 错误文字 */
	.err-text {
		font-size: var(--font-sm);
		color: var(--err);
	}
}

/* 商店卡片双列网格 */
.market-grid {
	display: grid;
	grid-template-columns: repeat(2, 1fr);
	gap: var(--h-padding);
}
</style>
