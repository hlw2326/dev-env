<template>
	<hlw-drawer v-model:visible="visible" title="Redis 数据" width="70%" :mask-close="true" :footer="true">
		<template #top>
			<p-toolbar>
				<template #start>
					<div class="toolbar">
						<p-select v-model="selectedDatabase" size="small" :options="databaseOptions" option-label="label" option-value="value"
							:loading="loadingDatabases" :disabled="loadingDatabases" @update:model-value="selectDatabase" />
						<p-input-group>
							<p-input-group-addon><i class="pi pi-search"></i></p-input-group-addon>
							<p-input-text v-model="pattern" size="small" placeholder="键名匹配规则，例如 cache:*" @keyup.enter="searchKeys" />
							<p-button size="small" label="查询" icon="pi pi-search" :loading="loadingKeys" @click="searchKeys" />
						</p-input-group>
					</div>
				</template>
				<template #end>
					<p-button size="small" severity="secondary" variant="outlined" label="刷新" icon="pi pi-refresh"
						:loading="loadingDatabases" @click="loadDatabases" />
				</template>
			</p-toolbar>
		</template>

		<p-splitter class="browser">
			<p-splitter-panel :size="42" :min-size="30">
				<p-panel header="键列表" :pt="panelLayout">
					<template #icons>
						<span class="text-xs text-muted">第 {{ pageIndex + 1 }} 页</span>
					</template>
					<p-data-table v-model:selection="selectedKey" :value="keys" data-key="name" selection-mode="single" size="small"
						:loading="loadingKeys" :table-style="{ width: '100%', tableLayout: 'fixed' }" @row-select="loadValue($event.data)">
						<p-column field="name" header="键名" />
						<p-column field="type" header="类型" :style="{ width: '5rem' }" />
						<p-column header="TTL" :style="{ width: '5rem' }">
							<template #body="{ data }">{{ formatTTL(data.ttl) }}</template>
						</p-column>
					</p-data-table>
					<div class="pagination">
						<p-button size="small" severity="secondary" variant="outlined" label="上一页" :disabled="pageIndex === 0 || loadingKeys"
							@click="previousKeys" />
						<p-button size="small" severity="secondary" variant="outlined" label="下一页" :disabled="nextCursor === 0 || loadingKeys"
							@click="nextKeys" />
					</div>
				</p-panel>
			</p-splitter-panel>

			<p-splitter-panel :size="58" :min-size="35">
				<p-panel header="键详情" :pt="panelLayout">
					<template #icons>
						<h-tag v-if="value !== null" :value="value.type" />
						<h-tag v-if="value !== null" color="info" :value="`TTL ${formatTTL(value.ttl)}`" />
					</template>
					<h-empty v-if="selectedKey === null" icon="pi pi-database" title="选择一个 Redis 键"
						description="键内容仅在选中后读取，当前数据浏览为只读模式。" />
					<p-message v-else-if="loadingValue" severity="secondary" size="small" icon="pi pi-spin pi-spinner" variant="simple">正在读取键内容…</p-message>
					<template v-else-if="value !== null">
						<p-message v-if="value.type === 'string' && value.encoding === 'base64'" severity="warn" size="small">
							当前值为二进制内容，已使用 Base64 展示。
						</p-message>
						<p-message v-if="value.truncated" severity="warn" size="small">
							内容超过 64 KB，仅展示前 64 KB。
						</p-message>
						<pre v-if="value.type === 'string'" class="value">{{ value.text }}</pre>
						<p-data-table v-else-if="value.entries.length > 0" :value="value.entries" size="small" :table-style="{ width: '100%', tableLayout: 'fixed' }">
							<p-column field="field" header="字段" :style="{ width: '30%' }" />
							<p-column field="value" header="值" />
							<p-column v-if="value.type === 'zset'" field="score" header="分数" :style="{ width: '7rem' }" />
						</p-data-table>
						<p-message v-else severity="secondary" size="small">当前键类型没有可展示的文本内容。</p-message>
					</template>
				</p-panel>
			</p-splitter-panel>
		</p-splitter>

		<template #footer>
			<p-button label="关闭" severity="secondary" variant="outlined" @click="visible = false" />
		</template>
	</hlw-drawer>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRedisData } from '@/composables'

const {
	visible,
	loadingDatabases,
	loadingKeys,
	loadingValue,
	databases,
	selectedDatabase,
	pattern,
	keys,
	pageIndex,
	nextCursor,
	selectedKey,
	value,
	loadDatabases,
	selectDatabase,
	searchKeys,
	previousKeys,
	nextKeys,
	loadValue,
} = useRedisData()

const databaseOptions = computed(() => databases.value.map((item) => ({
	label: `DB ${item.index} · ${item.keys} 个键`,
	value: item.index,
})))
const panelLayout = { content: { style: { display: 'grid', gap: 'var(--h-padding)', padding: 'var(--h-padding)' } } }

function formatTTL(value: number): string {
	if (value === -1) return '永久'
	if (value === -2) return '已过期'
	return `${value}s`
}
</script>

<style scoped lang="scss">
.toolbar,
.pagination {
	display: flex;
	align-items: center;
	gap: var(--h-padding);
}

.toolbar {
	flex-wrap: wrap;
}

.pagination {
	justify-content: flex-end;
}

.browser {
	min-height: 32rem;
}

.value {
	max-height: 24rem;
	overflow: auto;
	padding: var(--h-padding);
	white-space: pre-wrap;
	overflow-wrap: anywhere;
	background: var(--bg-app);
	border: 1px solid var(--border-color);
	border-radius: var(--h-radius);
	font-family: var(--font-mono);
	font-size: var(--font-xs);
	line-height: var(--leading-normal);
}
</style>
