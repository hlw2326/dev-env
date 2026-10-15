<template>
	<!-- 通用卡片容器 -->
	<div class="h-card" :class="{ fill }">
		<!-- 卡片头部栏 -->
		<div v-if="title || $slots.title || $slots.extra || $slots.actions" class="header">
			<div class="header-left">
				<slot name="title">
					<span class="title">{{ title }}</span>
				</slot>
			</div>
			<div v-if="$slots.extra || $slots.actions" class="header-extra">
				<slot name="extra"></slot>
				<slot name="actions"></slot>
			</div>
		</div>
		<!-- 卡片主体内容 -->
		<div class="body" :class="{ 'no-padding': noPadding }">
			<slot></slot>
		</div>
	</div>
</template>

<script setup lang="ts">
import type { HCardProps } from './types.d'

// 全局卡片组件名称
defineOptions({
	name: 'HCard',
})

// 接收卡片属性
defineProps<HCardProps>()
</script>

<style scoped lang="scss">
/* 卡片容器外框 */
.h-card {
	background-color: var(--bg-card);
	border: 1px solid var(--border-color);
	border-radius: var(--h-radius);
	overflow: hidden;
	display: flex;
	flex-direction: column;

	/* 高度自适应修饰 */
	&.fill {
		flex: 1;
		min-height: 0;
	}

	/* 卡片头部区域 */
	.header {
		height: 3.25rem;
		padding: 0 1.25rem;
		border-bottom: 1px dashed var(--border-color);
		display: flex;
		align-items: center;
		justify-content: space-between;

		/* 卡片主标题 */
		.title {
			font-size: var(--font-lg);
			color: var(--text-primary);
		}
	}

	/* 卡片内容区域 */
	.body {
		flex: 1;
		min-height: 0;
		display: flex;
		flex-direction: column;
		padding: var(--h-padding);

		/* 零内边距修饰 */
		&.no-padding {
			padding: 0;
		}
	}
}
</style>
