<template>
	<header class="titlebar">
		<div class="left">
			<div class="logo">
				<i class="pi pi-box"></i>
			</div>
			<span class="brand">DevEnv</span>
			<span class="sep">/</span>
			<span class="page font-mono">{{ routeName }}</span>
		</div>

		<div class="center" @dblclick="toggleMaxWin">
			<span class="title">本地集成开发环境管理平台</span>
		</div>

		<div class="right">
			<button
				type="button"
				class="theme-toggle"
				:title="modeTip"
				@click="toggleTheme"
			>
				<i v-if="theme.mode === 'dark'" class="pi pi-sun toggle-icon"></i>
				<i v-else class="pi pi-moon toggle-icon"></i>
			</button>
			<span class="env font-mono">Win64 · Go & Wails</span>
			<win-btns />
		</div>
	</header>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import { toggleMaxWin, loadSettings, saveSettings } from '@/api'
import { useTheme } from '@/composables/theme'
import type { ThemeMode } from '@/composables/theme'
import WinBtns from './win-btns.vue'

const route = useRoute()
const { theme, setThemeMode } = useTheme()

const routeName = computed(() => {
	if (route.meta.title) {
		return String(route.meta.title)
	}
	return String(route.name)
})

const modeTip = computed(() => {
	if (theme.mode === 'dark') {
		return '切换到浅色模式'
	}
	return '切换到深色模式'
})

async function toggleTheme() {
	let next: ThemeMode = 'light'
	if (theme.mode === 'light') {
		next = 'dark'
	}
	setThemeMode(next)
	try {
		const cfg = await loadSettings()
		await saveSettings({
			theme: next,
			root: cfg.root,
			port: cfg.port,
			hosts: cfg.hosts,
			autostart: cfg.autostart,
		})
	} catch {
		// ignore persistence error
	}
}
</script>

<style scoped lang="scss">
.titlebar {
	height: 2.375rem;
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

.left {
	display: flex;
	align-items: center;
	gap: 0.5rem;
	--wails-draggable: drag;

	.logo {
		width: 1.375rem;
		height: 1.375rem;
		border-radius: var(--radius-sm);
		background: var(--p-primary-color);
		color: var(--on-primary);
		display: flex;
		align-items: center;
		justify-content: center;
		font-size: var(--icon-xs);
	}

	.brand {
		font-size: var(--font-base);
		color: var(--text-primary);
	}

	.sep {
		color: var(--border-hover);
		font-size: var(--font-sm);
	}

	.page {
		font-size: var(--font-sm);
		color: var(--text-secondary);
	}
}

.center {
	flex: 1;
	height: 100%;
	display: flex;
	align-items: center;
	justify-content: center;
	--wails-draggable: drag;

	.title {
		font-size: var(--font-xs);
		color: var(--text-muted);
	}
}

.right {
	display: flex;
	align-items: center;
	height: 100%;
	gap: 0.5rem;

	.theme-toggle {
		width: 1.625rem;
		height: 1.625rem;
		border-radius: var(--radius-sm);
		border: 1px solid var(--border-color);
		background: var(--bg-app);
		color: var(--text-secondary);
		display: inline-flex;
		align-items: center;
		justify-content: center;
		cursor: pointer;
		--wails-draggable: no-drag;

		&:hover {
			border-color: var(--border-hover);
			color: var(--text-primary);
		}

		.toggle-icon {
			font-size: var(--icon-xs);
		}
	}

	.env {
		font-size: var(--font-xs);
		color: var(--text-secondary);
		background: var(--bg-app);
		padding: 0.125rem 0.5rem;
		border-radius: var(--radius-sm);
		border: 1px solid var(--border-color);
		--wails-draggable: no-drag;
	}
}
</style>
