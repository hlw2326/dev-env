<template>
	<h-modal
		v-if="item !== null"
		v-model:visible="visible"
		:title="`安装 ${item.name}`"
		width="30rem"
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
				>
					<div class="ver-info">
						<span class="ver-num font-mono">{{ ver.version }}</span>
						<span v-if="ver.recommended" class="rec-tag">推荐</span>
						<span v-if="ver.note !== ''" class="note-tag">{{ ver.note }}</span>
					</div>

					<div>
						<span v-if="ver.installed" class="installed-label">已安装</span>
						<p-button
							v-else
							size="small"
							icon="pi pi-download"
							label="安装"
							@click="emit('install', item.kind, ver.version)"
						/>
					</div>
				</div>
			</div>
		</div>
	</h-modal>
</template>

<script setup lang="ts">
import type { MarketItem, SoftwareVersion } from '@/types'

defineProps<{
	item: MarketItem | null
	versions: SoftwareVersion[]
	loading: boolean
	error: string
}>()

const visible = defineModel<boolean>('visible', { default: false })

const emit = defineEmits<{
	install: [kind: string, version: string]
	retry: [kind: string]
}>()
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
	gap: 0.5rem;
	max-height: 22rem;
	overflow-y: auto;
}

.version-row {
	display: flex;
	align-items: center;
	justify-content: space-between;
	padding: 0.625rem 0.875rem;
	background: var(--bg-app);
	border: 1px solid var(--border-color);
	border-radius: var(--radius-md);
}

.ver-info {
	display: flex;
	align-items: center;
	gap: 0.5rem;
}

.ver-num {
	font-size: var(--font-base);
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

.installed-label {
	font-size: var(--font-xs);
	color: var(--text-muted);
}
</style>
