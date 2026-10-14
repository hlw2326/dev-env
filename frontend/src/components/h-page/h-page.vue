<template>
	<div class="h-page" :class="{ 'no-scroll': scroll === false }">
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

		<div class="body" :class="{ 'with-padding': padding !== false }">
			<slot></slot>
		</div>
	</div>
</template>

<script setup lang="ts">
defineOptions({
	name: 'HPage',
})

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
			font-size: var(--font-xl);
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
