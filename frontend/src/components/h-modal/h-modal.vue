<template>
  <Dialog
    v-model:visible="visible"
    :modal="modal"
    :header="title"
    :closable="closable"
    :close-on-escape="closeOnEscape"
    :dismissable-mask="dismissableMask"
    :style="modalStyle"
    class="h-modal"
  >
    <template v-if="$slots.header" #header>
      <slot name="header"></slot>
    </template>

    <slot></slot>

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
          :icon="okIcon || undefined"
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

defineOptions({
  name: 'HModal'
})

const visible = defineModel<boolean>({ default: false })

const props = withDefaults(defineProps<HModalProps>(), {
  title: '',
  width: '26rem',
  modal: true,
  closable: true,
  closeOnEscape: true,
  dismissableMask: false,
  footer: true,
  okText: '保存',
  okIcon: '',
  cancelText: '取消',
  loading: false
})

const emit = defineEmits<HModalEmits>()

const modalStyle = computed(() => {
  if (!props.width) return {}
  return { width: typeof props.width === 'number' ? `${props.width}px` : props.width }
})

function onCancel(): void {
  visible.value = false
  emit('cancel')
}

defineExpose<HModalInst>({
  open: () => { visible.value = true },
  close: () => { visible.value = false }
})
</script>
