<template>
	<!-- 全局语义标签，内部复用 PrimeVue Tag -->
	<p-tag
		class="h-tag"
		:class="tagClasses"
		:severity="effectiveSeverity"
		:icon="icon"
		:rounded="radius === 'full'"
		:value="tagValue"
	>
		<template v-if="hasContent" #default>
			<span v-if="dot" class="dot" aria-hidden="true"></span>
			<slot>{{ value }}</slot>
		</template>
	</p-tag>
</template>

<script setup lang="ts">
import { computed, useSlots } from 'vue'
import type { HTagProps } from './types.d'

// 全局标签组件名称
defineOptions({
	name: 'HTag',
})

const props = withDefaults(defineProps<HTagProps>(), {
	severity: 'secondary',
	size: 'small',
	radius: 'small',
	dot: false,
})
const slots = useSlots()

// 将项目颜色映射到 PrimeVue severity
const effectiveSeverity = computed(() => {
	if (props.color === undefined) {
		return props.severity
	}
	if (props.color === 'warning') {
		return 'warn'
	}
	if (props.color === 'default') {
		return 'secondary'
	}
	return props.color
})

// 当标签带状态圆点或自定义插槽时，交给插槽渲染内容
const hasContent = computed(() => {
	if (props.dot) {
		return true
	}
	return Boolean(slots.default)
})

const tagValue = computed(() => {
	if (hasContent.value) {
		return undefined
	}
	return props.value
})

// 标签尺寸和圆角样式
const tagClasses = computed(() => [
	`tag-size-${props.size}`,
	`tag-radius-${props.radius}`,
])
</script>

<style scoped lang="scss">
/* 标签根节点 */
.h-tag {
	display: inline-flex !important;
	align-items: center;
	gap: 0.35rem;
	font-family: var(--font-sans) !important;
	font-weight: 400 !important;
	line-height: var(--leading-snug) !important;
	white-space: nowrap;
}

/* 标签尺寸 */
.tag-size-small {
	padding: 0.125rem 0.5rem !important;
	font-size: var(--font-xs) !important;
}

.tag-size-medium {
	padding: 0.2rem 0.625rem !important;
	font-size: var(--font-sm) !important;
}

.tag-size-large {
	padding: 0.3rem 0.75rem !important;
	font-size: var(--font-base) !important;
}

/* 标签圆角 */
.tag-radius-small {
	border-radius: var(--radius-sm) !important;
}

.tag-radius-large {
	border-radius: var(--radius-lg) !important;
}

.tag-radius-full {
	border-radius: var(--radius-full) !important;
}

/* 状态圆点 */
.dot {
	width: 0.375rem;
	height: 0.375rem;
	flex: 0 0 auto;
	border-radius: 50%;
	background: currentColor;
}
</style>
