<template>
	<!-- 全局设置页面 -->
	<h-page title="全局设置" description="配置默认站点目录、本地网关端口、hosts 自动写入与界面主题">
		<!-- 页面右上角保存按钮 -->
		<template #extra>
			<p-button
				icon="pi pi-check"
				label="保存设置"
				:loading="saving"
				@click="saveAll"
			/>
		</template>

		<!-- 网关与站点设置卡片 -->
		<comp-gateway />

		<!-- 外观与存储设置卡片 -->
		<comp-theme />
	</h-page>
</template>

<script setup lang="ts">
import { onMounted } from 'vue'
import { useSettings } from './use-settings'
import CompGateway from './comps/gateway.vue'
import CompTheme from './comps/theme.vue'

// 全局设置状态与操作
const { saving, fetchData, saveAll } = useSettings()

// 挂载时读取全局设置
onMounted(() => {
	fetchData()
})
</script>
