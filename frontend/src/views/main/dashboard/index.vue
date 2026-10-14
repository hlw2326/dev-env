<template>
	<h-page title="开发环境概览" desc="Windows 本地高效 PHP / Web 开发栈一站式控制中心">
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

		<dash-stats />

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

const { listening, gatewayPort, loadData } = useDashboard()

onMounted(() => {
	loadData()
})
</script>

<style scoped lang="scss">
.dash-grid {
	display: grid;
	grid-template-columns: 1fr 1fr;
	gap: var(--h-padding);
}
</style>
