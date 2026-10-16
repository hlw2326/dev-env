<template>
	<div class="project-row surface-card">
		<div class="left">
			<div class="kind-badge" :class="project.kind">
				<i :class="kindIcon"></i>
			</div>
			<div class="info">
				<div class="top-line">
					<span class="name">{{ project.name }}</span>
					<p-button
						v-if="project.hosts.length > 0"
						type="button"
						unstyled
						class="domain-link font-mono"
						@click="emit('open', project)"
					>
						<span>{{ project.hosts[0] }}</span>
						<i class="pi pi-external-link link-icon"></i>
					</p-button>
					<h-tag v-else color="default" value="本地调试" />
					<h-tag
						:color="stateColor"
						:value="stateLabel"
						dot
						radius="full"
					/>
					<CompGitStatus :status="gitStatus" @open="emit('git', project)" />
				</div>
				<div class="meta-line">
					<h-tag :value="kindLabel" />
					<h-tag v-if="project.version !== ''" class="font-mono" :value="project.version" />
					<div v-if="project.need !== ''" class="need-box">
						<h-tag color="warning" :value="`未安装 ${needLabels[project.need]} 环境`" />
						<p-button
							v-if="project.need !== 'rust'"
							size="small"
							:label="`一键安装 ${needLabels[project.need]}`"
							icon="pi pi-download"
							:loading="installingKind === project.need"
							@click="emit('install', project.need)"
						/>
					</div>
					<span class="path-text font-mono">{{ project.path }}</span>
				</div>
				<div v-if="project.traffic.requestCount > 0" class="traffic-line font-mono">
					<span>请求 {{ project.traffic.requestCount }}</span>
					<span>· 流量 {{ formatBytes(project.traffic.responseBytes) }}</span>
					<span>· 错误 {{ project.traffic.errorCount }}</span>
					<span v-if="project.traffic.averageResponseMs > 0">· 平均 {{ project.traffic.averageResponseMs.toFixed(0) }}ms</span>
					<span v-if="project.traffic.lastAccess !== ''">· 最近 {{ project.traffic.lastAccess }}</span>
				</div>
			</div>
		</div>
		<div class="actions">
			<p-button
				v-if="project.state === 'starting'"
				size="small"
				severity="secondary"
				label="启动中"
				:loading="true"
				disabled
			/>
			<p-button
				v-else-if="project.state === 'running'"
				size="small"
				severity="danger"
				variant="outlined"
				icon="pi pi-stop"
				label="停止"
				@click="emit('toggle', project)"
			/>
			<p-button
				v-else
				size="small"
				icon="pi pi-play"
				label="启动"
				:disabled="project.need !== ''"
				@click="emit('toggle', project)"
			/>
			<p-button
				v-if="canOpen"
				size="small"
				severity="secondary"
				variant="outlined"
				icon="pi pi-external-link"
				label="打开"
				@click="emit('open', project)"
			/>
			<p-button
				size="small"
				severity="secondary"
				variant="outlined"
				icon="pi pi-folder-open"
				label="目录"
				@click="emit('folder', project.root)"
			/>
			<p-button
				size="small"
				severity="secondary"
				variant="outlined"
				icon="pi pi-sliders-h"
				label="详情"
				@click="emit('detail', project)"
			/>
			<p-button
				size="small"
				severity="danger"
				variant="outlined"
				icon="pi pi-trash"
				label="删除"
				:disabled="project.state === 'starting'"
				@click="confirmDelete($event)"
			/>
		</div>
	</div>
</template>

<script setup lang="ts">
import { useConfirm } from 'primevue/useconfirm'
import { computed } from 'vue'
import CompGitStatus from './git-status.vue'
import { formatBytes } from '@/utils'
import type { ProjectGitStatus, ProjectTemplate, Project } from '@/types'

const props = defineProps<{
	project: Project
	installingKind: string
	templates: ProjectTemplate[]
	gitStatus?: ProjectGitStatus
}>()
const emit = defineEmits<{
	toggle: [project: Project]
	open: [project: Project]
	folder: [path: string]
	detail: [project: Project]
	git: [project: Project]
	delete: [project: Project]
	install: [kind: string]
}>()
const confirm = useConfirm()

const currentTemplate = computed<ProjectTemplate | null>(() => {
	for (const item of props.templates) {
		if (item.kind === props.project.kind) {
			return item
		}
	}
	return null
})

const kindLabel = computed(() => {
	if (currentTemplate.value !== null) {
		return currentTemplate.value.label
	}
	return props.project.kind
})

const kindIcon = computed(() => {
	if (currentTemplate.value !== null) {
		return currentTemplate.value.icon
	}
	return 'pi pi-code'
})

const canOpen = computed(() => {
	return currentTemplate.value !== null && currentTemplate.value.capabilities.includes('domain')
})
const stateLabel = computed(() => {
	if (props.project.state === 'starting') {
		return '启动中'
	}
	if (props.project.state === 'running') {
		return '运行中'
	}
	if (props.project.state === 'error') {
		return '异常'
	}
	return '已停止'
})
const stateColor = computed(() => {
	if (props.project.state === 'starting') {
		return 'info'
	}
	if (props.project.state === 'running') {
		return 'success'
	}
	if (props.project.state === 'error') {
		return 'danger'
	}
	return 'default'
})
const needLabels: Record<string, string> = {
	php: 'PHP',
	nodejs: 'Node.js',
	python: 'Python',
	go: 'Go',
	java: 'Java',
	rust: 'Rust/cargo',
}
function confirmDelete(event: MouseEvent): void {
	confirm.require({
		target: event.currentTarget as HTMLElement,
		message: '只删除项目记录，不会删除项目目录',
		icon: 'pi pi-exclamation-triangle',
		acceptLabel: '删除',
		rejectLabel: '取消',
		acceptProps: { severity: 'danger', size: 'small' },
		rejectProps: { severity: 'secondary', variant: 'outlined', size: 'small' },
		accept: () => emit('delete', props.project),
	})
}
</script>

<style scoped lang="scss">
.project-row {
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
	border-radius: var(--h-radius);
	display: flex;
	align-items: center;
	justify-content: center;
	font-size: var(--icon-lg);
	flex-shrink: 0;
	background: var(--bg-app);
	color: var(--text-secondary);
	border: 1px solid var(--border-color);
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
	&.miniprogram {
		background: var(--warn-bg);
		color: var(--warn);
		border: 1px solid var(--warn-border);
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
	border: 1px solid transparent;
	background: var(--bg-app);
	padding: 0.15rem 0.5rem;
	border-radius: var(--h-radius);
	color: var(--info);
	font-size: var(--font-xs);
	cursor: pointer;
	&:hover {
		border-color: var(--border-hover);
	}
	.link-icon {
		font-size: var(--icon-xs);
	}
}

.meta-line {
	display: flex;
	align-items: center;
	gap: 0.625rem;
	flex-wrap: wrap;
}
.need-box {
	display: inline-flex; align-items: center; gap: 0.5rem;
}
.path-text {
	font-size: var(--font-xs);
	color: var(--text-muted);
	overflow: hidden;
	text-overflow: ellipsis;
	white-space: nowrap;
}
.traffic-line {
	display: flex;
	align-items: center;
	gap: 0.4rem;
	color: var(--text-muted);
	font-size: var(--font-xs);
}
.actions {
	display: flex; align-items: center;
	gap: 0.45rem;
	flex-shrink: 0;
}
</style>
