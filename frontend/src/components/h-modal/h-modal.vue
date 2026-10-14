<template>
	<Dialog
		v-model:visible="visible"
		:modal="modal !== false"
		:header="title"
		:closable="closable !== false"
		:close-on-escape="closeOnEscape !== false"
		:dismissable-mask="Boolean(dismissableMask)"
		:style="modalStyle"
		class="h-modal"
	>
		<!-- 自定义头部 -->
		<template v-if="$slots.header" #header>
			<slot name="header"></slot>
		</template>

		<!-- 弹窗主体 -->
		<slot></slot>

		<!-- 弹窗底部按钮 -->
		<template v-if="footer !== false" #footer>
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
import type { HModalProps, HModalEmits, HModalInst } from './types'

// 全局弹窗组件名称
defineOptions({
	name: 'HModal',
})

// 接收属性与默认值
const props = withDefaults(defineProps<HModalProps>(), {
	modal: true,
	closable: true,
	closeOnEscape: true,
	footer: true,
	cancelText: '取消',
	okText: '确定',
	okIcon: 'pi pi-check',
})

// 弹窗显隐双向绑定
const visible = defineModel<boolean>('visible', { default: false })

// 触发事件
const emit = defineEmits<HModalEmits>()

// 计算弹窗宽度样式
const modalStyle = computed(() => {
	if (!props.width) {
		return { width: '26rem' }
	}
	if (typeof props.width === 'number') {
		return { width: `${props.width}px` }
	}
	return { width: props.width }
})

// 关闭弹窗
function onCancel(): void {
	visible.value = false
	emit('cancel')
}

// 暴露实例方法
defineExpose<HModalInst>({
	open: () => { visible.value = true },
	close: () => { visible.value = false },
})
</script>
