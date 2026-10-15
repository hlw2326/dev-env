<template>
	<!-- 窗口控制按钮组 -->
	<div class="window-controls">
		<!-- 最小化按钮 -->
		<p-button
			type="button"
			unstyled
			class="control-btn"
			title="最小化"
			@click="onMin"
		>
			<i class="pi pi-minus icon"></i>
		</p-button>
		<!-- 最大化与还原按钮 -->
		<p-button
			type="button"
			unstyled
			class="control-btn"
			:title="maxTip"
			@click="onToggle"
		>
			<i v-if="maximized" class="pi pi-clone icon"></i>
			<i v-else class="pi pi-stop icon"></i>
		</p-button>
		<!-- 隐藏到托盘按钮 -->
		<p-button
			type="button"
			unstyled
			class="control-btn danger"
			title="隐藏到托盘"
			@click="onHide"
		>
			<i class="pi pi-times icon"></i>
		</p-button>
	</div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { minWin, toggleWin, checkWin, hideWin } from '@/api'
import { toast, formatError } from '@/utils'

// 窗口是否处于最大化状态
const maximized = ref(false)

// 最大化按钮提示文案
const maxTip = computed(() => {
	if (maximized.value) {
		return '向下还原'
	}
	return '最大化'
})

// 同步窗口最大化状态
async function syncState(): Promise<void> {
	try {
		maximized.value = await checkWin()
	} catch (err: unknown) {
		toast.error('读取窗口状态失败', formatError(err))
	}
}

// 最小化窗口
function onMin(): void {
	minWin()
}

// 切换最大化或还原窗口
async function onToggle(): Promise<void> {
	toggleWin()
	await syncState()
}

// 隐藏窗口并继续在托盘运行
function onHide(): void {
	hideWin()
}

// 挂载时同步窗口状态
onMounted(() => {
	syncState()
})
</script>

<style scoped lang="scss">
/* 窗口按钮组容器 */
.window-controls {
	display: flex;
	align-items: center;
	height: 100%;
	--wails-draggable: no-drag;

	/* 单个控制按钮 */
	.control-btn {
		width: 2.75rem;
		height: 100%;
		border: none;
		background: transparent;
		color: var(--text-secondary);
		display: inline-flex;
		align-items: center;
		justify-content: center;
		cursor: pointer;
		transition: all 0.15s ease;

		/* 悬停态 */
		&:hover {
			background: var(--bg-app);
			color: var(--text-primary);
		}

		/* 关闭按钮悬停态 */
		&.danger:hover {
			background: var(--err);
			color: var(--on-primary);
		}

		/* 按钮图标 */
		.icon {
			font-size: var(--icon-xs);
		}
	}
}
</style>
