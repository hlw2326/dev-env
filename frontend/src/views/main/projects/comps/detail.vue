<template>
	<!-- 项目详情模态框 -->
	<h-modal
		v-if="project !== null"
		v-model:visible="visible"
		:title="`项目详情 · ${project.name}`"
		width="52rem"
		:modal="true"
		:closable="true"
		:mask-close="false"
		:footer="showFooter"
		ok-text="保存修改"
		ok-icon="pi pi-check"
		cancel-text="取消"
		:loading="false"
		@ok="saveCurrentTab"
	>
		<!-- 详情页签切换栏 -->
		<div class="panel-tabs detail-tabs">
			<button
				v-for="tab in detailTabs"
				:key="tab.value"
				type="button"
				class="panel-tab"
				:class="{ active: currentTab === tab.value }"
				@click="currentTab = tab.value"
			>
				{{ tab.label }}
			</button>
		</div>

		<!-- 概览设置页签 -->
		<comp-detail-overview
			v-if="currentTab === 'overview'"
			ref="overview"
			:project="project"
			:software-list="softwareList"
			:template="currentTemplate"
			@save="emit('save', $event)"
		/>

		<!-- 域名管理页签 -->
		<comp-detail-domains
			v-else-if="currentTab === 'domains'"
			:project="project"
			@save="emit('save', $event)"
		/>

		<comp-detail-nginx
			v-else-if="currentTab === 'nginx'"
			ref="nginxEditor"
			:project="project"
			@save="emit('save', $event)"
		/>

		<comp-detail-logs
			v-else-if="currentTab === 'logs'"
			:project="project"
			:template="currentTemplate"
		/>

		<comp-detail-plugin
			v-else-if="currentPlugin !== null"
			:project="project"
			:plugin="currentPlugin"
			:busy="pluginBusy === currentPlugin.id"
			@run="emit('plugin', $event)"
		/>
	</h-modal>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import CompDetailOverview from './detail-overview.vue'
import CompDetailDomains from './detail-domains.vue'
import CompDetailLogs from './detail-logs.vue'
import CompDetailNginx from './detail-nginx.vue'
import CompDetailPlugin from './detail-plugin.vue'
import type { ProjectPluginActionInput, ProjectPlugin, ProjectTemplate, Project, Software, ProjectUpdateInput } from '@/types'

// 接收项目与软件列表属性
const props = defineProps<{
	project: Project | null
	softwareList: Software[]
	plugins: ProjectPlugin[]
	pluginBusy: string
	templates: ProjectTemplate[]
}>()

// 弹窗显隐双向绑定
const visible = defineModel<boolean>('visible', { required: true })

// 保存项目设置事件
const emit = defineEmits<{
	save: [input: ProjectUpdateInput]
	plugin: [input: ProjectPluginActionInput]
}>()

// 详情弹窗页签配置
const detailTabs = computed(() => {
	const tabs = [
		{ value: 'overview', label: '概览' },
		{ value: 'logs', label: '日志' },
	]
	if (props.project !== null && hasCapability('domain')) {
		tabs.splice(1, 0, { value: 'domains', label: '域名' })
		tabs.splice(2, 0, { value: 'nginx', label: 'nginx' })
	}
	for (const plugin of props.plugins) {
		tabs.push({ value: `plugin:${plugin.id}`, label: plugin.name })
	}
	return tabs
})

// 详情页脚区域开关
const showFooter = computed(() => ['overview', 'nginx'].includes(currentTab.value))

// 当前激活页签
const currentTab = ref('overview')

const currentTemplate = computed<ProjectTemplate | null>(() => {
	if (props.project === null) {
		return null
	}
	for (const item of props.templates) {
		if (item.kind === props.project.kind) {
			return item
		}
	}
	return null
})

function hasCapability(capability: string): boolean {
	const template = currentTemplate.value
	return template !== null && template.capabilities.includes(capability)
}

// 概览表单组件实例
const overview = ref<InstanceType<typeof CompDetailOverview> | null>(null)
const nginxEditor = ref<InstanceType<typeof CompDetailNginx> | null>(null)

// 当前插件页签对应的插件
const currentPlugin = computed<ProjectPlugin | null>(() => {
	const prefix = 'plugin:'
	if (!currentTab.value.startsWith(prefix)) {
		return null
	}
	const pluginID = currentTab.value.slice(prefix.length)
	for (const item of props.plugins) {
		if (item.id === pluginID) {
			return item
		}
	}
	return null
})

watch(
	() => props.project,
	() => {
		currentTab.value = 'overview'
	},
)

// 通过弹窗底部按钮保存概览设置
function saveCurrentTab(): void {
	if (currentTab.value === 'overview' && overview.value !== null) {
		overview.value.save()
	} else if (currentTab.value === 'nginx' && nginxEditor.value !== null) {
		nginxEditor.value.save()
	}
}
</script>

<style scoped lang="scss">
/* 详情页签下边距 */
.detail-tabs {
	margin-bottom: 1rem;
}
</style>
