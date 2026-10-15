<template>
	<!-- 软件设置对话框 -->
	<h-modal
		v-if="item !== null"
		v-model:visible="visible"
		:title="`软件设置 · ${item.name} ${item.version}`"
		width="40rem"
		:modal="true"
		:closable="true"
		:mask-close="false"
		:footer="false"
		ok-text="确定"
		ok-icon="pi pi-check"
		cancel-text="取消"
		:loading="false"
	>
		<!-- PHP 配置面板 -->
		<comp-config-php
			v-if="item.kind === 'php'"
			:item="item"
			@main="emit('main', $event)"
			@folder="emit('folder', $event)"
		/>

		<!-- 守护进程服务配置面板 -->
		<comp-config-daemon
			v-else-if="item.daemon"
			:item="item"
			@start="emit('start', $event)"
			@stop="emit('stop', $event)"
			@main="emit('main', $event)"
			@autostart="(target, on) => emit('autostart', target, on)"
			@folder="emit('folder', $event)"
		/>

		<!-- 非守护进程软件概览面板 -->
		<comp-config-overview
			v-else
			:item="item"
			@main="emit('main', $event)"
			@folder="emit('folder', $event)"
		/>
	</h-modal>
</template>

<script setup lang="ts">
import CompConfigPhp from './config-php.vue'
import CompConfigDaemon from './config-daemon.vue'
import CompConfigOverview from './config-overview.vue'
import type { Software } from '@/types'

// 接收当前软件属性
defineProps<{
	item: Software | null
}>()

// 弹窗显隐双向绑定
const visible = defineModel<boolean>('visible', { required: true })

// 定义设置弹窗操作事件
const emit = defineEmits<{
	start: [item: Software]
	stop: [item: Software]
	main: [item: Software]
	autostart: [item: Software, on: boolean]
	folder: [id: number]
}>()
</script>
