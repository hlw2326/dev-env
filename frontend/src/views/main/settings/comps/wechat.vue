<template>
	<h-card title="微信开发者工具">
		<div class="settings-block">
			<div class="form-field">
				<label>微信开发者工具 CLI</label>
				<div class="path-row">
					<p-input-text v-model="form.developerToolsCLI" placeholder="例如微信开发者工具安装目录下的 cli.bat" class="flex-input" />
					<p-button severity="secondary" variant="outlined" label="选择" icon="pi pi-folder-open" @click="pickDeveloperToolsCLI" />
				</div>
				<span class="field-hint">小程序发布统一使用此 CLI。请在微信开发者工具“设置 → 安全设置”中开启服务端口。</span>
			</div>
		</div>
	</h-card>
</template>

<script setup lang="ts">
import { useSettings } from '@/composables'

const { form, pickDeveloperToolsCLI } = useSettings()
</script>

<style scoped lang="scss">
.settings-block, .form-field { display: flex; flex-direction: column; gap: calc(var(--h-padding) / 2); }
.path-row { display: flex; align-items: center; gap: calc(var(--h-padding) / 2); }
.flex-input { flex: 1; min-width: 0; }
.path-row :deep(button) { flex-shrink: 0; }
.field-hint { font-size: var(--font-sm); color: var(--text-secondary); }
</style>
