<template>
	<!-- 向导步骤1：选择网站类型 -->
	<div class="wizard-body">
		<p class="step-desc">
			请选择您要创建的网站或后端应用类型，系统将自动匹配最佳的 Nginx 与服务模板：
		</p>
		<div class="type-grid">
			<div
				v-for="t in wizardKinds"
				:key="t.kind"
				class="type-card"
				:class="{ selected: form.kind === t.kind }"
				@click="selectKind(t.kind)"
			>
				<div class="type-icon" :class="`kind-${t.kind}`">
					<i :class="t.icon"></i>
				</div>
				<div class="type-title">{{ t.title }}</div>
				<div class="type-mode">{{ t.mode }}</div>
				<div class="type-ex font-mono">{{ t.examples }}</div>
			</div>
		</div>
	</div>
</template>

<script setup lang="ts">
	import { wizardKinds, useWizard } from '../use-wizard'

	// 解构创建表单与类型切换方法
	const { form, selectKind } = useWizard()
</script>

<style scoped lang="scss">
	/* 向导步骤内容容器 */
	.wizard-body {
		display: flex;
		flex-direction: column;
		gap: 14px;
	}

	/* 顶部步骤引导文字 */
	.step-desc {
		font-size: var(--font-sm);
		color: var(--text-secondary);
		margin: 0;
	}

	/* 类型卡片网格 */
	.type-grid {
		display: grid;
		grid-template-columns: repeat(3, 1fr);
		gap: 12px;
	}

	/* 单个网站类型卡片 */
	.type-card {
		border: 1px solid var(--border-color);
		border-radius: var(--radius-md);
		padding: 14px;
		background: var(--bg-app);
		display: flex;
		flex-direction: column;
		gap: 6px;
		cursor: pointer;

		&.selected {
			border-color: var(--p-primary-color);
			background: var(--bg-card);
		}
	}

	/* 类型图标块 */
	.type-icon {
		width: 36px;
		height: 36px;
		border-radius: var(--radius-sm);
		display: flex;
		align-items: center;
		justify-content: center;
		font-size: 1.1rem;
		background: var(--info-bg);
		color: var(--info);

		&.kind-php {
			background: var(--indigo-bg);
			color: var(--indigo);
		}

		&.kind-nodejs {
			background: var(--ok-bg);
			color: var(--ok);
		}

		&.kind-python {
			background: var(--warn-bg);
			color: var(--warn);
		}
	}

	/* 类型名称 */
	.type-title {
		font-size: var(--font-base);
		color: var(--text-primary);
		margin-top: 2px;
	}

	/* 代理工作模式说明 */
	.type-mode {
		font-size: var(--font-xs);
		color: var(--text-secondary);
	}

	/* 框架示例文字 */
	.type-ex {
		font-size: var(--font-xs);
		color: var(--text-muted);
	}
</style>
