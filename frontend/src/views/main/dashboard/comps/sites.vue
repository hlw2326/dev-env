<template>
	<div class="panel surface-card">
		<div class="header">
			<div class="title">
				<i class="pi pi-globe icon"></i>
				<span>最近网站</span>
			</div>
			<p-button
				label="进入网站管理"
				icon="pi pi-arrow-right"
				icon-pos="right"
				size="small"
				@click="$router.push('/sites')"
			/>
		</div>

		<div v-if="sites.length === 0" class="empty">
			尚未创建任何网站
		</div>

		<div v-else class="list">
			<div
				v-for="site in recentSites"
				:key="site.id"
				class="row"
				@click="$router.push('/sites')"
			>
				<div class="info">
					<div class="title-line">
						<span class="name">{{ site.name }}</span>
						<span class="kind font-mono">{{ site.kind.toUpperCase() }}</span>
					</div>
					<span class="path font-mono" :title="site.hosts.join(', ')">
						{{ site.hosts.join(', ') }} · {{ site.root }}
					</span>
				</div>
				<p-tag
					v-if="site.state === 'running'"
					severity="success"
					value="运行中"
					size="small"
				/>
				<p-tag
					v-else-if="site.need !== ''"
					severity="warn"
					value="缺少环境"
					size="small"
				/>
				<p-tag
					v-else-if="site.state === 'error'"
					severity="danger"
					value="异常"
					size="small"
				/>
				<p-tag
					v-else
					severity="secondary"
					value="已停止"
					size="small"
				/>
			</div>
		</div>
	</div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useDashboard } from '../use-dashboard'

const { sites } = useDashboard()

const recentSites = computed(() => {
	return sites.value.slice(0, 5)
})
</script>

<style scoped lang="scss">
.panel {
	padding: var(--h-padding);
	display: flex;
	flex-direction: column;
	gap: var(--h-padding);
}

.header {
	display: flex;
	justify-content: space-between;
	align-items: center;

	.title {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		font-size: var(--font-lg);
		color: var(--text-primary);
	}

	.icon {
		color: var(--indigo);
	}
}

.list {
	display: flex;
	flex-direction: column;
	gap: 0.5rem;
}

.row {
	display: flex;
	justify-content: space-between;
	align-items: center;
	padding: 0.75rem 0.875rem;
	background: var(--bg-app);
	border: 1px solid var(--border-color);
	border-radius: var(--radius-sm);
	cursor: pointer;

	&:hover {
		background: var(--bg-card);
		border-color: var(--border-hover);
	}

	.info {
		display: flex;
		flex-direction: column;
		gap: 0.2rem;
	}

	.title-line {
		display: flex;
		align-items: center;
		gap: 0.5rem;
	}

	.name {
		font-size: var(--font-base);
		color: var(--text-primary);
	}

	.kind {
		font-size: var(--font-xs);
		color: var(--text-muted);
	}

	.path {
		font-size: var(--font-xs);
		color: var(--text-muted);
		max-width: 20rem;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}
}

.empty {
	text-align: center;
	padding: 2rem;
	color: var(--text-muted);
	font-size: var(--font-sm);
	background: var(--bg-app);
	border: 1px dashed var(--border-color);
	border-radius: var(--h-radius);
}
</style>
