<template>
	<div class="fields">
		<p class="hint">上传到微信后台开发版本，不会自动提审或正式上线。请先生成微信小程序构建产物，发布目录需包含 project.config.json，且其中 appid 与下方一致。</p>
		<label>小程序 AppID<p-input-text v-model="config.appId" placeholder="wx…" /></label>
		<label>微信代码上传私钥<div class="path"><p-input-text v-model="config.uploadKey" placeholder="从微信后台下载的代码上传密钥文件" /><p-button size="small" severity="secondary" variant="outlined" label="选择" @click="emit('pick', 'uploadKey')" /></div></label>
		<p class="hint">私钥不是 AppSecret。需在微信后台的开发设置中启用代码上传并配置本机出口 IP 白名单；不要将私钥提交到 Git。</p>
		<div class="pair"><label>上传版本号<p-input-text v-model="config.version" placeholder="例如 1.0.0" /></label><label>CI 机器人编号<p-input-number v-model="config.robot" :min="1" :max="30" :use-grouping="false" /></label></div>
		<label>版本说明<p-input-text v-model="config.description" placeholder="说明本次上传内容" /></label>
		<label>Node.js 可执行文件（留空从 PATH 查找）<div class="path"><p-input-text v-model="config.nodePath" placeholder="node.exe 的绝对路径" /><p-button size="small" severity="secondary" variant="outlined" label="选择" @click="emit('pick', 'nodePath')" /></div></label>
		<label>CI 工具目录（留空使用项目依赖）<div class="path"><p-input-text v-model="config.modulePath" placeholder="包含 node_modules/miniprogram-ci 的目录" /><p-button size="small" severity="secondary" variant="outlined" label="选择" @click="emit('pick', 'modulePath')" /></div></label>
		<p class="hint">需提前在项目或独立工具目录安装官方 miniprogram-ci。这里只检查、调用已安装工具，不自动下载软件或执行构建命令；uni-app / Taro 请填写构建后的微信目录。</p>
	</div>
</template>

<script setup lang="ts">
import type { DeploymentConfig } from '@/types'
const config = defineModel<DeploymentConfig>({ required: true })
const emit = defineEmits<{ pick: [field: 'uploadKey' | 'nodePath' | 'modulePath'] }>()
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
