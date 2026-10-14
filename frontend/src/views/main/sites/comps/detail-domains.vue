<template>
	<div class="detail-domains">
		<div class="add-row">
			<p-input-text
				v-model="newHost"
				placeholder="输入新域名，如 api.localhost"
				class="font-mono flex-input"
				@keyup.enter="onAdd"
			/>
			<p-button icon="pi pi-plus" label="添加域名" @click="onAdd" />
		</div>

		<div class="host-list">
			<div v-for="host in site.hosts" :key="host" class="host-item">
				<span class="host-name font-mono">{{ host }}</span>
				<p-button
					size="small"
					severity="danger"
					variant="outlined"
					icon="pi pi-trash"
					label="移除"
					:disabled="site.hosts.length <= 1"
					@click="onRemove(host)"
				/>
			</div>
		</div>
	</div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import type { Site, UpdateSiteInput } from '@/types'

const props = defineProps<{
	site: Site
}>()

const emit = defineEmits<{
	save: [input: UpdateSiteInput]
}>()

const newHost = ref('')

function onAdd() {
	const clean = newHost.value.trim().toLowerCase()
	if (clean === '') {
		return
	}
	if (props.site.hosts.includes(clean)) {
		newHost.value = ''
		return
	}
	const nextHosts = [...props.site.hosts, clean]
	newHost.value = ''
	emit('save', {
		id: props.site.id,
		hosts: nextHosts,
		root: props.site.root,
		run: props.site.run,
		software: props.site.software,
		command: props.site.command,
		autostart: props.site.autostart,
	})
}

function onRemove(target: string) {
	if (props.site.hosts.length <= 1) {
		return
	}
	const nextHosts = props.site.hosts.filter((h) => h !== target)
	emit('save', {
		id: props.site.id,
		hosts: nextHosts,
		root: props.site.root,
		run: props.site.run,
		software: props.site.software,
		command: props.site.command,
		autostart: props.site.autostart,
	})
}
</script>

<style scoped lang="scss">
.detail-domains {
	display: flex;
	flex-direction: column;
	gap: 0.875rem;
}

.add-row {
	display: flex;
	align-items: center;
	gap: 0.5rem;
}

.flex-input {
	flex: 1;
}

.host-list {
	display: flex;
	flex-direction: column;
	gap: 0.5rem;
}

.host-item {
	display: flex;
	align-items: center;
	justify-content: space-between;
	padding: 0.625rem 0.875rem;
	background: var(--bg-app);
	border: 1px solid var(--border-color);
	border-radius: var(--radius-md);
}

.host-name {
	font-size: var(--font-base);
	color: var(--text-primary);
}
</style>
