<template>
	<p-panel :header="report.name" toggleable collapsed :pt="panelLayout">
		<template #header>
			<div class="summary">
				<h-logo :kind="report.kind" size="small" />
				<span class="name text-sm">{{ report.name }}</span>
				<span class="version font-mono text-xs text-muted">{{ versions.join('、') }}</span>
				<div class="tags">
					<h-tag v-if="roles.length > 0" :value="roles.join('、')" />
					<h-tag v-if="hasIssues" value="入口需确认" color="warning" />
				</div>
			</div>
		</template>
		<p-message v-if="report.error !== ''" severity="error" size="small" icon="pi pi-exclamation-circle">{{ report.error }}</p-message>
		<p-data-table v-if="entries.length > 0" :value="entries" size="small" :table-style="{ width: '100%', tableLayout: 'fixed' }">
			<p-column header="版本 / 类型">
				<template #body="{ data }">
					<span class="font-mono text-sm">{{ data.versions.join('、') }}</span>
					<h-tag :value="data.role" />
					<h-tag v-if="data.active" value="PATH 首选" color="info" />
				</template>
			</p-column>
			<p-column header="程序路径" :body-style="{ overflowWrap: 'anywhere' }">
				<template #body="{ data }">
					<span class="font-mono text-xs">{{ data.path }}</span>
					<p-message v-if="data.error !== ''" severity="warn" size="small">{{ data.error }}</p-message>
				</template>
			</p-column>
			<p-column header="来源">
				<template #body="{ data }">
					<span>{{ data.source }}</span>
					<p v-if="data.serviceName !== ''" class="text-xs text-muted">{{ data.serviceName }} · {{ data.serviceState }}</p>
				</template>
			</p-column>
		</p-data-table>
		<p-message v-for="warning in report.warnings" :key="warning" severity="warn" size="small">{{ warning }}</p-message>
		<p-message v-if="report.conflict" severity="warn" size="small">全局 {{ report.command }} 入口与 DevEnv 默认版本 {{ report.defaultVersion }} 不一致；项目仍使用各自绑定的运行环境。</p-message>
	</p-panel>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { CommandEnvironment } from '@/types'

const props = defineProps<{ report: CommandEnvironment }>()
const entries = computed(() => props.report.entries.filter((entry) => !entry.managed))
const versions = computed(() => [...new Set(entries.value.flatMap((entry) => entry.versions))])
const roles = computed(() => [...new Set(entries.value.filter((entry) => entry.versions.length > 0).map((entry) => entry.role))])
const hasIssues = computed(() => props.report.error !== '' || props.report.warnings.length > 0 || entries.value.some((entry) => entry.error !== ''))
const panelLayout = {
	header: { style: { gap: 'var(--h-padding)', flexWrap: 'wrap', padding: 'var(--h-padding)' } },
	content: { style: { display: 'grid', gap: 'var(--h-padding)', padding: 'var(--h-padding)' } },
}
</script>

<style scoped lang="scss">
.summary {
	display: grid;
	grid-template-columns: auto minmax(0, 1fr) minmax(0, 1fr) minmax(0, 2fr);
	align-items: center;
	gap: var(--h-padding);
	flex: 1;
	min-width: 0;
}

.name,
.version,
.tags {
	min-width: 0;
}

.version {
	overflow: hidden;
	text-overflow: ellipsis;
	white-space: nowrap;
}

.tags {
	display: flex;
	align-items: center;
	gap: var(--h-padding);
	flex-wrap: wrap;
}
</style>
