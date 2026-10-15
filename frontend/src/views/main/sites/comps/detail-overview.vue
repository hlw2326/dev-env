<template>
	<!-- 站点详情概览页签 -->
	<div class="detail-overview">
		<!-- 站点根目录设置 -->
		<div class="form-field">
			<label>站点根目录</label>
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
		<comp-run-field v-model="run" :root="root" @change="saveForm" />

		<!-- 运行环境版本选择 -->
		<div v-if="site.kind !== 'static'" class="form-field">
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

		<!-- Node 启动命令配置 -->
		<div v-if="site.kind === 'node'" class="form-field">
			<label>Node 启动命令（支持 %PORT% 占位符，站点分配端口：{{ site.port }}）</label>
			<p-input-text
				v-model="command"
				placeholder="如 npm start 或 node server.js"
				class="font-mono"
			/>
		</div>

		<!-- 随应用启动开关 -->
		<div class="switch-row">
			<div class="switch-info">
				<span class="section-title">随应用启动</span>
				<span class="sub-hint">打开 DevEnv 时自动启动该站点</span>
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
import type { Site, Software, UpdateInput } from '@/types'

// 接收站点详情与软件列表
const props = defineProps<{
	site: Site
	softwareList: Software[]
}>()

// 保存更新事件
const emit = defineEmits<{
	save: [input: UpdateInput]
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

// 可选运行环境版本下拉项
const versionOptions = computed(() => {
	let targetKind = 'php'
	if (props.site.kind === 'node') {
		targetKind = 'nodejs'
	}
	return props.softwareList
		.filter((item) => item.kind === targetKind && item.state === 'installed')
		.map((item) => ({
			value: item.id,
			label: `${item.name} ${item.version}`,
		}))
})

// 选择站点根目录并立即保存
async function pickRoot(): Promise<void> {
	try {
		const picked = await pickFolder('选择站点根目录')
		if (picked !== '') {
			root.value = picked
			saveForm()
		}
	} catch (err: unknown) {
		toast.error('选择目录失败', formatError(err))
	}
}

// 提交保存站点配置
function saveForm(): void {
	emit('save', {
		id: props.site.id,
		hosts: props.site.hosts,
		root: root.value,
		run: run.value,
		software: softwareId.value,
		command: command.value,
		autostart: autostart.value,
	})
}

// 暴露弹窗底部保存入口
defineExpose({
	save: saveForm,
})

// 监听站点变化同步表单状态
watch(
	() => props.site,
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
