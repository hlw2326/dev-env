<template>
	<!-- 创建网站弹窗 -->
	<h-modal
		v-model:visible="visible"
		title="创建网站"
		width="32rem"
		ok-text="创建"
		:loading="submitting"
		@ok="submitForm"
	>
		<!-- 创建网站表单 -->
		<div class="create-form">
			<!-- 站点类型选择 -->
			<div class="form-field">
				<label>站点类型</label>
				<div class="kind-options">
					<button
						v-for="item in kindOptions"
						:key="item.value"
						type="button"
						class="kind-btn"
						:class="{ active: form.kind === item.value }"
						@click="selectKind(item.value)"
					>
						<i :class="[item.icon, 'kind-icon']"></i>
						<span>{{ item.label }}</span>
					</button>
				</div>
			</div>

			<!-- 网站名称输入 -->
			<div class="form-field">
				<label>网站名称</label>
				<p-input-text v-model="form.name" placeholder="如 blog" @input="syncName" />
			</div>

			<!-- 域名输入 -->
			<div class="form-field">
				<label>域名</label>
				<p-input-text
					v-model="form.host"
					placeholder="如 blog.localhost"
					class="font-mono"
					@input="hostTouched = true"
				/>
			</div>

			<!-- 根目录选择 -->
			<div class="form-field">
				<label>根目录</label>
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

			<!-- 运行目录选择 -->
			<comp-run-field v-model="form.run" :root="form.root" />

			<!-- Node 更多设置折叠区 -->
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
import { ref, reactive, watch } from 'vue'
import { pickFolder, findRun } from '@/api'
import { joinPath, trimSlash, toast, formatError } from '@/utils'
import CompRunField from './run-field.vue'
import type { CreateInput } from '@/types'

// 接收默认站点父目录
const props = defineProps<{
	defaultRoot: string
}>()

// 弹窗显隐双向绑定
const visible = defineModel<boolean>('visible', { default: false })

// 提交创建事件
const emit = defineEmits<{
	submit: [input: CreateInput]
}>()

// 站点类型选项列表
const kindOptions = [
	{ value: 'php', label: 'PHP', icon: 'pi pi-code' },
	{ value: 'node', label: 'Node.js', icon: 'pi pi-bolt' },
	{ value: 'static', label: '静态', icon: 'pi pi-file' },
]

// 创建表单数据
const form = reactive<CreateInput>({
	kind: 'php',
	name: '',
	host: '',
	root: '',
	run: '',
	command: '',
})

// 域名是否已手动修改
const hostTouched = ref(false)
// 根目录是否已手动修改
const rootTouched = ref(false)
// 更多设置展开状态
const moreOpen = ref(false)
// 提交加载状态
const submitting = ref(false)

// 切换站点类型并探测运行目录
function selectKind(kind: string): void {
	form.kind = kind
	detectRun()
}

// 根据站点名称自动同步域名与目录
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

// 选择站点根目录并自动回填名称
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

// 调用后端自动探测运行目录
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

// 提交创建表单
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
/* 表单外框 */
.create-form {
	display: flex;
	flex-direction: column;
}

/* 三列类型按钮网格 */
.kind-options {
	display: grid;
	grid-template-columns: repeat(3, 1fr);
	gap: 0.625rem;
}

/* 类型选择按钮 */
.kind-btn {
	display: flex;
	align-items: center;
	justify-content: center;
	gap: 0.5rem;
	height: 2.5rem;
	border: 1px solid var(--border-color);
	border-radius: var(--h-radius);
	background: var(--bg-app);
	color: var(--text-secondary);
	font-size: var(--font-sm);
	cursor: pointer;

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

	/* 类型图标 */
	.kind-icon {
		font-size: var(--icon-md);
	}
}

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

/* 更多设置容器 */
.more-box {
	margin-top: 0.25rem;
}

/* 更多设置折叠按钮 */
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

	/* 折叠箭头图标 */
	i {
		font-size: var(--icon-xs);
	}
}

/* 更多设置展开内容 */
.more-body {
	margin-top: 0.5rem;
}
</style>
