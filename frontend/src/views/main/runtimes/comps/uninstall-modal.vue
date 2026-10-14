<template>
	<!-- 卸载运行环境确认弹窗 -->
	<h-modal
		v-model:visible="visible"
		title="卸载运行环境确认"
		width="460px"
	>
		<div v-if="runtime" class="uninstall-body">
			<!-- 卸载警告说明 -->
			<div class="warn-box">
				<i class="pi pi-exclamation-triangle warn-icon"></i>
				<div>
					<div class="warn-title">
						即将卸载环境：{{ getTypeName(runtime.runtimeType) }} {{ runtime.version }}
					</div>
					<div class="warn-desc">此操作将从 DevEnv 中移除该运行时的注册信息。</div>
				</div>
			</div>

			<!-- 存在引用冲突提示 -->
			<div v-if="blockReason" class="err-box">
				<i class="pi pi-times-circle"></i>
				<span>{{ blockReason }}</span>
			</div>

			<!-- 最后一个版本的联动删除提示 -->
			<div v-else-if="isLastServiceVersion" class="warn-box">
				<i class="pi pi-info-circle warn-icon"></i>
				<div class="warn-desc">
					当前版本是 {{ getTypeName(runtime.runtimeType) }} 的唯一已安装版本，卸载后将同步移除对应的服务实例（请确保该服务实例已停止）。
				</div>
			</div>

			<!-- 删除物理文件勾选项 -->
			<label class="purge-row">
				<input v-model="purge" type="checkbox" />
				<span>同时永久删除物理文件目录（<code>{{ runtime.installPath }}</code>）</span>
			</label>
		</div>

		<template #footer>
			<p-button label="取消" severity="secondary" outlined @click="visible = false" />
			<p-button
				label="确认卸载"
				severity="danger"
				icon="pi pi-trash"
				:loading="loading"
				:disabled="Boolean(blockReason)"
				@click="emit('confirm')"
			/>
		</template>
	</h-modal>
</template>

<script setup lang="ts">
	import { computed } from 'vue'
	import type { RuntimeVersion } from '@/types'
	import { getTypeName } from '../use-runtimes'

	// 双向绑定弹窗可见性
	const visible = defineModel<boolean>('visible', { default: false })

	// 双向绑定是否删除物理文件
	const purge = defineModel<boolean>('purge', { default: false })

	// 接收待卸载运行时与加载状态
	const props = defineProps<{
		runtime: RuntimeVersion | null
		sameTypeCount?: number
		loading: boolean
	}>()

	// 定义确认卸载事件
	const emit = defineEmits<{
		confirm: []
	}>()

	// 阻止卸载的原因（网站引用 或 服务使用中且同类型还有其他版本）
	const blockReason = computed(() => {
		if (!props.runtime) {
			return ''
		}
		if ((props.runtime.siteCount ?? 0) > 0) {
			return `该版本正在被 ${props.runtime.siteCount} 个网站使用，请先在网站管理中解绑`
		}
		if ((props.runtime.serviceCount ?? 0) > 0 && (props.sameTypeCount ?? 1) > 1) {
			return '当前服务实例正在使用此版本，请先在服务实例页切换到其他版本后再卸载'
		}
		return ''
	})

	// 是否为当前类型的最后一个版本且关联了服务实例
	const isLastServiceVersion = computed(() => {
		if (!props.runtime) {
			return false
		}
		return (props.runtime.serviceCount ?? 0) > 0 && (props.sameTypeCount ?? 1) <= 1
	})
</script>

<style scoped lang="scss">
	/* 卸载弹窗内容容器 */
	.uninstall-body {
		display: flex;
		flex-direction: column;
		gap: 14px;
	}

	/* 警告提示框 */
	.warn-box {
		display: flex;
		align-items: flex-start;
		gap: 10px;
		padding: 12px;
		background: var(--warn-bg);
		border: 1px solid var(--warn-border);
		border-radius: var(--radius-sm);
		color: var(--warn);
	}

	/* 警告图标 */
	.warn-icon {
		font-size: 1.2rem;
		margin-top: 2px;
	}

	/* 警告主标题 */
	.warn-title {
		font-size: var(--font-sm);
		color: var(--text-primary);
		margin-bottom: 4px;
	}

	/* 警告副说明 */
	.warn-desc {
		font-size: var(--font-xs);
		color: var(--text-secondary);
	}

	/* 引用冲突报错框 */
	.err-box {
		display: flex;
		align-items: center;
		gap: 8px;
		padding: 10px 12px;
		background: var(--err-bg);
		border: 1px solid var(--err-border);
		border-radius: var(--radius-sm);
		color: var(--err);
		font-size: var(--font-xs);
	}

	/* 物理删除复选框行 */
	.purge-row {
		display: flex;
		align-items: center;
		gap: 8px;
		font-size: var(--font-sm);
		color: var(--text-secondary);
		cursor: pointer;
	}
</style>
