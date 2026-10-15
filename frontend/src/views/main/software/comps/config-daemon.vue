<template>
	<!-- 守护进程服务配置面板 -->
	<div class="config-panel">
		<!-- 顶部页签栏 -->
		<div class="panel-tabs">
			<p-button
				v-for="tab in tabs"
				:key="tab.value"
				type="button"
				unstyled
				class="panel-tab"
				:class="{ active: activeTab === tab.value }"
				@click="activeTab = tab.value"
			>
				{{ tab.label }}
			</p-button>
		</div>

		<!-- 概览页签 -->
		<comp-config-overview
			v-if="activeTab === 'overview'"
			:item="item"
			:busy="busy"
			@start="emit('start', $event)"
			@stop="emit('stop', $event)"
			@main="emit('main', $event)"
			@autostart="(target, on) => emit('autostart', target, on)"
			@folder="emit('folder', $event)"
		/>

		<!-- 常用配置页签 -->
		<comp-config-items
			v-else-if="activeTab === 'items'"
			:item="item"
		/>

		<!-- 配置文件全文编辑页签 -->
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
import { computed, ref } from 'vue'
import CompConfigOverview from './config-overview.vue'
import CompConfigItems from './config-items.vue'
import CompConfigFile from './config-file.vue'
import CompConfigLog from './config-log.vue'
import type { Software } from '@/types'

// 接收守护进程软件属性
const props = defineProps<{
	item: Software
	busy: boolean
}>()

// 定义操作事件
const emit = defineEmits<{
	start: [item: Software]
	stop: [item: Software]
	main: [item: Software]
	autostart: [item: Software, on: boolean]
	folder: [id: number]
}>()

// 配置文件名称查表
const FILE_NAMES: Record<string, string> = {
	mysql: 'my.ini',
	mariadb: 'my.ini',
	postgresql: 'postgresql.conf',
	mongodb: 'mongod.conf',
	redis: 'redis.conf',
	memcached: 'memcached.conf',
}

// 守护进程配置页签列表
const tabs = computed(() => {
	let fileName = '配置文件'
	const matched = FILE_NAMES[props.item.kind]
	if (matched) {
		fileName = matched
	}
	return [
		{ value: 'overview', label: '概览' },
		{ value: 'items', label: '常用配置' },
		{ value: 'file', label: fileName },
		{ value: 'log', label: '日志' },
	]
})

// 当前激活页签
const activeTab = ref('overview')
</script>

<style scoped lang="scss">
/* 配置面板外框 */
.config-panel {
	display: flex;
	flex-direction: column;
	gap: 0.875rem;
}
</style>
