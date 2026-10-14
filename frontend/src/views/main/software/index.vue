<template>
	<h-page title="软件商店" desc="安装、启停、配置 Nginx、PHP、Node.js、MySQL 与 Redis">
		<div class="top-bar surface-card">
			<div class="main-tabs">
				<button
					type="button"
					class="tab-btn"
					:class="{ active: tab === 'market' }"
					@click="tab = 'market'"
				>
					全部软件
				</button>
				<button
					type="button"
					class="tab-btn"
					:class="{ active: tab === 'installed' }"
					@click="tab = 'installed'"
				>
					已安装 ({{ installedCount }})
				</button>
			</div>

			<div class="right-filters">
				<div v-if="tab === 'market'" class="cat-tabs">
					<button
						v-for="cat in categories"
						:key="cat"
						type="button"
						class="cat-btn"
						:class="{ active: category === cat }"
						@click="category = cat"
					>
						{{ cat }}
					</button>
				</div>

				<div class="search-box">
					<i class="pi pi-search search-icon"></i>
					<p-input-text v-model="search" placeholder="搜索软件..." class="search-input" />
				</div>
			</div>
		</div>

		<div v-if="marketError !== ''" class="error-banner surface-card">
			<span class="err-text">{{ marketError }}</span>
			<p-button size="small" icon="pi pi-refresh" label="重试" @click="refreshAll" />
		</div>

		<div v-if="tab === 'market'" class="market-grid">
			<market-card
				v-for="item in filteredMarket"
				:key="item.kind"
				:item="item"
				:installed-list="installedList"
				:progress-map="progressMap"
				@open-install="openInstallModal"
				@cancel="handleCancelInstall"
				@config="openConfigByKind"
			/>
		</div>

		<div v-else class="installed-section">
			<div v-if="filteredInstalled.length === 0" class="empty-state surface-card">
				<i class="pi pi-shopping-bag empty-icon"></i>
				<span class="empty-title">还没有安装任何软件</span>
				<p class="empty-desc">切换到「全部软件」页签选择需要的环境或数据库一键安装</p>
				<p-button label="浏览全部软件" @click="tab = 'market'" />
			</div>

			<div v-else class="installed-list">
				<installed-row
					v-for="item in filteredInstalled"
					:key="item.id"
					:item="item"
					@start="handleStart"
					@stop="handleStop"
					@set-main="handleSetMain"
					@config="openConfigDrawer"
					@remove="handleRemove"
					@cancel="handleCancelInstall"
					@retry="handleInstall"
				/>
			</div>
		</div>

		<install-modal
			v-model:visible="installModalVisible"
			:item="activeMarketItem"
			:versions="versionOptions"
			:loading="versionsLoading"
			:error="versionsError"
			@install="handleInstall"
			@retry="fetchVersionsFor"
		/>

		<config-drawer
			v-model:visible="configVisible"
			:item="activeSoftware"
			@start="handleStart"
			@stop="handleStop"
			@set-main="handleSetMain"
			@autostart="handleSetAutostart"
			@open-folder="handleOpenFolder"
		/>
	</h-page>
</template>

<script setup lang="ts">
import { useSoftware } from './use-software'
import MarketCard from './comps/market-card.vue'
import InstalledRow from './comps/installed-row.vue'
import InstallModal from './comps/install-modal.vue'
import ConfigDrawer from './comps/config-drawer.vue'

const categories = ['全部', 'Web服务', '运行环境', '数据库', '缓存']

const {
	tab,
	category,
	search,
	filteredMarket,
	installedList,
	filteredInstalled,
	installedCount,
	marketError,
	progressMap,
	installModalVisible,
	activeMarketItem,
	versionOptions,
	versionsLoading,
	versionsError,
	configVisible,
	activeSoftware,
	refreshAll,
	openInstallModal,
	fetchVersionsFor,
	handleInstall,
	handleCancelInstall,
	handleRemove,
	handleSetMain,
	handleStart,
	handleStop,
	handleSetAutostart,
	handleOpenFolder,
	openConfigDrawer,
	openConfigByKind,
} = useSoftware()
</script>

<style scoped lang="scss">
.top-bar {
	display: flex;
	align-items: center;
	justify-content: space-between;
	padding: 0.625rem 0.875rem;
	gap: 1rem;
	flex-wrap: wrap;
}

.main-tabs,
.cat-tabs {
	display: flex;
	align-items: center;
	gap: 0.375rem;
}

.right-filters {
	display: flex;
	align-items: center;
	gap: 0.875rem;
}

.tab-btn,
.cat-btn {
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
	width: 14rem;

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

.error-banner {
	display: flex;
	align-items: center;
	justify-content: space-between;
	padding: 0.75rem 1rem;
	background: var(--err-bg);
	border-color: var(--err-border);

	.err-text {
		font-size: var(--font-sm);
		color: var(--err);
	}
}

.market-grid {
	display: grid;
	grid-template-columns: repeat(2, 1fr);
	gap: var(--h-padding);
}

.installed-list {
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
</style>
