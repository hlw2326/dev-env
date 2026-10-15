<template>
	<!-- 外观与存储设置卡片 -->
	<h-card title="外观与存储">
		<div class="settings-block">
			<!-- 色彩模式选择 -->
			<div class="form-field">
				<label>色彩模式（点击即生效）</label>
				<div class="theme-options">
					<button
						v-for="item in modeOptions"
						:key="item.value"
						type="button"
						class="theme-btn"
						:class="{ active: theme.mode === item.value }"
						@click="selectTheme(item.value)"
					>
						<i :class="[item.icon, 'theme-icon']"></i>
						<span>{{ item.label }}</span>
					</button>
				</div>
			</div>

			<!-- 主题强调色选择 -->
			<div class="form-field">
				<div class="color-label-row">
					<label>主题强调色</label>
					<span class="color-tag" :style="{ backgroundColor: theme.color }">
						{{ colorName }}
					</span>
				</div>
				<div class="color-options">
					<button
						v-for="item in THEME_COLORS"
						:key="item.value"
						type="button"
						class="color-wrap"
						:class="{ active: isSameColor(item.value) }"
						:title="item.name"
						@click="selectColor(item.value)"
					>
						<span class="color-dot" :style="{ backgroundColor: item.value }">
							<i v-if="isSameColor(item.value)" class="pi pi-check check-icon"></i>
						</span>
					</button>
					<label class="color-picker-btn" title="自定义强调色">
						<input
							type="color"
							:value="theme.color"
							class="color-input"
							@input="onPickColor"
						/>
						<i class="pi pi-plus plus-icon"></i>
					</label>
				</div>
			</div>

			<!-- 应用数据目录展示 -->
			<div class="form-field">
				<label>应用数据目录（只读）</label>
				<div class="folder-row">
					<p-input-text v-model="dataFolder" readonly class="font-mono flex-input" />
					<p-button
						severity="secondary"
						variant="outlined"
						icon="pi pi-folder-open"
						label="打开"
						@click="openData"
					/>
				</div>
				<span class="field-hint">存放数据库、已安装软件、MySQL 数据目录与运行日志</span>
			</div>
		</div>
	</h-card>
</template>

<script setup lang="ts">
import { useSettings } from '@/composables'
import type { ThemeMode } from '@/composables/theme'

// 外观与存储状态与方法
const {
	theme,
	THEME_COLORS,
	colorName,
	dataFolder,
	selectTheme,
	selectColor,
	openData,
} = useSettings()

// 色彩模式选项列表
const modeOptions: { label: string; value: ThemeMode; icon: string }[] = [
	{ label: '浅色模式', value: 'light', icon: 'pi pi-sun' },
	{ label: '深色模式', value: 'dark', icon: 'pi pi-moon' },
	{ label: '跟随系统', value: 'system', icon: 'pi pi-desktop' },
]

// 判断是否为当前选中强调色
function isSameColor(value: string): boolean {
	return theme.color.toLowerCase() === value.toLowerCase()
}

// 处理自定义拾色器输入
function onPickColor(event: Event): void {
	const target = event.target as HTMLInputElement
	selectColor(target.value)
}
</script>

<style scoped lang="scss">
/* 设置块容器 */
.settings-block {
	display: flex;
	flex-direction: column;
	gap: 0.75rem;
}

/* 色彩模式按钮组 */
.theme-options {
	display: flex;
	gap: 0.625rem;
}

/* 色彩模式切换按钮 */
.theme-btn {
	display: inline-flex;
	align-items: center;
	gap: 0.5rem;
	padding: 0.5rem 1rem;
	border: 1px solid var(--border-color);
	border-radius: var(--h-radius);
	background: var(--bg-app);
	color: var(--text-secondary);
	font-size: var(--font-sm);
	cursor: pointer;
	transition: all 0.15s ease;

	/* 悬停态 */
	&:hover {
		border-color: var(--border-hover);
		color: var(--text-primary);
	}

	/* 选中态 */
	&.active {
		background: var(--p-primary-color);
		color: var(--on-primary);
		border-color: var(--p-primary-color);
	}

	/* 模式图标 */
	.theme-icon {
		font-size: var(--icon-sm);
	}
}

/* 强调色标题行 */
.color-label-row {
	display: flex;
	align-items: center;
	gap: 0.5rem;
}

/* 当前强调色标签 */
.color-tag {
	font-size: var(--font-xs);
	color: var(--on-primary);
	padding: 0.1rem 0.45rem;
	border-radius: var(--h-radius);
}

/* 色板选项组 */
.color-options {
	display: flex;
	align-items: center;
	gap: 0.625rem;
	flex-wrap: wrap;
}

/* 单个色圈外框 */
.color-wrap {
	display: flex;
	align-items: center;
	justify-content: center;
	width: 2rem;
	height: 2rem;
	border-radius: 50%;
	cursor: pointer;
	border: 0.125rem solid transparent;
	background: transparent;

	/* 选中边框 */
	&.active {
		border-color: var(--text-primary);
	}

	/* 内部颜色圆点 */
	.color-dot {
		width: 1.375rem;
		height: 1.375rem;
		border-radius: 50%;
		display: flex;
		align-items: center;
		justify-content: center;
	}

	/* 选中勾选图标 */
	.check-icon {
		font-size: var(--icon-xs);
		color: var(--on-primary);
	}
}

/* 自定义拾色按钮 */
.color-picker-btn {
	width: 1.875rem;
	height: 1.875rem;
	border-radius: 50%;
	border: 1px dashed var(--border-color);
	display: flex;
	align-items: center;
	justify-content: center;
	cursor: pointer;
	position: relative;

	/* 隐藏原生拾色输入框 */
	.color-input {
		position: absolute;
		opacity: 0;
		width: 100%;
		height: 100%;
		cursor: pointer;
	}

	/* 加号图标 */
	.plus-icon {
		font-size: var(--icon-xs);
		color: var(--text-secondary);
	}
}

/* 目录行容器 */
.folder-row {
	display: flex;
	align-items: center;
	gap: 0.5rem;
}

/* 自适应输入框 */
.flex-input {
	flex: 1;
}

/* 字段说明文字 */
.field-hint {
	font-size: var(--font-sm);
	color: var(--text-secondary);
}
</style>
