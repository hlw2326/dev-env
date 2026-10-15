<template>
	<!-- 常用配置项表单面板 -->
	<div class="config-items">
		<!-- 配置项列表 -->
		<div class="item-list">
			<div v-for="field in configItems" :key="field.key" class="item-row">
				<div class="item-meta">
					<span class="item-label">{{ field.label }}</span>
					<span class="item-key font-mono">{{ field.key }} · {{ field.description }}</span>
				</div>
				<p-input-text v-model="field.value" class="font-mono item-input" />
			</div>
		</div>

		<!-- 底部保存按钮栏 -->
		<div class="footer-bar">
			<p-button
				v-if="item.daemon && item.running"
				icon="pi pi-refresh"
				label="保存并重启"
				:loading="saving"
				@click="saveItems(true)"
			/>
			<p-button
				v-else
				icon="pi pi-check"
				label="保存配置"
				:loading="saving"
				@click="saveItems(false)"
			/>
		</div>
	</div>
</template>

<script setup lang="ts">
import { watch } from 'vue'
import { useConfig } from '../use-config'
import type { Software } from '@/types'

// 接收软件条目属性
const props = defineProps<{
	item: Software
}>()

// 常用配置状态与操作方法
const { configItems, saving, fetchConfig, submitConfig } = useConfig()

// 保存常用配置项
async function saveItems(restartNow: boolean): Promise<void> {
	await submitConfig(props.item.id, restartNow, props.item.running)
}

// 监听软件切换读取配置项
watch(
	() => props.item.id,
	(id) => {
		fetchConfig(id)
	},
	{ immediate: true },
)
</script>

<style scoped lang="scss">
/* 常用配置面板容器 */
.config-items {
	display: flex;
	flex-direction: column;
	gap: 0.875rem;
}

/* 配置项列表 */
.item-list {
	display: flex;
	flex-direction: column;
	gap: 0.625rem;
}

/* 单个配置项行 */
.item-row {
	display: flex;
	align-items: center;
	justify-content: space-between;
	padding: 0.625rem 0.875rem;
	background: var(--bg-app);
	border: 1px solid var(--border-color);
	border-radius: var(--h-radius);
	gap: 1rem;
}

/* 左侧配置名称与键名 */
.item-meta {
	display: flex;
	flex-direction: column;
	gap: 0.2rem;
}

/* 配置项中文名 */
.item-label {
	font-size: var(--font-base);
	color: var(--text-primary);
}

/* 配置项键名与说明 */
.item-key {
	font-size: var(--font-xs);
	color: var(--text-muted);
}

/* 右侧配置输入框 */
.item-input {
	width: 12rem;
}

/* 底部操作栏 */
.footer-bar {
	display: flex;
	justify-content: flex-end;
}
</style>
