<template>
	<div class="fields">
		<div class="pair"><label>SSH 主机<p-input-text v-model="config.host" placeholder="服务器 IP 或主机名" /></label><label>端口<p-input-number v-model="config.port" :min="1" :max="65535" :use-grouping="false" /></label></div>
		<label>SSH 用户名<p-input-text v-model="config.username" placeholder="建议使用仅有项目目录权限的部署账号" /></label>
		<label>服务器目标目录<p-input-text v-model="config.remote" placeholder="/var/www/my-project（须提前创建）" /></label>
		<label>服务器指纹<p-input-text v-model="config.fingerprint" placeholder="SHA256:…" /></label>
		<p class="hint">必须通过服务器管理员等可信渠道核验 SSH 指纹。不自动接受未知主机，也不忽略指纹不匹配。</p>
		<label>认证方式
			<p-select
				v-model="authMethod"
				:options="authMethods"
				option-label="label"
				option-value="value"
				@change="changeAuthMethod"
			/>
		</label>
		<label v-if="authMethod === 'password'">本次连接密码<p-input-text v-model="password" type="password" autocomplete="off" placeholder="仅用于本次操作，不保存" /></label>
		<template v-else>
			<label>SSH 私钥文件
				<div class="path"><p-input-text v-model="config.privateKey" placeholder="本机私钥绝对路径，内容不会保存到数据库" /><p-button size="small" severity="secondary" variant="outlined" label="选择" icon="pi pi-folder-open" @click="emit('pick', 'privateKey')" /></div>
			</label>
			<label>私钥口令（加密私钥才需要）<p-input-text v-model="passphrase" type="password" autocomplete="off" placeholder="仅用于本次操作，不保存" /></label>
		</template>
		<p class="hint">测试后会清空密码和口令；发布前请重新填写。需要目标目录写权限及 SFTP POSIX rename 支持。不会执行服务器命令、安装依赖或重启 Node.js 服务。</p>
	</div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import type { DeploymentConfig } from '@/types'
const config = defineModel<DeploymentConfig>({ required: true })
const password = defineModel<string>('password', { required: true })
const passphrase = defineModel<string>('passphrase', { required: true })
const emit = defineEmits<{ pick: [field: 'privateKey'] }>()
const authMethods = [
	{ value: 'password', label: '密码' },
	{ value: 'privateKey', label: 'SSH 私钥' },
]
const authMethod = ref<'password' | 'privateKey'>('password')

watch(
	() => config.value.privateKey,
	(value) => { authMethod.value = value === '' ? 'password' : 'privateKey' },
	{ immediate: true },
)

function changeAuthMethod(): void {
	if (authMethod.value === 'password') {
		config.value.privateKey = ''
		passphrase.value = ''
		return
	}
	password.value = ''
}
</script>

<style scoped lang="scss">
.fields, label { display: flex; flex-direction: column; gap: calc(var(--h-padding) / 2); }
.fields { gap: var(--h-padding); }
label { font-size: var(--font-sm); color: var(--text-secondary); }
.pair { display: grid; grid-template-columns: minmax(0, 2fr) minmax(0, 1fr); gap: var(--h-padding); }
.path { display: flex; min-width: 0; gap: calc(var(--h-padding) / 2); }
.path :deep(input) { min-width: 0; flex: 1; }
.path :deep(button) { flex-shrink: 0; }
.hint { margin: 0; color: var(--text-muted); font-size: var(--font-xs); overflow-wrap: anywhere; }
</style>
