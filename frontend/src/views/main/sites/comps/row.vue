<template>
	<!-- 单个网站列表行 -->
	<div class="site-row surface-card">
		<!-- 左侧图标与站点信息 -->
		<div class="left">
			<div class="kind-badge" :class="site.kind">
				<i :class="kindIcons[site.kind]"></i>
			</div>
			<div class="info">
				<div class="top-line">
					<span class="name">{{ site.name }}</span>
					<button
						type="button"
						class="domain-link font-mono"
						@click="emit('open', site)"
					>
						<span>{{ site.hosts[0] }}</span>
						<i class="pi pi-external-link link-icon"></i>
					</button>
					<span class="status-pill" :class="site.state">
						<span class="dot"></span>
						<span>{{ stateLabels[site.state] }}</span>
					</span>
				</div>
				<div class="meta-line">
					<span class="tag">{{ kindLabels[site.kind] }}</span>
					<span v-if="site.version !== ''" class="tag font-mono">{{ site.version }}</span>
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

		<!-- 右侧站点操作按钮组 -->
		<div class="actions">
			<p-button
				v-if="site.state === 'running'"
				size="small"
				severity="danger"
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
				@click="confirmDelete($event)"
			/>
		</div>
	</div>
</template>

<script setup lang="ts">
import { useConfirm } from 'primevue/useconfirm'
import type { Site } from '@/types'

// 接收站点属性
const props = defineProps<{
	site: Site
	installingKind: string
}>()

// 定义行操作事件
const emit = defineEmits<{
	toggle: [site: Site]
	open: [site: Site]
	folder: [path: string]
	detail: [site: Site]
	delete: [site: Site]
	install: [kind: string]
}>()

// 确认弹窗服务
const confirm = useConfirm()

// 站点类型名称映射
const kindLabels: Record<string, string> = {
	php: 'PHP',
	node: 'Node.js',
	static: '静态',
}

// 站点类型图标映射
const kindIcons: Record<string, string> = {
	php: 'pi pi-code',
	node: 'pi pi-bolt',
	static: 'pi pi-file',
}

// 站点状态名称映射
const stateLabels: Record<string, string> = {
	running: '运行中',
	stopped: '已停止',
	error: '异常',
}

// 缺失环境名称映射
const needLabels: Record<string, string> = {
	php: 'PHP',
	nodejs: 'Node.js',
}

// 弹出删除站点确认框
function confirmDelete(event: MouseEvent): void {
	confirm.require({
		target: event.currentTarget as HTMLElement,
		message: '只删除站点记录，不会删除项目目录',
		icon: 'pi pi-exclamation-triangle',
		acceptLabel: '删除',
		rejectLabel: '取消',
		acceptProps: { severity: 'danger', size: 'small' },
		rejectProps: { severity: 'secondary', variant: 'outlined', size: 'small' },
		accept: () => emit('delete', props.site),
	})
}
</script>

<style scoped lang="scss">
/* 站点行外框 */
.site-row {
	display: flex;
	align-items: center;
	justify-content: space-between;
	padding: 0.875rem 1.125rem;
	gap: 1rem;
}
/* 左侧信息区 */
.left {
	display: flex;
	align-items: center;
	gap: 0.875rem;
	min-width: 0;
	flex: 1;
}
/* 类型图标色块 */
.kind-badge {
	width: 2.5rem;
	height: 2.5rem;
	border-radius: var(--h-radius);
	display: flex;
	align-items: center;
	justify-content: center;
	font-size: var(--icon-lg);
	flex-shrink: 0;
	/* PHP 站点色块 */
	&.php {
		background: var(--indigo-bg);
		color: var(--indigo);
		border: 1px solid var(--indigo-border);
	}
	/* Node 站点色块 */
	&.node {
		background: var(--ok-bg);
		color: var(--ok);
		border: 1px solid var(--ok-border);
	}
	/* 静态站点色块 */
	&.static {
		background: var(--info-bg);
		color: var(--info);
		border: 1px solid var(--info-border);
	}
}
/* 文本垂直区 */
.info {
	display: flex;
	flex-direction: column;
	gap: 0.35rem;
	min-width: 0;
	flex: 1;
}
/* 顶部名称与域名行 */
.top-line {
	display: flex;
	align-items: center;
	gap: 0.75rem;
	flex-wrap: wrap;
}
/* 站点名称 */
.name {
	font-size: var(--font-lg);
	color: var(--text-primary);
}
/* 可点击域名链接 */
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
	/* 悬停态 */
	&:hover {
		border-color: var(--border-hover);
	}
	/* 外链图标 */
	.link-icon {
		font-size: var(--icon-xs);
	}
}
/* 底部元数据行 */
.meta-line {
	display: flex;
	align-items: center;
	gap: 0.625rem;
	flex-wrap: wrap;
}
/* 通用元数据小标签 */
.tag {
	font-size: var(--font-xs);
	color: var(--text-secondary);
	background: var(--bg-app);
	padding: 0.1rem 0.45rem;
	border-radius: var(--h-radius);
	border: 1px solid var(--border-color);
}
/* 缺少环境提示区 */
.need-box {
	display: inline-flex;
	align-items: center;
	gap: 0.5rem;
}
/* 缺少环境文字 */
.need-text {
	font-size: var(--font-xs);
	color: var(--warn);
}
/* 完整路径文字 */
.path-text {
	font-size: var(--font-xs);
	color: var(--text-muted);
	overflow: hidden;
	text-overflow: ellipsis;
	white-space: nowrap;
}
/* 右侧操作按钮组 */
.actions {
	display: flex;
	align-items: center;
	gap: 0.45rem;
	flex-shrink: 0;
}
</style>
