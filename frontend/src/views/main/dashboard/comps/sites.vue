<template>
	<!-- 最近项目卡片 -->
	<h-card title="最近项目">
		<!-- 卡片右上角跳转按钮 -->
		<template #extra>
			<p-button
				label="进入项目管理"
				icon="pi pi-arrow-right"
				icon-pos="right"
				size="small"
				@click="$router.push('/sites')"
			/>
		</template>

		<!-- 空状态提示 -->
		<h-empty
			v-if="sites.length === 0"
			icon="pi pi-globe"
			description="尚未创建任何项目"
		/>

		<!-- 最近网站列表 -->
		<div v-else class="list">
			<comp-site-row
				v-for="site in recentSites"
				:key="site.id"
				:site="site"
			/>
		</div>
	</h-card>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useDashboard } from '../use-dashboard'
import CompSiteRow from './site-row.vue'

// 概览站点数据
const { sites } = useDashboard()

// 最近前五个网站
const recentSites = computed(() => {
	return sites.value.slice(0, 5)
})
</script>

<style scoped lang="scss">
/* 网站列表容器 */
.list {
	display: flex;
	flex-direction: column;
	gap: 0.5rem;
}
</style>
