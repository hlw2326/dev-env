<template>
	<!-- 软件概览信息面板 -->
	<div class="config-overview">
		<!-- 守护进程运行状态与启停 -->
		<div v-if="item.daemon" class="kv-row">
			<span class="kv-label">运行状态</span>
			<div class="row-actions">
				<span v-if="item.running" class="state-text ok">
					运行中 (PID {{ item.pid }})
				</span>
				<span v-else class="state-text">已停止</span>
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

		<!-- 守护进程监听端口 -->
		<div v-if="item.daemon" class="kv-row">
			<span class="kv-label">监听端口</span>
			<span class="kv-val font-mono">127.0.0.1:{{ item.port }}</span>
		</div>

		<!-- 软件版本信息 -->
		<div class="kv-row">
			<span class="kv-label">软件版本</span>
			<span class="kv-val font-mono">{{ item.name }} {{ item.version }}</span>
		</div>

		<!-- 安装或数据目录 -->
		<div class="kv-row">
			<span class="kv-label">{{ folderLabel }}</span>
			<div class="row-actions">
				<span class="kv-val font-mono" :title="item.path">{{ item.path }}</span>
				<p-button
					size="small"
					severity="secondary"
					variant="outlined"
					icon="pi pi-folder-open"
					label="打开"
					@click="emit('folder', item.id)"
				/>
			</div>
		</div>

		<!-- 默认版本设置 -->
		<div class="kv-row">
			<span class="kv-label">默认版本</span>
			<div class="row-actions">
				<span v-if="item.main" class="state-text ok">当前已设为默认版本</span>
				<p-button
					v-else
					size="small"
					label="设为默认版本"
					@click="emit('main', item)"
				/>
			</div>
		</div>

		<!-- 守护进程随应用启动开关 -->
		<div v-if="item.daemon" class="kv-row">
			<span class="kv-label">随应用启动</span>
			<p-toggle-switch
				v-model="autoFlag"
				@change="onAutostart"
			/>
		</div>
	</div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import type { Software } from '@/types'

// 接收软件条目属性
const props = defineProps<{
	item: Software
}>()

// 定义概览面板事件
const emit = defineEmits<{
	start: [item: Software]
	stop: [item: Software]
	main: [item: Software]
	autostart: [item: Software, on: boolean]
	folder: [id: number]
}>()

// 随应用启动本地状态
const autoFlag = ref(false)

// 包含独立数据目录的数据库类型集合
const DATA_KINDS = ['mysql', 'mariadb', 'postgresql', 'mongodb']

// 目录行标签名称
const folderLabel = computed(() => {
	if (DATA_KINDS.includes(props.item.kind)) {
		return '软件与数据目录'
	}
	return '安装路径'
})

// 切换随应用自动启动
function onAutostart(): void {
	emit('autostart', props.item, autoFlag.value)
}

// 同步软件随应用启动状态
watch(
	() => props.item.autostart,
	(value) => {
		autoFlag.value = value
	},
	{ immediate: true },
)
</script>

<style scoped lang="scss">
/* 概览列表容器 */
.config-overview {
	display: flex;
	flex-direction: column;
	gap: 0.625rem;
}

/* 键值对单行 */
.kv-row {
	display: flex;
	align-items: center;
	justify-content: space-between;
	padding: 0.75rem 0.875rem;
	background: var(--bg-app);
	border: 1px solid var(--border-color);
	border-radius: var(--h-radius);
	gap: 1rem;
}

/* 左侧字段名 */
.kv-label {
	font-size: var(--font-sm);
	color: var(--text-secondary);
	flex-shrink: 0;
}

/* 右侧字段值 */
.kv-val {
	font-size: var(--font-xs);
	color: var(--text-primary);
}

/* 右侧操作组合 */
.row-actions {
	display: flex;
	align-items: center;
	justify-content: flex-end;
	gap: 0.625rem;
	min-width: 0;
	flex: 1;

	/* 长内容限制在剩余空间内 */
	.kv-val {
		min-width: 0;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}
}

/* 状态文本 */
.state-text {
	font-size: var(--font-xs);
	color: var(--text-secondary);

	/* 正常色状态 */
	&.ok {
		color: var(--ok);
	}
}
</style>
