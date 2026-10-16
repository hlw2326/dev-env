<template>
	<!-- 项目详情右侧抽屉 -->
	<hlw-drawer
		v-if="project !== null"
		v-model:visible="visible"
		:title="`项目详情 · ${project.name}`"
		position="right"
		width="80%"
		:modal="true"
		:closable="true"
		:mask-close="true"
		:footer="showFooter"
		ok-text="保存修改"
		ok-icon="pi pi-check"
		cancel-text="取消"
		:loading="saving"
		@ok="saveCurrentTab"
	>
		<template #top>
			<comp-detail-tabs v-model="currentTab" :tabs="detailTabs" />
		</template>

		<!-- 概览设置页签 -->
		<comp-detail-info
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

		<comp-detail-git
			v-else-if="currentTab === 'git'"
			:project="project"
			@refresh="emit('refresh-git')"
		/>

		<comp-detail-plugin
			v-else-if="currentPlugin !== null"
			:project="project"
			:plugin="currentPlugin"
			:busy="pluginBusy === currentPlugin.id"
			@run="emit('plugin', $event)"
		/>

		<comp-detail-deployment v-else-if="currentTab === 'deployment'" :project="project" />
	</hlw-drawer>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import HlwDrawer from '@/components/hlw-drawer'
import CompDetailTabs from './comps/detail-tabs.vue'
import CompDetailInfo from './comps/detail-info.vue'
import CompDetailDomains from './comps/detail-domains.vue'
import CompDetailLogs from './comps/detail-logs.vue'
import CompDetailNginx from './comps/detail-nginx.vue'
import CompDetailPlugin from './comps/detail-plugin.vue'
import CompDetailGit from './comps/detail-git.vue'
import CompDetailDeployment from './comps/detail-deployment.vue'
import type { ProjectPluginActionInput, ProjectPlugin, ProjectTemplate, Project, Software, ProjectUpdateInput } from '@/types'

// 接收项目与软件列表属性
const props = withDefaults(defineProps<{
	project: Project | null
	softwareList: Software[]
	plugins: ProjectPlugin[]
	pluginBusy: string
	templates: ProjectTemplate[]
	initialTab?: string
	saving: boolean
}>(), {
	initialTab: 'overview',
	saving: false,
})

// 抽屉显隐双向绑定
const visible = defineModel<boolean>('visible', { required: true })

// 保存项目设置事件
const emit = defineEmits<{
	save: [input: ProjectUpdateInput]
	plugin: [input: ProjectPluginActionInput]
	'refresh-git': []
}>()

// 详情抽屉页签配置
const detailTabs = computed(() => {
	const tabs = [
		{ value: 'overview', label: '概览' },
		{ value: 'logs', label: '日志' },
		{ value: 'git', label: 'Git' },
	]
	if (props.project !== null && hasCapability('domain')) {
		tabs.splice(1, 0, { value: 'domains', label: '域名' })
		tabs.splice(2, 0, { value: 'nginx', label: 'nginx' })
	}
	for (const plugin of props.plugins) {
		tabs.push({ value: `plugin:${plugin.id}`, label: plugin.name })
	}
	if (hasCapability('server-sync')) tabs.push({ value: 'deployment', label: '服务器同步' })
	if (hasCapability('wechat-upload')) tabs.push({ value: 'deployment', label: '微信发布' })
	return tabs
})

// 详情页脚区域开关
const showFooter = computed(() => ['overview', 'nginx'].includes(currentTab.value))

// 当前激活页签
const currentTab = ref('overview')

watch(visible, (value) => { if (value) currentTab.value = props.initialTab })

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
const overview = ref<InstanceType<typeof CompDetailInfo> | null>(null)
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
	() => props.project?.id,
	() => {
		currentTab.value = props.initialTab
	},
)

watch(
	() => props.initialTab,
	(value) => {
		currentTab.value = value
	},
)

// 通过抽屉底部按钮保存当前设置
function saveCurrentTab(): void {
	if (props.saving) return
	if (currentTab.value === 'overview' && overview.value !== null) {
		overview.value.save()
	} else if (currentTab.value === 'nginx' && nginxEditor.value !== null) {
		nginxEditor.value.save()
	}
}
</script>

