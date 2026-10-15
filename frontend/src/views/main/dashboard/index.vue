<template>
	<!-- 环境概览页面 -->
	<h-page title="开发环境概览" description="Windows 本地高效 PHP / Web 开发栈一站式控制中心">
		<!-- 页面右上角 Nginx 监听状态 -->
		<template #extra>
			<p-tag
				v-if="listening"
				severity="success"
				:value="'Nginx: 监听 127.0.0.1:' + nginxPort"
				icon="pi pi-server"
			/>
			<p-tag
				v-else
				severity="danger"
				value="Nginx: 未监听"
				icon="pi pi-exclamation-triangle"
			/>
		</template>

		<!-- 统计卡片区 -->
		<comp-stats />

		<!-- 服务状态与最近项目双列网格 -->
		<div class="dash-grid">
			<comp-services />
			<comp-projects />
		</div>
	</h-page>
</template>

<script setup lang="ts">
import { onMounted } from 'vue'
import { useDashboard } from './use-dashboard'
import CompStats from './comps/stats.vue'
import CompServices from './comps/services.vue'
import CompProjects from './comps/projects.vue'

	// 概览页 Nginx 状态与加载函数
	const { listening, nginxPort, loadData } = useDashboard()

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
