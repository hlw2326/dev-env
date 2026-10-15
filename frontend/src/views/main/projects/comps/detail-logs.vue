<template>
	<!-- 项目详情日志查看页签 -->
	<div class="detail-logs">
		<!-- 日志类型切换与刷新栏 -->
		<div class="log-toolbar">
			<div class="log-tabs">
				<button
					v-if="hasAccessLog"
					type="button"
					class="tab-btn"
					:class="{ active: logKind === 'access' }"
					@click="switchKind('access')"
				>
					访问日志
				</button>
				<button
					v-if="hasRunLog"
					type="button"
					class="tab-btn"
					:class="{ active: logKind === 'run' }"
					@click="switchKind('run')"
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

		<!-- 日志内容输出框 -->
		<div class="log-box font-mono">
			<div v-if="lines.length === 0" class="empty-log">暂无日志记录</div>
			<div v-for="(line, index) in lines" :key="index" class="log-line">
				{{ line }}
			</div>
		</div>
	</div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { readLogs } from '@/api'
import { toast, formatError } from '@/utils'
import type { ProjectTemplate, Project } from '@/types'

// 接收项目属性
const props = defineProps<{
	project: Project
	template: ProjectTemplate | null
}>()

// 单次读取最大日志行数
const LOG_LIMIT = 200

// 当前查看的日志类型
const logKind = ref('access')

const hasAccessLog = computed(() => {
	return props.template !== null && props.template.capabilities.includes('access-log')
})

const hasRunLog = computed(() => {
	return props.template !== null && props.template.capabilities.includes('command')
})

// 日志文本行列表
const lines = ref<string[]>([])

// 拉取项目日志末尾行
async function fetchLogs(): Promise<void> {
	try {
		lines.value = await readLogs(props.project.id, logKind.value, LOG_LIMIT)
	} catch (err: unknown) {
		toast.error('读取日志失败', formatError(err))
	}
}

// 切换访问或运行日志
function switchKind(kind: string): void {
	logKind.value = kind
	fetchLogs()
}

// 监听项目切换自动加载访问日志
watch(
	() => props.project.id,
	() => {
		if (hasRunLog.value) {
			logKind.value = 'run'
		} else {
			logKind.value = 'access'
		}
		fetchLogs()
	},
	{ immediate: true },
)
</script>

<style scoped lang="scss">
/* 日志面板容器 */
.detail-logs {
	display: flex;
	flex-direction: column;
	gap: 0.75rem;
}

/* 日志顶部工具栏 */
.log-toolbar {
	display: flex;
	align-items: center;
	justify-content: space-between;
}

/* 日志类型页签组 */
.log-tabs {
	display: flex;
	gap: 0.5rem;
}

/* 页签切换按钮 */
.tab-btn {
	border: 1px solid var(--border-color);
	background: var(--bg-app);
	color: var(--text-secondary);
	font-size: var(--font-sm);
	padding: 0.35rem 0.75rem;
	border-radius: var(--h-radius);
	cursor: pointer;

	/* 激活页签态 */
	&.active {
		background: var(--p-primary-color);
		color: var(--on-primary);
		border-color: var(--p-primary-color);
	}
}

/* 日志输出黑框 */
.log-box {
	background: var(--code-bg);
	color: var(--code-text);
	border-radius: var(--h-radius);
	padding: 0.875rem;
	height: 20rem;
	overflow-y: auto;
	font-size: var(--font-xs);
	display: flex;
	flex-direction: column;
	gap: 0.25rem;
}

/* 空日志提示 */
.empty-log {
	color: var(--text-muted);
}

/* 单行日志文本 */
.log-line {
	white-space: pre-wrap;
	word-break: break-all;
}
</style>
