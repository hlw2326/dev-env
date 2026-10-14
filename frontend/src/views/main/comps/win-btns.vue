<template>
	<div class="win-controls">
		<!-- 最小化按钮 -->
		<button
			type="button"
			class="win-btn"
			title="最小化"
			@click="minWin"
		>
			<i class="pi pi-minus"></i>
		</button>
		<!-- 最大化/还原按钮 -->
		<button
			type="button"
			class="win-btn"
			:title="maxTitle"
			@click="onToggleMax"
		>
			<i v-if="isMax" class="pi pi-clone"></i>
			<i v-else class="pi pi-stop"></i>
		</button>
		<!-- 关闭按钮 -->
		<button
			type="button"
			class="win-btn close"
			title="关闭"
			@click="quitWin"
		>
			<i class="pi pi-times"></i>
		</button>
	</div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { minWin, toggleMaxWin, isMaxWin, quitWin } from '@/api'

// 是否最大化状态
const isMax = ref(false)

// 最大化按钮提示文案
const maxTitle = computed(() => {
	if (isMax.value) {
		return '还原'
	}
	return '最大化'
})

// 同步窗口最大化状态
async function syncMax() {
	isMax.value = await isMaxWin()
}

// 切换最大化状态
function onToggleMax() {
	toggleMaxWin()
	setTimeout(syncMax, 150)
}

// 挂载监听窗口变化
onMounted(() => {
	syncMax()
	window.addEventListener('resize', syncMax)
})

// 卸载移除监听
onUnmounted(() => {
	window.removeEventListener('resize', syncMax)
})
</script>

<style scoped lang="scss">
/* 窗口控制按钮组 */
.win-controls {
	display: flex;
	height: 100%;
	--wails-draggable: no-drag;
}

/* 单个窗口按钮 */
.win-btn {
	width: 2.75rem;
	height: 100%;
	border: none;
	background: transparent;
	color: var(--text-secondary);
	display: flex;
	align-items: center;
	justify-content: center;
	cursor: pointer;
	font-size: var(--font-xs);

	&:hover {
		background-color: var(--bg-app);
		color: var(--text-primary);
	}

	&.close:hover {
		background-color: var(--danger-color);
		color: var(--on-primary);
	}
}
</style>
