<template>
	<!-- 网站管理详情综合模态框壳 -->
	<h-modal
		v-model:visible="visible"
		:title="modalTitle"
		width="920px"
		:footer="false"
	>
		<div v-if="curSite" class="detail-box">
			<!-- 顶部详情页签导航栏 -->
			<div class="tabs-nav">
				<button
					v-for="t in tabs"
					:key="t.key"
					type="button"
					class="tab-btn"
					:class="{ active: curTab === t.key }"
					@click="switchTab(t.key)"
				>
					<i :class="t.icon"></i>
					<span>{{ t.label }}</span>
				</button>
			</div>

			<!-- 动态挂载当前激活的详情页签组件 -->
			<component :is="activeComp" />
		</div>
	</h-modal>
</template>

<script setup lang="ts">
	import { computed, defineAsyncComponent, type Component } from 'vue'
	import { useDetail, type DetailTab } from '../use-detail'

	// 双向绑定弹窗可见性
	const visible = defineModel<boolean>('visible', { default: false })

	// 解构当前网站与激活页签
	const { curSite, curTab, switchTab } = useDetail()

	// 详情页签配置与异步子组件表
	const tabs: Array<{
		key: DetailTab
		label: string
		icon: string
		comp: Component
	}> = [
		{
			key: 'info',
			label: '基本信息',
			icon: 'pi pi-sliders-h',
			comp: defineAsyncComponent(() => import('./detail-info.vue'))
		},
		{
			key: 'git',
			label: 'Git 仓库',
			icon: 'pi pi-github',
			comp: defineAsyncComponent(() => import('./detail-git.vue'))
		},
		{
			key: 'conf',
			label: 'Nginx 配置',
			icon: 'pi pi-code',
			comp: defineAsyncComponent(() => import('./detail-conf.vue'))
		},
		{
			key: 'ssl',
			label: 'SSL 证书',
			icon: 'pi pi-lock',
			comp: defineAsyncComponent(() => import('./detail-ssl.vue'))
		},
		{
			key: 'logs',
			label: '访问与错误日志',
			icon: 'pi pi-file',
			comp: defineAsyncComponent(() => import('./detail-logs.vue'))
		}
	]

	// 弹窗标题文本
	const modalTitle = computed(() => {
		if (curSite.value) {
			return `网站管理：${curSite.value.name}`
		}
		return '网站详情'
	})

	// 当前激活页签对应的动态组件
	const activeComp = computed(() => {
		const hit = tabs.find((t) => t.key === curTab.value)
		if (hit) {
			return hit.comp
		}
		return tabs[0].comp
	})
</script>

<style scoped lang="scss">
	/* 详情内容整体容器 */
	.detail-box {
		display: flex;
		flex-direction: column;
		gap: 14px;
	}

	/* 顶部页签切换条 */
	.tabs-nav {
		display: flex;
		gap: 6px;
		border-bottom: 1px solid var(--border-color);
		padding-bottom: 10px;
	}

	/* 单个页签按钮 */
	.tab-btn {
		display: inline-flex;
		align-items: center;
		gap: 6px;
		padding: 6px 14px;
		border-radius: var(--radius-sm);
		border: 1px solid transparent;
		background: transparent;
		color: var(--text-secondary);
		font-size: var(--font-sm);
		cursor: pointer;

		&:hover {
			background: var(--bg-app);
			color: var(--text-primary);
		}

		&.active {
			background: var(--p-primary-color);
			color: var(--on-primary);
		}
	}
</style>
