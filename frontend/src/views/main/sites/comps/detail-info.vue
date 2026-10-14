<template>
	<!-- 详情页签1：网站基本信息与域名绑定 -->
	<div v-if="curSite" class="detail-pane">
		<div class="info-grid">
			<!-- 左侧基本信息卡片 -->
			<div class="info-card">
				<div class="sec-title">基本信息</div>
				<div class="form-field">
					<label>网站名称</label>
					<p-input-text v-model="editForm.name" />
				</div>
				<div class="form-row">
					<div class="form-field flex-1">
						<label>网站类型</label>
						<p-input-text :model-value="curSite.kind.toUpperCase()" disabled />
					</div>
					<div class="form-field flex-1">
						<label>监听端口</label>
						<input
							v-model.number="editForm.port"
							type="number"
							class="p-inputtext p-component font-mono"
						/>
					</div>
				</div>
				<div class="form-field">
					<label>网站根目录</label>
					<div class="input-row">
						<p-input-text v-model="editForm.root" class="font-mono flex-1" />
						<p-button
							label="选择目录"
							icon="pi pi-folder-open"
							severity="secondary"
							outlined
							@click="pickEditRoot"
						/>
					</div>
				</div>
				<div class="form-row">
					<div class="form-field flex-1">
						<label>入口文件</label>
						<p-input-text v-model="editForm.entry" class="font-mono" />
					</div>
					<div v-if="curSite.kind !== 'static'" class="form-field flex-1">
						<label>{{ getPortLabel(curSite.kind) }}</label>
						<input
							v-model.number="editForm.upstream"
							type="number"
							class="p-inputtext p-component font-mono"
						/>
					</div>
				</div>
			</div>

			<!-- 右侧域名绑定与运行环境卡片 -->
			<div class="info-card">
				<div class="sec-title">域名绑定与运行环境</div>
				<div class="form-field">
					<label>已绑定域名（每行一个）</label>
					<textarea
						v-model="domainsText"
						rows="4"
						class="code-area font-mono"
						placeholder="blog.local&#10;www.blog.local"
					></textarea>
				</div>
				<div v-if="curSite.kind !== 'static'" class="form-field">
					<label>关联运行环境</label>
					<p-select
						v-model="editForm.runtimeId"
						:options="runtimeOpts"
						optionLabel="label"
						optionValue="value"
						placeholder="选择运行环境"
					/>
				</div>
				<div class="chk-line">
					<input id="edit_autostart" v-model="editForm.autoStart" type="checkbox" />
					<label for="edit_autostart">随 DevEnv 启动时自动启动该网站</label>
				</div>
			</div>
		</div>

		<!-- 底部保存按钮栏 -->
		<div class="pane-footer">
			<p-button
				label="保存基本配置并重载 Nginx"
				icon="pi pi-save"
				:loading="saving"
				@click="saveInfo"
			/>
		</div>
	</div>
</template>

<script setup lang="ts">
	import { getPortLabel } from '../use-sites'
	import { useDetail } from '../use-detail'

	// 解构网站编辑状态与操作方法
	const {
		curSite,
		saving,
		editForm,
		domainsText,
		runtimeOpts,
		pickEditRoot,
		saveInfo
	} = useDetail()
</script>

<style scoped lang="scss">
	/* 页签主体容器 */
	.detail-pane {
		display: flex;
		flex-direction: column;
		gap: 14px;
	}

	/* 双列信息网格 */
	.info-grid {
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: 14px;
	}

	/* 信息分块卡片 */
	.info-card {
		background: var(--bg-app);
		border: 1px solid var(--border-color);
		border-radius: var(--radius-md);
		padding: 14px;
		display: flex;
		flex-direction: column;
		gap: 12px;
	}

	/* 卡片小节标题 */
	.sec-title {
		font-size: var(--font-sm);
		font-weight: 500;
		color: var(--text-primary);
	}

	/* 并排表单行 */
	.form-row {
		display: flex;
		gap: 10px;
	}

	/* 输入与按钮同行 */
	.input-row {
		display: flex;
		align-items: center;
		gap: 8px;
	}

	/* 复选框行 */
	.chk-line {
		display: inline-flex;
		align-items: center;
		gap: 8px;
		font-size: var(--font-sm);
		color: var(--text-secondary);
		cursor: pointer;
	}

	/* 多行代码输入框 */
	.code-area {
		width: 100%;
		padding: 8px 10px;
		background: var(--bg-card);
		border: 1px solid var(--border-color);
		border-radius: var(--radius-sm);
		color: var(--text-primary);
		font-size: var(--font-xs);
		resize: vertical;
	}

	/* 底部操作行 */
	.pane-footer {
		display: flex;
		justify-content: flex-end;
	}

	/* 自适应填充宽度 */
	.flex-1 {
		flex: 1;
	}
</style>
