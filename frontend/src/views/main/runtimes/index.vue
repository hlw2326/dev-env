<template>
	<!-- 运行环境管理页面容器 -->
	<h-page title="运行环境" desc="统一管理已安装的 PHP、Nginx、MySQL 等本地开发环境版本">
		<!-- 页面右上角操作按钮组 -->
		<template #extra>
			<p-button label="刷新扫描" icon="pi pi-refresh" severity="secondary" outlined :loading="loading" @click="loadRuntimes" />
			<p-button label="导入本地环境" icon="pi pi-folder-open" severity="secondary" outlined @click="openImport" />
			<p-button label="软件商店" icon="pi pi-shopping-bag" @click="$router.push('/store')" />
		</template>

		<!-- 空状态引导区 -->
		<div v-if="runtimes.length === 0 && !loading" class="empty-state">
			<i class="pi pi-box empty-icon"></i>
			<p class="empty-title">暂未安装任何运行环境</p>
			<p class="empty-sub">前往软件商店可一键快速下载并安装 PHP、Nginx 等开发环境</p>
			<p-button label="前往软件商店安装" icon="pi pi-arrow-right" @click="$router.push('/store')" />
		</div>

		<!-- 已安装运行环境分组列表 -->
		<div v-else class="groups-box">
			<rt-group
				v-for="(items, type) in grouped"
				:key="type"
				:type="String(type)"
				:items="items"
				:expanded="isGroupOpen(String(type))"
				:busyId="busyId"
				@toggle="toggleGroup"
				@ext="openExt"
				@default="setDefault"
				@health="checkHealth"
				@uninstall="openUninstall"
			/>
		</div>

		<!-- PHP 扩展管理弹窗 -->
		<ext-modal v-model:visible="showExt" :runtime="curRt" :list="extensions" :loading="loadingExt" />

		<!-- 卸载确认弹窗 -->
		<uninstall-modal
			v-model:visible="showUninstall"
			v-model:purge="purgeFiles"
			:runtime="curRt"
			:same-type-count="curRtSameTypeCount"
			:loading="uninstalling"
			@confirm="confirmUninstall"
		/>

		<!-- 导入本地环境弹窗 -->
		<import-modal v-model:visible="showImport" v-model:form="importForm" :loading="importing" @pick-dir="pickImportDir" @submit="submitImport" />
	</h-page>
</template>

<script setup lang="ts">
	import RtGroup from './comps/group.vue'
	import ExtModal from './comps/ext-modal.vue'
	import UninstallModal from './comps/uninstall-modal.vue'
	import ImportModal from './comps/import-modal.vue'
	import { useRuntimes } from './use-runtimes'

	// 解构运行环境管理状态与方法
	const {
		loading, busyId, runtimes, grouped, showExt, curRt, curRtSameTypeCount, extensions, loadingExt,
		showUninstall, purgeFiles, uninstalling, showImport, importing, importForm,
		isGroupOpen, toggleGroup, loadRuntimes, setDefault, checkHealth,
		openExt, openUninstall, confirmUninstall, openImport, pickImportDir, submitImport
	} = useRuntimes()
</script>

<style scoped lang="scss">
	/* 分组列表垂直排列 */
	.groups-box {
		display: flex;
		flex-direction: column;
		gap: var(--h-padding);
	}

	/* 空状态引导框 */
	.empty-state {
		text-align: center;
		padding: 60px 20px;
		background: var(--bg-card);
		border: 1px dashed var(--border-color);
		border-radius: var(--h-radius);
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 8px;
	}

	/* 空状态图标与文字 */
	.empty-icon { font-size: 2.5rem; color: var(--text-muted); }
	.empty-title { font-size: var(--font-lg); color: var(--text-primary); margin: 0; }
	.empty-sub { font-size: var(--font-sm); color: var(--text-muted); margin: 0 0 8px 0; }
</style>
