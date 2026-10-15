<template>
	<!-- 通用模态对话框 -->
	<Dialog
		v-model:visible="visible"
		:modal="modal !== false"
		:header="title"
		:closable="closable !== false"
		:close-on-escape="true"
		:dismissable-mask="Boolean(maskClose)"
		:style="modalStyle"
		class="h-modal"
	>
		<!-- 自定义头部插槽 -->
		<template v-if="$slots.header" #header>
			<slot name="header"></slot>
		</template>

		<!-- 弹窗主体内容 -->
		<slot></slot>

		<!-- 弹窗底部操作区 -->
		<template v-if="footer !== false" #footer>
			<slot name="footer">
				<Button
					:label="cancelLabel"
					severity="secondary"
					variant="outlined"
					:disabled="loading"
					@click="onCancel"
				/>
				<Button
					:label="okLabel"
					:icon="confirmIcon"
					:loading="loading"
					@click="emit('ok')"
				/>
			</slot>
		</template>
	</Dialog>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import Dialog from 'primevue/dialog'
import Button from 'primevue/button'
import type { HModalProps, HModalEmits, HModalExpose } from './types'

// 全局弹窗组件名称
defineOptions({
	name: 'HModal',
})

// 接收弹窗属性
const props = defineProps<HModalProps>()

// 弹窗显隐双向绑定
const visible = defineModel<boolean>('visible', { default: false })

// 定义弹窗事件
const emit = defineEmits<HModalEmits>()

// 计算弹窗宽度样式
const modalStyle = computed(() => {
	if (props.width) {
		return { width: props.width }
	}
	return { width: '26rem' }
})

// 取消按钮文案
const cancelLabel = computed(() => {
	if (props.cancelText) {
		return props.cancelText
	}
	return '取消'
})

// 确认按钮文案
const okLabel = computed(() => {
	if (props.okText) {
		return props.okText
	}
	return '确定'
})

// 确认按钮图标
const confirmIcon = computed(() => {
	if (props.okIcon) {
		return props.okIcon
	}
	return 'pi pi-check'
})

// 关闭弹窗并触发取消事件
function onCancel(): void {
	visible.value = false
	emit('cancel')
}

// 暴露弹窗控制方法
defineExpose<HModalExpose>({
	open: () => {
		visible.value = true
	},
	close: () => {
		visible.value = false
	},
})
</script>
