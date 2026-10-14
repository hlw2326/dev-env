<template>
	<!-- 外观、主题与语言设置卡片 -->
	<h-card :title="$t('外观、主题与多语言')">
		<!-- 界面语言切换 -->
		<div class="setting-item">
			<label class="item-label">{{ $t('界面语言') }}</label>
			<div class="lang-group">
				<button type="button" class="lang-pill" :class="{ active: curLang === 'zh-CN' }" @click="emit('change-lang', 'zh-CN')">
					<i class="pi pi-globe"></i>
					<span>简体中文</span>
				</button>
				<button type="button" class="lang-pill" :class="{ active: curLang === 'en-US' }" @click="emit('change-lang', 'en-US')">
					<i class="pi pi-language"></i>
					<span>English</span>
				</button>
			</div>
		</div>

		<div class="divider"></div>

		<!-- 色彩模式选择 -->
		<div class="setting-item">
			<label class="item-label">{{ $t('色彩模式') }}</label>
			<div class="theme-options">
				<div class="theme-card" :class="{ active: theme.mode === 'light' }" @click="emit('change-mode', 'light')">
					<div class="theme-preview light-preview">
						<div class="bar"></div>
						<div class="line"></div>
					</div>
					<span class="theme-name">{{ $t('浅色') }}</span>
				</div>

				<div class="theme-card" :class="{ active: theme.mode === 'dark' }" @click="emit('change-mode', 'dark')">
					<div class="theme-preview dark-preview">
						<div class="bar"></div>
						<div class="line"></div>
					</div>
					<span class="theme-name">{{ $t('深色') }}</span>
				</div>

				<div class="theme-card" :class="{ active: theme.mode === 'system' }" @click="emit('change-mode', 'system')">
					<div class="theme-preview system-preview">
						<div class="half-light"></div>
						<div class="half-dark"></div>
					</div>
					<span class="theme-name">{{ $t('跟随系统') }}</span>
				</div>
			</div>
		</div>

		<div class="divider"></div>

		<!-- 主题强调色选择 -->
		<div class="setting-item">
			<div class="color-header">
				<label class="item-label">{{ $t('主题强调色') }}</label>
				<span class="color-tag" :style="{ backgroundColor: theme.color }">{{ colorName }}</span>
			</div>

			<div class="color-options">
				<div
					v-for="item in THEME_COLORS"
					:key="item.value"
					class="color-wrap"
					:class="{ active: theme.color.toLowerCase() === item.value.toLowerCase() }"
					:title="item.name"
					@click="setThemeColor(item.value)"
				>
					<div class="color-dot" :style="{ backgroundColor: item.value }">
						<i v-if="theme.color.toLowerCase() === item.value.toLowerCase()" class="pi pi-check check-icon"></i>
					</div>
				</div>

				<label class="color-btn" :title="$t('自定义')">
					<input type="color" :value="theme.color" class="color-input" @input="onColorPick" />
					<i class="pi pi-plus plus-icon"></i>
				</label>
			</div>
		</div>
	</h-card>
</template>

<script setup lang="ts">
	import { computed } from 'vue'
	import { $t, getLocale } from '@/i18n'
	import { useTheme, type ThemeMode } from '@/composables/theme'

	// 定义语言与模式切换事件
	const emit = defineEmits<{ 'change-lang': [lang: string]; 'change-mode': [mode: ThemeMode] }>()

	// 主题状态与色板
	const { theme, setThemeColor, THEME_COLORS } = useTheme()

	// 当前界面语言
	const curLang = computed(() => getLocale())

	// 当前强调色显示名称
	const colorName = computed(() => {
		const matched = THEME_COLORS.find((c) => c.value.toLowerCase() === theme.color.toLowerCase())
		if (matched) {
			return matched.name
		}
		return `${$t('自定义')} (${theme.color})`
	})

	// 自定义取色器回调
	function onColorPick(event: Event) {
		setThemeColor((event.target as HTMLInputElement).value)
	}
</script>

<style scoped lang="scss">
	/* 设置项与分割线 */
	.setting-item { display: flex; flex-direction: column; gap: 10px; }
	.item-label { font-size: var(--font-sm); font-weight: 500; color: var(--text-primary); }
	.divider { height: 1px; background: var(--border-color); margin: 12px 0; }

	/* 语言按钮组 */
	.lang-group { display: flex; gap: 10px; }
	.lang-pill {
		display: inline-flex; align-items: center; gap: 8px; padding: 8px 18px;
		border-radius: var(--radius-md); border: 1px solid var(--border-color);
		background: var(--bg-card); color: var(--text-secondary); font-size: var(--font-sm); cursor: pointer;
		&:hover { border-color: var(--p-primary-color); color: var(--text-primary); }
		&.active { border-color: var(--p-primary-color); background: var(--bg-app); color: var(--p-primary-color); font-weight: 500; }
	}

	/* 色彩模式选项组 */
	.theme-options { display: flex; gap: 1.25rem; }
	.theme-card {
		display: flex; flex-direction: column; align-items: center; gap: 0.5rem; cursor: pointer;
		.theme-preview {
			width: 6.5rem; height: 4rem; border-radius: var(--radius-sm);
			border: 1px solid var(--border-color); position: relative; overflow: hidden;
			.bar { height: 0.875rem; width: 100%; }
			.line { height: 0.375rem; width: 60%; margin: 0.75rem auto; border-radius: 2px; }
		}
		.light-preview { background: rgb(255, 255, 255); .bar { background: rgb(241, 245, 249); } .line { background: rgb(226, 232, 240); } }
		.dark-preview { background: rgb(15, 23, 42); .bar { background: rgb(30, 41, 59); } .line { background: rgb(51, 65, 85); } }
		.system-preview { display: flex; .half-light { width: 50%; height: 100%; background: rgb(255, 255, 255); } .half-dark { width: 50%; height: 100%; background: rgb(15, 23, 42); } }
		.theme-name { font-size: var(--font-xs); color: var(--text-secondary); }
		&.active {
			.theme-preview { border-color: var(--p-primary-color); box-shadow: 0 0 0 2px var(--p-primary-color); }
			.theme-name { color: var(--p-primary-color); font-weight: 500; }
		}
	}

	/* 强调色选择区 */
	.color-header {
		display: flex; align-items: center; gap: 10px;
		.color-tag { font-size: var(--font-xs); color: var(--on-primary); padding: 2px 8px; border-radius: 4px; font-weight: 500; }
	}
	.color-options {
		display: flex; align-items: center; gap: 0.75rem; flex-wrap: wrap;
		.color-wrap {
			display: flex; align-items: center; justify-content: center; width: 2rem; height: 2rem;
			border-radius: 50%; cursor: pointer; border: 2px solid transparent;
			&.active { border-color: var(--text-primary); }
			.color-dot {
				width: 1.375rem; height: 1.375rem; border-radius: 50%; display: flex; align-items: center; justify-content: center;
				.check-icon { font-size: 0.75rem; color: var(--on-primary); }
			}
		}
		.color-btn {
			width: 1.875rem; height: 1.875rem; border-radius: 50%; border: 1px dashed var(--border-color);
			display: flex; align-items: center; justify-content: center; cursor: pointer; position: relative;
			.color-input { position: absolute; opacity: 0; width: 100%; height: 100%; cursor: pointer; }
			.plus-icon { font-size: 0.75rem; color: var(--text-secondary); }
		}
	}
</style>
