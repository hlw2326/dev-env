<template>
	<h-modal
		v-if="item !== null"
		v-model:visible="visible"
		:title="`安装 ${item.name}`"
		width="34rem"
		:footer="false"
	>
		<div class="install-modal">
			<div v-if="loading" class="state-box">
				<i class="pi pi-spin pi-spinner state-icon"></i>
				<span class="state-text">正在从官方源获取版本清单...</span>
			</div>

			<div v-else-if="error !== ''" class="state-box error">
				<span class="error-msg">{{ error }}</span>
				<p-button
					size="small"
					icon="pi pi-refresh"
					label="重试"
					@click="emit('retry', item.kind)"
				/>
			</div>

			<div v-else class="version-list">
				<div
					v-for="ver in versions"
					:key="ver.version"
					class="version-row"
					:class="{ downloading: !!getActiveTask(ver.version) }"
				>
					<div class="row-main">
						<div class="ver-info">
							<span class="ver-num font-mono">{{ ver.version }}</span>
							<span v-if="ver.recommended" class="rec-tag">推荐</span>
							<span v-if="ver.note !== ''" class="note-tag">{{ ver.note }}</span>
						</div>

						<div class="ver-actions">
							<span v-if="ver.installed" class="installed-label">
								<i class="pi pi-check-circle"></i> 已安装
							</span>
							<template v-else-if="getActiveTask(ver.version)">
								<p-button
									size="small"
									severity="danger"
									variant="outlined"
									icon="pi pi-times"
									label="取消"
									@click="emit('cancel', item.kind, ver.version)"
								/>
							</template>
							<p-button
								v-else
								size="small"
								icon="pi pi-download"
								label="安装"
								:disabled="hasActiveDownloading && !getActiveTask(ver.version)"
								@click="emit('install', item.kind, ver.version)"
							/>
						</div>
					</div>

					<!-- 下载与安装实时进度条 -->
					<div v-if="getActiveTask(ver.version)" class="download-progress-box">
						<div class="progress-info">
							<span class="progress-msg">
								<i class="pi pi-spin pi-spinner"></i>
								{{ getActiveTask(ver.version)?.message }}
							</span>
							<span class="progress-pct font-mono">{{ getActiveTask(ver.version)?.progress }}%</span>
						</div>
						<p-progress-bar
							:value="getActiveTask(ver.version)?.progress"
							:show-value="false"
							class="download-bar"
						/>
					</div>
				</div>
			</div>
		</div>
	</h-modal>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { MarketItem, Software, SoftwareVersion } from '@/types'

const props = defineProps<{
	item: MarketItem | null
	versions: SoftwareVersion[]
	loading: boolean
	error: string
	progressMap?: Record<string, { progress: number; message: string }>
	installedList?: Software[]
}>()

const visible = defineModel<boolean>('visible', { default: false })

const emit = defineEmits<{
	install: [kind: string, version: string]
	cancel: [kind: string, version: string]
	retry: [kind: string]
}>()

function getActiveTask(version: string) {
	if (!props.item) return null
	const key = `${props.item.kind}:${version}`
	if (props.progressMap && props.progressMap[key]) {
		return props.progressMap[key]
	}
	if (props.installedList) {
		const found = props.installedList.find(
			(s) => s.kind === props.item?.kind && s.version === version && s.state === 'installing'
		)
		if (found) {
			return { progress: found.progress || 1, message: '正在准备下载...' }
		}
	}
	return null
}

const hasActiveDownloading = computed(() => {
	if (!props.item) return false
	if (props.progressMap) {
		for (const key of Object.keys(props.progressMap)) {
			if (key.startsWith(`${props.item.kind}:`)) {
				return true
			}
		}
	}
	if (props.installedList) {
		return props.installedList.some(
			(s) => s.kind === props.item?.kind && s.state === 'installing'
		)
	}
	return false
})
</script>

<style scoped lang="scss">
.install-modal {
	display: flex;
	flex-direction: column;
	gap: 0.75rem;
}

.state-box {
	display: flex;
	flex-direction: column;
	align-items: center;
	justify-content: center;
	padding: 2rem 1rem;
	gap: 0.75rem;

	.state-icon {
		font-size: var(--icon-lg);
		color: var(--text-secondary);
	}

	.state-text {
		font-size: var(--font-sm);
		color: var(--text-secondary);
	}

	&.error {
		background: var(--err-bg);
		border: 1px solid var(--err-border);
		border-radius: var(--radius-md);

		.error-msg {
			font-size: var(--font-sm);
			color: var(--err);
			text-align: center;
		}
	}
}

.version-list {
	display: flex;
	flex-direction: column;
	gap: 0.625rem;
	max-height: 24rem;
	overflow-y: auto;
	padding-right: 0.25rem;
}

.version-row {
	display: flex;
	flex-direction: column;
	gap: 0.5rem;
	padding: 0.75rem 0.875rem;
	background: var(--bg-app);
	border: 1px solid var(--border-color);
	border-radius: var(--radius-md);
	transition: all 0.2s ease;

	&.downloading {
		background: var(--bg-card);
		border-color: var(--p-primary-color);
		box-shadow: 0 0 0 1px var(--p-primary-color);
	}
}

.row-main {
	display: flex;
	align-items: center;
	justify-content: space-between;
	width: 100%;
}

.ver-info {
	display: flex;
	align-items: center;
	gap: 0.5rem;
	flex-wrap: wrap;
}

.ver-num {
	font-size: var(--font-base);
	font-weight: 600;
	color: var(--text-primary);
}

.rec-tag {
	font-size: var(--font-xs);
	color: var(--ok);
	background: var(--ok-bg);
	border: 1px solid var(--ok-border);
	padding: 0.1rem 0.45rem;
	border-radius: var(--radius-sm);
}

.note-tag {
	font-size: var(--font-xs);
	color: var(--text-secondary);
}

.ver-actions {
	display: flex;
	align-items: center;
	gap: 0.5rem;
}

.installed-label {
	display: flex;
	align-items: center;
	gap: 0.35rem;
	font-size: var(--font-xs);
	color: var(--ok);
	font-weight: 500;
}

.download-progress-box {
	display: flex;
	flex-direction: column;
	gap: 0.35rem;
	padding-top: 0.5rem;
	border-top: 1px dashed var(--border-color);
}

.progress-info {
	display: flex;
	justify-content: space-between;
	align-items: center;
	font-size: var(--font-xs);
	color: var(--text-secondary);
}

.progress-msg {
	display: flex;
	align-items: center;
	gap: 0.35rem;
	color: var(--text-primary);
}

.progress-pct {
	font-weight: 600;
	color: var(--p-primary-color);
}

.download-bar {
	height: 0.45rem;
	border-radius: var(--radius-full);
}
</style>
