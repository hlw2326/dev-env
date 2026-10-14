<template>
	<div class="detail-overview">
		<div class="form-field">
			<label>站点根目录</label>
			<div class="row-line">
				<p-input-text v-model="root" class="font-mono flex-input" />
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
			<label>运行目录（网关文档根）</label>
			<div class="row-line">
				<p-select
					v-model="runPreset"
					:options="runOptions"
					option-label="label"
					option-value="value"
					class="flex-input"
					@change="onRunChange"
				/>
				<p-input-text
					v-if="runPreset === 'custom'"
					v-model="customRun"
					placeholder="如 public"
					class="font-mono flex-input"
				/>
			</div>
			<span class="path-preview font-mono">生效完整路径：{{ fullPathPreview }}</span>
		</div>

		<div v-if="site.kind !== 'static'" class="form-field">
			<label>运行环境版本</label>
			<p-select
				v-model="softwareId"
				:options="versionOptions"
				option-label="label"
				option-value="value"
				placeholder="选择已安装的版本"
				@change="onSave"
			/>
		</div>

		<div v-if="site.kind === 'node'" class="form-field">
			<label>Node 启动命令（支持 %PORT% 占位符，站点分配端口：{{ site.port }}）</label>
			<p-input-text
				v-model="command"
				placeholder="如 npm start 或 node server.js"
				class="font-mono"
			/>
		</div>

		<div class="switch-row">
			<div class="switch-info">
				<span class="sec-title">随应用启动</span>
				<span class="sub-desc">打开 DevEnv 时自动启动该站点</span>
			</div>
			<p-toggle-switch v-model="autostart" @change="onSave" />
		</div>

		<div class="footer-bar">
			<p-button icon="pi pi-check" label="保存修改" @click="onSave" />
		</div>
	</div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { pickFolder } from '@/api'
import type { Site, Software, UpdateSiteInput } from '@/types'

const props = defineProps<{
	site: Site
	softwareList: Software[]
}>()

const emit = defineEmits<{
	save: [input: UpdateSiteInput]
}>()

const runOptions = [
	{ value: '', label: '/ (根目录本身)' },
	{ value: 'public', label: '/public' },
	{ value: 'web', label: '/web' },
	{ value: 'dist', label: '/dist' },
	{ value: 'custom', label: '自定义子目录' },
]

const root = ref('')
const runPreset = ref('')
const customRun = ref('')
const softwareId = ref<number | undefined>(undefined)
const command = ref('')
const autostart = ref(true)

const versionOptions = computed(() => {
	let targetKind = 'php'
	if (props.site.kind === 'node') {
		targetKind = 'nodejs'
	}
	return props.softwareList
		.filter((item) => item.kind === targetKind && item.state === 'installed')
		.map((item) => ({
			value: item.id,
			label: `${item.kind.toUpperCase()} ${item.version}`,
		}))
})

const effectiveRun = computed(() => {
	if (runPreset.value === 'custom') {
		return customRun.value.trim().replace(/^[/\\]+/, '')
	}
	return runPreset.value
})

const fullPathPreview = computed(() => {
	const base = root.value.trim().replace(/[\\/]+$/, '')
	if (effectiveRun.value === '') {
		return base
	}
	return `${base}\\${effectiveRun.value.replace(/\//g, '\\')}`
})

watch(
	() => props.site,
	(cur) => {
		root.value = cur.root
		softwareId.value = cur.software
		command.value = cur.command
		autostart.value = cur.autostart
		if (cur.run === '' || cur.run === 'public' || cur.run === 'web' || cur.run === 'dist') {
			runPreset.value = cur.run
			customRun.value = ''
		} else {
			runPreset.value = 'custom'
			customRun.value = cur.run
		}
	},
	{ immediate: true }
)

async function onPickRoot() {
	const picked = await pickFolder('选择站点根目录')
	if (picked !== '') {
		root.value = picked
		onSave()
	}
}

function onRunChange() {
	if (runPreset.value !== 'custom') {
		customRun.value = ''
		onSave()
	}
}

function onSave() {
	emit('save', {
		id: props.site.id,
		hosts: props.site.hosts,
		root: root.value,
		run: effectiveRun.value,
		software: softwareId.value,
		command: command.value,
		autostart: autostart.value,
	})
}
</script>

<style scoped lang="scss">
.detail-overview {
	display: flex;
	flex-direction: column;
	gap: 0.25rem;
}

.row-line {
	display: flex;
	align-items: center;
	gap: 0.5rem;
}

.flex-input {
	flex: 1;
}

.path-preview {
	font-size: var(--font-xs);
	color: var(--text-muted);
}

.switch-row {
	display: flex;
	align-items: center;
	justify-content: space-between;
	padding: 0.75rem 0;
	border-top: 1px solid var(--border-color);
	margin-top: 0.25rem;
}

.switch-info {
	display: flex;
	flex-direction: column;
	gap: 0.2rem;
}

.sec-title {
	font-size: var(--font-base);
	color: var(--text-primary);
}

.sub-desc {
	font-size: var(--font-sm);
	color: var(--text-secondary);
}

.footer-bar {
	display: flex;
	justify-content: flex-end;
	padding-top: 0.5rem;
}
</style>
