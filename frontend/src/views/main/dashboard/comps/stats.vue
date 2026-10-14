<template>
	<div class="stats-row">
		<div class="stat-box surface-card" @click="$router.push('/sites')">
			<div class="icon-bg info">
				<i class="pi pi-globe"></i>
			</div>
			<div class="content">
				<span class="num">{{ sites.length }}</span>
				<span class="label">网站总数</span>
			</div>
		</div>

		<div class="stat-box surface-card" @click="$router.push('/sites')">
			<div class="icon-bg warn">
				<i class="pi pi-check-circle"></i>
			</div>
			<div class="content">
				<span class="num">{{ runningSites }} / {{ sites.length }}</span>
				<span class="label">运行中网站</span>
			</div>
		</div>

		<div class="stat-box surface-card" @click="$router.push('/software')">
			<div class="icon-bg indigo">
				<i class="pi pi-box"></i>
			</div>
			<div class="content">
				<span class="num">{{ readySoftware.length }}</span>
				<span class="label">已安装软件</span>
			</div>
		</div>

		<div class="stat-box surface-card" @click="$router.push('/software')">
			<div class="icon-bg ok">
				<i class="pi pi-server"></i>
			</div>
			<div class="content">
				<span class="num">{{ runningDaemons }} / {{ daemonList.length }}</span>
				<span class="label">运行中服务</span>
			</div>
		</div>
	</div>
</template>

<script setup lang="ts">
import { useDashboard } from '../use-dashboard'

const { sites, readySoftware, daemonList, runningSites, runningDaemons } = useDashboard()
</script>

<style scoped lang="scss">
.stats-row {
	display: grid;
	grid-template-columns: repeat(4, 1fr);
	gap: var(--h-padding);
}

.stat-box {
	padding: var(--h-padding);
	display: flex;
	align-items: center;
	gap: 0.875rem;
	cursor: pointer;

	&:hover {
		background: var(--bg-app);
		border-color: var(--border-hover);
	}

	.icon-bg {
		width: 2.75rem;
		height: 2.75rem;
		border-radius: var(--radius-sm);
		display: flex;
		align-items: center;
		justify-content: center;
		font-size: var(--icon-lg);
		flex-shrink: 0;

		&.info {
			background: var(--info-bg);
			color: var(--info);
		}
		&.warn {
			background: var(--warn-bg);
			color: var(--warn);
		}
		&.indigo {
			background: var(--indigo-bg);
			color: var(--indigo);
		}
		&.ok {
			background: var(--ok-bg);
			color: var(--ok);
		}
	}

	.content {
		display: flex;
		flex-direction: column;
	}

	.num {
		font-size: var(--font-2xl);
		color: var(--text-primary);
		line-height: var(--leading-tight);
	}

	.label {
		font-size: var(--font-xs);
		color: var(--text-muted);
	}
}
</style>
