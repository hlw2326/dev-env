<template>
	<!-- 服务进程管理卡片 -->
	<h-card :title="$t('服务进程管理')">
		<!-- 自动拉起核心服务开关 -->
		<div class="toggle-row">
			<div class="toggle-info">
				<span class="toggle-title">{{ $t('启动客户端时自动拉起核心服务') }}</span>
				<span class="toggle-sub">{{ $t('打开软件后自动启动已标记自启的 Nginx / MySQL 进程') }}</span>
			</div>
			<p-toggle-switch v-model="model.autoStartServices" />
		</div>
	</h-card>
</template>

<script setup lang="ts">
	import { $t } from '@/i18n'
	import type { AppSettings } from '@/types'

	// 双向绑定设置对象
	const model = defineModel<AppSettings>({ required: true })
</script>

<style scoped lang="scss">
	/* 开关行容器 */
	.toggle-row {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: 10px 14px;
		background: var(--bg-app);
		border: 1px solid var(--border-color);
		border-radius: var(--radius-md);
	}

	/* 开关文字区 */
	.toggle-info {
		display: flex;
		flex-direction: column;
		gap: 2px;
	}

	/* 开关主标题 */
	.toggle-title {
		font-size: var(--font-base);
		color: var(--text-primary);
	}

	/* 开关副标题 */
	.toggle-sub {
		font-size: var(--font-xs);
		color: var(--text-muted);
	}
</style>
