<template>
	<!-- 服务实例管理页面 -->
	<h-page title="服务实例管理" desc="监控与管理本地 Web 服务器 (Nginx)、数据库 (MySQL)、缓存 (Redis) 与 PHP FastCGI 进程">
		<!-- 页面右上角刷新状态按钮 -->
		<template #extra>
			<p-button
				label="刷新状态"
				icon="pi pi-refresh"
				severity="secondary"
				outlined
				:loading="loading"
				@click="loadServices"
			/>
		</template>

		<!-- 空状态提示 -->
		<div v-if="services.length === 0 && !loading" class="empty-srv">
			<i class="pi pi-server empty-icon"></i>
			<p class="empty-title">暂无服务实例，请先在软件商店安装运行环境</p>
			<p class="empty-sub">安装 PHP、Nginx、MySQL 或 Redis 后将自动创建对应的服务实例</p>
			<p-button
				label="前往软件商店"
				icon="pi pi-shopping-bag"
				@click="$router.push('/store')"
			/>
		</div>

		<!-- 服务实例卡片网格 -->
		<div v-else class="srv-grid">
			<srv-card
				v-for="srv in services"
				:key="srv.id"
				:item="srv"
				:options="runtimeOptions[srv.id] || []"
				:busy="busyId === srv.id"
				@start="startSrv"
				@stop="stopSrv"
				@restart="restartSrv"
				@logs="openLogs"
				@remove="removeSrv"
				@switch-version="switchVersion"
			/>
		</div>

		<!-- 实时日志终端面板 -->
		<srv-logs
			:service="curSrv"
			:logs="logs"
			@refresh="fetchLogs"
			@close="closeLogs"
		/>
	</h-page>
</template>

<script setup lang="ts">
	import SrvCard from './comps/card.vue'
	import SrvLogs from './comps/logs.vue'
	import { useServices } from './use-services'

	// 解构服务列表与控制方法
	const {
		services,
		runtimeOptions,
		loading,
		busyId,
		curSrv,
		logs,
		loadServices,
		startSrv,
		stopSrv,
		restartSrv,
		removeSrv,
		switchVersion,
		openLogs,
		fetchLogs,
		closeLogs
	} = useServices()
</script>

<style scoped lang="scss">
	/* 服务实例网格布局 */
	.srv-grid {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(350px, 1fr));
		gap: 16px;
	}

	/* 空状态提示框 */
	.empty-srv {
		text-align: center;
		padding: 60px 20px;
		background: var(--bg-card);
		border: 1px dashed var(--border-color);
		border-radius: var(--h-radius);
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 8px;
	}

	.empty-icon {
		font-size: 2.5rem;
		color: var(--text-muted);
	}

	.empty-title {
		font-size: var(--font-lg);
		color: var(--text-primary);
		margin: 0;
	}

	.empty-sub {
		font-size: var(--font-sm);
		color: var(--text-muted);
		margin: 0 0 8px 0;
	}
</style>

