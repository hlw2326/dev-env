<template>
	<!-- 系统设置与工作路径页面 -->
	<h-page :title="$t('系统设置与工作路径')" :desc="$t('全局偏好设置、外观主题及开发环境基础路径')">
		<!-- 页面右上角保存与恢复默认按钮 -->
		<template #extra>
			<p-button
				:label="$t('恢复默认')"
				severity="secondary"
				outlined
				@click="resetConfig($event)"
			/>
			<p-button
				:label="$t('保存设置')"
				icon="pi pi-check"
				:loading="saving"
				@click="saveConfig"
			/>
		</template>

		<!-- 设置卡片双列网格 -->
		<div class="settings-grid">
			<!-- 外观、主题与多语言卡片 -->
			<appearance-card
				class="full-col"
				@change-lang="changeLang"
				@change-mode="changeMode"
			/>

			<!-- 存储与工作目录配置卡片 -->
			<paths-card
				v-model="settings"
				@pick-runtime="pickRuntimeDir"
				@pick-project="pickProjectDir"
			/>

			<!-- 服务进程管理卡片 -->
			<service-card v-model="settings" />
		</div>
	</h-page>
</template>

<script setup lang="ts">
	import { $t } from '@/i18n'
	import AppearanceCard from './comps/appearance.vue'
	import PathsCard from './comps/paths.vue'
	import ServiceCard from './comps/service.vue'
	import { useSettings } from './use-settings'

	// 解构设置页状态与方法
	const {
		settings,
		saving,
		pickRuntimeDir,
		pickProjectDir,
		changeLang,
		changeMode,
		saveConfig,
		resetConfig
	} = useSettings()
</script>

<style scoped lang="scss">
	/* 设置双列网格布局 */
	.settings-grid {
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: var(--h-padding);
	}

	/* 跨双列卡片 */
	.full-col {
		grid-column: span 2;
	}
</style>
