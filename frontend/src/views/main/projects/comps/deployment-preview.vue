<template>
	<section class="preview">
		<strong>发布预览 · {{ plan.files.length }} 个文件 · {{ (plan.bytes / 1024 / 1024).toFixed(2) }} MB</strong>
		<div class="target">{{ plan.source }} → {{ plan.target }}</div>
		<p class="hint">已排除 {{ plan.excluded }} 个文件或目录。预览 15 分钟内有效，文件或配置改变后须重新预览。以下是将上传的本地文件，不是服务器差异对比。</p>
		<p-input-text v-model="search" placeholder="输入文件路径筛选预览清单" />
		<div class="list"><div v-for="file in files" :key="file.path" class="file"><code>{{ file.path }}</code><span>{{ file.size }} B</span></div></div>
		<p class="hint">最多展示匹配的前 200 项。服务器同名文件会被覆盖，额外文件保留；中断后已完成的部分不会回滚。</p>
	</section>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import type { DeploymentPlan } from '@/types'
const props = defineProps<{ plan: DeploymentPlan }>()
const search = ref('')
const files = computed(() => props.plan.files.filter((file) => file.path.toLowerCase().includes(search.value.toLowerCase())).slice(0, 200))
</script>

<style scoped lang="scss">
.preview { display: flex; flex-direction: column; gap: calc(var(--h-padding) / 2); padding: var(--h-padding); background: var(--bg-app); border: 1px solid var(--border-color); border-radius: var(--h-radius); }
.target { color: var(--text-secondary); overflow-wrap: anywhere; font-size: var(--font-sm); }
.hint { margin: 0; color: var(--text-muted); font-size: var(--font-xs); }
.list { max-height: 14rem; overflow-y: auto; }
.file { display: flex; justify-content: space-between; gap: var(--h-padding); font-size: var(--font-xs); }
.file code { overflow-wrap: anywhere; min-width: 0; }
.file span { color: var(--text-muted); flex-shrink: 0; }
</style>
