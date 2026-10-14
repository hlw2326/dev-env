<template>
	<div class="detail-logs">
		<div class="log-toolbar">
			<div class="log-tabs">
				<button
					type="button"
					class="tab-btn"
					:class="{ active: logKind === 'access' }"
					@click="onSwitchKind('access')"
				>
					访问日志
				</button>
				<button
					v-if="site.kind === 'node'"
					type="button"
					class="tab-btn"
					:class="{ active: logKind === 'run' }"
					@click="onSwitchKind('run')"
				>
					运行日志
				</button>
			</div>
			<p-button
				size="small"
				severity="secondary"
				variant="outlined"
				icon="pi pi-refresh"
				label="刷新"
				@click="fetchLogs"
			/>
		</div>

		<div class="log-box font-mono">
			<div v-if="lines.length === 0" class="empty-log">暂无日志记录</div>
			<div v-for="(line, idx) in lines" :key="idx" class="log-line">{{ line }}</div>
		</div>
	</div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { loadSiteLogs } from '@/api'
import type { Site } from '@/types'

const props = defineProps<{
	site: Site
}>()

const logKind = ref('access')
const lines = ref<string[]>([])

async function fetchLogs() {
	lines.value = await loadSiteLogs(props.site.id, logKind.value, 200)
}

function onSwitchKind(kind: string) {
	logKind.value = kind
	fetchLogs()
}

watch(
	() => props.site.id,
	() => {
		logKind.value = 'access'
		fetchLogs()
	},
	{ immediate: true }
)
</script>

<style scoped lang="scss">
.detail-logs {
	display: flex;
	flex-direction: column;
	gap: 0.75rem;
}

.log-toolbar {
	display: flex;
	align-items: center;
	justify-content: space-between;
}

.log-tabs {
	display: flex;
	gap: 0.5rem;
}

.tab-btn {
	border: 1px solid var(--border-color);
	background: var(--bg-app);
	color: var(--text-secondary);
	font-size: var(--font-sm);
	padding: 0.35rem 0.75rem;
	border-radius: var(--radius-sm);
	cursor: pointer;

	&.active {
		background: var(--p-primary-color);
		color: var(--on-primary);
		border-color: var(--p-primary-color);
	}
}

.log-box {
	background: var(--code-bg);
	color: var(--code-text);
	border-radius: var(--radius-md);
	padding: 0.875rem;
	height: 20rem;
	overflow-y: auto;
	font-size: var(--font-xs);
	display: flex;
	flex-direction: column;
	gap: 0.25rem;
}

.empty-log {
	color: var(--text-muted);
}

.log-line {
	white-space: pre-wrap;
	word-break: break-all;
}
</style>
