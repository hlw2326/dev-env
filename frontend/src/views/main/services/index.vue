<template>
	<!-- 服务管理页面 -->
	<h-page title="服务管理" description="统一查看并控制已安装的数据库、缓存与 Web 服务">
		<!-- 页面右上角刷新按钮 -->
		<template #extra>
			<p-button
				icon="pi pi-refresh"
				label="刷新"
				size="small"
				severity="secondary"
				variant="outlined"
				:loading="loading"
				@click="loadData()"
			/>
		</template>

		<!-- 无已安装服务提示 -->
		<p-message v-if="!ready && checking" severity="secondary" size="small" icon="pi pi-spin pi-spinner">正在检测已安装服务的运行状态…</p-message>
		<p-message v-else-if="!ready && statusError !== ''" severity="error" size="small">检测失败：{{ statusError }}，请点击刷新重试。</p-message>
		<div v-else-if="services.length === 0" class="surface-card">
			<h-empty
				icon="pi pi-server"
				title="当前没有可管理的服务"
				description="可以前往软件商店安装数据库、缓存或 Web 服务"
			>
				<p-button size="small" icon="pi pi-shopping-bag" label="前往软件商店" @click="$router.push('/software')" />
			</h-empty>
		</div>

		<!-- 服务管理列表 -->
		<div v-else class="list">
			<h-service
				v-for="item in services"
				:key="item.id"
				:item="item"
				:busy="busy === item.id"
				:checking="checking"
				:unavailable="statusError !== ''"
				:restarting="busy === item.id && operation === 'restart'"
				show-restart
				@start="startService"
				@stop="stopService"
				@restart="restartService"
				@configure="openSettings"
				@data="openRedisData"
			/>
		</div>

		<h-service-config
			v-model:visible="settingsVisible"
			:item="activeService"
			:busy="checking || statusError !== '' || (activeService !== null && busy === activeService.id)"
			@start="startConfiguredService"
			@stop="stopConfiguredService"
			@main="markMain"
			@autostart="toggleAutostart"
			@folder="browseFolder"
		/>

		<comp-redis-data />
	</h-page>
</template>

<script setup lang="ts">
import { onMounted, onUnmounted, ref, watch } from 'vue'
import { useRedisData, useServices } from '@/composables'
import type { Software } from '@/types'
import CompRedisData from './comps/redis-data.vue'

// 服务管理数据与控制方法
const {
	services,
	ready,
	checking,
	statusError,
	loading,
	busy,
	operation,
	loadData,
	startService,
	stopService,
	restartService,
	startConfiguredService,
	stopConfiguredService,
	markMain,
	toggleAutostart,
	browseFolder,
	bindEvents,
} = useServices()
const { open: openRedisData } = useRedisData()

const settingsVisible = ref(false)
const activeService = ref<Software | null>(null)

function openSettings(item: Software): void {
	activeService.value = item
	settingsVisible.value = true
}

watch(services, (items) => {
	if (activeService.value === null) {
		return
	}
	const activeID = activeService.value.id
	const current = items.find((item) => item.id === activeID)
	if (current) {
		activeService.value = current
	}
})

// 服务事件解绑函数
let stopEvents: (() => void) | null = null
let refreshTimer: number | null = null

// 挂载时加载并监听服务变化
onMounted(() => {
	loadData()
	stopEvents = bindEvents()
	refreshTimer = window.setInterval(() => {
		void loadData(true)
	}, 5000)
})

// 卸载时移除服务事件监听
onUnmounted(() => {
	if (stopEvents) {
		stopEvents()
		stopEvents = null
	}
	if (refreshTimer !== null) {
		window.clearInterval(refreshTimer)
		refreshTimer = null
	}
})
</script>

<style scoped lang="scss">
/* 服务管理列表 */
.list {
	display: flex;
	flex-direction: column;
	gap: 0.625rem;
}
</style>
