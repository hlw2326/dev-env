<template>
	<div class="detail-nginx">
		<div class="config-toolbar">
			<span class="hint">当前网站运行时生成的独立配置文件，命名格式：website-网站ID-网站名称.conf</span>
			<p-button
				severity="secondary"
				variant="outlined"
				icon="pi pi-file-edit"
				label="打开网站配置文件"
				@click="openConfig"
			/>
		</div>
		<div class="form-field">
			<label>Nginx 高级配置</label>
			<p-textarea
				v-model="nginx"
				rows="14"
				placeholder="例如：client_max_body_size 20m;"
				class="nginx-editor font-mono"
			/>
			<span class="hint">填写 server 区块内的 Nginx 指令。保存后会合并到自动生成的站点配置，并通过 nginx -t 检查。</span>
		</div>
	</div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { openSiteNginxConfig } from '@/api'
import { formatError, toast } from '@/utils'
import type { Site, UpdateInput } from '@/types'

const props = defineProps<{
	site: Site
}>()

const emit = defineEmits<{
	save: [input: UpdateInput]
}>()

const nginx = ref('')

async function openConfig(): Promise<void> {
	try {
		const result = await openSiteNginxConfig(props.site.id)
		if (!result.ok) {
			toast.error('打开网站配置失败', result.message)
		}
	} catch (err: unknown) {
		toast.error('打开网站配置失败', formatError(err))
	}
}

function saveForm(): void {
	emit('save', {
		id: props.site.id,
		hosts: props.site.hosts,
		root: props.site.root,
		run: props.site.run,
		software: props.site.software,
		command: props.site.command,
		nginx: nginx.value,
		autostart: props.site.autostart,
	})
}

defineExpose({
	save: saveForm,
})

watch(
	() => props.site,
	(item) => {
		nginx.value = item.nginx
	},
	{ immediate: true },
)
</script>

<style scoped lang="scss">
.detail-nginx {
	display: flex;
	flex-direction: column;
	gap: 0.875rem;
}

.config-toolbar {
	display: flex;
	align-items: center;
	justify-content: space-between;
	gap: 0.75rem;
}

.nginx-editor {
	width: 100%;
	min-height: 18rem;
	resize: vertical;
}

.hint {
	display: block;
	margin-top: 0.5rem;
	color: var(--text-secondary);
	font-size: var(--font-sm);
	line-height: 1.5;
}
</style>
