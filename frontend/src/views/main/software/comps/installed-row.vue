<template>
	<div class="installed-row surface-card">
		<div class="left">
			<div class="pkg-icon" :class="item.kind">
				<i :class="kindIcons[item.kind]"></i>
			</div>
			<div class="info">
				<div class="top-line">
					<span class="item-name">{{ kindNames[item.kind] }} {{ item.version }}</span>
					<span v-if="item.main" class="main-tag">默认</span>

					<span v-if="item.state === 'installing'" class="status-pill installing">
						<i class="pi pi-spin pi-spinner status-icon"></i>
						<span>安装中 {{ activeProgress.progress }}%</span>
					</span>
					<span v-else-if="item.state === 'failed'" class="status-pill failed">
						<span class="dot"></span>
						<span>安装失败</span>
					</span>
					<span v-else-if="isDaemon && item.running" class="status-pill running">
						<span class="dot"></span>
						<span>运行中 (PID {{ item.pid }})</span>
					</span>
					<span v-else-if="isDaemon && !item.running" class="status-pill stopped">
						<span class="dot"></span>
						<span>已停止</span>
					</span>
					<span v-else class="status-pill ok">
						<span class="dot"></span>
						<span>可用</span>
					</span>

					<span v-if="isDaemon && item.port > 0" class="port-badge font-mono">
						端口 {{ item.port }}
					</span>
				</div>

				<div class="sub-line">
					<div v-if="item.state === 'installing'" class="row-progress-box">
						<div class="row-progress-info">
							<span class="progress-msg">{{ activeProgress.message }}</span>
							<span class="progress-pct font-mono">{{ activeProgress.progress }}%</span>
						</div>
						<p-progress-bar :value="activeProgress.progress" :show-value="false" class="row-progress-bar" />
					</div>
					<span v-else-if="item.state === 'failed'" class="error-text">{{ item.error }}</span>
					<span v-else class="path-text font-mono">{{ item.path }}</span>
				</div>
			</div>
		</div>

		<div class="actions">
			<template v-if="item.state === 'installing'">
				<p-button
					size="small"
					severity="danger"
					variant="outlined"
					icon="pi pi-times"
					label="取消"
					@click="emit('cancel', item.kind, item.version)"
				/>
			</template>

			<template v-else-if="item.state === 'failed'">
				<p-button
					size="small"
					icon="pi pi-refresh"
					label="重试"
					@click="emit('retry', item.kind, item.version)"
				/>
				<p-button
					size="small"
					severity="danger"
					variant="outlined"
					icon="pi pi-trash"
					label="清理"
					@click="emit('remove', item)"
				/>
			</template>

			<template v-else>
				<template v-if="isDaemon">
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
				</template>

				<p-button
					v-if="!item.main"
					size="small"
					severity="secondary"
					variant="outlined"
					label="设为默认"
					@click="emit('setMain', item)"
				/>

				<p-button
					size="small"
					severity="secondary"
					variant="outlined"
					icon="pi pi-cog"
					label="设置"
					@click="emit('config', item)"
				/>

				<p-button
					size="small"
					severity="danger"
					variant="outlined"
					icon="pi pi-trash"
					label="卸载"
					@click="onConfirmRemove($event)"
				/>
			</template>
		</div>
	</div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useConfirm } from 'primevue/useconfirm'
import type { Software } from '@/types'

const props = defineProps<{
	item: Software
	progressMap?: Record<string, { progress: number; message: string }>
}>()

const activeProgress = computed(() => {
	const key = `${props.item.kind}:${props.item.version}`
	if (props.progressMap && props.progressMap[key]) {
		return props.progressMap[key]
	}
	return {
		progress: props.item.progress || 0,
		message: '正在安装...',
	}
})

const emit = defineEmits<{
	start: [item: Software]
	stop: [item: Software]
	setMain: [item: Software]
	config: [item: Software]
	remove: [item: Software]
	cancel: [kind: string, version: string]
	retry: [kind: string, version: string]
}>()

const confirm = useConfirm()

const kindNames: Record<string, string> = {
	nginx: 'Nginx',
	apache: 'Apache',
	php: 'PHP',
	nodejs: 'Node.js',
	mysql: 'MySQL',
	redis: 'Redis',
}

const kindIcons: Record<string, string> = {
	nginx: 'pi pi-compass',
	apache: 'pi pi-globe',
	php: 'pi pi-code',
	nodejs: 'pi pi-bolt',
	mysql: 'pi pi-database',
	redis: 'pi pi-server',
}

const isDaemon = computed(() => {
	return props.item.kind === 'mysql' || props.item.kind === 'redis' || props.item.kind === 'nginx' || props.item.kind === 'apache'
})

function onConfirmRemove(event: MouseEvent) {
	let message = `确认卸载 ${kindNames[props.item.kind]} ${props.item.version} 吗？`
	if (props.item.kind === 'mysql') {
		message = `卸载只删除软件目录，数据库数据目录 dbdata\\mysql-${props.item.version} 将完整保留`
	}
	confirm.require({
		target: event.currentTarget as HTMLElement,
		message,
		icon: 'pi pi-exclamation-triangle',
		acceptLabel: '卸载',
		rejectLabel: '取消',
		acceptProps: { severity: 'danger', size: 'small' },
		rejectProps: { severity: 'secondary', variant: 'outlined', size: 'small' },
		accept: () => {
			emit('remove', props.item)
		},
	})
}
</script>

<style scoped lang="scss">
.installed-row {
	display: flex;
	align-items: center;
	justify-content: space-between;
	padding: 0.875rem 1.125rem;
	gap: 1rem;
}

.left {
	display: flex;
	align-items: center;
	gap: 0.875rem;
	min-width: 0;
	flex: 1;
}

.pkg-icon {
	width: 2.5rem;
	height: 2.5rem;
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

	&.apache {
		background: #fff1f2;
		color: #e11d48;
		border: 1px solid #fecdd3;
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

.info {
	display: flex;
	flex-direction: column;
	gap: 0.3rem;
	min-width: 0;
	flex: 1;
}

.top-line {
	display: flex;
	align-items: center;
	gap: 0.625rem;
	flex-wrap: wrap;
}

.item-name {
	font-size: var(--font-base);
	color: var(--text-primary);
}

.main-tag {
	font-size: var(--font-xs);
	color: var(--info);
	background: var(--info-bg);
	border: 1px solid var(--info-border);
	padding: 0.1rem 0.45rem;
	border-radius: var(--radius-sm);
}

.status-pill {
	display: inline-flex;
	align-items: center;
	gap: 0.35rem;
	font-size: var(--font-xs);
	padding: 0.125rem 0.5rem;
	border-radius: var(--radius-full);

	.dot {
		width: 6px;
		height: 6px;
		border-radius: 50%;
	}

	.status-icon {
		font-size: var(--icon-xs);
	}

	&.running,
	&.ok {
		color: var(--ok);
		background: var(--ok-bg);
		border: 1px solid var(--ok-border);

		.dot {
			background: var(--ok);
		}
	}

	&.stopped {
		color: var(--text-secondary);
		background: var(--bg-app);
		border: 1px solid var(--border-color);

		.dot {
			background: var(--text-muted);
		}
	}

	&.installing {
		color: var(--info);
		background: var(--info-bg);
		border: 1px solid var(--info-border);
	}

	&.failed {
		color: var(--err);
		background: var(--err-bg);
		border: 1px solid var(--err-border);

		.dot {
			background: var(--err);
		}
	}
}

.port-badge {
	font-size: var(--font-xs);
	color: var(--text-secondary);
	background: var(--bg-app);
	border: 1px solid var(--border-color);
	padding: 0.1rem 0.45rem;
	border-radius: var(--radius-sm);
}

.sub-line {
	display: flex;
	align-items: center;
	width: 100%;
}

.row-progress-box {
	display: flex;
	flex-direction: column;
	gap: 0.35rem;
	width: 100%;
	max-width: 26rem;
	margin-top: 0.25rem;
}

.row-progress-info {
	display: flex;
	align-items: center;
	justify-content: space-between;
	font-size: var(--font-xs);
	color: var(--text-secondary);
}

.progress-msg {
	color: var(--text-secondary);
	overflow: hidden;
	text-overflow: ellipsis;
	white-space: nowrap;
}

.progress-pct {
	font-weight: 600;
	color: var(--p-primary-color);
	margin-left: 0.5rem;
}

.row-progress-bar {
	height: 0.45rem;
	border-radius: var(--radius-full);
}

.path-text {
	font-size: var(--font-xs);
	color: var(--text-muted);
}

.error-text {
	font-size: var(--font-xs);
	color: var(--err);
}

.actions {
	display: flex;
	align-items: center;
	gap: 0.45rem;
	flex-shrink: 0;
}
</style>
