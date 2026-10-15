<template>
	<p-button v-if="versions.length === 0" size="small" severity="secondary" variant="text" icon="pi pi-search"
		label="查找本机已有版本" :loading="loading && activeKind === kind" :disabled="loading" @click="scanExisting(kind)" />
	<template v-else>
		<p-message severity="info" size="small">本机已有 {{ versions.join('、') }}，可继续安装独立的 DevEnv 版本。</p-message>
		<p-button size="small" severity="secondary" variant="text" label="查看本机环境" icon="pi pi-list" @click="openEnvironments(kind)" />
	</template>
	<p-message v-if="error !== ''" severity="error" size="small" :pt="messageLayout">{{ error }}</p-message>
	<p-message v-else-if="report !== undefined && versions.length === 0 && !stale" severity="secondary" size="small">
		未找到与当前安装包对应的外部版本，可正常安装，或在“本机环境”中选择程序文件。
	</p-message>
	<p-button v-if="report !== undefined && versions.length === 0" size="small" severity="secondary" variant="text"
		label="打开本机环境" @click="openEnvironments(kind)" />
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useLocalEnvironments, externalVersions } from '@/composables'

const props = defineProps<{ kind: string }>()
const { loading, activeKind, stale, findReport, findError, scanExisting, openEnvironments } = useLocalEnvironments()
const versions = computed(() => externalVersions(props.kind))
const report = computed(() => findReport(props.kind))
const error = computed(() => findError(props.kind))
const messageLayout = { text: { style: { whiteSpace: 'pre-wrap', overflowWrap: 'anywhere' } } }
</script>
