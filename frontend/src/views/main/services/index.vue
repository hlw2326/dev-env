<template>
	<!-- 运行中的服务页面 -->
	<h-page title="运行服务" description="查看当前运行的数据库与缓存服务">
		<!-- 页面右上角刷新按钮 -->
		<template #extra>
			<p-button
				icon="pi pi-refresh"
				label="刷新"
				size="small"
				severity="secondary"
				variant="outlined"
				:loading="loading"
				@click="loadData"
			/>
		</template>

		<!-- 无运行服务提示 -->
		<div v-if="running.length === 0" class="surface-card">
			<h-empty
				icon="pi pi-server"
				title="当前没有运行中的服务"
				description="可以前往软件商店启动已安装的数据库或缓存服务"
			>
				<p-button icon="pi pi-shopping-bag" label="前往软件商店" @click="$router.push('/software')" />
			</h-empty>
		</div>

		<!-- 运行服务列表 -->
		<div v-else class="list">
			<h-service
				v-for="item in running"
				:key="item.id"
				:item="item"
				:busy="busy === item.id"
				@stop="stopService"
			/>
		</div>
	</h-page>
</template>

<script setup lang="ts">
import { onMounted, onUnmounted } from 'vue'
import { useServices } from './use-services'

// 运行服务数据与控制方法
const { running, loading, busy, loadData, stopService, bindEvents } = useServices()

// 服务事件解绑函数
let stopEvents: (() => void) | null = null

// 挂载时加载并监听服务变化
onMounted(() => {
	loadData()
	stopEvents = bindEvents()
})

// 卸载时移除服务事件监听
onUnmounted(() => {
	if (stopEvents) {
		stopEvents()
		stopEvents = null
	}
})
</script>

<style scoped lang="scss">
/* 运行服务列表 */
.list {
	display: flex;
	flex-direction: column;
	gap: 0.625rem;
}
</style>
