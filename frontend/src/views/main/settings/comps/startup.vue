<template>
	<h-card title="软件启动">
		<p-message v-if="error !== ''" severity="error" size="small">
			{{ error }}
			<p-button label="重试" size="small" variant="text" :loading="busy" @click="load" />
		</p-message>
		<p-toolbar :pt="{ root: { style: { border: '0', padding: '0', gap: 'var(--h-padding)' } } }">
			<template #start>
				<div>
					<label for="application-startup">开机启动 DevEnv</label>
					<p class="text-xs text-muted">登录 Windows 后自动打开本软件，点击即生效；不影响 Nginx、数据库等服务的开机自启。</p>
				</div>
			</template>
			<template #end>
				<i v-if="busy" class="pi pi-spin pi-spinner" aria-label="正在更新" />
				<p-toggle-switch input-id="application-startup" :model-value="enabled" :disabled="busy || !ready" @update:model-value="change" />
			</template>
		</p-toolbar>
	</h-card>
</template>

<script setup lang="ts">
import { onMounted } from 'vue'
import { useStartup } from '@/composables/settings/use-startup'

const { enabled, ready, busy, error, load, change } = useStartup()
onMounted(() => void load())
</script>
