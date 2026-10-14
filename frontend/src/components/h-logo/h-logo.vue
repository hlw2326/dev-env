<template>
	<div class="h-logo" :class="[`size-${size}`]" :style="style">
		<img
			v-if="hasImage"
			:src="meta.logo"
			:alt="kind"
			class="logo-img"
			@error="onImageError"
		/>
		<span v-else class="logo-text font-mono font-bold">
			{{ meta.short }}
		</span>
	</div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import type { HLogoProps, LogoMeta } from './types.d'

const metaMap: Record<string, LogoMeta> = {
	nginx: { short: 'NG', color: '#009639', logo: '/logos/nginx.svg' },
	apache: { short: 'AP', color: '#d22128', logo: '/logos/apache.svg' },
	php: { short: 'PHP', color: '#777bb4', logo: '/logos/php.svg' },
	nodejs: { short: 'Node', color: '#539e43', logo: '/logos/nodejs.svg' },
	mysql: { short: 'SQL', color: '#00758f', logo: '/logos/mysql.svg' },
	redis: { short: 'Redis', color: '#dc382d', logo: '/logos/redis.svg' },
}

const defaultMeta: LogoMeta = {
	short: 'DEV',
	color: '#64748b',
	logo: '',
}

const props = withDefaults(defineProps<HLogoProps>(), {
	size: 'medium',
})

const meta = computed(() => {
	const cur = metaMap[props.kind]
	if (cur) {
		return cur
	}
	return defaultMeta
})

const hasError = ref(false)

watch(
	() => props.kind,
	() => {
		hasError.value = false
	}
)

function onImageError() {
	hasError.value = true
}

const hasImage = computed(() => {
	if (hasError.value) {
		return false
	}
	if (meta.value.logo.length === 0) {
		return false
	}
	return true
})

const style = computed(() => {
	if (hasImage.value) {
		return {
			background: 'var(--surface-card, #ffffff)',
			border: '1px solid var(--border)',
		}
	}
	return {
		backgroundColor: `${meta.value.color}15`,
		borderColor: `${meta.value.color}40`,
		color: meta.value.color,
	}
})
</script>

<style scoped lang="scss">
.h-logo {
	display: inline-flex;
	align-items: center;
	justify-content: center;
	border-radius: var(--radius-md);
	border-width: 1px;
	border-style: solid;
	flex-shrink: 0;
	overflow: hidden;
	user-select: none;

	&.size-small {
		width: 1.875rem;
		height: 1.875rem;
		font-size: 0.7rem;
	}

	&.size-medium {
		width: 2.5rem;
		height: 2.5rem;
		font-size: 0.8125rem;
	}

	&.size-large {
		width: 2.75rem;
		height: 2.75rem;
		font-size: 0.875rem;
	}

	.logo-img {
		width: 100%;
		height: 100%;
		object-fit: contain;
		padding: 0.25rem;
	}

	.logo-text {
		letter-spacing: -0.02em;
	}
}
</style>
