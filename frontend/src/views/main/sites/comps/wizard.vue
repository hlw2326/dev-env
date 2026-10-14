<template>
	<!-- 4步添加网站向导弹窗壳 -->
	<h-modal
		v-model:visible="visible"
		:title="`添加网站 - 步骤 ${step}/4：${stepTitle}`"
		width="680px"
	>
		<!-- 步骤进度指示条 -->
		<div class="steps-bar">
			<div
				v-for="s in 4"
				:key="s"
				class="step-pill"
				:class="{ active: step === s, done: step > s }"
				@click="jumpStep(s)"
			>
				<span class="step-num font-mono">{{ s }}</span>
				<span class="step-label">{{ stepLabels[s - 1] }}</span>
			</div>
		</div>

		<!-- 步骤1：选择类型 -->
		<wizard-kind v-if="step === 1" />

		<!-- 步骤2：基本信息 -->
		<wizard-basic v-if="step === 2" />

		<!-- 步骤3：运行环境 -->
		<wizard-env v-if="step === 3" />

		<!-- 步骤4：高级配置 -->
		<wizard-adv v-if="step === 4" />

		<!-- 向导底部操作按钮 -->
		<template #footer>
			<div class="wizard-footer">
				<p-button
					label="取消"
					severity="secondary"
					outlined
					@click="visible = false"
				/>
				<div class="nav-btns">
					<p-button
						v-if="step > 1"
						label="← 上一步"
						severity="secondary"
						outlined
						@click="step--"
					/>
					<p-button
						v-if="step < 4"
						label="下一步 →"
						@click="nextStep"
					/>
					<p-button
						v-else
						label="完成创建"
						icon="pi pi-check"
						:loading="submitting"
						@click="submitCreate"
					/>
				</div>
			</div>
		</template>
	</h-modal>
</template>

<script setup lang="ts">
	import WizardKind from './wizard-kind.vue'
	import WizardBasic from './wizard-basic.vue'
	import WizardEnv from './wizard-env.vue'
	import WizardAdv from './wizard-adv.vue'
	import { stepLabels, useWizard } from '../use-wizard'

	// 双向绑定弹窗可见性
	const visible = defineModel<boolean>('visible', { default: false })

	// 解构向导步骤与提交动作
	const {
		step,
		stepTitle,
		submitting,
		jumpStep,
		nextStep,
		submitCreate
	} = useWizard()
</script>

<style scoped lang="scss">
	/* 步骤指示条容器 */
	.steps-bar {
		display: grid;
		grid-template-columns: repeat(4, 1fr);
		gap: 8px;
		margin-bottom: 16px;
	}

	/* 单个步骤胶囊 */
	.step-pill {
		display: flex;
		align-items: center;
		gap: 8px;
		padding: 8px 10px;
		border-radius: var(--radius-sm);
		background: var(--bg-app);
		border: 1px solid var(--border-color);
		color: var(--text-muted);
		font-size: var(--font-xs);
		cursor: pointer;

		&.active {
			border-color: var(--p-primary-color);
			color: var(--p-primary-color);
			background: var(--bg-card);
		}

		&.done {
			border-color: var(--ok-border);
			color: var(--ok);
			background: var(--ok-bg);
		}
	}

	/* 步骤圆圈序号 */
	.step-num {
		width: 20px;
		height: 20px;
		border-radius: 50%;
		display: flex;
		align-items: center;
		justify-content: center;
		background: var(--bg-card);
		border: 1px solid currentColor;
	}

	/* 向导底部按钮栏 */
	.wizard-footer {
		width: 100%;
		display: flex;
		justify-content: space-between;
		align-items: center;
	}

	/* 上下步按钮组 */
	.nav-btns {
		display: flex;
		gap: 8px;
	}
</style>
