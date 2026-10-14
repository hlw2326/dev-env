<template>
	<div class="panel surface-card">
		<div class="header">
			<div class="title">
				<i class="pi pi-bolt icon"></i>
				<span>核心软件与服务状态</span>
			</div>
			<p-button
				icon="pi pi-refresh"
				label="刷新"
				size="small"
				severity="secondary"
				variant="outlined"
				:loading="loading"
				@click="loadData"
			/>
		</div>

		<div v-if="readySoftware.length === 0" class="empty">
			<span>暂无已安装的软件与服务</span>
		</div>

		<div v-else class="list">
			<div v-for="item in readySoftware" :key="item.id" class="row">
				<div class="info">
					<div class="title-line">
						<span class="name">{{ formatKind(item.kind) }} {{ item.version }}</span>
						<p-tag
							v-if="item.main"
							severity="info"
							value="默认"
							size="small"
						/>
						<template v-if="isDaemon(item.kind)">
							<p-tag
								v-if="item.running"
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
						</template>
						<p-tag
							v-else
							severity="success"
							value="就绪"
							size="small"
						/>
					</div>
					<span class="meta font-mono">
						<template v-if="isDaemon(item.kind)">
							<span>端口: {{ item.port }}</span>
							<span v-if="item.running"> · PID: {{ item.pid }}</span>
							<span v-else> · 空闲</span>
						</template>
						<span v-else>{{ item.path }}</span>
					</span>
				</div>

				<div class="actions">
					<template v-if="isDaemon(item.kind)">
						<p-button
							v-if="!item.running"
							label="启动"
							icon="pi pi-play"
							size="small"
							severity="success"
							:loading="busyId === item.id"
							@click="handleStart(item.id)"
						/>
						<p-button
							v-else
							label="停止"
							icon="pi pi-stop"
							size="small"
							severity="danger"
							variant="outlined"
							:loading="busyId === item.id"
							@click="handleStop(item.id)"
						/>
					</template>
					<p-button
						v-else
						label="管理"
						icon="pi pi-cog"
						size="small"
						severity="secondary"
						variant="outlined"
						@click="$router.push('/software')"
					/>
				</div>
			</div>
		</div>
	</div>
</template>

<script setup lang="ts">
import { useDashboard } from '../use-dashboard'

const { readySoftware, loading, busyId, loadData, handleStart, handleStop } = useDashboard()

function isDaemon(kind: string): boolean {
	return kind === 'mysql' || kind === 'redis' || kind === 'nginx'
}

function formatKind(kind: string): string {
	if (kind === 'nginx') {
		return 'Nginx'
	}
	if (kind === 'php') {
		return 'PHP'
	}
	if (kind === 'nodejs') {
		return 'Node.js'
	}
	if (kind === 'mysql') {
		return 'MySQL'
	}
	if (kind === 'redis') {
		return 'Redis'
	}
	return kind
}
</script>

<style scoped lang="scss">
.panel {
	padding: var(--h-padding);
	display: flex;
	flex-direction: column;
	gap: var(--h-padding);
}

.header {
	display: flex;
	justify-content: space-between;
	align-items: center;

	.title {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		font-size: var(--font-lg);
		color: var(--text-primary);
	}

	.icon {
		color: var(--info);
	}
}

.list {
	display: flex;
	flex-direction: column;
	gap: 0.5rem;
}

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
		min-width: 0;
	}

	.title-line {
		display: flex;
		align-items: center;
		gap: 0.5rem;
	}

	.name {
		font-size: var(--font-base);
		color: var(--text-primary);
	}

	.meta {
		font-size: var(--font-xs);
		color: var(--text-muted);
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
		max-width: 22rem;
	}
}

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
