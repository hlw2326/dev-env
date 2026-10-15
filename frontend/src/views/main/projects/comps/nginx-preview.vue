<template>
	<section class="nginx-preview">
		<div class="preview-toolbar">
			<div>
				<strong>完整 Nginx 配置</strong>
				<span class="hint">系统根据项目、域名和附加配置生成，内容只读</span>
			</div>
			<div class="preview-actions">
				<p-button
					v-if="!editing"
					severity="secondary"
					variant="outlined"
					icon="pi pi-refresh"
					label="刷新配置"
					:loading="loading"
					@click="loadConfig(true)"
				/>
				<p-button
					v-if="!editing"
					severity="secondary"
					variant="outlined"
					icon="pi pi-pencil"
					label="编辑"
					:disabled="config === ''"
					@click="beginEdit"
				/>
				<template v-else>
					<p-button severity="secondary" variant="outlined" label="取消" @click="cancelEdit" />
					<p-button icon="pi pi-save" label="保存" :loading="saving" @click="saveConfig" />
				</template>
			</div>
		</div>
		<div v-if="config !== ''" class="preview-shell">
			<div class="line-numbers" aria-hidden="true">
				<span v-for="line in lineCount" :key="line">{{ line }}</span>
			</div>
			<textarea
				v-model="draft"
				class="preview-editor font-mono"
				:readonly="!editing"
				wrap="off"
				aria-label="完整 Nginx 配置"
				@keydown="handleEditorKeydown"
			/>
		</div>
		<div v-else class="preview-empty">
			<i class="pi pi-info-circle" />
			<span>{{ message || '项目尚未生成 Nginx 配置，请先启动项目' }}</span>
		</div>
	</section>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { readProjectNginxConfig, writeProjectNginxConfig } from '@/api'
import { formatError, toast } from '@/utils'

const props = defineProps<{ projectId: number }>()
const config = ref('')
const draft = ref('')
const message = ref('')
const loading = ref(false)
const editing = ref(false)
const saving = ref(false)
const lineCount = computed(() => Math.max(1, draft.value.split('\n').length))

async function loadConfig(showToast = false): Promise<void> {
	loading.value = true
	message.value = ''
	try {
		config.value = await readProjectNginxConfig(props.projectId)
		draft.value = config.value
		editing.value = false
	} catch (err: unknown) {
		config.value = ''
		draft.value = ''
		message.value = formatError(err)
		if (showToast) toast.warn('读取完整 Nginx 配置失败', message.value)
	} finally {
		loading.value = false
	}
}

function beginEdit(): void {
	draft.value = config.value
	message.value = ''
	editing.value = true
}

function cancelEdit(): void {
	draft.value = config.value
	message.value = ''
	editing.value = false
}

async function saveConfig(): Promise<void> {
	if (draft.value.trim() === '') {
		toast.error('保存 Nginx 配置失败', '配置不能为空')
		return
	}
	saving.value = true
	try {
		const result = await writeProjectNginxConfig(props.projectId, draft.value)
		if (!result.ok) {
			toast.error('保存 Nginx 配置失败', result.message)
			return
		}
		config.value = draft.value
		editing.value = false
		toast.success('Nginx 配置已保存', '已通过 nginx -t 校验')
	} catch (err: unknown) {
		toast.error('保存 Nginx 配置失败', formatError(err))
	} finally {
		saving.value = false
	}
}

function handleEditorKeydown(event: KeyboardEvent): void {
	if (!editing.value || event.key !== 'Tab') return
	event.preventDefault()
	const target = event.target
	if (!(target instanceof HTMLTextAreaElement)) return
	target.setRangeText('\t', target.selectionStart, target.selectionEnd, 'end')
	draft.value = target.value
}

watch(() => props.projectId, () => loadConfig(), { immediate: true })
</script>

<style scoped lang="scss">
.nginx-preview {
	display: flex;
	flex-direction: column;
	gap: 0.75rem;
}

.preview-toolbar {
	display: flex;
	align-items: center;
	justify-content: space-between;
	gap: 0.75rem;
}

.preview-actions {
	display: flex;
	align-items: center;
	gap: 0.5rem;
}

.preview-toolbar strong {
	display: block;
	color: var(--text-primary);
}

.hint {
	display: block;
	margin-top: 0.25rem;
	color: var(--text-secondary);
	font-size: var(--font-sm);
	line-height: 1.5;
}

.preview-shell {
	display: grid;
	grid-template-columns: 3rem minmax(0, 1fr);
	max-height: 24rem;
	overflow: hidden;
	border: 1px solid var(--border-color);
	border-radius: var(--h-radius);
	background: var(--code-bg);
}

.line-numbers {
	overflow: hidden;
	padding: 0.85rem 0.65rem;
	background: color-mix(in srgb, var(--code-bg) 86%, var(--border-color));
	color: var(--text-muted);
	font: var(--font-xs)/1.55 var(--font-mono);
	text-align: right;
	user-select: none;
}

.line-numbers span {
	display: block;
	height: 1.55em;
}

.preview-editor {
	width: 100%;
	min-width: 0;
	min-height: 16rem;
	padding: 0.85rem;
	resize: none;
	overflow: auto;
	border: 0;
	outline: 0;
	background: transparent;
	color: var(--code-text);
	font-size: var(--font-sm);
	line-height: 1.55;
	white-space: pre;
	scrollbar-gutter: stable;
}

.preview-editor:not([readonly]) {
	background: color-mix(in srgb, var(--code-bg) 88%, var(--primary-color));
}

.preview-empty {
	display: flex;
	align-items: center;
	gap: 0.5rem;
	min-height: 5rem;
	padding: 1rem;
	border: 1px dashed var(--border-color);
	border-radius: var(--h-radius);
	color: var(--text-secondary);
}
</style>
