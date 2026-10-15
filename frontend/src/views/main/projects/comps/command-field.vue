<template>
	<div class="form-field">
		<label>{{ commandLabel }}</label>
		<p-select
			v-if="isPackageScriptTemplate"
			v-model="command"
			:options="commandOptions"
			:placeholder="commandPlaceholder"
			editable
			class="command-input font-mono"
		/>
		<p-input-text
			v-else
			v-model="command"
			:placeholder="commandPlaceholder"
			class="command-input font-mono"
		/>
		<span v-if="isPackageScriptTemplate && commandOptions.length === 0" class="hint-text">
			未找到可识别的运行脚本，可直接输入命令。
		</span>
	</div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { findCommands } from '@/api'
import { formatError, toast } from '@/utils'
import type { ProjectTemplate } from '@/types'

const props = defineProps<{
	kind: string
	root: string
	template: ProjectTemplate | null
}>()

const command = defineModel<string>({ required: true })
const commandOptions = ref<string[]>([])

const isPackageScriptTemplate = computed(() => {
	return props.template !== null && props.template.commandMode === 'package-script'
})

const commandLabel = computed(() => {
	if (props.template === null) {
		return '启动命令'
	}
	return props.template.commandLabel
})

const commandPlaceholder = computed(() => {
	if (props.template === null) {
		return ''
	}
	return props.template.commandPlaceholder
})

async function loadCommandOptions(): Promise<void> {
	commandOptions.value = []
	if (!isPackageScriptTemplate.value || props.root.trim() === '') {
		return
	}
	try {
		const options = await findCommands(props.root.trim(), props.kind)
		commandOptions.value = options
		if (options.length > 0 && !options.includes(command.value)) {
			command.value = options[0]
		}
	} catch (err: unknown) {
		toast.error('读取 package.json 失败', formatError(err))
	}
}

watch(
	() => [props.kind, props.root, props.template],
	() => {
		void loadCommandOptions()
	},
	{ immediate: true },
)
</script>

<style scoped lang="scss">
.form-field label {
	font-size: var(--font-sm);
	color: var(--text-secondary);
}

.command-input {
	width: 100%;
	font-size: var(--font-sm);
}

.command-input :deep(.p-select-label),
.command-input :deep(.p-inputtext) {
	font-size: var(--font-sm) !important;
}

.hint-text {
	font-size: var(--font-xs);
	color: var(--text-muted);
}
</style>
