<template>
	<div class="deployment">
		<p class="notice">仅手动发布：保存配置、启动项目、提交 Git 和重启 DevEnv 都不会自动同步或上传。</p>
		<div v-if="error !== ''" class="error">{{ error }}<p-button size="small" label="重试" @click="load" /></div>
		<div v-else-if="config === null" class="hint"><i class="pi pi-spin pi-spinner"></i> 正在读取发布配置…</div>
		<template v-else>
			<fieldset :disabled="disabled" class="form">
				<label>发布源目录（相对项目根目录）<p-input-text v-model="config.source" placeholder=". 或 dist；小程序示例 dist/build/mp-weixin" /></label>
				<p class="hint">项目根目录：{{ project.root }}。需要构建的项目请先自行完成构建，再预览并同步产物。</p>
				<comp-deployment-server v-if="config.provider === 'sftp'" v-model="config" v-model:password="password" v-model:passphrase="passphrase" @pick="pick" />
				<comp-deployment-wechat v-else-if="config.provider === 'wechat'" v-model="config" @pick="pick" />
				<label>额外排除规则（每行一条）<p-textarea v-model="config.exclude" rows="3" placeholder="tests/&#10;*.map&#10;public/cache/" /></label>
				<p class="hint">始终排除 .git、.env*、node_modules、SSH 目录、常见私钥、凭据、数据库文件及 SQL 导出。服务器同步还排除 storage、runtime、logs、uploads；这些目录不会被覆盖。规则支持目录名、相对路径和 *.ext，不支持 **。</p>
			</fieldset>
			<p v-if="dirty" class="warning">配置尚未保存，请保存后检查或生成预览。</p>
			<div class="actions">
				<p-button label="保存配置" size="small" icon="pi pi-save" :disabled="disabled" :loading="busy === 'save'" @click="save" />
				<p-button label="检查连接 / 环境" size="small" severity="secondary" variant="outlined" :disabled="disabled || dirty" :loading="busy === 'test'" @click="test" />
				<p-button label="生成发布预览" size="small" severity="secondary" variant="outlined" icon="pi pi-list" :disabled="disabled || dirty" :loading="busy === 'preview'" @click="preview" />
				<p-button v-if="config.provider === 'wechat'" label="微信后台" size="small" severity="secondary" variant="outlined" icon="pi pi-external-link" @click="openWeChat" />
			</div>
			<comp-deployment-preview v-if="plan !== null" :plan="plan" />
			<div class="actions">
				<p-button v-if="plan !== null" label="确认并手动发布" severity="warn" icon="pi pi-upload" :disabled="disabled || dirty" :loading="busy === 'start'" @click="confirmStart" />
				<template v-if="running"><span class="hint"><i class="pi pi-spin pi-spinner"></i> 发布中，仅任务执行时刷新进度</span><p-button label="取消发布" severity="danger" variant="outlined" size="small" :loading="busy === 'cancel'" @click="cancel" /></template>
			</div>
			<comp-deployment-history :history="history" @refresh="refreshHistory" />
		</template>
	</div>
</template>

<script setup lang="ts">
import { useDeployment } from '@/composables'
import type { Project } from '@/types'
import CompDeploymentServer from './deployment-server.vue'
import CompDeploymentWechat from './deployment-wechat.vue'
import CompDeploymentPreview from './deployment-preview.vue'
import CompDeploymentHistory from './deployment-history.vue'
const props = defineProps<{ project: Project }>()
const { config, history, plan, password, passphrase, busy, error, running, disabled, dirty, load, save, preview, test, confirmStart, cancel, pick, refreshHistory, openWeChat } = useDeployment(() => props.project)
</script>

<style scoped lang="scss">
.deployment, .form, label { display: flex; flex-direction: column; gap: var(--h-padding); min-width: 0; }
.form { border: 0; padding: 0; margin: 0; }
label { gap: calc(var(--h-padding) / 2); font-size: var(--font-sm); color: var(--text-secondary); }
.notice { margin: 0; padding: var(--h-padding); border: 1px solid var(--border-color); border-radius: var(--h-radius); background: var(--bg-app); font-size: var(--font-sm); }
.hint { margin: 0; color: var(--text-muted); font-size: var(--font-xs); overflow-wrap: anywhere; }
.actions { display: flex; flex-wrap: wrap; align-items: center; gap: calc(var(--h-padding) / 2); }
.actions :deep(button) { flex-shrink: 0; }
.warning { margin: 0; color: var(--warn); font-size: var(--font-sm); }
.error { color: var(--err); overflow-wrap: anywhere; }
</style>
