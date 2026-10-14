<template>
	<!-- 向导步骤3：运行环境与服务实例 -->
	<div class="wizard-body">
		<!-- 静态站点无需语言运行环境提示 -->
		<template v-if="form.kind === 'static'">
			<div class="info-banner">
				<i class="pi pi-check-circle text-ok"></i>
				<div>
					<div>静态网站由 Nginx 直接高效代理，无需额外语言进程服务。</div>
					<div class="text-xs text-muted">
						系统将自动为您生成支持 SPA 路由回退与静态缓存的 Nginx 虚拟主机配置。
					</div>
				</div>
			</div>
		</template>

		<!-- 动态语言环境与转发配置 -->
		<template v-else>
			<div class="form-field">
				<label>选择 {{ form.kind.toUpperCase() }} 运行环境版本</label>
				<div v-if="matchingRuntimes.length === 0" class="empty-inline">
					<span>暂未检测到已安装的 {{ form.kind.toUpperCase() }} 运行环境</span>
					<p-button
						label="前往软件商店安装"
						icon="pi pi-shopping-bag"
						size="small"
						severity="secondary"
						outlined
						@click="goStore"
					/>
				</div>
				<div v-else class="rt-list">
					<label
						v-for="rt in matchingRuntimes"
						:key="rt.id"
						class="rt-card"
						:class="{ active: form.runtimeId === rt.id }"
					>
						<input
							v-model="form.runtimeId"
							type="radio"
							:value="rt.id"
						/>
						<div class="rt-info">
							<div class="rt-title">
								<span>{{ rt.runtimeType.toUpperCase() }} {{ rt.version }} ({{ rt.architecture }})</span>
								<p-tag v-if="rt.isDefault" value="默认" severity="info" size="small" />
							</div>
							<div class="rt-path font-mono">{{ rt.installPath }}</div>
						</div>
					</label>
				</div>
			</div>

			<!-- 关联后台服务或自动创建端口实例 -->
			<div class="form-field">
				<label>关联后台服务 / 转发端口配置</label>
				<div class="radio-row">
					<label class="radio-opt">
						<input v-model="form.autoService" type="radio" :value="true" />
						<span>自动创建/绑定端口实例</span>
					</label>
					<label class="radio-opt">
						<input v-model="form.autoService" type="radio" :value="false" />
						<span>使用现有服务实例</span>
					</label>
				</div>
			</div>

			<div v-if="!form.autoService" class="form-field">
				<label>选择现有服务实例</label>
				<p-select
					v-model="form.serviceId"
					:options="serviceOptions"
					optionLabel="label"
					optionValue="value"
					placeholder="请选择已创建的服务实例"
				/>
			</div>

			<!-- 转发端口与启动命令 -->
			<div class="form-row">
				<div class="form-field flex-1">
					<label>{{ getPortLabel(form.kind) }}</label>
					<input
						v-model.number="form.upstream"
						type="number"
						class="p-inputtext p-component font-mono"
						min="1"
						max="65535"
					/>
				</div>
				<div v-if="form.kind !== 'php'" class="form-field flex-2">
					<label>应用启动命令 / 入口</label>
					<p-input-text
						v-model="form.cmd"
						class="font-mono"
						:placeholder="cmdHint"
					/>
				</div>
			</div>
		</template>
	</div>
</template>

<script setup lang="ts">
	import { useRouter } from 'vue-router'
	import { getPortLabel } from '../use-sites'
	import { useWizard } from '../use-wizard'

	// 路由实例
	const router = useRouter()

	// 解构向导运行环境状态
	const {
		showWizard,
		form,
		matchingRuntimes,
		serviceOptions,
		cmdHint
	} = useWizard()

	// 跳转到软件商店页面
	function goStore() {
		showWizard.value = false
		router.push('/store')
	}
</script>

<style scoped lang="scss">
	/* 向导步骤内容容器 */
	.wizard-body {
		display: flex;
		flex-direction: column;
		gap: 14px;
	}

	/* 静态站点提示横幅 */
	.info-banner {
		display: flex;
		align-items: flex-start;
		gap: 10px;
		padding: 14px;
		background: var(--ok-bg);
		border: 1px solid var(--ok-border);
		border-radius: var(--radius-sm);
		color: var(--text-primary);
		font-size: var(--font-sm);
	}

	/* 未安装运行环境提示框 */
	.empty-inline {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: 12px;
		background: var(--bg-app);
		border: 1px dashed var(--border-color);
		border-radius: var(--radius-sm);
		font-size: var(--font-sm);
		color: var(--text-muted);
	}

	/* 运行环境单选列表 */
	.rt-list {
		display: flex;
		flex-direction: column;
		gap: 8px;
		max-height: 180px;
		overflow-y: auto;
	}

	/* 单个运行环境选项卡片 */
	.rt-card {
		display: flex;
		align-items: center;
		gap: 10px;
		padding: 10px 12px;
		border: 1px solid var(--border-color);
		border-radius: var(--radius-sm);
		cursor: pointer;

		&.active {
			border-color: var(--p-primary-color);
			background: var(--bg-app);
		}
	}

	/* 运行环境标题行 */
	.rt-title {
		display: flex;
		align-items: center;
		gap: 8px;
		font-size: var(--font-sm);
		color: var(--text-primary);
	}

	/* 运行环境路径 */
	.rt-path {
		font-size: var(--font-xs);
		color: var(--text-muted);
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

	/* 并排表单行 */
	.form-row {
		display: flex;
		gap: 12px;
	}

	/* 1倍宽度 */
	.flex-1 {
		flex: 1;
	}

	/* 2倍宽度 */
	.flex-2 {
		flex: 2;
	}

	/* 成功绿字 */
	.text-ok {
		color: var(--ok);
	}

	/* 小号文字 */
	.text-xs {
		font-size: var(--font-xs);
	}

	/* 弱化文字 */
	.text-muted {
		color: var(--text-muted);
	}
</style>
