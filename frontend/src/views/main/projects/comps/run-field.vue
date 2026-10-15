<template>
	<!-- 运行目录选择与路径预览字段 -->
	<div class="form-field">
		<label>运行目录（Nginx 文档根）</label>
		<div class="row-line">
			<p-select
				v-model="runPreset"
				:options="runOptions"
				option-label="label"
				option-value="value"
				class="flex-input"
				@change="onPreset"
			/>
			<p-input-text
				v-if="runPreset === 'custom'"
				v-model="customRun"
				placeholder="相对子目录，如 public"
				class="font-mono flex-input"
				@input="onCustom"
			/>
		</div>
		<span class="hint-text font-mono">文档根路径：{{ previewPath }}</span>
	</div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { resolvePath } from '@/utils'

// 接收项目根目录路径
const props = defineProps<{
	root: string
}>()

// 相对运行目录双向绑定
const run = defineModel<string>({ required: true })

// 运行目录变更通知事件
const emit = defineEmits<{
	change: []
}>()

// 运行目录预设列表
const runOptions = [
	{ value: '', label: '/ (根目录本身)' },
	{ value: 'public', label: '/public' },
	{ value: 'web', label: '/web' },
	{ value: 'dist', label: '/dist' },
	{ value: 'custom', label: '自定义子目录' },
]

// 预设下拉选中值
const runPreset = ref('')

// 自定义子目录输入值
const customRun = ref('')

// 生效文档根完整路径预览
const previewPath = computed(() => resolvePath(props.root, run.value))

// 切换预设下拉项
function onPreset(): void {
	if (runPreset.value === 'custom') {
		run.value = customRun.value.trim().replace(/^[/\\]+/, '')
	} else {
		customRun.value = ''
		run.value = runPreset.value
		emit('change')
	}
}

// 输入自定义运行子目录
function onCustom(): void {
	run.value = customRun.value.trim().replace(/^[/\\]+/, '')
}

// 同步外部运行目录值到下拉与输入框
watch(
	run,
	(value) => {
		if (value === '' || value === 'public' || value === 'web' || value === 'dist') {
			runPreset.value = value
			customRun.value = ''
		} else {
			runPreset.value = 'custom'
			customRun.value = value
		}
	},
	{ immediate: true },
)
</script>

<style scoped lang="scss">
/* 水平输入行 */
.row-line {
	display: flex;
	align-items: center;
	gap: 0.5rem;
}

/* 自适应输入框 */
.flex-input {
	flex: 1;
}

/* 运行目录选择器沿用表单字号，避免 PrimeVue 默认字号偏大 */
.row-line :deep(.p-select),
.row-line :deep(.p-select-label),
.row-line :deep(.p-inputtext) {
	font-size: var(--font-sm) !important;
}

/* 路径预览说明 */
.hint-text {
	font-size: var(--font-xs);
	color: var(--text-muted);
}
</style>
