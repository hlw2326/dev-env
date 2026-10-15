<template>
	<hlw-drawer v-model:visible="visible" title="本机环境" width="70%" :mask-close="true" :footer="true">
		<template #top>
			<p-toolbar :pt="toolbarLayout">
				<template #start>
					<div class="filter">
						<label for="local-environment-kind" class="text-sm text-muted">软件类型</label>
						<p-select v-model="selectedOption" input-id="local-environment-kind" size="small" :options="options"
							option-label="name" option-value="value" :disabled="loading" :pt="selectLayout" />
					</div>
				</template>
				<template #end>
					<div class="actions">
						<p-button size="small" label="选择程序" severity="secondary" variant="outlined" icon="pi pi-folder-open"
							:disabled="selectedKind === '' || loading" @click="selectExecutable(selectedKind)" />
						<p-button size="small" label="查找已有环境" icon="pi pi-search" :loading="loading" :disabled="loading" @click="scanExisting(selectedKind)" />
					</div>
				</template>
			</p-toolbar>
		</template>
		<p-panel header="查询结果" :pt="resultsLayout">
			<template #icons>
				<h-tag v-if="scanned && selectedKind === '' && detectedCount > 0" color="success" :value="`已发现 ${detectedCount} 类环境`" />
			</template>
			<p-message v-if="managerError !== ''" severity="error" icon="pi pi-exclamation-circle" :pt="messageLayout">{{ managerError }}</p-message>
			<p-message v-if="stale" severity="warn" size="small" icon="pi pi-info-circle">当前显示上次查询结果，重新查找后更新。</p-message>
			<p-message v-if="loading" severity="secondary" size="small" icon="pi pi-spin pi-spinner" variant="simple">正在读取本机已有环境，请稍候…</p-message>
			<h-empty v-else-if="!scanned && managerError === ''" icon="pi pi-desktop" title="查找已有的开发环境"
				description="查看电脑上已安装的软件版本、程序位置和服务状态。">
				<p-button size="small" label="开始查找" icon="pi pi-search" variant="outlined" @click="scanExisting(selectedKind)" />
			</h-empty>
			<h-empty v-else-if="scanned && displayedReports.length === 0 && managerError === ''" icon="pi pi-search" title="未发现外部环境"
				description="可以手动选择已安装的程序文件，也可以返回软件商店安装。">
				<p-button v-if="selectedKind !== ''" size="small" label="选择已有程序" icon="pi pi-folder-open" severity="secondary"
					variant="outlined" @click="selectExecutable(selectedKind)" />
			</h-empty>
			<comp-environment-result v-for="report in displayedReports" :key="report.kind" :report="report" />
		</p-panel>
		<template #footer>
			<p-button label="关闭" severity="secondary" variant="outlined" @click="visible = false" />
		</template>
	</hlw-drawer>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useLocalEnvironments } from '@/composables'
import type { MarketItem } from '@/types'
import CompEnvironmentResult from './environment-result.vue'

const ALL_SOFTWARE_VALUE = '__all_software__'
const props = defineProps<{ items: MarketItem[] }>()
const { visible, selectedKind, loading, scanned, stale, displayedReports, detectedCount, managerError, scanExisting, selectExecutable } = useLocalEnvironments()
const options = computed(() => [{ name: '全部软件', value: ALL_SOFTWARE_VALUE }, ...props.items.map((item) => ({ name: item.name, value: item.kind }))])
const selectedOption = computed({
	get(): string {
		if (selectedKind.value === '') return ALL_SOFTWARE_VALUE
		return selectedKind.value
	},
	set(value: string): void {
		if (value === ALL_SOFTWARE_VALUE) {
			selectedKind.value = ''
			return
		}
		selectedKind.value = value
	},
})
const toolbarLayout = {
	root: { style: { border: 'none', borderBottom: '1px solid var(--border-color)', borderRadius: 0, padding: 'var(--h-padding)', gap: 'var(--h-padding)' } },
}
const selectLayout = { root: { style: { minWidth: '12rem' } } }
const resultsLayout = {
	root: { style: { fontSize: 'var(--font-sm)' } },
	header: { style: { padding: 'var(--h-padding)', color: 'var(--text-secondary)' } },
	content: { style: { display: 'grid', gap: 'var(--h-padding)', padding: 'var(--h-padding)' } },
}
const messageLayout = { text: { style: { whiteSpace: 'pre-wrap', overflowWrap: 'anywhere', lineHeight: 'var(--leading-normal)' } } }
</script>

<style scoped lang="scss">
.filter,
.actions {
	display: flex;
	align-items: center;
	gap: var(--h-padding);
}

.actions {
	flex-wrap: wrap;
	justify-content: flex-end;
}
</style>
