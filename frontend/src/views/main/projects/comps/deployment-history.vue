<template>
	<section class="history">
		<div class="toolbar"><strong>最近发布记录</strong><p-button size="small" severity="secondary" variant="outlined" icon="pi pi-refresh" label="刷新记录" @click="emit('refresh')" /></div>
		<p v-if="history.length === 0" class="hint">尚无发布记录。启动项目、Git 提交或重启 DevEnv 均不会自动发布。</p>
		<div v-for="task in history" :key="task.id" class="task">
			<div class="toolbar"><span :class="task.state">{{ states[task.state] }} · {{ task.completed }}/{{ task.total }} 文件</span><span class="hint">{{ new Date(task.created).toLocaleString() }}</span></div>
			<code>{{ task.target }}</code><p class="message">{{ task.message }}</p>
		</div>
	</section>
</template>

<script setup lang="ts">
import type { DeploymentTask } from '@/types'
defineProps<{ history: DeploymentTask[] }>()
const emit = defineEmits<{ refresh: [] }>()
const states: Record<string, string> = { running: '执行中', success: '已完成', failed: '失败', cancelled: '已取消或超时', interrupted: '上次执行中断' }
</script>

<style scoped lang="scss">
.history { display: flex; flex-direction: column; gap: var(--h-padding); }
.toolbar { display: flex; align-items: center; justify-content: space-between; gap: var(--h-padding); }
.task { padding: var(--h-padding); border: 1px solid var(--border-color); border-radius: var(--h-radius); font-size: var(--font-sm); }
.task code, .message { display: block; overflow-wrap: anywhere; white-space: pre-wrap; }
.message { margin-bottom: 0; max-height: 12rem; overflow-y: auto; }
.hint { margin: 0; color: var(--text-muted); font-size: var(--font-xs); }
.running { color: var(--info); }
.success { color: var(--ok); }
.failed { color: var(--err); }
.cancelled, .interrupted { color: var(--warn); }
</style>
