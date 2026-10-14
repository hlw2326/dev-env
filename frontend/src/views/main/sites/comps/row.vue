<template>
	<div class="site-row surface-card">
		<div class="left">
			<div class="kind-badge" :class="site.kind">
				<i :class="kindIcons[site.kind]"></i>
			</div>
			<div class="info">
				<div class="top-line">
					<span class="name">{{ site.name }}</span>
					<button type="button" class="domain-link font-mono" @click="emit('open', site)">
						<span>{{ site.hosts[0] }}</span>
						<i class="pi pi-external-link link-icon"></i>
					</button>
					<span class="state-badge" :class="site.state">
						<span class="dot"></span>
						<span>{{ stateLabels[site.state] }}</span>
					</span>
				</div>
				<div class="meta-line">
					<span class="kind-tag">{{ kindLabels[site.kind] }}</span>
					<span v-if="site.version !== ''" class="version-tag font-mono">{{ site.version }}</span>
					<div v-if="site.need !== ''" class="need-box">
						<span class="need-text">未安装 {{ needLabels[site.need] }} 环境</span>
						<p-button
							size="small"
							:label="`一键安装 ${needLabels[site.need]}`"
							icon="pi pi-download"
							:loading="installingKind === site.need"
							@click="emit('install', site.need)"
						/>
					</div>
					<span class="path-text font-mono">{{ site.path }}</span>
				</div>
			</div>
		</div>

		<div class="actions">
			<p-button
				v-if="site.state === 'running'"
				size="small"
				severity="secondary"
				variant="outlined"
				icon="pi pi-stop"
				label="停止"
				@click="emit('toggle', site)"
			/>
			<p-button
				v-else
				size="small"
				icon="pi pi-play"
				label="启动"
				:disabled="site.need !== ''"
				@click="emit('toggle', site)"
			/>
			<p-button
				size="small"
				severity="secondary"
				variant="outlined"
				icon="pi pi-external-link"
				label="打开"
				@click="emit('open', site)"
			/>
			<p-button
				size="small"
				severity="secondary"
				variant="outlined"
				icon="pi pi-folder-open"
				label="目录"
				@click="emit('folder', site.root)"
			/>
			<p-button
				size="small"
				severity="secondary"
				variant="outlined"
				icon="pi pi-sliders-h"
				label="详情"
				@click="emit('detail', site)"
			/>
			<p-button
				size="small"
				severity="danger"
				variant="outlined"
				icon="pi pi-trash"
				label="删除"
				@click="onConfirmDelete($event)"
			/>
		</div>
	</div>
</template>

<script setup lang="ts">
import { useConfirm } from 'primevue/useconfirm'
import type { Site } from '@/types'

const props = defineProps<{
	site: Site
	installingKind: string
}>()

const emit = defineEmits<{
	toggle: [site: Site]
	open: [site: Site]
	folder: [path: string]
	detail: [site: Site]
	delete: [site: Site]
	install: [kind: string]
}>()

const confirm = useConfirm()

const kindLabels: Record<string, string> = {
	php: 'PHP',
	node: 'Node.js',
	static: '静态',
}

const kindIcons: Record<string, string> = {
	php: 'pi pi-code',
	node: 'pi pi-bolt',
	static: 'pi pi-file',
}

const stateLabels: Record<string, string> = {
	running: '运行中',
	stopped: '已停止',
	error: '异常',
}

const needLabels: Record<string, string> = {
	php: 'PHP',
	nodejs: 'Node.js',
}

function onConfirmDelete(event: MouseEvent) {
	confirm.require({
		target: event.currentTarget as HTMLElement,
		message: '只删除站点记录，不会删除项目目录',
		icon: 'pi pi-exclamation-triangle',
		acceptLabel: '删除',
		rejectLabel: '取消',
		acceptProps: { severity: 'danger', size: 'small' },
		rejectProps: { severity: 'secondary', variant: 'outlined', size: 'small' },
		accept: () => {
			emit('delete', props.site)
		},
	})
}
</script>

<style scoped lang="scss">
.site-row {
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

.kind-badge {
	width: 2.5rem;
	height: 2.5rem;
	border-radius: var(--radius-md);
	display: flex;
	align-items: center;
	justify-content: center;
	font-size: var(--icon-lg);
	flex-shrink: 0;

	&.php {
		background: var(--indigo-bg);
		color: var(--indigo);
		border: 1px solid var(--indigo-border);
	}

	&.node {
		background: var(--ok-bg);
		color: var(--ok);
		border: 1px solid var(--ok-border);
	}

	&.static {
		background: var(--info-bg);
		color: var(--info);
		border: 1px solid var(--info-border);
	}
}

.info {
	display: flex;
	flex-direction: column;
	gap: 0.35rem;
	min-width: 0;
	flex: 1;
}

.top-line {
	display: flex;
	align-items: center;
	gap: 0.75rem;
	flex-wrap: wrap;
}

.name {
	font-size: var(--font-lg);
	color: var(--text-primary);
}

.domain-link {
	display: inline-flex;
	align-items: center;
	gap: 0.35rem;
	border: none;
	background: var(--bg-app);
	padding: 0.15rem 0.5rem;
	border-radius: var(--radius-sm);
	color: var(--info);
	font-size: var(--font-xs);
	cursor: pointer;

	&:hover {
		background: var(--bg-hover);
	}

	.link-icon {
		font-size: var(--icon-xs);
	}
}

.state-badge {
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

	&.running {
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

	&.error {
		color: var(--err);
		background: var(--err-bg);
		border: 1px solid var(--err-border);

		.dot {
			background: var(--err);
		}
	}
}

.meta-line {
	display: flex;
	align-items: center;
	gap: 0.625rem;
	flex-wrap: wrap;
}

.kind-tag,
.version-tag {
	font-size: var(--font-xs);
	color: var(--text-secondary);
	background: var(--bg-app);
	padding: 0.1rem 0.45rem;
	border-radius: var(--radius-sm);
	border: 1px solid var(--border-color);
}

.need-box {
	display: inline-flex;
	align-items: center;
	gap: 0.5rem;
}

.need-text {
	font-size: var(--font-xs);
	color: var(--warn);
}

.path-text {
	font-size: var(--font-xs);
	color: var(--text-muted);
	overflow: hidden;
	text-overflow: ellipsis;
	white-space: nowrap;
}

.actions {
	display: flex;
	align-items: center;
	gap: 0.45rem;
	flex-shrink: 0;
}
</style>
