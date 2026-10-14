<template>
	<!-- 单个网站展示卡片 -->
	<div class="site-card">
		<!-- 卡片顶部名称、状态与快捷操作 -->
		<div class="site-header">
			<div class="title-row">
				<div class="kind-badge" :class="`type-${item.kind}`">
					<i :class="getKindIcon(item.kind)"></i>
				</div>
				<span class="site-name">{{ item.name }}</span>
				<p-tag :severity="statusTones[item.status]" :value="statusLabels[item.status]" size="small" />
				<p-tag v-if="item.tls" severity="info" value="HTTPS" icon="pi pi-lock" size="small" />
			</div>

			<div class="top-actions">
				<p-button label="管理" icon="pi pi-sliders-h" size="small" severity="secondary" outlined @click="openDetail(item, 'info')" />
				<p-button label="访问" icon="pi pi-external-link" size="small" severity="secondary" outlined @click="openBrowser(item)" />
				<p-button icon="pi pi-trash" size="small" severity="danger" text rounded title="删除网站" @click="removeItem($event, item)" />
			</div>
		</div>

		<!-- 域名、运行环境与物理根目录网格 -->
		<div class="info-grid">
			<div class="info-item">
				<i class="pi pi-globe text-info"></i>
				<span class="domain-links font-mono">
					<a v-for="domain in item.domains" :key="domain" href="javascript:void(0)" class="domain-chip" @click="openDomain(item, domain)">
						{{ formatDomainPort(domain, item.port) }}
					</a>
				</span>
			</div>

			<div class="info-item">
				<i class="pi pi-box text-indigo"></i>
				<span class="rt-tag">{{ getRuntimeLabel(item) }}</span>
				<span v-if="item.upstream" class="port-tag font-mono">:{{ item.upstream }}</span>
			</div>

			<div class="info-item">
				<i class="pi pi-folder text-muted"></i>
				<span class="path-text font-mono" :title="item.root">{{ item.root }}</span>
			</div>
		</div>

		<!-- 访问统计与 Git 仓库状态栏 -->
		<div class="metrics-bar">
			<div class="metric-group">
				<span class="metric-item"><i class="pi pi-chart-line"></i> 今日访问: <span class="font-mono">{{ stats.visits }}</span></span>
				<span class="metric-sep">|</span>
				<span class="metric-item"><i class="pi pi-exclamation-circle"></i> 错误: <span class="font-mono">{{ stats.errors }}</span></span>
				<span class="metric-sep">|</span>
				<span class="metric-item"><i class="pi pi-clock"></i> 平均响应: <span class="font-mono">{{ stats.latency }}ms</span></span>
			</div>

			<div class="git-group">
				<template v-if="item.git">
					<i class="pi pi-github"></i>
					<span class="font-mono">{{ item.branch }} 分支</span>
					<span v-if="item.commit" class="commit-sha font-mono">#{{ item.commit }}</span>
					<span class="git-link" @click="openDetail(item, 'git')">Git 仓库 ▸</span>
				</template>
				<template v-else>
					<i class="pi pi-info-circle text-muted"></i>
					<span class="text-muted">Git: 未启用</span>
				</template>
			</div>
		</div>

		<!-- 底部启停与配置入口栏 -->
		<div class="footer-actions">
			<div class="left-btns">
				<p-button
					v-if="item.status !== 'running'"
					label="启动"
					icon="pi pi-play"
					size="small"
					severity="success"
					:loading="busyId === item.id"
					@click="startItem(item)"
				/>
				<p-button
					v-else
					label="停止"
					icon="pi pi-stop"
					size="small"
					severity="danger"
					outlined
					:loading="busyId === item.id"
					@click="stopItem(item)"
				/>
				<p-button
					label="重启"
					icon="pi pi-refresh"
					size="small"
					severity="secondary"
					outlined
					:loading="busyId === item.id"
					@click="restartItem(item)"
				/>
			</div>

			<div class="right-btns">
				<p-button label="Nginx 配置" icon="pi pi-code" size="small" severity="secondary" text @click="openDetail(item, 'conf')" />
				<p-button label="SSL 证书" icon="pi pi-lock" size="small" severity="secondary" text @click="openDetail(item, 'ssl')" />
				<p-button label="日志分析" icon="pi pi-file" size="small" severity="secondary" text @click="openDetail(item, 'logs')" />
			</div>
		</div>
	</div>
</template>

<script setup lang="ts">
	import { computed } from 'vue'
	import type { Site } from '@/types'
	import {
		statusTones, statusLabels, getKindIcon, formatDomainPort,
		getRuntimeLabel, getSiteStats, useSites
	} from '../use-sites'
	import { openDetail } from '../use-detail'

	// 接收网站对象
	const props = defineProps<{ item: Site }>()

	// 解构网站启停与访问操作
	const { busyId, startItem, stopItem, restartItem, removeItem, openDomain, openBrowser } = useSites()

	// 网站访问统计数据
	const stats = computed(() => getSiteStats(props.item))
</script>

<style scoped lang="scss">
	/* 网站卡片外层容器 */
	.site-card {
		background: var(--bg-card); border: 1px solid var(--border-color);
		border-radius: var(--h-radius); padding: 16px 18px;
		display: flex; flex-direction: column; gap: 12px;
	}

	/* 顶部标题与按钮行 */
	.site-header { display: flex; justify-content: space-between; align-items: center; }
	.title-row { display: flex; align-items: center; gap: 8px; }
	.top-actions { display: flex; align-items: center; gap: 6px; }
	.site-name { font-size: var(--font-lg); color: var(--text-primary); }

	/* 网站类型徽章 */
	.kind-badge {
		width: 28px; height: 28px; border-radius: var(--radius-sm);
		display: flex; align-items: center; justify-content: center;
		background: var(--info-bg); color: var(--info);
		&.type-php { background: var(--indigo-bg); color: var(--indigo); }
		&.type-nodejs { background: var(--ok-bg); color: var(--ok); }
		&.type-python { background: var(--warn-bg); color: var(--warn); }
	}

	/* 三列信息网格 */
	.info-grid {
		display: grid; grid-template-columns: repeat(3, 1fr); gap: 10px;
		background: var(--bg-app); padding: 10px 12px;
		border-radius: var(--radius-sm); border: 1px solid var(--border-color);
	}
	.info-item { display: flex; align-items: center; gap: 8px; font-size: var(--font-xs); color: var(--text-secondary); overflow: hidden; }
	.domain-links { display: flex; gap: 6px; flex-wrap: wrap; }
	.domain-chip { color: var(--info); text-decoration: none; &:hover { text-decoration: underline; } }
	.port-tag { background: var(--bg-card); border: 1px solid var(--border-color); padding: 1px 5px; border-radius: var(--radius-sm); }
	.path-text { white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }

	/* 指标与 Git 状态条 */
	.metrics-bar { display: flex; justify-content: space-between; align-items: center; font-size: var(--font-xs); color: var(--text-secondary); }
	.metric-group, .git-group { display: flex; align-items: center; gap: 8px; }
	.metric-sep { color: var(--border-color); }
	.git-link { color: var(--p-primary-color); cursor: pointer; }

	/* 底部按钮栏 */
	.footer-actions { display: flex; justify-content: space-between; align-items: center; padding-top: 8px; border-top: 1px solid var(--border-color); }
	.left-btns, .right-btns { display: flex; align-items: center; gap: 6px; }
	.text-info { color: var(--info); }
	.text-indigo { color: var(--indigo); }
	.text-muted { color: var(--text-muted); }
</style>
