<template>
	<!-- 网关与站点设置卡片 -->
	<h-card title="Nginx 与站点">
		<div class="settings-block">
			<!-- Nginx 监听状态 -->
			<div class="gateway-status">
				<span class="section-title">Nginx 状态</span>
				<div class="status-right">
					<span v-if="listening" class="status-pill ok">
						<span class="dot"></span>
						<span>Nginx 正在监听 127.0.0.1:{{ form.port }}</span>
					</span>
					<span v-else class="status-pill error">
						<span class="dot"></span>
						<span>Nginx 未监听</span>
					</span>
				</div>
			</div>

			<div class="config-row">
				<div class="toggle-info">
					<span class="section-title">Nginx 配置文件</span>
				<span class="field-hint">打开 Nginx 主配置文件 nginx.conf</span>
				</div>
				<p-button
					severity="secondary"
					variant="outlined"
					icon="pi pi-file-edit"
					label="打开配置文件"
					@click="openNginxConfig"
				/>
			</div>

			<!-- 网关启动错误提示 -->
			<div v-if="nginxError !== ''" class="gateway-err">
				{{ nginxError }}
			</div>

			<!-- 新站点默认父目录 -->
			<div class="form-field">
				<label>新站点默认父目录</label>
				<div class="folder-row">
					<p-input-text v-model="form.root" class="font-mono flex-input" />
					<p-button
						severity="secondary"
						variant="outlined"
						icon="pi pi-folder-open"
						label="选择"
						@click="pickRoot"
					/>
				</div>
				<span class="field-hint">创建新网站时默认使用该目录作为父目录</span>
			</div>

			<!-- Nginx 监听端口 -->
			<div class="form-field">
				<label>Nginx 监听端口</label>
				<p-input-number
					v-model="form.port"
					:use-grouping="false"
					:min="1"
					:max="65535"
					class="port-input"
				/>
				<span class="field-hint">所有网站通过 Nginx 共用此端口（默认 80），端口被占用时请修改并保存</span>
			</div>

			<!-- 自动写入系统 hosts 开关 -->
			<div class="toggle-row">
				<div class="toggle-info">
					<span class="section-title">自动写入系统 hosts</span>
					<span class="field-hint">
						*.localhost 域名无需写入 hosts；仅自定义后缀（如 .test）需要开启此项并以管理员权限运行
					</span>
				</div>
				<p-toggle-switch v-model="form.hosts" />
			</div>
		</div>
	</h-card>
</template>

<script setup lang="ts">
import { useSettings } from '../use-settings'

// 网关与站点设置状态与方法
	const { form, listening, nginxError, pickRoot, openNginxConfig } = useSettings()
</script>

<style scoped lang="scss">
/* 设置块容器 */
.settings-block {
	display: flex;
	flex-direction: column;
	gap: 0.75rem;
}

/* 网关状态横幅 */
.gateway-status {
	display: flex;
	align-items: center;
	justify-content: space-between;
	padding: 0.75rem 0.875rem;
	background: var(--bg-app);
	border: 1px solid var(--border-color);
	border-radius: var(--h-radius);
}

/* 小节标题文字 */
.section-title {
	font-size: var(--font-base);
	color: var(--text-primary);
}

/* 网关错误提示框 */
.gateway-err {
	padding: 0.625rem 0.875rem;
	background: var(--err-bg);
	border: 1px solid var(--err-border);
	border-radius: var(--h-radius);
	color: var(--err);
	font-size: var(--font-sm);
}

/* 目录选择行 */
.folder-row {
	display: flex;
	align-items: center;
	gap: 0.5rem;
}

/* 自适应输入框 */
.flex-input {
	flex: 1;
}

/* 端口输入框宽度 */
.port-input {
	width: 10rem;
}

/* 字段说明文字 */
.field-hint {
	font-size: var(--font-sm);
	color: var(--text-secondary);
}

/* 开关配置行 */
.toggle-row {
	display: flex;
	align-items: center;
	justify-content: space-between;
	padding-top: 0.5rem;
	border-top: 1px solid var(--border-color);
	gap: 1rem;
}

.config-row {
	display: flex;
	align-items: center;
	justify-content: space-between;
	padding-top: 0.5rem;
	border-top: 1px solid var(--border-color);
	gap: 1rem;
}

/* 开关左侧文字区 */
.toggle-info {
	display: flex;
	flex-direction: column;
	gap: 0.2rem;
}
</style>
