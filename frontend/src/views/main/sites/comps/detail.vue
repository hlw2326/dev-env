<template>
	<!-- 站点详情模态框 -->
	<h-modal
		v-if="site !== null"
		v-model:visible="visible"
		:title="`站点详情 · ${site.name}`"
		width="52rem"
		:modal="true"
		:closable="true"
		:mask-close="false"
		:footer="currentTab === 'overview' || currentTab === 'nginx'"
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
			:site="site"
			:software-list="softwareList"
			@save="emit('save', $event)"
		/>

		<!-- 域名管理页签 -->
		<comp-detail-domains
			v-else-if="currentTab === 'domains'"
			:site="site"
			@save="emit('save', $event)"
		/>

		<comp-detail-nginx
			v-else-if="currentTab === 'nginx'"
			ref="nginxEditor"
			:site="site"
			@save="emit('save', $event)"
		/>

		<comp-detail-logs
			v-else-if="currentTab === 'logs'"
			:site="site"
		/>
	</h-modal>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import CompDetailOverview from './detail-overview.vue'
import CompDetailDomains from './detail-domains.vue'
import CompDetailLogs from './detail-logs.vue'
import CompDetailNginx from './detail-nginx.vue'
import type { Site, Software, UpdateInput } from '@/types'

// 接收站点与软件列表属性
defineProps<{
	site: Site | null
	softwareList: Software[]
}>()

// 弹窗显隐双向绑定
const visible = defineModel<boolean>('visible', { required: true })

// 保存站点设置事件
const emit = defineEmits<{
	save: [input: UpdateInput]
}>()

// 详情弹窗页签配置
const detailTabs = [
	{ value: 'overview', label: '概览' },
	{ value: 'domains', label: '域名' },
	{ value: 'nginx', label: 'Nginx 高级配置' },
	{ value: 'logs', label: '日志' },
]

// 当前激活页签
const currentTab = ref('overview')

// 概览表单组件实例
const overview = ref<InstanceType<typeof CompDetailOverview> | null>(null)
const nginxEditor = ref<InstanceType<typeof CompDetailNginx> | null>(null)

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
