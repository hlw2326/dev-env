<template>
	<!-- 软件商店顶部筛选与搜索工具栏 -->
	<div class="top-bar surface-card">
		<!-- 全部软件与已安装页签 -->
		<div class="main-tabs">
			<button
				type="button"
				class="tab-btn"
				:class="{ active: tab === 'market' }"
				@click="tab = 'market'"
			>
				全部软件
			</button>
			<button
				type="button"
				class="tab-btn"
				:class="{ active: tab === 'installed' }"
				@click="tab = 'installed'"
			>
				已安装 ({{ installedCount }})
			</button>
		</div>

		<!-- 右侧分类筛选与搜索框 -->
		<div class="right-filters">
			<div v-if="tab === 'market'" class="category-tabs">
				<button
					v-for="item in categories"
					:key="item"
					type="button"
					class="tab-btn"
					:class="{ active: category === item }"
					@click="category = item"
				>
					{{ item }}
				</button>
			</div>

			<div class="search-box">
				<i class="pi pi-search search-icon"></i>
				<p-input-text
					v-model="search"
					placeholder="搜索软件..."
					class="search-input"
				/>
			</div>
		</div>
	</div>
</template>

<script setup lang="ts">
import { useSoftware } from '@/composables'

// 软件商店筛选状态
const { tab, category, search, installedCount } = useSoftware()

// 软件分类筛选列表
const categories = ['全部', '运行环境', 'Web 服务', '数据库', '缓存', '工具']
</script>

<style scoped lang="scss">
/* 顶部工具栏外框 */
.top-bar {
	display: flex;
	align-items: center;
	justify-content: space-between;
	padding: 0.625rem 0.875rem;
	gap: 1rem;
	flex-wrap: wrap;
}

/* 页签与分类按钮组 */
.main-tabs,
.category-tabs {
	display: flex;
	align-items: center;
	gap: 0.375rem;
}

.category-tabs {
	flex-wrap: wrap;
}

/* 右侧筛选组合 */
.right-filters {
	display: flex;
	align-items: center;
	gap: 0.875rem;
}

/* 切换按钮通用样式 */
.tab-btn {
	border: 1px solid transparent;
	background: transparent;
	color: var(--text-secondary);
	font-size: var(--font-sm);
	padding: 0.35rem 0.75rem;
	border-radius: var(--h-radius);
	cursor: pointer;
	transition: all 0.15s ease;

	/* 悬停态 */
	&:hover {
		background: var(--bg-app);
		color: var(--text-primary);
	}

	/* 选中态 */
	&.active {
		background: var(--p-primary-color);
		color: var(--on-primary);
	}
}

/* 搜索框容器 */
.search-box {
	position: relative;
	display: flex;
	align-items: center;
	width: 14rem;

	/* 搜索图标 */
	.search-icon {
		position: absolute;
		left: 0.75rem;
		font-size: var(--icon-sm);
		color: var(--text-muted);
	}

	/* 搜索输入框 */
	.search-input {
		width: 100%;
		padding-left: 2.15rem !important;
	}
}
</style>
