<template>
	<!-- 存储与工作目录配置卡片 -->
	<h-card :title="$t('存储与工作目录配置')">
		<!-- 运行环境安装根目录 -->
		<div class="form-field">
			<label>{{ $t('运行环境安装根目录 (Runtimes Root)') }}</label>
			<div class="path-row">
				<p-input-text
					v-model="model.runtimeRootPath"
					class="font-mono flex-1"
					placeholder="如 D:\DevEnv\runtimes"
				/>
				<p-button
					label="选择目录"
					icon="pi pi-folder-open"
					severity="secondary"
					outlined
					@click="emit('pick-runtime')"
				/>
			</div>
		</div>

		<!-- 默认网站项目根目录 -->
		<div class="form-field">
			<label>{{ $t('默认网站/项目根目录 (WWW Root)') }}</label>
			<div class="path-row">
				<p-input-text
					v-model="model.defaultProjectPath"
					class="font-mono flex-1"
					placeholder="如 G:\www 或 D:\projects"
				/>
				<p-button
					label="选择目录"
					icon="pi pi-folder-open"
					severity="secondary"
					outlined
					@click="emit('pick-project')"
				/>
			</div>
		</div>
	</h-card>
</template>

<script setup lang="ts">
	import { $t } from '@/i18n'
	import type { AppSettings } from '@/types'

	// 双向绑定设置对象
	const model = defineModel<AppSettings>({ required: true })

	// 定义目录选择事件
	const emit = defineEmits<{
		'pick-runtime': []
		'pick-project': []
	}>()
</script>

<style scoped lang="scss">
	/* 路径输入与按钮同行布局 */
	.path-row {
		display: flex;
		gap: 8px;
		align-items: center;
	}

	/* 自适应填充宽度 */
	.flex-1 {
		flex: 1;
	}
</style>
