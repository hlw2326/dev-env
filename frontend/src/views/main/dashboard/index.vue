<template>
	<h-page title="开发环境概览" description="Windows 本地高效 PHP / Web 开发栈一站式控制中心">
		<template #extra>
			<h-tag v-if="!initialized && dashboardError !== ''" color="danger" value="Nginx: 检测失败" icon="pi pi-exclamation-triangle" />
			<h-tag v-else-if="!initialized" color="info" value="Nginx: 检测中" icon="pi pi-spin pi-spinner" />
			<h-tag
				v-else-if="listening"
				color="success"
				:value="'Nginx: 监听 127.0.0.1:' + nginxPort"
				icon="pi pi-server"
			/>
			<h-tag
				v-else
				color="danger"
				value="Nginx: 未监听"
				icon="pi pi-exclamation-triangle"
			/>
		</template>

		<comp-stats />
		<comp-resources />

		<div class="dash-grid">
			<comp-services />
			<comp-projects />
		</div>
	</h-page>
</template>

<script setup lang="ts">
import { onMounted, onUnmounted } from 'vue'
import { listenEvent } from '@/api'
import { useDashboard } from '@/composables'
import CompStats from './comps/stats.vue'
import CompResources from './comps/resources.vue'
import CompServices from './comps/services.vue'
import CompProjects from './comps/projects.vue'

const { listening, nginxPort, loadData, initialized, dashboardError } = useDashboard()
let refreshTimer: number | null = null
let stopServiceEvents: (() => void) | null = null
let stopProjectEvents: (() => void) | null = null

onMounted(() => {
	stopServiceEvents = listenEvent('service:status-changed', () => void loadData(true))
	stopProjectEvents = listenEvent('project:status-changed', () => void loadData(true))
	void loadData()
	refreshTimer = window.setInterval(() => {
		void loadData(true)
	}, 5000)
})

onUnmounted(() => {
	if (stopServiceEvents !== null) stopServiceEvents()
	if (stopProjectEvents !== null) stopProjectEvents()
	if (refreshTimer !== null) {
		window.clearInterval(refreshTimer)
		refreshTimer = null
	}
})
</script>

<style scoped lang="scss">
.dash-grid {
	display: grid;
	grid-template-columns: 1fr 1fr;
	gap: var(--h-padding);
}
</style>
