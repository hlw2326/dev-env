<template>
	<!-- 软件服务状态行 -->
	<div class="h-service">
		<!-- 左侧图标与信息 -->
		<div class="left">
			<h-logo :kind="item.kind" size="small" />
			<div class="info">
				<div class="title">
					<span class="name">{{ item.name }} {{ item.version }}</span>
					<h-tag
						v-if="item.main"
						color="info"
						value="默认"
					/>
					<h-tag
						v-if="item.service && checking && !busy"
						color="info"
						value="检测中"
					/>
					<h-tag
						v-else-if="item.service && unavailable"
						color="warning"
						value="待检测"
					/>
					<h-tag
						v-else-if="item.service && item.running"
						color="success"
						value="运行中"
						dot
						radius="full"
					/>
					<h-tag
						v-else-if="item.service"
						color="default"
						value="已停止"
						dot
						radius="full"
					/>
					<h-tag
						v-else
						color="success"
						value="就绪"
					/>
				</div>
				<span class="meta font-mono">
					<template v-if="item.service">
						<span>端口: {{ item.port }}</span>
						<span v-if="item.running && item.pid > 0"> · PID: {{ item.pid }}</span>
						<span v-else> · 空闲</span>
					</template>
					<span v-else>{{ item.path }}</span>
				</span>
				<span v-if="item.service" class="resource-meta font-mono">
					<span>CPU {{ item.metrics.cpuPercent.toFixed(1) }}%</span>
					<span>· 内存 {{ formatBytes(item.metrics.memoryBytes) }}</span>
					<span>· I/O {{ formatBytes(item.metrics.readBytes) }}/{{ formatBytes(item.metrics.writeBytes) }}</span>
					<span>· 运行 {{ formatDuration(item.metrics.uptimeSeconds) }}</span>
				</span>
			</div>
		</div>

		<!-- 右侧服务操作 -->
		<div class="actions">
			<template v-if="item.service">
				<p-button v-if="checking && !busy" label="检测中" size="small" severity="secondary" :loading="true" disabled />
				<p-button v-else-if="unavailable" label="待检测" size="small" severity="secondary" disabled />
				<template v-else>
				<p-button
					v-if="item.kind === 'redis'"
					label="数据"
					icon="pi pi-database"
					size="small"
					severity="secondary"
					variant="outlined"
					:disabled="!item.running"
					@click="emit('data', item)"
				/>
				<p-button
					v-if="configurable"
					label="设置"
					icon="pi pi-cog"
					size="small"
					severity="secondary"
					variant="outlined"
					@click="emit('configure', item)"
				/>
				<p-button
					v-if="(item.running || restarting) && showRestart"
					label="重启"
					icon="pi pi-refresh"
					size="small"
					severity="secondary"
					variant="outlined"
					:loading="restarting"
					:disabled="busy"
					@click="emit('restart', item.id)"
				/>
				<p-button
					v-if="!item.running && !restarting"
					label="启动"
					icon="pi pi-play"
					size="small"
					severity="success"
					:loading="busy"
					@click="emit('start', item.id)"
				/>
				<p-button
					v-else
					label="停止"
					icon="pi pi-stop"
					size="small"
					severity="danger"
					:loading="busy && !restarting"
					:disabled="busy"
					@click="emit('stop', item.id)"
				/>
				</template>
			</template>
			<p-button
				v-else-if="!item.service"
				label="管理"
				icon="pi pi-cog"
				size="small"
				severity="secondary"
				variant="outlined"
				@click="emit('manage')"
			/>
		</div>
	</div>
</template>

<script setup lang="ts">
import { formatBytes, formatDuration } from '@/utils'
import type { HServiceProps } from './types.d'

// 全局服务行组件名称
defineOptions({
	name: 'HService',
})

// 接收服务数据与操作状态
withDefaults(defineProps<HServiceProps>(), {
	configurable: true,
	showRestart: false,
	restarting: false,
	checking: false,
	unavailable: false,
})

// 定义服务控制事件
const emit = defineEmits<{
	start: [id: number]
	stop: [id: number]
	restart: [id: number]
	manage: []
	configure: [item: HServiceProps['item']]
	data: [item: HServiceProps['item']]
}>()
</script>

<style scoped lang="scss">
/* 服务行容器 */
.h-service {
	display: flex;
	justify-content: space-between;
	align-items: center;
	padding: 0.75rem 0.875rem;
	background: var(--bg-app);
	border: 1px solid var(--border-color);
	border-radius: var(--h-radius);
	transition: all 0.15s ease;

	/* 悬停状态 */
	&:hover {
		background: var(--bg-card);
		border-color: var(--border-hover);
	}
}

/* 左侧信息组合 */
.left {
	display: flex;
	align-items: center;
	gap: 0.75rem;
	min-width: 0;
}

/* 文字垂直区 */
.info {
	display: flex;
	flex-direction: column;
	gap: 0.2rem;
	min-width: 0;
}

/* 名称与状态标签行 */
.title {
	display: flex;
	align-items: center;
	gap: 0.5rem;
}

/* 软件服务名称 */
.name {
	font-size: var(--font-base);
	color: var(--text-primary);
}

/* 服务运行元数据 */
.meta {
	font-size: var(--font-xs);
	color: var(--text-muted);
	overflow: hidden;
	text-overflow: ellipsis;
	white-space: nowrap;
	max-width: 22rem;
}

.resource-meta {
	display: flex;
	align-items: center;
	gap: 0.5rem;
	font-size: var(--font-xs);
	color: var(--text-muted);
}

/* 右侧操作区 */
.actions {
	display: flex;
	align-items: center;
	gap: 0.5rem;
	flex-shrink: 0;
}
</style>
