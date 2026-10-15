<template>
	<div class="kind-options">
		<button
			v-for="item in templates"
			:key="item.kind"
			type="button"
			class="kind-btn"
			:class="{ active: model === item.kind }"
			@click="model = item.kind"
		>
			<i :class="[item.icon, 'kind-icon']"></i>
			<span>{{ item.label }}</span>
		</button>
	</div>
</template>

<script setup lang="ts">
import type { ProjectTemplate } from '@/types'

const model = defineModel<string>({ required: true })

const { templates } = defineProps<{
	templates: ProjectTemplate[]
}>()
</script>

<style scoped lang="scss">
.kind-options {
	display: grid;
	grid-template-columns: repeat(4, 1fr);
	gap: 0.625rem;
}

.kind-btn {
	display: flex;
	align-items: center;
	justify-content: center;
	gap: 0.5rem;
	height: 2.5rem;
	border: 1px solid var(--border-color);
	border-radius: var(--h-radius);
	background: var(--bg-app);
	color: var(--text-secondary);
	font-size: var(--font-sm);
	cursor: pointer;

	&:hover {
		border-color: var(--border-hover);
		color: var(--text-primary);
	}

	&.active {
		background: var(--p-primary-color);
		color: var(--on-primary);
		border-color: var(--p-primary-color);
	}

	.kind-icon {
		font-size: var(--icon-md);
	}
}
</style>
