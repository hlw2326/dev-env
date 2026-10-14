<template>
	<!-- PHP 扩展管理弹窗 -->
	<h-modal
		v-model:visible="visible"
		:title="modalTitle"
		width="560px"
	>
		<div class="ext-body">
			<!-- 扩展数量与重启提示 -->
			<div class="ext-head">
				<span class="ext-count">共发现 {{ list.length }} 个可用扩展</span>
				<p-tag severity="info" value="修改 php.ini 后需重启对应服务生效" />
			</div>

			<!-- 加载中状态 -->
			<div v-if="loading" class="loading-box">
				<i class="pi pi-spin pi-spinner"></i>
				<span>正在加载扩展列表...</span>
			</div>

			<!-- 扩展卡片网格 -->
			<div v-else class="ext-grid">
				<div
					v-for="ext in list"
					:key="ext.name"
					class="ext-card"
					:class="{ enabled: ext.enabled }"
				>
					<div class="ext-info">
						<span class="ext-name font-mono">{{ ext.name }}</span>
						<span v-if="ext.enabled" class="ext-status font-mono">已启用</span>
						<span v-else class="ext-status font-mono">未开启</span>
					</div>
					<p-tag v-if="ext.enabled" severity="success" value="Active" />
					<p-tag v-else severity="secondary" value="Off" />
				</div>
			</div>
		</div>

		<template #footer>
			<p-button label="关闭" severity="secondary" outlined @click="visible = false" />
		</template>
	</h-modal>
</template>

<script setup lang="ts">
	import { computed } from 'vue'
	import type { RuntimeVersion, PHPExtension } from '@/types'

	// 双向绑定弹窗可见性
	const visible = defineModel<boolean>('visible', { default: false })

	// 接收运行时与扩展列表
	const props = defineProps<{
		runtime: RuntimeVersion | null
		list: PHPExtension[]
		loading: boolean
	}>()

	// 弹窗标题文本
	const modalTitle = computed(() => {
		if (props.runtime) {
			return `PHP ${props.runtime.version} 扩展管理`
		}
		return 'PHP 扩展管理'
	})
</script>

<style scoped lang="scss">
	/* 弹窗主体容器 */
	.ext-body {
		display: flex;
		flex-direction: column;
		gap: 14px;
	}

	/* 顶部统计行 */
	.ext-head {
		display: flex;
		justify-content: space-between;
		align-items: center;
	}

	/* 数量统计文字 */
	.ext-count {
		font-size: var(--font-sm);
		color: var(--text-secondary);
	}

	/* 加载中状态 */
	.loading-box {
		display: flex;
		align-items: center;
		justify-content: center;
		gap: 8px;
		height: 160px;
		color: var(--text-muted);
	}

	/* 扩展网格布局 */
	.ext-grid {
		display: grid;
		grid-template-columns: repeat(2, 1fr);
		gap: 10px;
		max-height: 320px;
		overflow-y: auto;
	}

	/* 单个扩展卡片 */
	.ext-card {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: 10px 12px;
		border: 1px solid var(--border-color);
		border-radius: var(--radius-sm);
		background: var(--bg-app);

		&.enabled {
			border-color: var(--ok-border);
			background: var(--ok-bg);
		}
	}

	/* 扩展信息列 */
	.ext-info {
		display: flex;
		flex-direction: column;
		gap: 2px;
	}

	/* 扩展名称 */
	.ext-name {
		font-size: var(--font-sm);
		color: var(--text-primary);
	}

	/* 扩展状态描述 */
	.ext-status {
		font-size: var(--font-xs);
		color: var(--text-muted);
	}
</style>
