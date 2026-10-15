<template>
	<!-- PHP 配置面板 -->
	<div class="config-panel">
		<!-- 顶部页签栏 -->
		<div class="panel-tabs">
			<button
				v-for="tab in tabs"
				:key="tab.value"
				type="button"
				class="panel-tab"
				:class="{ active: activeTab === tab.value }"
				@click="activeTab = tab.value"
			>
				{{ tab.label }}
			</button>
		</div>

		<!-- 概览页签 -->
		<comp-config-overview
			v-if="activeTab === 'overview'"
			:item="item"
			:busy="false"
			@main="emit('main', $event)"
			@folder="emit('folder', $event)"
		/>

		<!-- 常用配置页签 -->
		<comp-config-items
			v-else-if="activeTab === 'items'"
			:item="item"
		/>

		<!-- PHP 扩展开关页签 -->
		<div v-else-if="activeTab === 'extensions'" class="extension-grid">
			<div
				v-for="entry in extensions"
				:key="entry.name"
				class="extension-row"
			>
				<span class="extension-name font-mono">{{ entry.name }}</span>
				<p-toggle-switch
					v-model="entry.enabled"
					@change="onToggle(entry.name, entry.enabled)"
				/>
			</div>
		</div>

		<!-- php.ini 全文编辑页签 -->
		<comp-config-file
			v-else-if="activeTab === 'file'"
			:software-id="item.id"
		/>

		<!-- 运行日志页签 -->
		<comp-config-log
			v-else
			:software-id="item.id"
		/>
	</div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { useConfig } from '../use-config'
import CompConfigOverview from './config-overview.vue'
import CompConfigItems from './config-items.vue'
import CompConfigFile from './config-file.vue'
import CompConfigLog from './config-log.vue'
import type { Software } from '@/types'

// 接收 PHP 软件属性
const props = defineProps<{
	item: Software
}>()

// 定义操作事件
const emit = defineEmits<{
	main: [item: Software]
	folder: [id: number]
}>()

// 扩展管理状态与方法
const { extensions, fetchExtensions, toggleExtension } = useConfig()

// PHP 配置页签列表
const tabs = [
	{ value: 'overview', label: '概览' },
	{ value: 'items', label: '常用配置' },
	{ value: 'extensions', label: '扩展管理' },
	{ value: 'file', label: 'php.ini' },
	{ value: 'log', label: '日志' },
]

// 当前激活页签
const activeTab = ref('overview')

// 切换指定 PHP 扩展启用状态
async function onToggle(name: string, enabled: boolean): Promise<void> {
	await toggleExtension(props.item.id, name, enabled)
}

// 切换到扩展管理页签时加载扩展列表
watch(
	() => [props.item.id, activeTab.value] as const,
	([id, tab]) => {
		if (tab === 'extensions') {
			fetchExtensions(id)
		}
	},
	{ immediate: true },
)
</script>

<style scoped lang="scss">
/* 配置面板外框 */
.config-panel {
	display: flex;
	flex-direction: column;
	gap: 0.875rem;
}

/* 扩展双列网格 */
.extension-grid {
	display: grid;
	grid-template-columns: repeat(2, 1fr);
	gap: 0.625rem;
	max-height: 20rem;
	overflow-y: auto;
}

/* 单个扩展开关行 */
.extension-row {
	display: flex;
	align-items: center;
	justify-content: space-between;
	padding: 0.625rem 0.875rem;
	background: var(--bg-app);
	border: 1px solid var(--border-color);
	border-radius: var(--h-radius);
}

/* 扩展名称文本 */
.extension-name {
	font-size: var(--font-base);
	color: var(--text-primary);
}
</style>
