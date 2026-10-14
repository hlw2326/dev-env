<template>
	<div class="cfg-nginx">
		<div class="sub-tabs">
			<button
				v-for="t in tabs"
				:key="t.value"
				type="button"
				class="sub-tab"
				:class="{ active: currentTab === t.value }"
				@click="onSwitchTab(t.value)"
			>
				{{ t.label }}
			</button>
		</div>

		<div v-if="currentTab === 'overview'" class="overview-pane">
			<div class="kv-row">
				<span class="kv-label">运行状态</span>
				<div class="row-actions">
					<span v-if="item.running" class="status-ok">运行中 (PID {{ item.pid }})</span>
					<span v-else class="status-stop">已停止</span>
					<p-button
						v-if="item.running"
						size="small"
						severity="secondary"
						variant="outlined"
						icon="pi pi-stop"
						label="停止"
						@click="emit('stop', item)"
					/>
					<p-button
						v-else
						size="small"
						icon="pi pi-play"
						label="启动"
						@click="emit('start', item)"
					/>
				</div>
			</div>

			<div class="kv-row">
				<span class="kv-label">监听端口</span>
				<span class="kv-val font-mono">{{ item.port }}</span>
			</div>

			<div class="kv-row">
				<span class="kv-label">安装目录</span>
				<div class="row-actions">
					<span class="kv-val font-mono">{{ item.path }}</span>
					<p-button
						size="small"
						severity="secondary"
						variant="outlined"
						icon="pi pi-folder-open"
						label="打开"
						@click="emit('openFolder', item.id)"
					/>
				</div>
			</div>

			<div class="kv-row">
				<span class="kv-label">随应用启动</span>
				<p-toggle-switch
					:model-value="item.autostart"
					@update:model-value="(val: boolean) => emit('autostart', item, val)"
				/>
			</div>
		</div>

		<div v-else-if="currentTab === 'common'" class="common-pane">
			<div v-for="cfg in configItems" :key="cfg.key" class="form-field">
				<label>{{ cfg.label }} <span class="key-hint font-mono">({{ cfg.key }})</span></label>
				<p-input-text v-model="cfg.value" class="font-mono" />
				<span class="desc-text">{{ cfg.description }}</span>
			</div>
			<div class="footer-bar">
				<p-button
					severity="secondary"
					variant="outlined"
					icon="pi pi-check"
					label="仅保存"
					:loading="saving"
					@click="submitConfig(item.id, false, item.running)"
				/>
				<p-button
					icon="pi pi-refresh"
					label="保存并重启"
					:loading="saving"
					@click="submitConfig(item.id, true, item.running)"
				/>
			</div>
		</div>

		<cfg-file
			v-else-if="currentTab === 'file'"
			v-model="rawText"
			:file-path="configFile"
			:saving="saving"
			@save="submitRawFile(item.id)"
		/>

		<cfg-log
			v-else-if="currentTab === 'log'"
			:lines="logLines"
			@refresh="fetchLogs(item.id)"
		/>
	</div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { useConfig } from '../use-config'
import CfgFile from './cfg-file.vue'
import CfgLog from './cfg-log.vue'
import type { Software } from '@/types'

const props = defineProps<{
	item: Software
}>()

const emit = defineEmits<{
	start: [item: Software]
	stop: [item: Software]
	autostart: [item: Software, on: boolean]
	openFolder: [id: number]
}>()

const tabs = [
	{ value: 'overview', label: '概览' },
	{ value: 'common', label: '常用配置' },
	{ value: 'file', label: '配置文件' },
	{ value: 'log', label: '日志' },
]

const currentTab = ref('overview')

const {
	configItems,
	configFile,
	rawText,
	logLines,
	saving,
	fetchConfig,
	submitConfig,
	fetchRawFile,
	submitRawFile,
	fetchLogs,
} = useConfig()

function onSwitchTab(tab: string) {
	currentTab.value = tab
	if (tab === 'common') {
		fetchConfig(props.item.id)
	} else if (tab === 'file') {
		fetchConfig(props.item.id)
		fetchRawFile(props.item.id)
	} else if (tab === 'log') {
		fetchLogs(props.item.id)
	}
}

watch(
	() => props.item.id,
	() => {
		currentTab.value = 'overview'
	},
	{ immediate: true }
)
</script>

<style scoped lang="scss">
.cfg-nginx {
	display: flex;
	flex-direction: column;
	gap: 1rem;
}

.sub-tabs {
	display: flex;
	gap: 0.5rem;
	border-bottom: 1px solid var(--border-color);
	padding-bottom: 0.75rem;
}

.sub-tab {
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

.overview-pane {
	display: flex;
	flex-direction: column;
	gap: 0.75rem;
}

.kv-row {
	display: flex;
	align-items: center;
	justify-content: space-between;
	padding: 0.75rem 0.875rem;
	background: var(--bg-app);
	border: 1px solid var(--border-color);
	border-radius: var(--radius-md);
}

.kv-label {
	font-size: var(--font-sm);
	color: var(--text-secondary);
}

.kv-val {
	font-size: var(--font-xs);
	color: var(--text-primary);
}

.row-actions {
	display: flex;
	align-items: center;
	gap: 0.625rem;
}

.status-ok {
	font-size: var(--font-xs);
	color: var(--ok);
}

.status-stop {
	font-size: var(--font-xs);
	color: var(--text-muted);
}

.common-pane {
	display: flex;
	flex-direction: column;
}

.key-hint {
	font-size: var(--font-xs);
	color: var(--text-muted);
}

.desc-text {
	font-size: var(--font-xs);
	color: var(--text-muted);
}

.footer-bar {
	display: flex;
	justify-content: flex-end;
	gap: 0.5rem;
	margin-top: 0.5rem;
}
</style>
