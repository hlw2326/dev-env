<template>
	<h-page title="网站" desc="管理本地 PHP、Node.js 与静态网站，所有站点共用本地网关">
		<template #extra>
			<p-button icon="pi pi-plus" label="创建网站" @click="createVisible = true" />
		</template>

		<div class="toolbar surface-card">
			<div class="filter-tabs">
				<button
					v-for="tab in filterOptions"
					:key="tab.value"
					type="button"
					class="filter-btn"
					:class="{ active: kindFilter === tab.value }"
					@click="kindFilter = tab.value"
				>
					{{ tab.label }}
				</button>
			</div>

			<div class="search-box">
				<i class="pi pi-search search-icon"></i>
				<p-input-text v-model="search" placeholder="搜索站点名称或域名..." class="search-input" />
			</div>
		</div>

		<div v-if="filteredSites.length === 0" class="empty-state surface-card">
			<i class="pi pi-globe empty-icon"></i>
			<span class="empty-title">还没有创建任何网站</span>
			<p class="empty-desc">选择站点类型、填写名称和域名即可一键创建本地网站</p>
			<p-button icon="pi pi-plus" label="创建第一个网站" @click="createVisible = true" />
		</div>

		<div v-else class="site-list">
			<site-row
				v-for="item in filteredSites"
				:key="item.id"
				:site="item"
				:installing-kind="installingKind"
				@toggle="handleToggleState"
				@open="handleOpenSite"
				@folder="handleOpenFolder"
				@detail="openDetail"
				@delete="handleDelete"
				@install="handleQuickInstall"
			/>
		</div>

		<site-create
			v-model:visible="createVisible"
			:default-root="defaultRoot"
			@submit="onCreateSubmit"
		/>

		<h-modal
			v-if="activeSite !== null"
			v-model:visible="detailVisible"
			:title="`站点详情 · ${activeSite.name}`"
			width="38rem"
			:footer="false"
		>
			<div class="detail-tabs">
				<button
					v-for="tab in detailTabs"
					:key="tab.value"
					type="button"
					class="detail-tab"
					:class="{ active: currentTab === tab.value }"
					@click="currentTab = tab.value"
				>
					{{ tab.label }}
				</button>
			</div>

			<detail-overview
				v-if="currentTab === 'overview'"
				:site="activeSite"
				:software-list="softwareList"
				@save="handleUpdate"
			/>
			<detail-domains
				v-if="currentTab === 'domains'"
				:site="activeSite"
				@save="handleUpdate"
			/>
			<detail-logs
				v-if="currentTab === 'logs'"
				:site="activeSite"
			/>
		</h-modal>
	</h-page>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useSites } from './use-sites'
import SiteRow from './comps/row.vue'
import SiteCreate from './comps/create.vue'
import DetailOverview from './comps/detail-overview.vue'
import DetailDomains from './comps/detail-domains.vue'
import DetailLogs from './comps/detail-logs.vue'
import type { CreateSiteInput } from '@/types'

const {
	filteredSites,
	softwareList,
	defaultRoot,
	search,
	kindFilter,
	createVisible,
	activeSite,
	detailVisible,
	installingKind,
	handleCreate,
	handleUpdate,
	handleToggleState,
	handleDelete,
	handleQuickInstall,
	handleOpenSite,
	handleOpenFolder,
	openDetail,
} = useSites()

const filterOptions = [
	{ value: 'all', label: '全部' },
	{ value: 'php', label: 'PHP' },
	{ value: 'node', label: 'Node.js' },
	{ value: 'static', label: '静态' },
]

const detailTabs = [
	{ value: 'overview', label: '概览' },
	{ value: 'domains', label: '域名' },
	{ value: 'logs', label: '日志' },
]

const currentTab = ref('overview')

async function onCreateSubmit(input: CreateSiteInput) {
	const ok = await handleCreate(input)
	if (ok) {
		createVisible.value = false
	}
}
</script>

<style scoped lang="scss">
.toolbar {
	display: flex;
	align-items: center;
	justify-content: space-between;
	padding: 0.625rem 0.875rem;
	gap: 1rem;
}

.filter-tabs {
	display: flex;
	align-items: center;
	gap: 0.375rem;
}

.filter-btn {
	border: 1px solid transparent;
	background: transparent;
	color: var(--text-secondary);
	font-size: var(--font-sm);
	padding: 0.35rem 0.75rem;
	border-radius: var(--radius-sm);
	cursor: pointer;

	&:hover {
		background: var(--bg-app);
		color: var(--text-primary);
	}

	&.active {
		background: var(--p-primary-color);
		color: var(--on-primary);
	}
}

.search-box {
	position: relative;
	display: flex;
	align-items: center;
	width: 16rem;

	.search-icon {
		position: absolute;
		left: 0.75rem;
		font-size: var(--icon-sm);
		color: var(--text-muted);
	}

	.search-input {
		width: 100%;
		padding-left: 2.15rem !important;
	}
}

.site-list {
	display: flex;
	flex-direction: column;
	gap: 0.625rem;
}

.empty-state {
	display: flex;
	flex-direction: column;
	align-items: center;
	justify-content: center;
	padding: 3.5rem 1.5rem;
	gap: 0.625rem;

	.empty-icon {
		font-size: var(--icon-2xl);
		color: var(--text-muted);
	}

	.empty-title {
		font-size: var(--font-lg);
		color: var(--text-primary);
	}

	.empty-desc {
		font-size: var(--font-sm);
		color: var(--text-secondary);
		margin-bottom: 0.5rem;
	}
}

.detail-tabs {
	display: flex;
	gap: 0.5rem;
	border-bottom: 1px solid var(--border-color);
	padding-bottom: 0.75rem;
	margin-bottom: 1rem;
}

.detail-tab {
	border: 1px solid var(--border-color);
	background: var(--bg-app);
	color: var(--text-secondary);
	font-size: var(--font-sm);
	padding: 0.35rem 0.875rem;
	border-radius: var(--radius-sm);
	cursor: pointer;

	&.active {
		background: var(--p-primary-color);
		color: var(--on-primary);
		border-color: var(--p-primary-color);
	}
}
</style>
