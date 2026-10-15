<template>
	<div class="detail-nginx">
		<div class="config-toolbar">
			<div>
				<strong>Nginx 高级配置</strong>
				<span class="hint">仅编辑自动生成的 server 区块内容</span>
			</div>
			<p-button
				severity="secondary"
				variant="outlined"
				icon="pi pi-file-edit"
				label="打开项目配置文件"
				@click="openConfig"
			/>
		</div>
		<div v-if="project.kind === 'php'" class="environment-field">
			<label for="php-environment-file">PHP 环境变量文件</label>
			<p-select
				id="php-environment-file"
				v-model="environmentFile"
				:options="environmentFileOptions"
				placeholder="自动使用 .env.dev"
				:disabled="environmentFileOptions.length === 0"
			/>
			<span class="hint">启动 PHP 项目时，DevEnv 会将文件内容作为 FastCGI 环境变量传给应用；优先自动读取 .env.dev。</span>
		</div>
		<comp-nginx-preview :project-id="project.id" />
		<div class="editor-toolbar">
			<label for="nginx-snippet">插入常用指令</label>
			<p-select
				id="nginx-snippet"
				v-model="selectedSnippet"
				:options="snippetOptions"
				option-label="label"
				option-value="value"
				placeholder="选择指令模板"
				class="snippet-select"
				@change="insertSelectedSnippet"
			/>
			<span class="editor-meta">{{ lineCount }} 行 · {{ nginx.length }} 个字符</span>
		</div>

		<div class="editor-shell" :class="{ invalid: validationMessage !== '' }">
			<div ref="lineNumbers" class="line-numbers" aria-hidden="true">
				<span v-for="line in lineCount" :key="line">{{ line }}</span>
			</div>
			<textarea
				ref="editor"
				v-model="nginx"
				class="nginx-editor font-mono"
				spellcheck="false"
				wrap="off"
				aria-label="Nginx 高级配置"
				placeholder="例如：client_max_body_size 20m;"
				@keydown="handleEditorKeydown"
				@scroll="syncScroll"
			/>
		</div>

		<div class="editor-status" :class="{ error: validationMessage !== '' }">
			<i :class="validationMessage === '' ? 'pi pi-check-circle' : 'pi pi-exclamation-circle'" />
			<span>{{ validationMessage === '' ? '配置格式正确，可以保存' : validationMessage }}</span>
		</div>
		<span class="hint">
			支持 Tab 缩进、Enter 自动缩进和常用指令模板。内容会合并到自动生成的 server 区块，不能写入 server、http 或 events 区块。
		</span>
	</div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { openProjectNginxConfig } from '@/api'
import { formatError, toast } from '@/utils'
import type { Project, ProjectUpdateInput } from '@/types'
import CompNginxPreview from './nginx-preview.vue'

const props = defineProps<{ project: Project }>()
const emit = defineEmits<{ save: [input: ProjectUpdateInput] }>()

const nginx = ref('')
const environmentFile = ref('')
const selectedSnippet = ref('')
const editor = ref<HTMLTextAreaElement | null>(null)
const lineNumbers = ref<HTMLDivElement | null>(null)

const snippets = [
	{ label: '限制请求体大小', value: 'client_max_body_size 20m;' },
	{ label: '开启 Gzip', value: 'gzip on;\ngzip_types text/plain text/css application/json application/javascript;' },
	{ label: '添加安全响应头', value: 'add_header X-Content-Type-Options "nosniff" always;' },
	{ label: '静态资源缓存', value: 'location ~* \\.(css|js|png|jpg|jpeg|gif|svg|woff2)$ {\n\t expires 7d;\n\t add_header Cache-Control "public, max-age=604800";\n}' },
]

const snippetOptions = computed(() => [
	{ label: '选择指令模板', value: '' },
	...snippets,
])

const environmentFileOptions = computed(() => {
	const options = [...props.project.environmentFiles]
	if (environmentFile.value !== '' && !options.includes(environmentFile.value)) {
		options.unshift(environmentFile.value)
	}
	return options
})

const lineCount = computed(() => Math.max(1, nginx.value.split('\n').length))
const validationMessage = computed(() => validateNginxConfig(nginx.value))
function validateNginxConfig(value: string): string {
	let depth = 0
	let quote = ''
	const lines = value.split('\n')

	for (let index = 0; index < lines.length; index += 1) {
		const source = lines[index] ?? ''
		const trimmed = source.trim()
		if (trimmed === '' || trimmed.startsWith('#')) {
			continue
		}
		if (/^(server|http|events)(?:\s|\{|$)/i.test(trimmed)) {
			return `第 ${index + 1} 行不能声明 ${trimmed.split(/\s|\{/)[0]} 区块`
		}

		let comment = false
		for (const character of source) {
			if (comment) {
				continue
			}
			if (character === '#' && quote === '') {
				comment = true
				continue
			}
			if ((character === '"' || character === "'") && quote === '') {
				quote = character
				continue
			}
			if (character === quote) {
				quote = ''
				continue
			}
			if (quote !== '') {
				continue
			}
			if (character === '{') depth += 1
			if (character === '}') depth -= 1
			if (depth < 0) return `第 ${index + 1} 行的大括号没有匹配`
		}

		const code = source.replace(/\s+#.*$/, '').trim()
		if (code !== '' && !code.endsWith(';') && !code.endsWith('{') && code !== '}') {
			return `第 ${index + 1} 行需要以分号或左大括号结尾`
		}
	}

	if (quote !== '') return '配置中包含未闭合的引号'
	if (depth !== 0) return '配置中的大括号没有匹配'
	return ''
}
function insertSelectedSnippet(): void {
	if (selectedSnippet.value === '' || editor.value === null) return
	const target = editor.value
	const start = target.selectionStart
	const end = target.selectionEnd
	const before = nginx.value.slice(0, start)
	const prefix = before !== '' && !before.endsWith('\n') ? '\n' : ''
	target.setRangeText(`${prefix}${selectedSnippet.value}\n`, start, end, 'end')
	nginx.value = target.value
	selectedSnippet.value = ''
	target.focus()
}

function handleEditorKeydown(event: KeyboardEvent): void {
	if (editor.value === null) return
	const target = editor.value
	if (event.key === 'Tab') {
		event.preventDefault()
		target.setRangeText('\t', target.selectionStart, target.selectionEnd, 'end')
		nginx.value = target.value
		return
	}
	if (event.key !== 'Enter') return

	event.preventDefault()
	const before = target.value.slice(0, target.selectionStart)
	const currentLine = before.slice(before.lastIndexOf('\n') + 1)
	const indent = currentLine.match(/^\s*/)?.[0] ?? ''
	const extraIndent = currentLine.trimEnd().endsWith('{') ? '\t' : ''
	target.setRangeText(`\n${indent}${extraIndent}`, target.selectionStart, target.selectionEnd, 'end')
	nginx.value = target.value
}

function syncScroll(): void {
	if (editor.value !== null && lineNumbers.value !== null) {
		lineNumbers.value.scrollTop = editor.value.scrollTop
	}
}

async function openConfig(): Promise<void> {
	try {
		const result = await openProjectNginxConfig(props.project.id)
		if (!result.ok) toast.error('打开项目配置失败', result.message)
	} catch (err: unknown) {
		toast.error('打开项目配置失败', formatError(err))
	}
}

function saveForm(): void {
	if (validationMessage.value !== '') {
		toast.error('Nginx 配置格式错误', validationMessage.value)
		return
	}
	emit('save', {
		id: props.project.id,
		hosts: props.project.hosts,
		root: props.project.root,
		run: props.project.run,
		software: props.project.software,
		command: props.project.command,
		nginx: nginx.value,
		environmentFile: environmentFile.value,
		autostart: props.project.autostart,
	})
}

defineExpose({ save: saveForm })

watch(
	() => props.project,
	(item) => {
		nginx.value = item.nginx
		environmentFile.value = item.environmentFile
	},
	{ immediate: true },
)
</script>
<style scoped lang="scss">
.detail-nginx {
	display: flex;
	flex-direction: column;
	gap: 0.75rem;
}

.config-toolbar,
.editor-toolbar {
	display: flex;
	align-items: center;
	justify-content: space-between;
	gap: 0.75rem;
}

.environment-field {
	display: flex;
	flex-direction: column;
	gap: 0.35rem;
}

.hint {
	display: block;
	margin-top: 0.25rem;
	color: var(--text-secondary);
	font-size: var(--font-sm);
	line-height: 1.5;
}

.editor-toolbar .snippet-select {
	min-width: 12rem;
}

.editor-meta {
	margin-left: auto;
	color: var(--text-muted);
	font-size: var(--font-xs);
}

.editor-shell {
	display: grid;
	grid-template-columns: 3rem minmax(0, 1fr);
	min-height: 22rem;
	max-height: 30rem;
	overflow: hidden;
	border: 1px solid var(--border-color);
	border-radius: var(--h-radius);
	background: var(--code-bg);
}

.editor-shell.invalid {
	border-color: var(--err);
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

.nginx-editor {
	width: 100%;
	min-width: 0;
	min-height: 22rem;
	padding: 0.85rem;
	resize: none;
	overflow: auto;
	border: 0;
	outline: 0;
	background: transparent;
	color: var(--code-text);
	font-size: var(--font-sm);
	line-height: 1.55;
	tab-size: 4;
	white-space: pre;
	scrollbar-gutter: stable;
}

.editor-status {
	display: flex;
	align-items: center;
	gap: 0.4rem;
	color: var(--ok);
	font-size: var(--font-sm);
}

.editor-status.error {
	color: var(--err);
}
</style>
