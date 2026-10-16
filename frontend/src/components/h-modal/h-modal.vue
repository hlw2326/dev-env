<template>
	<!-- 通用模态对话框 -->
	<Dialog
		v-model:visible="visible"
		:modal="modal"
		:header="title"
		:closable="closable && !loading"
		:close-on-escape="!loading"
		:dismissable-mask="maskClose"
		:style="{ width }"
		class="h-modal"
	>
		<!-- 自定义头部插槽 -->
		<template v-if="$slots.header" #header>
			<slot name="header"></slot>
		</template>

		<!-- 弹窗主体内容 -->
		<slot></slot>

		<!-- 弹窗底部操作区 -->
		<template v-if="footer" #footer>
			<slot name="footer">
				<Button
					:label="cancelText"
					severity="secondary"
					variant="outlined"
					:disabled="loading"
					@click="onCancel"
				/>
				<Button
					:label="okText"
					:icon="okIcon"
					:loading="loading"
					@click="confirm"
				/>
			</slot>
		</template>
	</Dialog>
</template>

<script setup lang="ts">
import Dialog from 'primevue/dialog'
import Button from 'primevue/button'
import type { HModalProps, HModalEmits, HModalExpose } from './types'

// 全局弹窗组件名称
defineOptions({
	name: 'HModal',
})

// 接收完整弹窗属性
const { title, width, modal, closable, maskClose, footer, okText, okIcon, cancelText, loading } = defineProps<HModalProps>()

// 弹窗显隐双向绑定
const visible = defineModel<boolean>('visible', { required: true })

// 定义弹窗事件
const emit = defineEmits<HModalEmits>()

// 触发确认事件
function confirm(): void {
	emit('ok')
}

// 关闭弹窗并触发取消事件
function onCancel(): void {
	if (loading) {
		return
	}
	visible.value = false
	emit('cancel')
}

// 打开弹窗
function open(): void {
	visible.value = true
}

// 关闭弹窗
function close(): void {
	if (loading) {
		return
	}
	visible.value = false
}

// 暴露弹窗控制方法
defineExpose<HModalExpose>({
	open,
	close,
})
</script>

<style scoped lang="scss">
/* 弹窗正文由内容组件自行控制内边距 */
:deep(.p-dialog-content) {
	padding: 0 !important;
}
</style>
