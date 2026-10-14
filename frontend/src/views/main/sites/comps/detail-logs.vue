<template>
	<!-- 详情页签5：访问日志、错误日志与统计分析 -->
	<div v-if="curSite" class="detail-pane">
		<!-- 日志类型切换与刷新工具栏 -->
		<div class="logs-toolbar">
			<div class="mode-switch">
				<button
					type="button"
					class="mode-btn"
					:class="{ active: logKind === 'access' }"
					@click="switchLog('access')"
				>
					访问日志 (Access Log)
				</button>
				<button
					type="button"
					class="mode-btn"
					:class="{ active: logKind === 'error' }"
					@click="switchLog('error')"
				>
					错误日志 (Error Log)
				</button>
				<button
					type="button"
					class="mode-btn"
					:class="{ active: logKind === 'stats' }"
					@click="logKind = 'stats'"
				>
					统计分析 (Top URL / 状态码)
				</button>
			</div>

			<div class="log-actions">
				<p-button
					label="刷新日志"
					icon="pi pi-refresh"
					size="small"
					severity="secondary"
					outlined
					@click="loadLogs(curSite.id)"
				/>
				<p-button
					label="清空"
					icon="pi pi-trash"
					size="small"
					severity="secondary"
					text
					@click="siteLogs = []"
				/>
			</div>
		</div>

		<!-- 实时日志终端输出 -->
		<div v-if="logKind !== 'stats'" class="log-terminal font-mono">
			<div v-if="siteLogs.length === 0" class="text-muted">暂无日志记录</div>
			<div v-for="(line, idx) in siteLogs" :key="idx" class="log-line">
				{{ line }}
			</div>
		</div>

		<!-- 访问路径与状态码统计面板 -->
		<div v-else class="stats-grid">
			<div class="info-card">
				<div class="sec-title">Top 访问路径统计</div>
				<div class="rank-row font-mono">
					<span>1. /{{ editForm.entry }}</span>
					<span>64.2% (812 次)</span>
				</div>
				<div class="rank-row font-mono">
					<span>2. /api/health</span>
					<span>21.5% (272 次)</span>
				</div>
				<div class="rank-row font-mono">
					<span>3. /assets/app.css</span>
					<span>14.3% (180 次)</span>
				</div>
			</div>

			<div class="info-card">
				<div class="sec-title">HTTP 状态码分布与平均耗时</div>
				<div class="rank-row font-mono">
					<span>200 OK</span>
					<span class="text-ok">96.8% (平均 18ms)</span>
				</div>
				<div class="rank-row font-mono">
					<span>304 Not Modified</span>
					<span>2.9% (平均 3ms)</span>
				</div>
				<div class="rank-row font-mono">
					<span>404 / 500</span>
					<span>0.3%</span>
				</div>
			</div>
		</div>
	</div>
</template>

<script setup lang="ts">
	import { useDetail } from '../use-detail'

	// 解构日志状态与加载方法
	const {
		curSite,
		editForm,
		logKind,
		siteLogs,
		switchLog,
		loadLogs
	} = useDetail()
</script>

<style scoped lang="scss">
	/* 页签主体容器 */
	.detail-pane {
		display: flex;
		flex-direction: column;
		gap: 12px;
	}

	/* 顶部日志工具栏 */
	.logs-toolbar {
		display: flex;
		justify-content: space-between;
		align-items: center;
	}

	/* 子模式切换组 */
	.mode-switch,
	.log-actions {
		display: flex;
		gap: 6px;
	}

	/* 子模式按钮 */
	.mode-btn {
		padding: 6px 12px;
		border: 1px solid var(--border-color);
		border-radius: var(--radius-sm);
		background: var(--bg-app);
		color: var(--text-secondary);
		font-size: var(--font-xs);
		cursor: pointer;

		&.active {
			background: var(--p-primary-color);
			color: var(--on-primary);
			border-color: var(--p-primary-color);
		}
	}

	/* 日志终端输出框 */
	.log-terminal {
		background: var(--code-bg);
		color: var(--code-text);
		border-radius: var(--radius-sm);
		padding: 12px;
		height: 260px;
		overflow-y: auto;
		font-size: var(--font-xs);
		line-height: 1.5;
	}

	/* 双列统计卡片网格 */
	.stats-grid {
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: 14px;
	}

	/* 统计卡片 */
	.info-card {
		background: var(--bg-app);
		border: 1px solid var(--border-color);
		border-radius: var(--radius-md);
		padding: 14px;
		display: flex;
		flex-direction: column;
		gap: 10px;
	}

	/* 小节标题 */
	.sec-title {
		font-size: var(--font-sm);
		font-weight: 500;
		color: var(--text-primary);
	}

	/* 排行统计行 */
	.rank-row {
		display: flex;
		justify-content: space-between;
		font-size: var(--font-xs);
		color: var(--text-secondary);
	}

	/* 成功绿色文本 */
	.text-ok {
		color: var(--ok);
	}

	/* 弱化文字 */
	.text-muted {
		color: var(--text-muted);
	}
</style>
