<template>
	<!-- 软件服务状态行 -->
	<div class="h-service">
		<!-- 左侧图标与信息 -->
		<div class="left">
			<h-logo :kind="item.kind" size="small" />
			<div class="info">
				<div class="title">
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

		<!-- 右侧服务操作 -->
		<div class="actions">
			<template v-if="item.daemon">
				<p-button
					v-if="!item.running"
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
					variant="outlined"
					:loading="busy"
					@click="emit('stop', item.id)"
				/>
			</template>
			<p-button
				v-else
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
import type { HServiceProps } from './types.d'

// 全局服务行组件名称
defineOptions({
	name: 'HService',
})

// 接收服务数据与操作状态
defineProps<HServiceProps>()

// 定义服务控制事件
const emit = defineEmits<{
	start: [id: number]
	stop: [id: number]
	manage: []
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

/* 右侧操作区 */
.actions {
	display: flex;
	align-items: center;
	gap: 0.5rem;
	flex-shrink: 0;
}
</style>
