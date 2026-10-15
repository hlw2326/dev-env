<template>
	<h-card title="主机资源状态" class="dashboard-resources">
		<div class="resource-grid">
			<div class="resource-item">
				<div class="resource-head">
					<span>CPU 使用率</span>
					<strong>{{ systemMetrics.cpuPercent.toFixed(1) }}%</strong>
				</div>
				<p-progress-bar :value="systemMetrics.cpuPercent" :show-value="false" />
			</div>
			<div class="resource-item">
				<div class="resource-head">
					<span>内存使用率</span>
					<strong>{{ systemMetrics.memoryPercent.toFixed(1) }}%</strong>
				</div>
				<p-progress-bar :value="systemMetrics.memoryPercent" :show-value="false" />
				<span class="resource-detail">
					{{ formatBytes(systemMetrics.memoryUsedBytes) }} / {{ formatBytes(systemMetrics.memoryTotalBytes) }}
				</span>
			</div>
			<div class="resource-item">
				<div class="resource-head">
					<span>网站访问流量</span>
					<strong>{{ totalTraffic.requestCount }} 次请求</strong>
				</div>
				<span class="resource-detail">
					{{ formatBytes(totalTraffic.responseBytes) }} · 错误 {{ totalTraffic.errorCount }} · 平均 {{ totalTraffic.averageResponseMs.toFixed(0) }}ms
				</span>
			</div>
		</div>
	</h-card>
</template>

<script setup lang="ts">
import { useDashboard } from '@/composables'
import { formatBytes } from '@/utils'

const { systemMetrics, totalTraffic } = useDashboard()
</script>

<style scoped lang="scss">
.resource-grid {
	display: grid;
	grid-template-columns: repeat(3, minmax(0, 1fr));
	gap: 1.25rem;
}

.resource-item {
	display: flex;
	flex-direction: column;
	gap: 0.5rem;
}

.resource-head {
	display: flex;
	justify-content: space-between;
	align-items: center;
	font-size: var(--font-sm);
	color: var(--text-secondary);
}

.resource-head strong {
	font-family: var(--font-mono);
	font-size: var(--font-base);
	font-weight: 500;
	color: var(--text-primary);
}

.resource-detail {
	font-size: var(--font-xs);
	color: var(--text-muted);
}

/* 仪表盘资源卡片不显示头部分隔线，避免与下方服务卡片产生多余横线。 */
.dashboard-resources :deep(.header) {
	border-bottom: none;
}

@media (max-width: 900px) {
	.resource-grid {
		grid-template-columns: 1fr;
	}
}
</style>
