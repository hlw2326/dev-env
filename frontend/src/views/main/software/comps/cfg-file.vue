<template>
	<div class="cfg-file">
		<div class="file-header">
			<span class="file-path font-mono">{{ filePath }}</span>
			<span class="file-tip">保存前将自动备份为 .bak 文件</span>
		</div>
		<textarea
			v-model="model"
			class="code-editor font-mono"
			spellcheck="false"
		></textarea>
		<div class="footer-bar">
			<p-button
				icon="pi pi-check"
				label="保存配置文件"
				:loading="saving"
				@click="emit('save')"
			/>
		</div>
	</div>
</template>

<script setup lang="ts">
defineProps<{
	filePath: string
	saving: boolean
}>()

const model = defineModel<string>({ default: '' })

const emit = defineEmits<{
	save: []
}>()
</script>

<style scoped lang="scss">
.cfg-file {
	display: flex;
	flex-direction: column;
	gap: 0.625rem;
}

.file-header {
	display: flex;
	align-items: center;
	justify-content: space-between;
}

.file-path {
	font-size: var(--font-xs);
	color: var(--text-secondary);
}

.file-tip {
	font-size: var(--font-xs);
	color: var(--text-muted);
}

.code-editor {
	width: 100%;
	height: 20rem;
	background: var(--code-bg);
	color: var(--code-text);
	border: 1px solid var(--border-color);
	border-radius: var(--radius-md);
	padding: 0.875rem;
	font-size: var(--font-xs);
	line-height: var(--leading-normal);
	resize: vertical;
}

.footer-bar {
	display: flex;
	justify-content: flex-end;
}
</style>
