<template>
	<!-- 服务状态卡片 -->
	<h-card title="服务状态">
		<!-- 卡片右上角刷新按钮 -->
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

		<!-- 空状态提示 -->
		<h-empty
			v-if="daemonList.length === 0"
			icon="pi pi-server"
			description="暂无已安装的服务"
		/>

		<!-- 服务列表 -->
		<div v-else class="list">
			<h-service
				v-for="item in daemonList"
				:key="item.id"
				:item="item"
				:busy="busyId === item.id"
				:configurable="false"
				@start="startService"
				@stop="stopService"
				@manage="$router.push('/software')"
			/>
		</div>
	</h-card>
</template>

<script setup lang="ts">
import { onMounted, onUnmounted } from 'vue'
import { listenEvent } from '@/api'
import { useDashboard } from '@/composables'
import type { ServiceEvent } from '@/types'

// 概览软件数据与刷新方法
const { daemonList, loading, busyId, loadData, startService, stopService } = useDashboard()

let stopEvents: (() => void) | null = null

onMounted(() => {
	stopEvents = listenEvent('service:status-changed', (payload: ServiceEvent) => {
		// Start/Stop are asynchronous. Refresh after the backend has finished so
		// the dashboard reflects the actual process state instead of the request state.
		if (payload.id > 0) {
			void loadData(true)
		}
	})
})

onUnmounted(() => {
	if (stopEvents !== null) {
		stopEvents()
		stopEvents = null
	}
})
</script>

<style scoped lang="scss">
/* 服务列表容器 */
.list {
	display: flex;
	flex-direction: column;
	gap: 0.5rem;
}
</style>
