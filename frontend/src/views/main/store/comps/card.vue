<template>
	<!-- 单个软件包展示卡片 -->
	<div class="pkg-card" @click="emit('select', item)">
		<!-- 顶部图标与软件名称 -->
		<div class="pkg-top">
			<div class="pkg-icon" :style="{ backgroundColor: item.brandColor }">
				<i :class="item.iconClass"></i>
			</div>
			<div class="pkg-info">
				<div class="pkg-name-row">
					<span class="pkg-name">{{ item.displayName }}</span>
					<p-tag
						v-if="item.isInstalled"
						severity="success"
						value="已安装"
						icon="pi pi-check"
						class="installed-tag"
					/>
				</div>
				<span class="pkg-ver font-mono">最新/推荐: v{{ item.recommendedVersion }}</span>
			</div>
		</div>

		<!-- 软件包简介 -->
		<p class="pkg-desc">{{ item.description }}</p>

		<!-- 底部元信息与操作按钮 -->
		<div class="pkg-footer">
			<div class="pkg-meta">
				<span class="cat-badge">{{ catLabel }}</span>
				<span class="sync-dot" title="支持在线动态拉取最新版本">
					<span class="dot-inner"></span> 在线同步
				</span>
			</div>
			<div class="pkg-actions" @click.stop>
				<p-button
					v-if="item.isInstalled"
					label="版本管理"
					icon="pi pi-list"
					size="small"
					severity="secondary"
					outlined
					@click="emit('select', item)"
				/>
				<p-button
					v-else
					label="获取版本"
					icon="pi pi-download"
					size="small"
					@click="emit('select', item)"
				/>
			</div>
		</div>
	</div>
</template>

<script setup lang="ts">
	import { computed } from 'vue'
	import type { PackageInfo } from '@/types'
	import { categories } from '../use-store'

	// 接收软件包信息
	const props = defineProps<{
		item: PackageInfo
	}>()

	// 定义点击选择事件
	const emit = defineEmits<{
		select: [item: PackageInfo]
	}>()

	// 分类中文标签
	const catLabel = computed(() => {
		const hit = categories.find((c) => c.key === props.item.category)
		if (hit) {
			return hit.label
		}
		return props.item.category
	})
</script>

<style scoped lang="scss">
	/* 软件包卡片容器 */
	.pkg-card {
		background: var(--bg-card);
		border: 1px solid var(--border-color);
		border-radius: var(--h-radius);
		padding: var(--h-padding);
		display: flex;
		flex-direction: column;
		gap: 12px;
		cursor: pointer;

		&:hover {
			background: var(--bg-hover);
			border-color: var(--border-hover);
		}
	}

	/* 顶部信息行 */
	.pkg-top {
		display: flex;
		align-items: center;
		gap: 12px;
	}

	/* 品牌色图标容器 */
	.pkg-icon {
		width: 44px;
		height: 44px;
		border-radius: var(--radius-sm);
		display: flex;
		align-items: center;
		justify-content: center;
		color: var(--on-primary);
		font-size: 1.3rem;
		flex-shrink: 0;
	}

	/* 标题与推荐版本列 */
	.pkg-info {
		display: flex;
		flex-direction: column;
		gap: 2px;
		flex: 1;
	}

	/* 名称与已安装标签行 */
	.pkg-name-row {
		display: flex;
		align-items: center;
		gap: 8px;
	}

	/* 软件名称 */
	.pkg-name {
		font-size: var(--font-lg);
		color: var(--text-primary);
	}

	/* 已安装小标签 */
	.installed-tag {
		font-size: var(--font-xs);
		padding: 1px 6px;
	}

	/* 推荐版本号 */
	.pkg-ver {
		font-size: var(--font-xs);
		color: var(--text-muted);
	}

	/* 软件功能简介 */
	.pkg-desc {
		font-size: var(--font-sm);
		color: var(--text-secondary);
		line-height: 1.5;
		margin: 0;
		min-height: 38px;
	}

	/* 卡片底部栏 */
	.pkg-footer {
		display: flex;
		align-items: center;
		justify-content: space-between;
		margin-top: auto;
		padding-top: 8px;
		border-top: 1px solid var(--border-color);
	}

	/* 左侧分类与在线状态 */
	.pkg-meta {
		display: flex;
		align-items: center;
		gap: 8px;
	}

	/* 分类徽章 */
	.cat-badge {
		font-size: var(--font-xs);
		color: var(--text-muted);
		background: var(--bg-app);
		padding: 2px 8px;
		border-radius: var(--radius-sm);
		border: 1px solid var(--border-color);
	}

	/* 在线同步指示器 */
	.sync-dot {
		display: flex;
		align-items: center;
		gap: 4px;
		font-size: var(--font-xs);
		color: var(--ok);
	}

	/* 绿点圆心 */
	.dot-inner {
		width: 6px;
		height: 6px;
		border-radius: 50%;
		background: var(--ok);
	}

	/* 操作按钮容器 */
	.pkg-actions {
		display: flex;
		gap: 6px;
	}
</style>
