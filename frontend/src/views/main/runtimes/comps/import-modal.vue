<template>
	<!-- 导入本地运行环境弹窗 -->
	<h-modal
		v-model:visible="visible"
		title="导入本地运行环境"
		width="520px"
	>
		<div class="import-body">
			<!-- 环境类型、版本与架构选择 -->
			<div class="form-row">
				<div class="form-field flex-1">
					<label>环境类型</label>
					<p-select
						v-model="form.kind"
						:options="kindOpts"
						optionLabel="label"
						optionValue="value"
					/>
				</div>
				<div class="form-field flex-1">
					<label>版本号 *</label>
					<p-input-text v-model="form.version" class="font-mono" placeholder="如 8.3.12" />
				</div>
				<div class="form-field flex-1">
					<label>系统架构</label>
					<p-select
						v-model="form.arch"
						:options="archOpts"
						optionLabel="label"
						optionValue="value"
					/>
				</div>
			</div>

			<!-- 本地安装目录选择 -->
			<div class="form-field">
				<label>本地安装目录路径 *</label>
				<div class="path-row">
					<p-input-text
						v-model="form.dir"
						class="font-mono flex-1"
						placeholder="如 D:\php-8.3.12-Win32-vs16-x64"
					/>
					<p-button
						label="选择目录"
						icon="pi pi-folder-open"
						severity="secondary"
						outlined
						@click="emit('pick-dir')"
					/>
				</div>
			</div>
		</div>

		<template #footer>
			<p-button label="取消" severity="secondary" outlined @click="visible = false" />
			<p-button
				label="确认导入"
				icon="pi pi-check"
				:loading="loading"
				@click="emit('submit')"
			/>
		</template>
	</h-modal>
</template>

<script setup lang="ts">
	// 双向绑定弹窗可见性
	const visible = defineModel<boolean>('visible', { default: false })

	// 双向绑定导入表单数据
	const form = defineModel<{
		kind: string
		version: string
		arch: string
		dir: string
	}>('form', { required: true })

	// 接收导入加载状态
	defineProps<{
		loading: boolean
	}>()

	// 定义选择目录与提交事件
	const emit = defineEmits<{
		'pick-dir': []
		submit: []
	}>()

	// 环境类型下拉选项
	const kindOpts = [
		{ label: 'PHP', value: 'php' },
		{ label: 'Nginx', value: 'nginx' },
		{ label: 'MySQL', value: 'mysql' },
		{ label: 'Redis', value: 'redis' },
		{ label: 'PostgreSQL', value: 'postgresql' },
		{ label: 'Node.js', value: 'nodejs' }
	]

	// 系统架构下拉选项
	const archOpts = [
		{ label: 'x64 (64位)', value: 'x64' },
		{ label: 'x86 (32位)', value: 'x86' }
	]
</script>

<style scoped lang="scss">
	/* 导入表单内容区 */
	.import-body {
		display: flex;
		flex-direction: column;
		gap: 14px;
	}

	/* 三列并排表单行 */
	.form-row {
		display: flex;
		gap: 10px;
	}

	/* 路径选择输入行 */
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
