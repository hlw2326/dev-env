<template>
  <div class="h-card" :class="{ fill }">
    <div v-if="title || $slots.title || $slots.actions" class="header">
      <div class="header-left">
        <slot name="title">
          <span class="title">{{ title }}</span>
        </slot>
      </div>
      <div v-if="$slots.actions" class="header-actions">
        <slot name="actions"></slot>
      </div>
    </div>
    <div class="body" :class="{ 'no-padding': noPadding }">
      <slot></slot>
    </div>
  </div>
</template>

<script setup lang="ts">
defineOptions({
  name: 'HCard'
})

interface Props {
  title?: string
  fill?: boolean
  noPadding?: boolean
}

withDefaults(defineProps<Props>(), {
  title: '',
  fill: false,
  noPadding: false
})
</script>

<style scoped lang="scss">
.h-card {
  background-color: var(--bg-card);
  border: 1px solid var(--border-color);
  border-radius: var(--h-radius);
  overflow: hidden;
  display: flex;
  flex-direction: column;
  
  &.fill {
    flex: 1;
    min-height: 0;
  }
  
  .header {
    height: 3.25rem;
    padding: 0 1.25rem;
    border-bottom: 1px dashed var(--border-color);
    display: flex;
    align-items: center;
    justify-content: space-between;
    
    .title {
      font-size: var(--font-md);
      font-weight: 400;
      color: var(--text-primary);
    }
  }
  
  .body {
    flex: 1;
    min-height: 0;
    display: flex;
    flex-direction: column;
    padding: var(--h-padding);

    &.no-padding {
      padding: 0;
    }
  }
}
</style>
