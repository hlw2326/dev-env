<template>
	<div class="h-page" :class="{ 'no-scroll': scroll === false }">
		<!-- 页面头部 -->
		<slot name="header">
			<div v-if="title || $slots.title || $slots.extra" class="header">
				<div class="header-left">
					<slot name="title">
						<span class="title">{{ title }}</span>
					</slot>
					<slot name="desc">
						<p v-if="desc" class="desc">{{ desc }}</p>
					</slot>
				</div>
				<div v-if="$slots.extra" class="header-extra">
					<slot name="extra"></slot>
				</div>
			</div>
		</slot>

		<!-- 页面主体 -->
		<div class="body" :class="{ 'with-padding': padding !== false }">
			<slot></slot>
		</div>
	</div>
</template>

<script setup lang="ts">
// 全局页面组件名称
defineOptions({
	name: 'HPage',
})

// 组件接收属性与默认值
withDefaults(
	defineProps<{
		title?: string
		desc?: string
		padding?: boolean
		scroll?: boolean
	}>(),
	{
		padding: true,
		scroll: true,
	}
)
</script>

<style scoped lang="scss">
/* 页面外层容器 */
.h-page {
	display: flex;
	flex-direction: column;
	min-height: 100%;
	width: 100%;

	&.no-scroll {
		height: 100%;
		overflow: hidden;
	}
}

/* 页面顶部栏 */
.header {
	display: flex;
	justify-content: space-between;
	align-items: center;
	padding: 0.875rem var(--h-padding);
	background: var(--bg-card);
	border-bottom: 1px solid var(--border-color);
	flex-shrink: 0;

	.header-left {
		display: flex;
		flex-direction: column;
		gap: 0.2rem;

		.title {
			font-size: 1rem;
			color: var(--text-primary);
		}

		.desc {
			font-size: var(--font-xs);
			color: var(--text-secondary);
		}
	}

	.header-extra {
		display: flex;
		align-items: center;
		gap: 0.625rem;
	}
}

/* 页面内容区 */
.body {
	flex: 1;
	display: flex;
	flex-direction: column;

	&.with-padding {
		padding: var(--h-padding);
		gap: var(--h-padding);
	}
}
</style>
