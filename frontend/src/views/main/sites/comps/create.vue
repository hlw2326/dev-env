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
				<comp-kind-select v-model="form.kind" @update:model-value="selectKind" />
			</div>

			<!-- 项目名称输入 -->
			<div class="form-field">
				<label>项目名称</label>
				<p-input-text v-model="form.name" placeholder="如 blog" @input="syncName" />
			</div>

			<!-- 域名输入，Electron 项目不使用 Nginx -->
			<div v-if="form.kind !== 'electron'" class="form-field">
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

			<!-- Web 项目运行目录选择 -->
			<comp-run-field v-if="form.kind !== 'electron'" v-model="form.run" :root="form.root" />

			<!-- 进程项目启动命令设置 -->
			<div v-if="commandKinds.includes(form.kind)" class="more-box">
				<button type="button" class="more-toggle" @click="moreOpen = !moreOpen">
					<i class="pi" :class="{ 'pi-chevron-down': moreOpen, 'pi-chevron-right': !moreOpen }"></i>
					<span>启动设置</span>
				</button>
				<div v-if="moreOpen" class="more-body">
					<div class="form-field">
						<label>{{ commandLabel }}</label>
						<p-input-text
							v-model="form.command"
						:placeholder="commandPlaceholder"
							class="font-mono"
						/>
					</div>
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
import type { CreateInput } from '@/types'

const props = defineProps<{
	defaultRoot: string
}>()

const visible = defineModel<boolean>('visible', { required: true })

const emit = defineEmits<{
	submit: [input: CreateInput]
}>()

const commandKinds = ['node', 'python', 'go', 'rust', 'java', 'electron']

const form = reactive<CreateInput>({
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

const commandPlaceholder = computed(() => {
	if (form.kind === 'electron') {
		return 'pnpm dev'
	}
	if (form.kind === 'python') {
		return '如 python app.py'
	}
	if (form.kind === 'go') {
		return '如 go run .'
	}
	if (form.kind === 'rust') {
		return '如 cargo run'
	}
	if (form.kind === 'java') {
		return '如 java Main.java 或 mvn spring-boot:run'
	}
	return '如 npm start 或 node server.js'
})

const commandLabel = computed(() => {
	if (form.kind === 'electron') {
		return 'Electron 启动命令'
	}
	if (form.kind === 'python') {
		return 'Python 启动命令（留空则自动探测）'
	}
	if (form.kind === 'go') {
		return 'Go 启动命令（留空则自动探测 go.mod）'
	}
	if (form.kind === 'rust') {
		return 'Rust 启动命令（留空则自动探测 Cargo.toml）'
	}
	if (form.kind === 'java') {
		return 'Java 启动命令（留空则自动探测）'
	}
	return 'Node 启动命令（留空则自动探测 package.json）'
})

function selectKind(kind: string): void {
	form.kind = kind
	if (kind === 'electron' && form.command.trim() === '') {
		form.command = 'pnpm dev'
	}
	if (kind !== 'electron' && form.command === 'pnpm dev') {
		form.command = ''
	}
	detectRun()
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
		const picked = await pickFolder('选择站点根目录')
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
