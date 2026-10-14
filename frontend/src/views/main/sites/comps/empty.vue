<template>
	<!-- 暂无网站空状态引导卡片 -->
	<div class="empty-sites">
		<i class="pi pi-globe empty-icon"></i>
		<p class="empty-title">暂无符合条件的网站</p>
		<p class="empty-sub">
			点击右上角「添加网站」可一键配置项目目录、域名映射、PHP/Node.js/Go/Python 运行环境与 Git 仓库
		</p>
		<p-button
			label="立即添加网站"
			icon="pi pi-plus"
			@click="emit('create')"
		/>
	</div>
</template>

<script setup lang="ts">
	// 定义立即创建网站事件
	const emit = defineEmits<{
		create: []
	}>()
</script>

<style scoped lang="scss">
	/* 空状态外层容器 */
	.empty-sites {
		text-align: center;
		padding: 60px 20px;
		background: var(--bg-card);
		border: 1px dashed var(--border-color);
		border-radius: var(--h-radius);
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 8px;
	}

	/* 空状态地球图标 */
	.empty-icon {
		font-size: 2.5rem;
		color: var(--text-muted);
	}

	/* 空状态主标题 */
	.empty-title {
		font-size: var(--font-lg);
		color: var(--text-primary);
		margin: 0;
	}

	/* 空状态副标题说明 */
	.empty-sub {
		font-size: var(--font-sm);
		color: var(--text-muted);
		margin: 0 0 8px 0;
	}
</style>
