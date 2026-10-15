<template>
	<div class="plugin-panel">
		<div class="plugin-header">
			<div>
				<div class="plugin-title">{{ plugin.name }}</div>
				<div class="plugin-description">{{ plugin.description }}</div>
			</div>
			<h-tag v-if="plugin.available" color="success" value="已就绪" />
			<h-tag v-else color="danger" value="不可用" />
		</div>

		<div v-if="plugin.version !== ''" class="plugin-version font-mono">
			运行版本：{{ plugin.version }}
		</div>

		<div v-if="!plugin.available" class="plugin-warning">
			<i class="pi pi-info-circle"></i>
			<span>{{ plugin.reason }}</span>
		</div>

		<div v-else class="plugin-actions">
			<p-button
				v-for="action in plugin.actions"
				:key="action.id"
				:label="action.name"
				icon="pi pi-play"
				size="small"
				:loading="busy"
				@click="runAction(action.id)"
			/>
		</div>
	</div>
</template>

<script setup lang="ts">
import type { ProjectPluginActionInput, ProjectPlugin, Project } from '@/types'

// 接收项目插件信息与运行状态
const props = defineProps<{
	project: Project
	plugin: ProjectPlugin
	busy: boolean
}>()

// 执行插件动作
const emit = defineEmits<{
	run: [input: ProjectPluginActionInput]
}>()

function runAction(actionID: string): void {
	emit('run', {
		projectId: props.project.id,
		pluginId: props.plugin.id,
		actionId: actionID,
	})
}
</script>

<style scoped lang="scss">
.plugin-panel {
	display: flex;
	flex-direction: column;
	gap: 0.875rem;
}

.plugin-header {
	display: flex;
	align-items: flex-start;
	justify-content: space-between;
	gap: 1rem;
	padding: 0.875rem;
	background: var(--bg-app);
	border: 1px solid var(--border-color);
	border-radius: var(--h-radius);
}

.plugin-title {
	font-size: var(--font-base);
	color: var(--text-primary);
}

.plugin-description,
.plugin-version {
	font-size: var(--font-xs);
	color: var(--text-secondary);
}

.plugin-description {
	margin-top: 0.25rem;
}

.plugin-warning {
	display: flex;
	align-items: center;
	gap: 0.5rem;
	padding: 0.75rem 0.875rem;
	color: var(--text-secondary);
	background: var(--bg-app);
	border: 1px solid var(--border-color);
	border-radius: var(--h-radius);
}

.plugin-actions {
	display: flex;
	flex-wrap: wrap;
	gap: 0.5rem;
}
</style>
