<template>
	<!-- 单个服务实例卡片 -->
	<div
		class="srv-card surface-card"
		:class="{ 'is-running': item.status === 'running' }"
	>
		<!-- 服务标题与运行状态 -->
		<div class="srv-header">
			<div class="srv-title">
				<span class="type-pill" :class="item.serviceType">
					{{ item.serviceType.toUpperCase() }}
				</span>
				<span class="srv-name">{{ item.name }}</span>
			</div>

			<div class="header-right">
				<p-tag
					v-if="item.status === 'running'"
					severity="success"
					value="运行中"
					icon="pi pi-check-circle"
				/>
				<p-tag
					v-else
					severity="secondary"
					value="已停止"
					icon="pi pi-minus-circle"
				/>
				<p-button
					icon="pi pi-trash"
					size="small"
					severity="danger"
					text
					rounded
					title="删除服务实例"
					:loading="busy"
					@click="emit('remove', $event, item)"
				/>
			</div>
		</div>

		<!-- 运行版本切换选择器 -->
		<div class="srv-version-row">
			<span class="ver-label">版本</span>
			<div
				ref="verWrapRef"
				class="ver-select-wrap"
				:title="options.length <= 1 ? '在软件商店安装更多版本后可在此切换' : '切换服务运行版本'"
			>
				<p-select
					:model-value="item.runtimeId"
					:options="selectOptions"
					option-label="label"
					option-value="value"
					size="small"
					class="ver-select"
					:disabled="busy || options.length <= 1"
					@update:model-value="onVersionChange"
				/>
			</div>
			<span
				v-if="options.length <= 1"
				class="ver-store-link"
				@click="router.push('/store')"
			>
				安装更多版本
			</span>
		</div>

		<!-- 服务端口、进程与自启信息 -->
		<div class="srv-metrics">
			<div class="metric-col">
				<span class="m-lbl">监听端口</span>
				<span v-if="item.port" class="m-val font-mono">{{ item.port }}</span>
				<span v-else class="m-val font-mono">无</span>
			</div>
			<div class="metric-col">
				<span class="m-lbl">进程 PID</span>
				<span v-if="item.pid" class="m-val font-mono">{{ item.pid }}</span>
				<span v-else class="m-val font-mono">-</span>
			</div>
			<div class="metric-col">
				<span class="m-lbl">开机自启</span>
				<span v-if="item.autoStart" class="m-val">开启</span>
				<span v-else class="m-val">关闭</span>
			</div>
		</div>

		<!-- 服务操作按钮栏 -->
		<div class="srv-actions">
			<p-button
				v-if="item.status !== 'running'"
				label="启动服务"
				icon="pi pi-play"
				severity="success"
				class="flex-1"
				:loading="busy"
				@click="emit('start', item.id)"
			/>

			<template v-else>
				<p-button
					label="停止"
					icon="pi pi-stop"
					severity="danger"
					outlined
					class="flex-1"
					:loading="busy"
					@click="emit('stop', item.id)"
				/>
				<p-button
					label="重启"
					icon="pi pi-replay"
					severity="secondary"
					outlined
					class="flex-1"
					:loading="busy"
					@click="emit('restart', item.id)"
				/>
			</template>

			<p-button
				icon="pi pi-terminal"
				label="日志"
				severity="secondary"
				text
				@click="emit('logs', item)"
			/>
		</div>
	</div>
</template>

<script setup lang="ts">
	import { ref, computed } from 'vue'
	import { useRouter } from 'vue-router'
	import type { ServiceInstance, RuntimeOption } from '@/types'

	const router = useRouter()
	const verWrapRef = ref<HTMLElement | null>(null)

	// 接收服务实例数据、可选版本列表与繁忙状态
	const props = defineProps<{
		item: ServiceInstance
		options: RuntimeOption[]
		busy: boolean
	}>()

	// 定义服务控制与版本切换事件
	const emit = defineEmits<{
		start: [id: number]
		stop: [id: number]
		restart: [id: number]
		logs: [item: ServiceInstance]
		remove: [event: Event, item: ServiceInstance]
		switchVersion: [item: ServiceInstance, runtimeId: number, targetEl?: HTMLElement | null]
	}>()

	// 构建下拉框选项列表
	const selectOptions = computed(() => {
		if (props.options && props.options.length > 0) {
			return props.options.map((opt) => ({
				label: `${opt.version} (${opt.architecture})`,
				value: opt.runtimeId
			}))
		}
		return [
			{
				label: props.item.runtimeVersion
					? `${props.item.runtimeVersion}`
					: `Runtime #${props.item.runtimeId}`,
				value: props.item.runtimeId
			}
		]
	})

	// 触发版本切换
	function onVersionChange(newRuntimeId: number) {
		if (!newRuntimeId || newRuntimeId === props.item.runtimeId) {
			return
		}
		emit('switchVersion', props.item, newRuntimeId, verWrapRef.value)
	}
</script>

<style scoped lang="scss">
	/* 服务卡片容器 */
	.srv-card {
		padding: 20px;
		display: flex;
		flex-direction: column;
		gap: 14px;
		background: var(--bg-card);
		border: 1px solid var(--border-color);
		border-radius: var(--radius-lg);

		&.is-running {
			border-color: var(--ok-border);
		}
	}

	/* 顶部状态行 */
	.srv-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
	}

	/* 服务标题组合与右侧操作 */
	.srv-title,
	.header-right {
		display: flex;
		align-items: center;
		gap: 8px;
	}

	/* 服务类型标签 */
	.type-pill {
		padding: 3px 8px;
		border-radius: var(--radius-sm);
		font-size: var(--font-xs);
		background: var(--bg-app);
		color: var(--text-secondary);
		border: 1px solid var(--border-color);

		&.php {
			background: var(--info-bg);
			color: var(--info);
			border-color: var(--info-border);
		}

		&.nginx {
			background: var(--ok-bg);
			color: var(--ok);
			border-color: var(--ok-border);
		}

		&.mysql {
			background: var(--warn-bg);
			color: var(--warn);
			border-color: var(--warn-border);
		}

		&.redis {
			background: var(--err-bg);
			color: var(--err);
			border-color: var(--err-border);
		}
	}

	/* 服务名称 */
	.srv-name {
		font-size: var(--font-base);
		font-weight: 600;
		color: var(--text-primary);
	}

	/* 版本选择行 */
	.srv-version-row {
		display: flex;
		align-items: center;
		gap: 10px;
	}

	/* 版本文字标签 */
	.ver-label {
		font-size: var(--font-xs);
		color: var(--text-muted);
		flex-shrink: 0;
	}

	/* 下拉框容器 */
	.ver-select-wrap {
		flex: 1;
		min-width: 0;
	}

	/* 版本下拉框 */
	.ver-select {
		width: 100%;
	}

	/* 前往软件商店安装更多版本链接 */
	.ver-store-link {
		font-size: var(--font-xs);
		color: var(--primary);
		cursor: pointer;
		flex-shrink: 0;
		user-select: none;

		&:hover {
			text-decoration: underline;
		}
	}

	/* 指标网格 */
	.srv-metrics {
		display: grid;
		grid-template-columns: repeat(3, 1fr);
		gap: 8px;
		background: var(--bg-app);
		padding: 10px 12px;
		border-radius: var(--radius-md);
		border: 1px solid var(--border-color);
	}

	/* 单个指标列 */
	.metric-col {
		display: flex;
		flex-direction: column;
	}

	/* 指标标题 */
	.m-lbl {
		font-size: var(--font-xs);
		color: var(--text-muted);
	}

	/* 指标数值 */
	.m-val {
		font-size: var(--font-sm);
		color: var(--text-primary);
		margin-top: 2px;
	}

	/* 底部操作行 */
	.srv-actions {
		display: flex;
		gap: 8px;
		margin-top: auto;
	}

	/* 自适应宽度 */
	.flex-1 {
		flex: 1;
	}
</style>

