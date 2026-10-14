<template>
  <div class="h-page" :class="{ 'no-scroll': !scroll }">
    <!-- 统一页面头部 -->
    <div v-if="title || $slots.title || $slots.extra" class="h-page-header">
      <div class="header-left">
        <slot name="title">
          <span class="page-title">{{ title }}</span>
        </slot>
        <slot name="desc">
          <p v-if="desc" class="page-desc">{{ desc }}</p>
        </slot>
      </div>
      <div v-if="$slots.extra" class="header-extra">
        <slot name="extra"></slot>
      </div>
    </div>

    <!-- 页面主体内容区 -->
    <div class="h-page-body" :class="{ 'with-padding': padding }">
      <slot></slot>
    </div>
  </div>
</template>

<script setup lang="ts">
defineOptions({
  name: 'HPage'
})

interface Props {
  title?: string
  desc?: string
  padding?: boolean
  scroll?: boolean
}

withDefaults(defineProps<Props>(), {
  title: '',
  desc: '',
  padding: true,
  scroll: true
})
</script>

<style scoped lang="scss">
.h-page {
  display: flex;
  flex-direction: column;
  min-height: 100%;
  width: 100%;

  &.no-scroll {
    height: 100%;
    overflow: hidden;
  }
}

.h-page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 14px var(--h-padding);
  background: #ffffff;
  border: none;
  border-bottom: 1px solid var(--border-color);
  border-radius: 0;
  margin: 0;
  flex-shrink: 0;
  box-shadow: none;

  .header-left {
    display: flex;
    flex-direction: column;
    gap: 3px;

    .page-title {
      font-size: 1rem;
      font-weight: 400;
      color: var(--text-primary);
      letter-spacing: -0.01em;
    }

    .page-desc {
      font-size: var(--font-xs);
      color: var(--text-secondary);
    }
  }

  .header-extra {
    display: flex;
    align-items: center;
    gap: 10px;
  }
}

.h-page-body {
  flex: 1;
  display: flex;
  flex-direction: column;

  &.with-padding {
    padding: var(--h-padding);
    gap: var(--h-padding);
  }
}
</style>
