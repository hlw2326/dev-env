<template>
	<p-button type="button" unstyled class="git-status" :class="tone" @click="emit('open')"><i class="pi pi-github"></i>{{ label }}</p-button>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { ProjectGitStatus } from '@/types'
const props = defineProps<{ status?: ProjectGitStatus }>()
const emit = defineEmits<{ open: [] }>()
const label = computed(() => {
	const value = props.status
	if (value === undefined) return '查看 Git'
	if (!value.available) return 'Git 不可用'
	if (value.error !== '') return 'Git 检查失败'
	if (value.nested) return '嵌套仓库'
	if (!value.repository) return '未启用 Git'
	const parts: string[] = []
	if (value.changed > 0) parts.push(`未提交 ${value.changed} 项`)
	if (value.ahead > 0) parts.push(`待推送 ${value.ahead}`)
	if (value.behind > 0) parts.push(`落后 ${value.behind}`)
	if (parts.length > 0) return parts.join(' · ')
	if (value.upstream === '') return '工作区干净 · 未关联远端分支'
	return '工作区干净 · 本地记录已同步'
})
const tone = computed(() => {
	const value = props.status
	if (value === undefined) return 'muted'
	if (!value.available || value.error !== '' || value.nested || value.changed > 0 || value.ahead > 0 || value.behind > 0) return 'warning'
	if (!value.repository || value.upstream === '') return 'muted'
	return 'clean'
})
</script>

<style scoped lang="scss">
.git-status { display: inline-flex; align-items: center; gap: 0.3rem; padding: 0; border: 0; background: transparent; font-size: var(--font-xs); cursor: pointer; }
.git-status:hover { text-decoration: underline; }
.clean { color: var(--ok); }
.warning { color: var(--warn); }
.muted { color: var(--text-muted); }
</style>
