<template>
	<div class="fields">
		<p class="hint">通过微信开发者工具 CLI 上传到微信后台开发版本，不会自动提审或正式上线。请先生成微信小程序构建产物，发布目录需包含 project.config.json；发布时会为每个绑定的 AppID 临时替换 appid。</p>
		<label>小程序 AppID<p-input-chips v-model="config.appIds" placeholder="输入 wx... 后按 Enter 添加" :allow-duplicate="false" /></label>
		<p class="hint">支持绑定多个小程序，输入一个 AppID 后按 Enter 添加为标签；发布时会逐个上传到对应的小程序后台。</p>
		<p class="hint">微信开发者工具 CLI 已在“全局设置”中统一配置。请在微信开发者工具的“设置 → 安全设置”中开启服务端口，并登录具有上传权限的账号。DevEnv 只调用本机 CLI，不自动安装工具或执行构建命令。</p>
		<div class="pair"><label>上传版本号<p-input-text v-model="config.version" placeholder="例如 1.0.0" /></label><label>机器人编号<p-input-number v-model="config.robot" :min="1" :max="30" :use-grouping="false" /></label></div>
		<label>版本说明<p-input-text v-model="config.description" placeholder="说明本次上传内容" /></label>
		<p class="hint">uni-app / Taro 请先编译项目，再上传项目根目录下的 dist/dev/mp-weixin；上传成功后仍需到微信后台手动提审和发布。</p>
	</div>
</template>

<script setup lang="ts">
import type { DeploymentConfig } from '@/types'
const config = defineModel<DeploymentConfig>({ required: true })
</script>

<style scoped lang="scss">
.fields, label { display: flex; flex-direction: column; gap: calc(var(--h-padding) / 2); }
.fields { gap: var(--h-padding); }
label { font-size: var(--font-sm); color: var(--text-secondary); }
.pair { display: grid; grid-template-columns: minmax(0, 2fr) minmax(0, 1fr); gap: var(--h-padding); }
.hint { margin: 0; color: var(--text-muted); font-size: var(--font-xs); overflow-wrap: anywhere; }
</style>
