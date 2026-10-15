<template>
	<!-- 创建项目弹窗 -->
	<h-modal
		v-model:visible="visible"
		title="创建项目"
		width="32rem"
		:modal="true"
		:closable="true"
		:mask-close="false"
		:footer="true"
		ok-text="创建"
		ok-icon="pi pi-check"
		cancel-text="取消"
		:loading="submitting"
		@ok="submitForm"
	>
		<!-- 创建项目表单 -->
		<div class="create-form">
			<!-- 项目类型选择 -->
			<div class="form-field">
				<label>项目类型</label>
				<comp-kind-select v-model="form.kind" :templates="templates" @update:model-value="selectKind" />
			</div>

			<!-- 项目名称输入 -->
			<div class="form-field">
				<label>项目名称</label>
				<p-input-text v-model="form.name" placeholder="如 blog" @input="syncName" />
			</div>

			<!-- 域名输入，本地调试项目不使用 Nginx -->
			<div v-if="hasCapability('domain')" class="form-field">
				<label>域名</label>
				<p-input-text
					v-model="form.host"
					placeholder="如 blog.localhost"
					class="font-mono"
					@input="hostTouched = true"
				/>
			</div>

			<!-- 项目目录选择 -->
			<div class="form-field">
				<label>项目目录</label>
				<div class="row-line">
					<p-input-text
						v-model="form.root"
						placeholder="项目代码所在目录"
						class="font-mono flex-input"
						@input="rootTouched = true"
						@blur="detectRun"
					/>
					<p-button
						severity="secondary"
						variant="outlined"
						icon="pi pi-folder-open"
						label="选择"
						@click="pickRoot"
					/>
				</div>
			</div>

			<!-- Web 项目运行目录选择，本地调试项目直接使用项目根目录 -->
			<comp-run-field v-if="hasCapability('run-directory')" v-model="form.run" :root="form.root" />

			<!-- 进程项目启动命令设置 -->
			<div v-if="hasCapability('command')" class="more-box">
				<p-button type="button" unstyled class="more-toggle" @click="moreOpen = !moreOpen">
					<i class="pi" :class="{ 'pi-chevron-down': moreOpen, 'pi-chevron-right': !moreOpen }"></i>
					<span>启动设置</span>
				</p-button>
				<div v-if="moreOpen" class="more-body">
					<comp-command-field
						v-model="form.command"
						:kind="form.kind"
						:root="form.root"
						:template="currentTemplate"
					/>
				</div>
			</div>
		</div>
	</h-modal>
</template>

<script setup lang="ts">
import { computed, ref, reactive, watch } from 'vue'
import { pickFolder, findRun } from '@/api'
import { joinPath, trimSlash, toast, formatError } from '@/utils'
import CompRunField from './run-field.vue'
import CompKindSelect from './kind-select.vue'
import CompCommandField from './command-field.vue'
import type { ProjectCreateInput, ProjectTemplate } from '@/types'

const props = defineProps<{
	defaultRoot: string
	templates: ProjectTemplate[]
}>()

const visible = defineModel<boolean>('visible', { required: true })

const emit = defineEmits<{
	submit: [input: ProjectCreateInput]
}>()

const form = reactive<ProjectCreateInput>({
	kind: 'php',
	name: '',
	host: '',
	root: '',
	run: '',
	command: '',
})

const hostTouched = ref(false)
const rootTouched = ref(false)
const moreOpen = ref(false)
const submitting = ref(false)
const previousKind = ref('php')

function findTemplate(kind: string): ProjectTemplate | null {
	for (const item of props.templates) {
		if (item.kind === kind) {
			return item
		}
	}
	return null
}

const currentTemplate = computed<ProjectTemplate | null>(() => {
	return findTemplate(form.kind)
})

function hasCapability(capability: string): boolean {
	const template = currentTemplate.value
	return template !== null && template.capabilities.includes(capability)
}

function selectKind(kind: string): void {
	const oldTemplate = findTemplate(previousKind.value)
	if (oldTemplate !== null && (oldTemplate.commandMode === 'package-script' || oldTemplate.defaultCommand === form.command)) {
		form.command = ''
	}
	form.kind = kind
	const newTemplate = findTemplate(kind)
	if (newTemplate !== null && newTemplate.defaultCommand !== '' && form.command.trim() === '') {
		form.command = newTemplate.defaultCommand
	}
	previousKind.value = kind
	if (newTemplate !== null && newTemplate.commandMode === 'package-script') {
		moreOpen.value = true
	}
	void detectRun()
}

function syncName(): void {
	const clean = form.name.trim()
	if (!hostTouched.value) {
		if (clean === '') {
			form.host = ''
		} else {
			form.host = `${clean.toLowerCase()}.localhost`
		}
	}
	if (!rootTouched.value && props.defaultRoot !== '') {
		form.root = joinPath(props.defaultRoot, clean)
	}
}

async function pickRoot(): Promise<void> {
	try {
		const picked = await pickFolder('选择项目根目录')
		if (picked === '') {
			return
		}
		form.root = picked
		rootTouched.value = true
		if (form.name.trim() === '') {
			const parts = trimSlash(picked).split(/[\\/]/)
			const folderName = parts[parts.length - 1]
			if (folderName !== '') {
				form.name = folderName
				if (!hostTouched.value) {
					form.host = `${folderName.toLowerCase()}.localhost`
				}
			}
		}
	await detectRun()
	} catch (err: unknown) {
		toast.error('选择目录失败', formatError(err))
	}
}

async function detectRun(): Promise<void> {
	const root = form.root.trim()
	if (root === '') {
		return
	}
	try {
		form.run = await findRun(root, form.kind)
	} catch (err: unknown) {
		toast.error('探测运行目录失败', formatError(err))
	}
}

function submitForm(): void {
	emit('submit', { ...form })
}

// 打开弹窗时重置表单默认值
watch(visible, (open) => {
	if (open) {
		Object.assign(form, { kind: 'php', name: '', host: '', root: props.defaultRoot, run: '', command: '' })
		hostTouched.value = false
		rootTouched.value = false
		moreOpen.value = false
		previousKind.value = 'php'
	}
})
</script>

<style scoped lang="scss">
.create-form {
	display: flex;
	flex-direction: column;
}

.row-line {
	display: flex;
	align-items: center;
	gap: 0.5rem;
}

.flex-input {
	flex: 1;
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
