<template>
	<h-modal
		v-model:visible="visible"
		title="创建网站"
		width="32rem"
		ok-text="创建"
		:loading="submitting"
		@ok="onSubmit"
	>
		<div class="create-form">
			<div class="form-field">
				<label>站点类型</label>
				<div class="kind-options">
					<button
						v-for="opt in kindOptions"
						:key="opt.value"
						type="button"
						class="kind-btn"
						:class="{ active: form.kind === opt.value }"
						@click="onSelectKind(opt.value)"
					>
						<i :class="[opt.icon, 'kind-icon']"></i>
						<span>{{ opt.label }}</span>
					</button>
				</div>
			</div>

			<div class="form-field">
				<label>网站名称</label>
				<p-input-text
					v-model="form.name"
					placeholder="如 blog"
					@input="onNameInput"
				/>
			</div>

			<div class="form-field">
				<label>域名</label>
				<p-input-text
					v-model="form.host"
					placeholder="如 blog.localhost"
					class="font-mono"
					@input="hostTouched = true"
				/>
			</div>

			<div class="form-field">
				<label>根目录</label>
				<div class="folder-row">
					<p-input-text
						v-model="form.root"
						placeholder="项目代码所在目录"
						class="font-mono flex-input"
						@input="rootTouched = true"
						@blur="onDetectRun"
					/>
					<p-button
						severity="secondary"
						variant="outlined"
						icon="pi pi-folder-open"
						label="选择"
						@click="onPickRoot"
					/>
				</div>
			</div>

			<div class="form-field">
				<label>运行目录</label>
				<div class="run-row">
					<p-select
						v-model="runPreset"
						:options="runOptions"
						option-label="label"
						option-value="value"
						class="flex-input"
						@change="changePreset"
					/>
					<p-input-text
						v-if="runPreset === 'custom'"
						v-model="customRun"
						placeholder="相对子目录，如 public"
						class="font-mono flex-input"
					/>
				</div>
				<span class="hint-text font-mono">文档根路径：{{ resolvedPreview }}</span>
			</div>

			<div v-if="form.kind === 'node'" class="more-box">
				<button type="button" class="more-toggle" @click="moreOpen = !moreOpen">
					<i class="pi" :class="{ 'pi-chevron-down': moreOpen, 'pi-chevron-right': !moreOpen }"></i>
					<span>更多设置（启动命令）</span>
				</button>
				<div v-if="moreOpen" class="more-body">
					<div class="form-field">
						<label>启动命令（留空则自动探测 package.json）</label>
						<p-input-text
							v-model="form.command"
							placeholder="如 npm start 或 node server.js"
							class="font-mono"
						/>
					</div>
				</div>
			</div>
		</div>
	</h-modal>
</template>

<script setup lang="ts">
import { ref, reactive, computed, watch } from 'vue'
import { pickFolder, findRun } from '@/api'
import type { CreateSiteInput } from '@/types'

const props = defineProps<{
	defaultRoot: string
}>()

const visible = defineModel<boolean>('visible', { default: false })

const emit = defineEmits<{
	submit: [input: CreateSiteInput]
}>()

const kindOptions = [
	{ value: 'php', label: 'PHP', icon: 'pi pi-code' },
	{ value: 'node', label: 'Node.js', icon: 'pi pi-bolt' },
	{ value: 'static', label: '静态', icon: 'pi pi-file' },
]

const runOptions = [
	{ value: '', label: '/ (根目录本身)' },
	{ value: 'public', label: '/public' },
	{ value: 'web', label: '/web' },
	{ value: 'dist', label: '/dist' },
	{ value: 'custom', label: '自定义子目录' },
]

const form = reactive<CreateSiteInput>({
	kind: 'php',
	name: '',
	host: '',
	root: '',
	run: '',
	command: '',
})

const runPreset = ref('')
const customRun = ref('')
const hostTouched = ref(false)
const rootTouched = ref(false)
const moreOpen = ref(false)
const submitting = ref(false)

const effectiveRun = computed(() => {
	if (runPreset.value === 'custom') {
		return customRun.value.trim().replace(/^[/\\]+/, '')
	}
	return runPreset.value
})

const resolvedPreview = computed(() => {
	const base = form.root.trim().replace(/[\\/]+$/, '')
	if (base === '') {
		return ''
	}
	if (effectiveRun.value === '') {
		return base
	}
	return `${base}\\${effectiveRun.value.replace(/\//g, '\\')}`
})

watch(visible, (open) => {
	if (open) {
		form.kind = 'php'
		form.name = ''
		form.host = ''
		form.root = props.defaultRoot
		form.run = ''
		form.command = ''
		runPreset.value = ''
		customRun.value = ''
		hostTouched.value = false
		rootTouched.value = false
		moreOpen.value = false
	}
})

function onSelectKind(kind: string) {
	form.kind = kind
	onDetectRun()
}

function onNameInput() {
	const clean = form.name.trim()
	if (!hostTouched.value) {
		if (clean === '') {
			form.host = ''
		} else {
			form.host = `${clean.toLowerCase()}.localhost`
		}
	}
	if (!rootTouched.value && props.defaultRoot !== '') {
		if (clean === '') {
			form.root = props.defaultRoot
		} else {
			form.root = `${props.defaultRoot.replace(/[\\/]+$/, '')}\\${clean}`
		}
	}
}

async function onPickRoot() {
	const picked = await pickFolder('选择站点根目录')
	if (picked !== '') {
		form.root = picked
		rootTouched.value = true
		if (form.name.trim() === '') {
			const parts = picked.replace(/[\\/]+$/, '').split(/[\\/]/)
			const folderName = parts[parts.length - 1]
			if (folderName !== '') {
				form.name = folderName
				if (!hostTouched.value) {
					form.host = `${folderName.toLowerCase()}.localhost`
				}
			}
		}
		await onDetectRun()
	}
}

async function onDetectRun() {
	const root = form.root.trim()
	if (root === '') {
		return
	}
	const detected = await findRun(root, form.kind)
	if (detected === '' || detected === 'public' || detected === 'web' || detected === 'dist') {
		runPreset.value = detected
		customRun.value = ''
	} else {
		runPreset.value = 'custom'
		customRun.value = detected
	}
}

function changePreset() {
	if (runPreset.value !== 'custom') {
		customRun.value = ''
	}
}

function onSubmit() {
	form.run = effectiveRun.value
	emit('submit', {
		kind: form.kind,
		name: form.name,
		host: form.host,
		root: form.root,
		run: form.run,
		command: form.command,
	})
}
</script>

<style scoped lang="scss">
.create-form {
	display: flex;
	flex-direction: column;
}

.kind-options {
	display: grid;
	grid-template-columns: repeat(3, 1fr);
	gap: 0.625rem;
}

.kind-btn {
	display: flex;
	align-items: center;
	justify-content: center;
	gap: 0.5rem;
	height: 2.5rem;
	border: 1px solid var(--border-color);
	border-radius: var(--radius-md);
	background: var(--bg-app);
	color: var(--text-secondary);
	font-size: var(--font-sm);
	cursor: pointer;

	&:hover {
		border-color: var(--border-hover);
		color: var(--text-primary);
	}

	&.active {
		background: var(--p-primary-color);
		color: var(--on-primary);
		border-color: var(--p-primary-color);
	}

	.kind-icon {
		font-size: var(--icon-md);
	}
}

.folder-row,
.run-row {
	display: flex;
	align-items: center;
	gap: 0.5rem;
}

.flex-input {
	flex: 1;
}

.hint-text {
	font-size: var(--font-xs);
	color: var(--text-muted);
}

.more-box {
	margin-top: 0.25rem;
}

.more-toggle {
	display: inline-flex;
	align-items: center;
	gap: 0.4rem;
	border: none;
	background: transparent;
	color: var(--text-secondary);
	font-size: var(--font-sm);
	cursor: pointer;
	padding: 0.25rem 0;

	i {
		font-size: var(--icon-xs);
	}
}

.more-body {
	margin-top: 0.5rem;
}
</style>
