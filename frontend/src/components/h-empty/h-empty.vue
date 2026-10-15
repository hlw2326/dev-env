<template>
	<!-- 通用空状态容器 -->
	<div class="h-empty">
		<!-- 空状态图标 -->
		<i v-if="icon" :class="icon" class="icon"></i>
		<!-- 空状态主标题 -->
		<span v-if="title" class="title">{{ title }}</span>
		<!-- 空状态说明文字 -->
		<p v-if="description" class="subtitle">{{ description }}</p>
		<!-- 空状态操作插槽 -->
		<div v-if="$slots.default" class="actions">
			<slot></slot>
		</div>
	</div>
</template>

<script setup lang="ts">
// 全局空状态组件名称
defineOptions({
	name: 'HEmpty',
})

// 接收空状态属性
defineProps<{
	icon?: string
	title?: string
	description?: string
}>()
</script>

<style scoped lang="scss">
/* 空状态容器 */
.h-empty {
	display: flex;
	flex-direction: column;
	align-items: center;
	justify-content: center;
	gap: 0.5rem;
	padding: 2.5rem 1.25rem;
	text-align: center;
	color: var(--text-muted);

	/* 顶部大图标 */
	.icon {
		font-size: var(--icon-2xl);
		color: var(--text-muted);
		opacity: 0.45;
		margin-bottom: 0.25rem;
	}

	/* 空状态标题 */
	.title {
		font-size: var(--font-lg);
		color: var(--text-primary);
	}

	/* 空状态说明文字 */
	.subtitle {
		font-size: var(--font-sm);
		color: var(--text-muted);
	}

	/* 底部操作按钮区 */
	.actions {
		margin-top: 0.5rem;
		display: flex;
		align-items: center;
		gap: 0.5rem;
	}
}
</style>
