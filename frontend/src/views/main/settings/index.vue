<template>
	<!-- 全局设置页面 -->
	<h-page title="全局设置" description="配置默认项目目录、Nginx 监听端口、微信开发者工具 CLI、hosts 自动写入与界面主题">
		<!-- 页面右上角保存按钮 -->
		<template #extra>
			<p-button
				icon="pi pi-check"
				label="保存设置"
				:loading="saving"
				@click="saveAll"
			/>
		</template>

		<!-- Nginx 与项目设置卡片 -->
		<div class="settings-content">
			<comp-gateway />
			<comp-wechat />
			<comp-theme />
		</div>
	</h-page>
</template>

<script setup lang="ts">
import { onMounted } from 'vue'
import { useSettings } from '@/composables'
import CompGateway from './comps/gateway.vue'
import CompWechat from './comps/wechat.vue'
import CompTheme from './comps/theme.vue'

// 全局设置状态与操作
const { saving, fetchData, saveAll } = useSettings()

// 挂载时读取全局设置
onMounted(() => {
	fetchData()
})
</script>

<style scoped lang="scss">
.settings-content {
	display: flex;
	flex: 0 0 auto;
	flex-direction: column;
	gap: var(--h-padding);
	min-height: max-content;
	width: 100%;
}
</style>
