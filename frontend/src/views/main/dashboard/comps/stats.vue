<template>
	<!-- 概览统计卡片网格 -->
	<div class="stats-row">
		<!-- 单个统计卡片 -->
		<div
			v-for="card in statCards"
			:key="card.label"
			class="stat-box surface-card"
			@click="$router.push(card.route)"
		>
			<div class="icon-bg" :class="card.tone">
				<i :class="card.icon"></i>
			</div>
			<div class="content">
				<span class="value">{{ card.value }}</span>
				<span class="label">{{ card.label }}</span>
			</div>
		</div>
	</div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useDashboard } from '../use-dashboard'

// 概览统计状态
const { sites, readySoftware, daemonList, runningSites, runningDaemons } = useDashboard()

// 统计卡片配置列表
const statCards = computed(() => [
	{
		label: '项目总数',
		value: `${sites.value.length}`,
		icon: 'pi pi-globe',
		tone: 'info',
		route: '/sites',
	},
	{
		label: '运行中项目',
		value: `${runningSites.value} / ${sites.value.length}`,
		icon: 'pi pi-check-circle',
		tone: 'warn',
		route: '/sites',
	},
	{
		label: '已安装软件',
		value: `${readySoftware.value.length}`,
		icon: 'pi pi-box',
		tone: 'indigo',
		route: '/software',
	},
	{
		label: '运行中服务',
		value: `${runningDaemons.value} / ${daemonList.value.length}`,
		icon: 'pi pi-server',
		tone: 'ok',
		route: '/services',
	},
])
</script>

<style scoped lang="scss">
/* 四列统计网格 */
.stats-row {
	display: grid;
	grid-template-columns: repeat(4, 1fr);
	gap: var(--h-padding);
}

/* 统计卡片容器 */
.stat-box {
	padding: var(--h-padding);
	display: flex;
	align-items: center;
	gap: 0.875rem;
	cursor: pointer;
	transition: all 0.15s ease;

	/* 悬停态 */
	&:hover {
		background: var(--bg-app);
		border-color: var(--border-hover);
	}

	/* 图标色块 */
	.icon-bg {
		width: 2.75rem;
		height: 2.75rem;
		border-radius: var(--h-radius);
		display: flex;
		align-items: center;
		justify-content: center;
		font-size: var(--icon-lg);
		flex-shrink: 0;

		/* 信息色调 */
		&.info {
			background: var(--info-bg);
			color: var(--info);
		}

		/* 警告色调 */
		&.warn {
			background: var(--warn-bg);
			color: var(--warn);
		}

		/* 靛青色调 */
		&.indigo {
			background: var(--indigo-bg);
			color: var(--indigo);
		}

		/* 成功色调 */
		&.ok {
			background: var(--ok-bg);
			color: var(--ok);
		}
	}

	/* 数值与标签区 */
	.content {
		display: flex;
		flex-direction: column;
	}

	/* 统计大字数值 */
	.value {
		font-size: var(--font-2xl);
		color: var(--text-primary);
		line-height: var(--leading-tight);
	}

	/* 统计标签文本 */
	.label {
		font-size: var(--font-xs);
		color: var(--text-muted);
	}
}
</style>
