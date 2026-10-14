<template>
	<div class="market-card surface-card">
		<div class="card-top">
			<div class="pkg-badge" :class="item.kind">
				<i :class="kindIcons[item.kind]"></i>
			</div>
			<div class="title-col">
				<div class="name-line">
					<span class="pkg-name">{{ item.name }}</span>
					<span class="cat-badge">{{ item.category }}</span>
					<span v-if="item.count > 0" class="installed-badge">已安装 {{ item.count }} 个版本</span>
				</div>
				<p class="pkg-desc">{{ item.description }}</p>
			</div>
		</div>

		<div v-if="activeTask !== null" class="progress-area">
			<div class="progress-info">
				<span class="progress-msg">{{ activeTask.version }} · {{ activeTask.message }}</span>
				<span class="progress-num font-mono">{{ activeTask.progress }}%</span>
			</div>
			<p-progress-bar :value="activeTask.progress" :show-value="false" class="bar" />
			<div class="progress-actions">
				<p-button
					size="small"
					severity="danger"
					variant="outlined"
					icon="pi pi-times"
					label="取消"
					@click="emit('cancel', item.kind, activeTask.version)"
				/>
			</div>
		</div>

		<div v-else class="card-footer">
			<span v-if="item.recommended !== ''" class="rec-hint font-mono">
				推荐版本：{{ item.recommended }}
			</span>
			<span v-else class="rec-hint">点击选择版本安装</span>

			<div class="btn-group">
				<p-button
					v-if="item.count > 0"
					size="small"
					severity="secondary"
					variant="outlined"
					icon="pi pi-cog"
					label="设置"
					@click="emit('config', item.kind)"
				/>
				<p-button
					v-if="item.recommended !== ''"
					size="small"
					icon="pi pi-download"
					:label="`安装 ${item.recommended}`"
					@click="emit('openInstall', item)"
				/>
				<p-button
					v-else
					size="small"
					icon="pi pi-download"
					label="选择版本安装"
					@click="emit('openInstall', item)"
				/>
			</div>
		</div>
	</div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { MarketItem, Software } from '@/types'

const props = defineProps<{
	item: MarketItem
	installedList: Software[]
	progressMap: Record<string, { progress: number; message: string }>
}>()

const emit = defineEmits<{
	openInstall: [item: MarketItem]
	cancel: [kind: string, version: string]
	config: [kind: string]
}>()

const kindIcons: Record<string, string> = {
	nginx: 'pi pi-compass',
	php: 'pi pi-code',
	nodejs: 'pi pi-bolt',
	mysql: 'pi pi-database',
	redis: 'pi pi-server',
}

const activeTask = computed(() => {
	for (const sw of props.installedList) {
		if (sw.kind === props.item.kind && sw.state === 'installing') {
			const key = `${sw.kind}:${sw.version}`
			const live = props.progressMap[key]
			if (live) {
				return { version: sw.version, progress: live.progress, message: live.message }
			}
			return { version: sw.version, progress: sw.progress, message: '安装中...' }
		}
	}
	for (const [key, val] of Object.entries(props.progressMap)) {
		if (key.startsWith(`${props.item.kind}:`)) {
			const ver = key.slice(props.item.kind.length + 1)
			return { version: ver, progress: val.progress, message: val.message }
		}
	}
	return null
})
</script>

<style scoped lang="scss">
.market-card {
	display: flex;
	flex-direction: column;
	justify-content: space-between;
	padding: 1.125rem;
	gap: 1rem;
}

.card-top {
	display: flex;
	align-items: flex-start;
	gap: 0.875rem;
}

.pkg-badge {
	width: 2.75rem;
	height: 2.75rem;
	border-radius: var(--radius-md);
	display: flex;
	align-items: center;
	justify-content: center;
	font-size: var(--icon-lg);
	flex-shrink: 0;

	&.nginx {
		background: var(--warn-bg);
		color: var(--warn);
		border: 1px solid var(--warn-border);
	}

	&.php {
		background: var(--indigo-bg);
		color: var(--indigo);
		border: 1px solid var(--indigo-border);
	}

	&.nodejs {
		background: var(--ok-bg);
		color: var(--ok);
		border: 1px solid var(--ok-border);
	}

	&.mysql {
		background: var(--info-bg);
		color: var(--info);
		border: 1px solid var(--info-border);
	}

	&.redis {
		background: var(--err-bg);
		color: var(--err);
		border: 1px solid var(--err-border);
	}
}

.title-col {
	display: flex;
	flex-direction: column;
	gap: 0.35rem;
	flex: 1;
}

.name-line {
	display: flex;
	align-items: center;
	gap: 0.5rem;
	flex-wrap: wrap;
}

.pkg-name {
	font-size: var(--font-lg);
	color: var(--text-primary);
}

.cat-badge {
	font-size: var(--font-xs);
	color: var(--text-secondary);
	background: var(--bg-app);
	border: 1px solid var(--border-color);
	padding: 0.1rem 0.45rem;
	border-radius: var(--radius-sm);
}

.installed-badge {
	font-size: var(--font-xs);
	color: var(--ok);
	background: var(--ok-bg);
	border: 1px solid var(--ok-border);
	padding: 0.1rem 0.45rem;
	border-radius: var(--radius-sm);
}

.pkg-desc {
	font-size: var(--font-sm);
	color: var(--text-secondary);
}

.progress-area {
	display: flex;
	flex-direction: column;
	gap: 0.5rem;
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

.bar {
	height: 0.45rem;
}

.progress-actions {
	display: flex;
	justify-content: flex-end;
}

.card-footer {
	display: flex;
	align-items: center;
	justify-content: space-between;
	padding-top: 0.75rem;
	border-top: 1px dashed var(--border-color);
}

.rec-hint {
	font-size: var(--font-xs);
	color: var(--text-muted);
}

.btn-group {
	display: flex;
	align-items: center;
	gap: 0.5rem;
}
</style>
