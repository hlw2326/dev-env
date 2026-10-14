<template>
	<!-- 顶部统计指标卡片行 -->
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

		<div class="stat-box surface-card" @click="$router.push('/runtimes')">
			<div class="icon-bg indigo">
				<i class="pi pi-box"></i>
			</div>
			<div class="content">
				<span class="num">{{ runtimes.length }}</span>
				<span class="label">运行环境</span>
			</div>
		</div>

		<div class="stat-box surface-card" @click="$router.push('/services')">
			<div class="icon-bg ok">
				<i class="pi pi-server"></i>
			</div>
			<div class="content">
				<span class="num">{{ runningServices }} / {{ services.length }}</span>
				<span class="label">运行中服务</span>
			</div>
		</div>
	</div>
</template>

<script setup lang="ts">
import { useDashboard } from '../use-dashboard'

// 获取控制台共享状态
const { sites, runtimes, services, runningSites, runningServices } = useDashboard()
</script>

<style scoped lang="scss">
/* 统计卡片四列网格 */
.stats-row {
	display: grid;
	grid-template-columns: repeat(4, 1fr);
	gap: var(--h-padding);
}

/* 单个统计卡片 */
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
		font-size: 1.2rem;
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
		font-size: 1.4rem;
		color: var(--text-primary);
		line-height: 1.2;
	}

	.label {
		font-size: 0.8rem;
		color: var(--text-muted);
	}
}
</style>
