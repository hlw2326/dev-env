<template>
	<!-- 向导步骤2：基本信息与 Git 配置 -->
	<div class="wizard-body">
		<!-- 网站名称与监听端口 -->
		<div class="form-row">
			<div class="form-field flex-2">
				<label>网站名称 *</label>
				<p-input-text
					v-model="form.name"
					placeholder="如：我的博客 / 商城 API 服务"
					@input="onNameInput"
				/>
			</div>
			<div class="form-field flex-1">
				<label>监听端口</label>
				<input
					v-model.number="form.port"
					type="number"
					class="p-inputtext p-component font-mono"
					min="1"
					max="65535"
				/>
			</div>
		</div>

		<!-- 域名绑定列表 -->
		<div class="form-field">
			<label>域名绑定 *（默认 localhost，自定义域名将自动写入系统 hosts 文件）</label>
			<div class="domain-list">
				<div v-for="(_, idx) in form.domains" :key="idx" class="input-row">
					<p-input-text
						v-model="form.domains[idx]"
						class="font-mono flex-1"
						placeholder="如 localhost 或 test.com"
					/>
					<p-button
						v-if="form.domains.length > 1"
						icon="pi pi-times"
						severity="danger"
						text
						@click="form.domains.splice(idx, 1)"
					/>
				</div>
			</div>
			<div>
				<p-button
					label="+ 添加更多域名"
					size="small"
					severity="secondary"
					text
					@click="form.domains.push('')"
				/>
			</div>
		</div>

		<!-- 网站目录来源模式 -->
		<div class="form-field">
			<label>网站目录来源</label>
			<div class="radio-row">
				<label class="radio-opt">
					<input v-model="dirMode" type="radio" value="new" />
					<span>新建目录</span>
				</label>
				<label class="radio-opt">
					<input v-model="dirMode" type="radio" value="existing" />
					<span>使用现有目录</span>
				</label>
				<label class="radio-opt">
					<input v-model="dirMode" type="radio" value="git" />
					<span>从 Git 仓库克隆</span>
				</label>
			</div>
		</div>

		<!-- 网站根目录物理路径选择 -->
		<div class="form-field">
			<label>网站根目录物理路径 *</label>
			<div class="input-row">
				<p-input-text
					v-model="form.root"
					class="font-mono flex-1"
					placeholder="如 G:\www\blog"
				/>
				<p-button
					label="选择目录"
					icon="pi pi-folder-open"
					severity="secondary"
					outlined
					@click="pickCreateRoot"
				/>
			</div>
		</div>

		<!-- Git 远程仓库地址与分支 -->
		<div v-if="dirMode === 'git'" class="git-box">
			<div class="form-row">
				<div class="form-field flex-2">
					<label>Git 远程仓库地址 *</label>
					<p-input-text
						v-model="form.remote"
						class="font-mono"
						placeholder="https://github.com/user/repo.git"
					/>
				</div>
				<div class="form-field flex-1">
					<label>分支</label>
					<p-input-text
						v-model="form.branch"
						class="font-mono"
						placeholder="main"
					/>
				</div>
			</div>
		</div>
	</div>
</template>

<script setup lang="ts">
	import { useWizard } from '../use-wizard'

	// 解构向导表单与目录选择方法
	const {
		form,
		dirMode,
		onNameInput,
		pickCreateRoot
	} = useWizard()
</script>

<style scoped lang="scss">
	/* 向导步骤内容容器 */
	.wizard-body {
		display: flex;
		flex-direction: column;
		gap: 14px;
	}

	/* 并排表单行 */
	.form-row {
		display: flex;
		gap: 12px;
	}

	/* 域名列表垂直容器 */
	.domain-list {
		display: flex;
		flex-direction: column;
		gap: 8px;
	}

	/* 输入框与按钮同行 */
	.input-row {
		display: flex;
		align-items: center;
		gap: 8px;
	}

	/* 单选框水平组 */
	.radio-row {
		display: flex;
		gap: 16px;
		align-items: center;
	}

	/* 单个单选项 */
	.radio-opt {
		display: inline-flex;
		align-items: center;
		gap: 6px;
		font-size: var(--font-sm);
		color: var(--text-secondary);
		cursor: pointer;
	}

	/* Git 克隆附加框 */
	.git-box {
		background: var(--bg-app);
		border: 1px solid var(--border-color);
		border-radius: var(--radius-sm);
		padding: 12px;
	}

	/* 1倍宽度 */
	.flex-1 {
		flex: 1;
	}

	/* 2倍宽度 */
	.flex-2 {
		flex: 2;
	}
</style>
