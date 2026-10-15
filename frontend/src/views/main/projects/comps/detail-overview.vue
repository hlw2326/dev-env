<template>
	<!-- 项目详情概览页签 -->
	<div class="detail-overview">
		<!-- 项目根目录设置 -->
		<div class="form-field">
			<label>项目目录</label>
			<div class="row-line">
				<p-input-text v-model="root" class="font-mono flex-input" />
				<p-button
					severity="secondary"
					variant="outlined"
					icon="pi pi-folder-open"
					label="选择"
					@click="pickRoot"
				/>
			</div>
		</div>

		<!-- 运行目录设置 -->
		<comp-run-field v-if="hasCapability('run-directory')" v-model="run" :root="root" @change="saveForm" />

		<!-- 运行环境版本选择 -->
		<div v-if="hasCapability('runtime')" class="form-field">
			<label>运行环境版本</label>
			<p-select
				v-model="softwareId"
				:options="versionOptions"
				option-label="label"
				option-value="value"
				placeholder="选择已安装的版本"
				@change="saveForm"
			/>
		</div>

		<!-- 进程项目启动命令配置 -->
		<comp-command-field
			v-if="hasCapability('command')"
			v-model="command"
			:kind="project.kind"
			:root="root"
			:template="template"
		/>

		<!-- Web 项目随应用启动开关，本地调试项目由用户手动启动 -->
		<div v-if="hasCapability('autostart')" class="switch-row">
			<div class="switch-info">
				<span class="section-title">随应用启动</span>
				<span class="sub-hint">打开 DevEnv 时自动启动该项目</span>
			</div>
			<p-toggle-switch v-model="autostart" @change="saveForm" />
		</div>

	</div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { pickFolder } from '@/api'
import { toast, formatError } from '@/utils'
import CompRunField from './run-field.vue'
import CompCommandField from './command-field.vue'
import type { ProjectTemplate, Project, Software, ProjectUpdateInput } from '@/types'

// 接收项目详情与软件列表
const props = defineProps<{
	project: Project
	softwareList: Software[]
	template: ProjectTemplate | null
}>()

// 保存更新事件
const emit = defineEmits<{
	save: [input: ProjectUpdateInput]
}>()

// 根目录路径
const root = ref('')
// 相对运行目录
const run = ref('')
// 绑定软件版本标识
const softwareId = ref<number | undefined>(undefined)
// Node 启动命令
const command = ref('')
// 随应用自动启动开关
const autostart = ref(true)

function hasCapability(capability: string): boolean {
	return props.template !== null && props.template.capabilities.includes(capability)
}

// 可选运行环境版本下拉项
const versionOptions = computed(() => {
	let targetKind = ''
	if (props.template !== null) {
		targetKind = props.template.runtimeKind
	}
	return props.softwareList
		.filter((item) => item.kind === targetKind && item.state === 'installed')
		.map((item) => ({
			value: item.id,
			label: `${item.name} ${item.version}`,
		}))
})

// 选择项目根目录并立即保存
async function pickRoot(): Promise<void> {
	try {
		const picked = await pickFolder('选择项目根目录')
		if (picked !== '') {
			root.value = picked
			saveForm()
		}
	} catch (err: unknown) {
		toast.error('选择目录失败', formatError(err))
	}
}

// 提交保存项目配置
function saveForm(): void {
	emit('save', {
		id: props.project.id,
		hosts: props.project.hosts,
		root: root.value,
		run: run.value,
		software: softwareId.value,
		command: command.value,
		nginx: props.project.nginx,
		autostart: autostart.value,
	})
}

// 暴露弹窗底部保存入口
defineExpose({
	save: saveForm,
})

// 监听项目变化同步表单状态
watch(
	() => props.project,
	(item) => {
		root.value = item.root
		run.value = item.run
		softwareId.value = item.software
		command.value = item.command
		autostart.value = item.autostart
	},
	{ immediate: true },
)
</script>

<style scoped lang="scss">
/* 概览表单容器 */
.detail-overview {
	display: flex;
	flex-direction: column;
	gap: 0.25rem;
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

/* 开关配置行 */
.switch-row {
	display: flex;
	align-items: center;
	justify-content: space-between;
	padding: 0.75rem 0;
	border-top: 1px solid var(--border-color);
	margin-top: 0.25rem;
}

/* 开关左侧文字区 */
.switch-info {
	display: flex;
	flex-direction: column;
	gap: 0.2rem;
}

/* 小节标题文字 */
.section-title {
	font-size: var(--font-base);
	color: var(--text-primary);
}

/* 辅助说明文字 */
.sub-hint {
	font-size: var(--font-sm);
	color: var(--text-secondary);
}

</style>
