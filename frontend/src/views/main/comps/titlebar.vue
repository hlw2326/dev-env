<template>
	<!-- 顶部自定义标题栏 -->
	<header class="titlebar">
		<!-- 左侧品牌与当前路由名称 -->
		<div class="left">
			<span
				class="logo"
				role="img"
				aria-label="DevEnv"
				:style="{ display: 'inline-flex', color: 'var(--brand-color)' }"
				v-html="appIcon"
			></span>
			<span class="brand">DevEnv</span>
			<span class="separator">/</span>
			<span class="page font-mono">{{ routeName }}</span>
		</div>

		<!-- 中间拖拽区与平台副标题 -->
		<div class="center" @dblclick="onToggle">
			<span class="title">本地集成开发环境管理平台</span>
		</div>

		<!-- 右侧主题切换与窗口控制 -->
		<div class="right">
			<p-button
				type="button"
				unstyled
				class="theme-toggle"
				:title="modeTip"
				@click="toggleTheme"
			>
				<i v-if="theme.mode === 'dark'" class="pi pi-sun toggle-icon"></i>
				<i v-else class="pi pi-moon toggle-icon"></i>
			</p-button>
			<comp-window />
		</div>
	</header>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import { toggleWin, loadSettings, saveSettings } from '@/api'
import { useTheme } from '@/composables/theme'
import { toast, formatError } from '@/utils'
import devLogo from '@/assets/dev.svg'
import CompWindow from './window.vue'
import appIcon from '@/assets/dev.svg?raw'
import type { ThemeMode } from '@/composables/theme'

// 当前路由对象
const route = useRoute()

// 主题状态与切换方法
const { theme, setMode } = useTheme()

// 当前页面名称
const routeName = computed(() => {
	if (route.meta.title) {
		return String(route.meta.title)
	}
	return String(route.name)
})

// 主题切换提示文字
const modeTip = computed(() => {
	if (theme.mode === 'dark') {
		return '切换到浅色模式'
	}
	return '切换到深色模式'
})

// 双击标题栏切换窗口最大化
function onToggle(): void {
	toggleWin()
}

// 切换明暗主题并持久化
async function toggleTheme(): Promise<void> {
	let next: ThemeMode = 'light'
	if (theme.mode === 'light') {
		next = 'dark'
	}
	setMode(next)
	try {
		const settings = await loadSettings()
		await saveSettings({
			theme: next,
			root: settings.root,
			port: settings.port,
			hosts: settings.hosts,
			developerToolsCLI: settings.developerToolsCLI,
		})
	} catch (err: unknown) {
		toast.error('保存主题设置失败', formatError(err))
	}
}
</script>

<style scoped lang="scss">
/* 标题栏容器 */
.titlebar {
	height: 2.75rem;
	min-height: 2.75rem;
	background: var(--bg-card);
	border-bottom: 1px solid var(--border-color);
	display: flex;
	align-items: center;
	justify-content: space-between;
	padding-left: 0.625rem;
	user-select: none;
	--wails-draggable: drag;
	flex-shrink: 0;
	z-index: 1000;
}

/* 左侧品牌信息区 */
.left {
	display: flex;
	align-items: center;
	gap: 0.5rem;
	--wails-draggable: drag;

	/* 品牌图标 */
	.logo {
		width: 1.375rem;
		height: 1.375rem;
	}

	/* 品牌名称 */
	.brand {
		font-size: var(--font-base);
		color: var(--text-primary);
	}

	/* 分隔斜杠 */
	.separator {
		color: var(--border-hover);
		font-size: var(--font-sm);
	}

	/* 当前页面标题 */
	.page {
		font-size: var(--font-sm);
		color: var(--text-secondary);
	}
}

:global(html.dark) .logo-img {
	filter: brightness(0) invert(1);
}

/* 中间拖拽区域 */
.center {
	flex: 1;
	height: 100%;
	display: flex;
	align-items: center;
	justify-content: center;
	--wails-draggable: drag;

	/* 平台描述文本 */
	.title {
		font-size: var(--font-xs);
		color: var(--text-muted);
	}
}

/* 右侧工具与窗口按钮区 */
.right {
	display: flex;
	align-items: center;
	height: 100%;
	gap: 0.5rem;

	/* 主题切换按钮 */
	.theme-toggle {
		width: 1.625rem;
		height: 1.625rem;
		border-radius: var(--h-radius);
		border: 1px solid var(--border-color);
		background: var(--bg-app);
		color: var(--text-secondary);
		display: inline-flex;
		align-items: center;
		justify-content: center;
		cursor: pointer;
		--wails-draggable: no-drag;

		/* 按钮悬停态 */
		&:hover {
			border-color: var(--border-hover);
			color: var(--text-primary);
		}

		/* 切换图标 */
		.toggle-icon {
			font-size: var(--icon-xs);
		}
	}

}
</style>
