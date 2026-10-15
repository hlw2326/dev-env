<template>
	<!-- 最近网站条目行 -->
	<div class="site-row" @click="$router.push('/sites')">
		<!-- 左侧网站名称与路径 -->
		<div class="info">
			<div class="title-line">
				<span class="name">{{ site.name }}</span>
				<span class="kind font-mono">{{ site.kind.toUpperCase() }}</span>
			</div>
			<span class="path font-mono" :title="hostText">
				{{ hostText }} · {{ site.root }}
			</span>
		</div>

		<!-- 右侧状态标签 -->
		<p-tag
			v-if="site.state === 'running'"
			severity="success"
			value="运行中"
			size="small"
		/>
		<p-tag
			v-else-if="site.need !== ''"
			severity="warn"
			value="缺少环境"
			size="small"
		/>
		<p-tag
			v-else-if="site.state === 'error'"
			severity="danger"
			value="异常"
			size="small"
		/>
		<p-tag
			v-else
			severity="secondary"
			value="已停止"
			size="small"
		/>
	</div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { Site } from '@/types'

// 接收站点属性
const props = defineProps<{
	site: Site
}>()

// 域名拼接文本
const hostText = computed(() => {
	if (props.site.hosts.length === 0) {
		return '本地调试'
	}
	return props.site.hosts.join(', ')
})
</script>

<style scoped lang="scss">
/* 站点行容器 */
.site-row {
	display: flex;
	justify-content: space-between;
	align-items: center;
	padding: 0.75rem 0.875rem;
	background: var(--bg-app);
	border: 1px solid var(--border-color);
	border-radius: var(--h-radius);
	cursor: pointer;
	transition: all 0.15s ease;

	/* 悬停态 */
	&:hover {
		background: var(--bg-card);
		border-color: var(--border-hover);
	}

	/* 左侧文字区 */
	.info {
		display: flex;
		flex-direction: column;
		gap: 0.2rem;
		min-width: 0;
	}

	/* 站点名与类型行 */
	.title-line {
		display: flex;
		align-items: center;
		gap: 0.5rem;
	}

	/* 站点名称 */
	.name {
		font-size: var(--font-base);
		color: var(--text-primary);
	}

	/* 站点类型标识 */
	.kind {
		font-size: var(--font-xs);
		color: var(--text-muted);
	}

	/* 域名与目录路径 */
	.path {
		font-size: var(--font-xs);
		color: var(--text-muted);
		max-width: 20rem;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}
}
</style>
