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
		:footer="false"
		ok-text="确定"
		ok-icon="pi pi-check"
		cancel-text="取消"
		:loading="false"
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
	{ value: 'logs', label: '日志' },
]

// 当前激活页签
const currentTab = ref('overview')
</script>

<style scoped lang="scss">
/* 详情页签下边距 */
.detail-tabs {
	margin-bottom: 1rem;
}
</style>
