<template>
	<!-- 页面通用容器 -->
	<div class="h-page">
		<!-- 页面顶部标题栏 -->
		<slot name="header">
			<div v-if="title || $slots.title || $slots.extra" class="header">
				<div class="header-left">
					<slot name="title">
						<span class="title">{{ title }}</span>
					</slot>
					<slot name="description">
						<p v-if="description" class="subtitle">{{ description }}</p>
					</slot>
				</div>
				<div v-if="$slots.extra" class="header-extra">
					<slot name="extra"></slot>
				</div>
			</div>
		</slot>

		<!-- 页面内容区 -->
		<div class="body" :class="{ 'with-padding': padding, scrollable: scroll }">
			<slot></slot>
		</div>
	</div>
</template>

<script setup lang="ts">
import type { HPageProps } from './types.d'

// 全局页面组件名称
defineOptions({
	name: 'HPage',
})

// 接收页面属性
const {
	title,
	description,
	padding = true,
	scroll = true,
} = defineProps<HPageProps>()
</script>

<style scoped lang="scss">
/* 页面根容器 */
.h-page {
	display: flex;
	flex-direction: column;
	height: 100%;
	min-height: 0;
	width: 100%;
	overflow: hidden;
}

/* 页面头部栏 */
.header {
	display: flex;
	justify-content: space-between;
	align-items: center;
	padding: 0.875rem var(--page-padding);
	background: var(--bg-card);
	border-bottom: 1px solid var(--border-color);
	flex-shrink: 0;

	/* 头部左侧文案区 */
	.header-left {
		display: flex;
		flex-direction: column;
		gap: 0.2rem;

		/* 页面主标题 */
		.title {
			font-size: var(--font-xl);
			color: var(--text-primary);
		}

		/* 页面副标题 */
		.subtitle {
			font-size: var(--font-xs);
			color: var(--text-secondary);
		}
	}

	/* 头部右侧扩展区 */
	.header-extra {
		display: flex;
		align-items: center;
		gap: 0.625rem;
	}
}

/* 页面主体区域 */
.body {
	flex: 1;
	min-height: 0;
	display: flex;
	flex-direction: column;
	overflow: hidden;

	/* 页面正文独立滚动 */
	&.scrollable {
		overflow-y: auto;
	}

	/* 标准内边距修饰 */
	&.with-padding {
		padding: var(--page-padding);
		gap: var(--h-padding);
	}
}
</style>
