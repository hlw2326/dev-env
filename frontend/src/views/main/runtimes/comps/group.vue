<template>
	<!-- 单个运行环境类型折叠分组卡片 -->
	<div class="rt-group">
		<!-- 分组头部栏 -->
		<div
			class="group-header"
			:class="{ expanded }"
			@click="emit('toggle', type)"
		>
			<div class="header-left">
				<div class="group-icon" :class="`brand-${type}`">
					<i :class="getTypeIcon(type)"></i>
				</div>
				<span class="group-name">{{ getTypeName(type) }}</span>
				<span class="count-pill font-mono">{{ items.length }} 个版本</span>
			</div>

			<div class="header-right">
				<span v-if="expanded" class="hint-text">收起</span>
				<span v-else class="hint-text">展开</span>
				<i v-if="expanded" class="pi pi-chevron-up arrow-icon"></i>
				<i v-else class="pi pi-chevron-down arrow-icon"></i>
			</div>
		</div>

		<!-- 分组展开后的版本列表 -->
		<div v-if="expanded" class="group-list">
			<rt-card
				v-for="rt in items"
				:key="rt.id"
				:type="type"
				:item="rt"
				:busy="busyId === rt.id"
				@ext="emit('ext', $event)"
				@default="emit('default', $event)"
				@health="emit('health', $event)"
				@uninstall="emit('uninstall', $event)"
			/>
		</div>
	</div>
</template>

<script setup lang="ts">
	import type { RuntimeVersion } from '@/types'
	import RtCard from './card.vue'
	import { getTypeName, getTypeIcon } from '../use-runtimes'

	// 接收分组类型、版本列表与展开状态
	defineProps<{
		type: string
		items: RuntimeVersion[]
		expanded: boolean
		busyId: number | null
	}>()

	// 定义分组与版本项操作事件
	const emit = defineEmits<{
		toggle: [type: string]
		ext: [item: RuntimeVersion]
		default: [item: RuntimeVersion]
		health: [item: RuntimeVersion]
		uninstall: [item: RuntimeVersion]
	}>()
</script>

<style scoped lang="scss">
	/* 一体化分组卡片容器 */
	.rt-group {
		background: var(--bg-card);
		border: 1px solid var(--border-color);
		border-radius: var(--h-radius);
		overflow: hidden;
	}

	/* 分组顶部标题栏 */
	.group-header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: 12px 16px;
		background: var(--bg-app);
		cursor: pointer;
		user-select: none;
		transition: all 0.15s ease;

		&.expanded {
			border-bottom: 1px solid var(--border-color);
		}

		&:hover {
			background: var(--bg-hover);
		}
	}

	/* 左侧图标与名称 */
	.header-left {
		display: flex;
		align-items: center;
		gap: 10px;
	}

	/* 分组类型品牌图标 */
	.group-icon {
		width: 26px;
		height: 26px;
		border-radius: var(--radius-sm);
		display: flex;
		align-items: center;
		justify-content: center;
		color: var(--on-primary);
		font-size: var(--font-sm);
		background: var(--text-secondary);

		&.brand-php {
			background: var(--indigo);
		}

		&.brand-nginx,
		&.brand-nodejs {
			background: var(--ok);
		}

		&.brand-mysql,
		&.brand-postgresql {
			background: var(--info);
		}

		&.brand-redis,
		&.brand-apache {
			background: var(--err);
		}
	}

	/* 分组名称 */
	.group-name {
		font-size: var(--font-base);
		font-weight: 600;
		color: var(--text-primary);
	}

	/* 版本数量胶囊 */
	.count-pill {
		font-size: var(--font-xs);
		padding: 2px 8px;
		border-radius: var(--radius-sm);
		background: var(--bg-card);
		border: 1px solid var(--border-color);
		color: var(--text-secondary);
	}

	/* 右侧折叠指示区 */
	.header-right {
		display: flex;
		align-items: center;
		gap: 6px;
		color: var(--text-muted);
	}

	/* 展开收起文字 */
	.hint-text {
		font-size: var(--font-xs);
	}

	/* 折叠箭头图标 */
	.arrow-icon {
		font-size: var(--font-xs);
	}

	/* 版本条目列表容器 */
	.group-list {
		display: flex;
		flex-direction: column;
	}
</style>
