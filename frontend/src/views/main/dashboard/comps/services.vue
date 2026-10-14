<template>
	<div class="panel surface-card">
		<!-- 面板头部 -->
		<div class="header">
			<div class="title">
				<i class="pi pi-bolt icon"></i>
				<span>核心服务状态与控制</span>
			</div>
			<p-button
				icon="pi pi-refresh"
				label="刷新"
				size="small"
				severity="secondary"
				outlined
				:loading="loading"
				@click="loadData"
			/>
		</div>

		<!-- 空状态 -->
		<div v-if="services.length === 0" class="empty">
			暂无已配置的服务
		</div>

		<!-- 服务列表 -->
		<div v-else class="list">
			<div v-for="srv in services" :key="srv.id" class="row">
				<div class="info">
					<div class="title-line">
						<span class="name">{{ srv.name }}</span>
						<p-tag
							v-if="srv.status === 'running'"
							severity="success"
							value="运行中"
							size="small"
						/>
						<p-tag
							v-else
							severity="secondary"
							value="已停止"
							size="small"
						/>
					</div>
					<span class="meta font-mono">
						<span v-if="srv.port">端口: {{ srv.port }}</span>
						<span v-else>端口: 无</span>
						<span v-if="srv.status === 'running'"> · PID: {{ srv.pid }}</span>
						<span v-else> · 空闲</span>
					</span>
				</div>

				<div class="actions">
					<p-button
						v-if="srv.status !== 'running'"
						label="启动"
						icon="pi pi-play"
						size="small"
						severity="success"
						:loading="busyId === srv.id"
						@click="startSrv(srv.id)"
					/>
					<p-button
						v-else
						label="停止"
						icon="pi pi-stop"
						size="small"
						severity="danger"
						outlined
						:loading="busyId === srv.id"
						@click="stopSrv(srv.id)"
					/>
				</div>
			</div>
		</div>
	</div>
</template>

<script setup lang="ts">
import { useDashboard } from '../use-dashboard'

// 获取控制台共享状态与操作
const { services, loading, busyId, loadData, startSrv, stopSrv } = useDashboard()
</script>

<style scoped lang="scss">
/* 服务状态面板 */
.panel {
	padding: var(--h-padding);
	display: flex;
	flex-direction: column;
	gap: var(--h-padding);
}

/* 面板顶部栏 */
.header {
	display: flex;
	justify-content: space-between;
	align-items: center;

	.title {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		font-size: 1rem;
		color: var(--text-primary);
	}

	.icon {
		color: var(--info);
	}
}

/* 服务项列表 */
.list {
	display: flex;
	flex-direction: column;
	gap: 0.5rem;
}

/* 单个服务行 */
.row {
	display: flex;
	justify-content: space-between;
	align-items: center;
	padding: 0.75rem 0.875rem;
	background: var(--bg-app);
	border: 1px solid var(--border-color);
	border-radius: var(--radius-sm);

	&:hover {
		background: var(--bg-card);
		border-color: var(--border-hover);
	}

	.info {
		display: flex;
		flex-direction: column;
		gap: 0.2rem;
	}

	.title-line {
		display: flex;
		align-items: center;
		gap: 0.5rem;
	}

	.name {
		font-size: 0.9rem;
		color: var(--text-primary);
	}

	.meta {
		font-size: var(--font-xs);
		color: var(--text-muted);
	}
}

/* 空状态提示 */
.empty {
	text-align: center;
	padding: 2rem;
	color: var(--text-muted);
	font-size: var(--font-sm);
	background: var(--bg-app);
	border: 1px dashed var(--border-color);
	border-radius: var(--h-radius);
}
</style>
