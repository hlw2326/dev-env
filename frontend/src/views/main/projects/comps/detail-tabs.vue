<template>
	<div class="panel-tabs detail-tabs">
		<p-button
			v-for="tab in tabs"
			:key="tab.value"
			type="button"
			unstyled
			class="panel-tab"
			:class="{ active: currentTab === tab.value }"
			@click="currentTab = tab.value"
		>
			{{ tab.label }}
		</p-button>
	</div>
</template>

<script setup lang="ts">
defineOptions({ name: 'CompDetailTabs' })

defineProps<{
	tabs: Array<{ value: string; label: string }>
}>()

const currentTab = defineModel<string>({ required: true })
</script>

<style scoped lang="scss">
.detail-tabs {
	padding: var(--h-padding);
	flex-wrap: wrap;
	margin-bottom: var(--h-padding);
	background: var(--bg-card);
}
</style>
