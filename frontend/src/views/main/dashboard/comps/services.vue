<template>
	<!-- 核心软件与服务状态卡片 -->
	<h-card title="核心软件与服务状态">
		<!-- 卡片右上角刷新按钮 -->
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

		<!-- 空状态提示 -->
		<h-empty
			v-if="readySoftware.length === 0"
			icon="pi pi-box"
			description="暂无已安装的软件与服务"
		/>

		<!-- 软件与服务列表 -->
		<div v-else class="list">
			<h-service
				v-for="item in readySoftware"
				:key="item.id"
				:item="item"
				:busy="busyId === item.id"
				@start="startService"
				@stop="stopService"
				@manage="$router.push('/software')"
			/>
		</div>
	</h-card>
</template>

<script setup lang="ts">
import { useDashboard } from '../use-dashboard'

// 概览软件数据与刷新方法
const { readySoftware, loading, busyId, loadData, startService, stopService } = useDashboard()
</script>

<style scoped lang="scss">
/* 服务列表容器 */
.list {
	display: flex;
	flex-direction: column;
	gap: 0.5rem;
}
</style>
