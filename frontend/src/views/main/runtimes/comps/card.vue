<template>
	<!-- 单个运行时版本列表行 -->
	<div class="rt-row">
		<!-- 左侧：版本号、标签与安装路径元信息 -->
		<div class="rt-main">
			<div class="rt-title-line">
				<span class="rt-title font-mono">{{ getTypeName(type) }} {{ item.version }}</span>
				<span v-if="type === 'php' && item.variant" class="rt-badge font-mono">
					{{ item.variant.toUpperCase() }}
				</span>
				<span class="rt-badge font-mono">{{ item.architecture }}</span>
				<p-tag
					:severity="statusTone"
					:value="statusText"
					size="small"
				/>
				<p-tag
					v-if="item.isDefault"
					severity="info"
					value="默认 CLI"
					icon="pi pi-star-fill"
					size="small"
				/>
				<p-tag
					v-if="(item.serviceCount ?? 0) > 0"
					severity="warn"
					value="服务使用中"
					icon="pi pi-cog"
					size="small"
					class="service-active-tag"
					title="当前服务实例正在使用此版本，点击前往服务实例"
					@click="router.push('/services')"
				/>
			</div>

			<div class="rt-meta-line">
				<span class="meta-item font-mono" :title="item.installPath">
					<i class="pi pi-folder"></i>
					<span>{{ item.installPath }}</span>
				</span>
				<span class="meta-sep">·</span>
				<span class="meta-item">
					<i class="pi pi-link"></i>
					<span>被 {{ item.siteCount ?? 0 }} 个网站使用</span>
				</span>
			</div>
		</div>

		<!-- 右侧：快捷操作按钮组 -->
		<div class="rt-actions">
			<template v-if="type === 'php'">
				<p-button
					label="配置扩展"
					icon="pi pi-sliders-h"
					size="small"
					severity="secondary"
					outlined
					@click="emit('ext', item)"
				/>
				<p-button
					v-if="!item.isDefault"
					label="设为默认"
					icon="pi pi-star"
					size="small"
					severity="secondary"
					outlined
					:loading="busy"
					@click="emit('default', item)"
				/>
			</template>

			<template v-else-if="type === 'nginx'">
				<p-button
					label="测试配置"
					icon="pi pi-check-circle"
					size="small"
					severity="secondary"
					outlined
					@click="emit('health', item)"
				/>
			</template>

			<template v-else-if="type === 'redis'">
				<p-button
					label="测试连接"
					icon="pi pi-wifi"
					size="small"
					severity="secondary"
					outlined
					@click="emit('health', item)"
				/>
			</template>

			<p-button
				v-else
				label="健康检查"
				icon="pi pi-check-circle"
				size="small"
				severity="secondary"
				outlined
				@click="emit('health', item)"
			/>

			<p-button
				icon="pi pi-trash"
				size="small"
				severity="danger"
				text
				rounded
				title="卸载此版本"
				@click="emit('uninstall', item)"
			/>
		</div>
	</div>
</template>

<script setup lang="ts">
	import { computed } from 'vue'
	import { useRouter } from 'vue-router'
	import type { RuntimeVersion } from '@/types'
	import { getTypeName } from '../use-runtimes'

	const router = useRouter()

	// 接收运行时版本数据
	const props = defineProps<{
		type: string
		item: RuntimeVersion
		busy: boolean
	}>()

	// 定义运行时操作事件
	const emit = defineEmits<{
		ext: [item: RuntimeVersion]
		default: [item: RuntimeVersion]
		health: [item: RuntimeVersion]
		uninstall: [item: RuntimeVersion]
	}>()

	// 状态色彩映射
	const toneMap: Record<string, string> = {
		available: 'success',
		installed: 'success',
		uninitialized: 'warn',
		in_use: 'info',
		error: 'danger',
		failed: 'danger'
	}

	// 状态文案映射
	const labelMap: Record<string, string> = {
		available: '可用',
		installed: '可用',
		uninitialized: '未初始化',
		in_use: '使用中',
		installing: '安装中',
		error: '不可用',
		failed: '不可用'
	}

	// 当前状态色调
	const statusTone = computed(() => {
		const hit = toneMap[props.item.status]
		if (hit) {
			return hit
		}
		return 'secondary'
	})

	// 当前状态文案
	const statusText = computed(() => {
		const hit = labelMap[props.item.status]
		if (hit) {
			return hit
		}
		return '就绪'
	})
</script>

<style scoped lang="scss">
	/* 单个版本水平行容器 */
	.rt-row {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 16px;
		padding: 14px 16px;
		background: var(--bg-card);
		border-bottom: 1px solid var(--border-color);
		transition: all 0.15s ease;

		&:last-child {
			border-bottom: none;
		}

		&:hover {
			background: var(--bg-hover);
		}
	}

	/* 左侧主信息列 */
	.rt-main {
		display: flex;
		flex-direction: column;
		gap: 6px;
		min-width: 0;
		flex: 1;
	}

	/* 版本号与状态徽章行 */
	.rt-title-line {
		display: flex;
		align-items: center;
		gap: 8px;
		flex-wrap: wrap;
	}

	/* 版本名称 */
	.rt-title {
		font-size: var(--font-base);
		font-weight: 600;
		color: var(--text-primary);
	}

	/* 架构与线程安全徽章 */
	.rt-badge {
		font-size: var(--font-xs);
		background: var(--bg-app);
		border: 1px solid var(--border-color);
		padding: 1px 6px;
		border-radius: var(--radius-sm);
		color: var(--text-secondary);
	}

	/* 服务使用中可点击标签 */
	.service-active-tag {
		cursor: pointer;
	}

	/* 底部路径与引用统计行 */
	.rt-meta-line {
		display: flex;
		align-items: center;
		gap: 8px;
		font-size: var(--font-xs);
		color: var(--text-muted);
		flex-wrap: wrap;
	}

	/* 单项元信息 */
	.meta-item {
		display: inline-flex;
		align-items: center;
		gap: 5px;
	}

	/* 分隔圆点 */
	.meta-sep {
		color: var(--border-color);
	}

	/* 右侧操作按钮组 */
	.rt-actions {
		display: flex;
		align-items: center;
		gap: 8px;
		flex-shrink: 0;
	}
</style>

