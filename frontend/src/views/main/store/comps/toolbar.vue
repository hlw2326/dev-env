<template>
	<!-- 软件商店分类筛选与搜索同步工具栏 -->
	<div class="store-toolbar">
		<!-- 左侧分类切换标签 -->
		<div class="cat-tabs">
			<button
				v-for="cat in categories"
				:key="cat.key"
				type="button"
				class="cat-btn"
				:class="{ active: category === cat.key }"
				@click="category = cat.key"
			>
				{{ cat.label }}
			</button>
		</div>

		<!-- 右侧版本同步与关键字搜索 -->
		<div class="extra-row">
			<div class="search-box">
				<i class="pi pi-search search-icon"></i>
				<p-input-text
					v-model="keyword"
					placeholder="搜索软件包名称或描述..."
					class="search-input"
				/>
			</div>
			<p-button
				icon="pi pi-refresh"
				label="检查与同步版本"
				size="small"
				severity="secondary"
				outlined
				:loading="syncing"
				@click="emit('sync')"
			/>
		</div>
	</div>
</template>

<script setup lang="ts">
	import { categories } from '../use-store'

	// 接收同步加载状态
	defineProps<{
		syncing: boolean
	}>()

	// 定义同步事件
	const emit = defineEmits<{
		sync: []
	}>()

	// 双向绑定当前选中分类与搜索关键字
	const category = defineModel<string>('category', { required: true })
	const keyword = defineModel<string>('keyword', { required: true })
</script>

<style scoped lang="scss">
	/* 工具栏外层容器（消除左右边距，使底部分割线通栏显示） */
	.store-toolbar {
		display: flex;
		justify-content: space-between;
		align-items: center;
		gap: 12px;
		flex-wrap: wrap;
		margin: calc(-1 * var(--h-padding)) calc(-1 * var(--h-padding)) 0;
		padding: 10px var(--h-padding);
		background: var(--bg-card);
		border-bottom: 1px solid var(--border-color);
	}

	/* 分类切换标签容器 */
	.cat-tabs {
		display: flex;
		align-items: center;
		gap: 8px;
		flex-wrap: wrap;
	}

	/* 单个分类按钮 */
	.cat-btn {
		background: transparent;
		border: none;
		padding: 6px 14px;
		font-size: var(--font-sm);
		color: var(--text-secondary);
		border-radius: var(--h-radius);
		cursor: pointer;

		&:hover {
			background: var(--bg-hover);
			color: var(--text-primary);
		}

		&.active {
			background: var(--p-primary-color);
			color: var(--on-primary);
		}
	}

	/* 右侧搜索与同步按钮行 */
	.extra-row {
		display: flex;
		align-items: center;
		gap: 10px;
	}

	/* 搜索输入框外层 */
	.search-box {
		position: relative;
		display: flex;
		align-items: center;
		width: 240px;
	}

	/* 搜索放大镜图标 */
	.search-icon {
		position: absolute;
		left: 10px;
		color: var(--text-muted);
		font-size: var(--font-sm);
		pointer-events: none;
	}

	/* 搜索输入框 */
	.search-input {
		width: 100%;
		padding-left: 32px;
		height: 32px;
		font-size: var(--font-sm);
		border-radius: var(--h-radius);
	}
</style>
