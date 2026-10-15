<template>
	<!-- 环境概览页面 -->
	<h-page title="开发环境概览" description="Windows 本地高效 PHP / Web 开发栈一站式控制中心">
		<!-- 页面右上角网关监听状态 -->
		<template #extra>
			<p-tag
				v-if="listening"
				severity="success"
				:value="'本地网关: 监听 127.0.0.1:' + gatewayPort"
				icon="pi pi-server"
			/>
			<p-tag
				v-else
				severity="danger"
				value="本地网关: 未监听"
				icon="pi pi-exclamation-triangle"
			/>
		</template>

		<!-- 统计卡片区 -->
		<comp-stats />

		<!-- 服务状态与最近站点双列网格 -->
		<div class="dash-grid">
			<comp-services />
			<comp-sites />
		</div>
	</h-page>
</template>

<script setup lang="ts">
import { onMounted } from 'vue'
import { useDashboard } from './use-dashboard'
import CompStats from './comps/stats.vue'
import CompServices from './comps/services.vue'
import CompSites from './comps/sites.vue'

// 概览页网关状态与加载函数
const { listening, gatewayPort, loadData } = useDashboard()

// 挂载时加载概览数据
onMounted(() => {
	loadData()
})
</script>

<style scoped lang="scss">
/* 概览主面板双列网格 */
.dash-grid {
	display: grid;
	grid-template-columns: 1fr 1fr;
	gap: var(--h-padding);
}
</style>
