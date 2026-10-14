<template>
	<h-page title="开发环境概览" desc="Windows 本地高效 PHP / Web 开发栈一站式控制中心">
		<!-- 头部右侧状态徽章 -->
		<template #extra>
			<p-tag severity="info" value="进程管理器: 运行中" icon="pi pi-server" />
		</template>

		<!-- 顶部统计卡 -->
		<dash-stats />

		<!-- 核心服务与最近网站双栏 -->
		<div class="dash-grid">
			<dash-services />
			<dash-sites />
		</div>
	</h-page>
</template>

<script setup lang="ts">
import { onMounted } from 'vue'
import { useDashboard } from './use-dashboard'
import DashStats from './comps/stats.vue'
import DashServices from './comps/services.vue'
import DashSites from './comps/sites.vue'

// 控制台状态钩子
const { loadData } = useDashboard()

// 页面挂载加载数据
onMounted(() => {
	loadData()
})
</script>

<style scoped lang="scss">
/* 控制台双栏网格 */
.dash-grid {
	display: grid;
	grid-template-columns: 1fr 1fr;
	gap: var(--h-padding);
}
</style>
