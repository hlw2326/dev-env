<template>
	<!-- 配置文件全文编辑器 -->
	<div class="config-file">
		<!-- 顶部文件路径与刷新按钮 -->
		<div class="file-bar">
			<span class="file-path font-mono">{{ configFile }}</span>
			<p-button
				size="small"
				severity="secondary"
				variant="outlined"
				icon="pi pi-refresh"
				label="重新读取"
				@click="reloadFile"
			/>
		</div>

		<!-- 多行文本编辑框 -->
		<p-textarea
			v-model="rawText"
			rows="14"
			class="editor-area font-mono"
		/>

		<!-- 底部保存按钮 -->
		<div class="footer-bar">
			<p-button
				icon="pi pi-check"
				label="保存配置文件"
				:loading="saving"
				@click="saveFile"
			/>
		</div>
	</div>
</template>

<script setup lang="ts">
import { watch } from 'vue'
import { useConfig } from '../use-config'

// 接收软件标识属性
const props = defineProps<{
	softwareId: number
}>()

// 配置文件状态与操作
const { configFile, rawText, saving, fetchConfig, fetchRaw, submitRaw } = useConfig()

// 重新读取配置文件内容
async function reloadFile(): Promise<void> {
	await Promise.all([fetchConfig(props.softwareId), fetchRaw(props.softwareId)])
}

// 保存配置文件全文
async function saveFile(): Promise<void> {
	await submitRaw(props.softwareId)
}

// 监听软件标识变化加载配置文件
watch(
	() => props.softwareId,
	() => {
		reloadFile()
	},
	{ immediate: true },
)
</script>

<style scoped lang="scss">
/* 全文编辑容器 */
.config-file {
	display: flex;
	flex-direction: column;
	gap: 0.75rem;
}

/* 顶部路径栏 */
.file-bar {
	display: flex;
	align-items: center;
	justify-content: space-between;
	gap: 0.5rem;
}

/* 配置文件路径文本 */
.file-path {
	font-size: var(--font-xs);
	color: var(--text-muted);
	overflow: hidden;
	text-overflow: ellipsis;
	white-space: nowrap;
}

/* 多行代码文本框 */
.editor-area {
	width: 100%;
	font-size: var(--font-xs);
	background: var(--code-bg);
	color: var(--code-text);
}

/* 底部操作栏 */
.footer-bar {
	display: flex;
	justify-content: flex-end;
}
</style>
