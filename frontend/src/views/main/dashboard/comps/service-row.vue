<template>
	<!-- 核心软件与服务行 -->
	<div class="service-row">
		<!-- 左侧图标与信息 -->
		<div class="row-left">
			<h-logo :kind="item.kind" size="small" />
			<div class="info">
				<div class="title-line">
					<span class="name">{{ item.name }} {{ item.version }}</span>
					<p-tag
						v-if="item.main"
						severity="info"
						value="默认"
						size="small"
					/>
					<p-tag
						v-if="item.daemon && item.running"
						severity="success"
						value="运行中"
						size="small"
					/>
					<p-tag
						v-else-if="item.daemon"
						severity="secondary"
						value="已停止"
						size="small"
					/>
					<p-tag
						v-else
						severity="success"
						value="就绪"
						size="small"
					/>
				</div>
				<span class="meta font-mono">
					<template v-if="item.daemon">
						<span>端口: {{ item.port }}</span>
						<span v-if="item.running"> · PID: {{ item.pid }}</span>
						<span v-else> · 空闲</span>
					</template>
					<span v-else>{{ item.path }}</span>
				</span>
			</div>
		</div>

		<!-- 右侧操作按钮 -->
		<div class="actions">
			<template v-if="item.daemon">
				<p-button
					v-if="!item.running"
					label="启动"
					icon="pi pi-play"
					size="small"
					severity="success"
					:loading="busyId === item.id"
					@click="startService(item.id)"
				/>
				<p-button
					v-else
					label="停止"
					icon="pi pi-stop"
					size="small"
					severity="danger"
					variant="outlined"
					:loading="busyId === item.id"
					@click="stopService(item.id)"
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
</template>

<script setup lang="ts">
import { useDashboard } from '../use-dashboard'
import type { Software } from '@/types'

// 接收软件条目属性
defineProps<{
	item: Software
}>()

// 守护进程控制操作
const { busyId, startService, stopService } = useDashboard()
</script>

<style scoped lang="scss">
/* 服务行容器 */
.service-row {
	display: flex;
	justify-content: space-between;
	align-items: center;
	padding: 0.75rem 0.875rem;
	background: var(--bg-app);
	border: 1px solid var(--border-color);
	border-radius: var(--h-radius);
	transition: all 0.15s ease;

	/* 悬停态 */
	&:hover {
		background: var(--bg-card);
		border-color: var(--border-hover);
	}

	/* 左侧信息组合 */
	.row-left {
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
	.title-line {
		display: flex;
		align-items: center;
		gap: 0.5rem;
	}

	/* 软件名称 */
	.name {
		font-size: var(--font-base);
		color: var(--text-primary);
	}

	/* 元数据文本 */
	.meta {
		font-size: var(--font-xs);
		color: var(--text-muted);
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
		max-width: 22rem;
	}
}
</style>
