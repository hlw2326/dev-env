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
		<p-message v-if="!servicesReady && checking" severity="secondary" size="small" icon="pi pi-spin pi-spinner">正在检测服务运行状态…</p-message>
		<p-message v-else-if="!servicesReady && serviceError !== ''" severity="error" size="small">检测失败：{{ serviceError }}，请点击刷新重试。</p-message>
		<h-empty
			v-else-if="daemonList.length === 0"
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
				:checking="checking"
				:unavailable="serviceError !== ''"
				:configurable="false"
				@start="startService"
				@stop="stopService"
				@manage="$router.push('/software')"
			/>
		</div>
	</h-card>
</template>

<script setup lang="ts">
import { useDashboard } from '@/composables'

// 概览软件数据与刷新方法
const { daemonList, loading, busyId, loadData, startService, stopService, servicesReady, checking, serviceError } = useDashboard()

</script>

<style scoped lang="scss">
/* 服务列表容器 */
.list {
	display: flex;
	flex-direction: column;
	gap: 0.5rem;
}
</style>
