<template>
	<!-- 选择软件版本安装弹窗 -->
	<h-modal
		v-if="item !== null"
		v-model:visible="visible"
		:title="`安装 ${item.name}`"
		width="34rem"
		:modal="true"
		:closable="true"
		:mask-close="false"
		:footer="false"
		ok-text="确定"
		ok-icon="pi pi-check"
		cancel-text="取消"
		:loading="false"
	>
		<div class="install-modal">
			<!-- 加载版本清单状态 -->
			<div v-if="loading" class="state-box">
				<i class="pi pi-spin pi-spinner state-icon"></i>
				<span class="state-text">正在从官方源获取版本清单...</span>
			</div>

			<!-- 获取版本清单失败状态 -->
			<div v-else-if="error !== ''" class="state-box error">
				<span class="error-text">{{ error }}</span>
				<p-button
					size="small"
					icon="pi pi-refresh"
					label="重试"
					@click="emit('retry', item.kind)"
				/>
			</div>

			<!-- 可安装版本列表 -->
			<div v-else class="version-list">
				<comp-version-row
					v-for="row in versions"
					:key="row.version"
					:row="row"
					:task="findTask(row.version)"
					:busy="isBusy"
					@install="emit('install', item.kind, $event)"
					@cancel="emit('cancel', item.kind, $event)"
				/>
			</div>
		</div>
	</h-modal>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import CompVersionRow from './version-row.vue'
import type { MarketItem, Software, SoftwareVersion } from '@/types'

// 接收商店项与版本列表属性
const props = defineProps<{
	item: MarketItem | null
	versions: SoftwareVersion[]
	loading: boolean
	error: string
	progressMap: Record<string, { progress: number; message: string }>
	installedList: Software[]
}>()

// 弹窗显隐双向绑定
const visible = defineModel<boolean>('visible', { required: true })

// 定义安装、取消与重试事件
const emit = defineEmits<{
	install: [kind: string, version: string]
	cancel: [kind: string, version: string]
	retry: [kind: string]
}>()

// 当前软件是否有正在进行的安装任务
const isBusy = computed(() => {
	const current = props.item
	if (!current) {
		return false
	}
	for (const key of Object.keys(props.progressMap)) {
		if (key.startsWith(`${current.kind}:`)) {
			return true
		}
	}
	return props.installedList.some(
		(entry) => entry.kind === current.kind && entry.state === 'installing',
	)
})

// 查找指定版本的实时安装进度
function findTask(version: string): { progress: number; message: string } | null {
	const current = props.item
	if (!current) {
		return null
	}
	const key = `${current.kind}:${version}`
	if (props.progressMap[key]) {
		return props.progressMap[key]
	}
	const found = props.installedList.find(
		(entry) => entry.kind === current.kind && entry.version === version && entry.state === 'installing',
	)
	if (found) {
		return { progress: found.progress, message: '正在准备下载...' }
	}
	return null
}
</script>

<style scoped lang="scss">
/* 安装弹窗容器 */
.install-modal {
	display: flex;
	flex-direction: column;
	gap: 0.75rem;
}

/* 加载或错误状态盒 */
.state-box {
	display: flex;
	flex-direction: column;
	align-items: center;
	justify-content: center;
	padding: 2rem 1rem;
	gap: 0.75rem;

	/* 状态图标 */
	.state-icon {
		font-size: var(--icon-lg);
		color: var(--text-secondary);
	}

	/* 状态说明文本 */
	.state-text {
		font-size: var(--font-sm);
		color: var(--text-secondary);
	}

	/* 错误态样式 */
	&.error {
		background: var(--err-bg);
		border: 1px solid var(--err-border);
		border-radius: var(--h-radius);

		/* 错误详情文字 */
		.error-text {
			font-size: var(--font-sm);
			color: var(--err);
			text-align: center;
		}
	}
}

/* 版本滚动列表 */
.version-list {
	display: flex;
	flex-direction: column;
	gap: 0.625rem;
	max-height: 24rem;
	overflow-y: auto;
	padding-right: 0.25rem;
}
</style>
