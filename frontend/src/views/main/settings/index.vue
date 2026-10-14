<template>
	<h-page title="全局设置" desc="配置默认站点目录、本地网关端口、hosts 自动写入与界面主题">
		<template #extra>
			<p-button icon="pi pi-check" label="保存设置" :loading="saving" @click="save" />
		</template>

		<h-card title="网关与站点">
			<div class="settings-block">
				<div class="gateway-status">
					<span class="sec-title">本地网关状态</span>
					<div class="status-right">
						<span v-if="listening" class="status-badge ok">
							<span class="dot"></span>
							<span>正在监听 127.0.0.1:{{ form.port }}</span>
						</span>
						<span v-else class="status-badge err">
							<span class="dot"></span>
							<span>未监听</span>
						</span>
					</div>
				</div>

				<div v-if="gatewayError !== ''" class="gateway-err">
					{{ gatewayError }}
				</div>

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
					<span class="field-desc">创建新网站时默认使用该目录作为父目录</span>
				</div>

				<div class="form-field">
					<label>本地网关端口</label>
					<p-input-number v-model="form.port" :use-grouping="false" :min="1" :max="65535" class="port-input" />
					<span class="field-desc">所有站点共用此网关端口（默认 80），端口被占用时请修改此项并保存</span>
				</div>

				<div class="toggle-row">
					<div class="toggle-info">
						<span class="sec-title">自动写入系统 hosts</span>
						<span class="field-desc">
							*.localhost 域名无需写入 hosts；仅自定义后缀（如 .test）需要开启此项并以管理员权限运行
						</span>
					</div>
					<p-toggle-switch v-model="form.hosts" />
				</div>

				<div class="toggle-row">
					<div class="toggle-info">
						<span class="sec-title">开机自动启动 DevEnv</span>
						<span class="field-desc">
							电脑开机登录后自动在后台启动，并自动拉起配置了“随应用启动”的软件服务和网站
						</span>
					</div>
					<p-toggle-switch v-model="form.autostart" />
				</div>
			</div>
		</h-card>

		<h-card title="外观与存储">
			<div class="settings-block">
				<div class="form-field">
					<label>色彩模式（点击即生效）</label>
					<div class="theme-options">
						<button
							type="button"
							class="theme-btn"
							:class="{ active: theme.mode === 'light' }"
							@click="selectTheme('light')"
						>
							<i class="pi pi-sun theme-icon"></i>
							<span>浅色模式</span>
						</button>
						<button
							type="button"
							class="theme-btn"
							:class="{ active: theme.mode === 'dark' }"
							@click="selectTheme('dark')"
						>
							<i class="pi pi-moon theme-icon"></i>
							<span>深色模式</span>
						</button>
						<button
							type="button"
							class="theme-btn"
							:class="{ active: theme.mode === 'system' }"
							@click="selectTheme('system')"
						>
							<i class="pi pi-desktop theme-icon"></i>
							<span>跟随系统</span>
						</button>
					</div>
				</div>

				<div class="form-field">
					<div class="color-label-row">
						<label>主题强调色</label>
						<span class="color-tag" :style="{ backgroundColor: theme.color }">{{ currentColorName }}</span>
					</div>
					<div class="color-options">
						<button
							v-for="item in THEME_COLORS"
							:key="item.value"
							type="button"
							class="color-wrap"
							:class="{ active: theme.color.toLowerCase() === item.value.toLowerCase() }"
							:title="item.name"
							@click="selectColor(item.value)"
						>
							<span class="color-dot" :style="{ backgroundColor: item.value }">
								<i
									v-if="theme.color.toLowerCase() === item.value.toLowerCase()"
									class="pi pi-check check-icon"
								></i>
							</span>
						</button>
						<label class="color-picker-btn" title="自定义强调色">
							<input
								type="color"
								:value="theme.color"
								class="color-input"
								@input="(e) => selectColor((e.target as HTMLInputElement).value)"
							/>
							<i class="pi pi-plus plus-icon"></i>
						</label>
					</div>
				</div>

				<div class="form-field">
					<label>应用数据目录（只读）</label>
					<div class="folder-row">
						<p-input-text :model-value="dataFolder" readonly class="font-mono flex-input" />
						<p-button
							severity="secondary"
							variant="outlined"
							icon="pi pi-folder-open"
							label="打开"
							@click="openData"
						/>
					</div>
					<span class="field-desc">存放数据库、已安装软件、MySQL 数据目录与运行日志</span>
				</div>
			</div>
		</h-card>
	</h-page>
</template>

<script setup lang="ts">
import { useSettings } from './use-settings'

const {
	form,
	theme,
	THEME_COLORS,
	currentColorName,
	listening,
	gatewayError,
	dataFolder,
	saving,
	selectTheme,
	selectColor,
	pickRoot,
	save,
	openData,
} = useSettings()
</script>

<style scoped lang="scss">
.settings-block {
	display: flex;
	flex-direction: column;
	gap: 0.75rem;
}

.gateway-status {
	display: flex;
	align-items: center;
	justify-content: space-between;
	padding: 0.75rem 0.875rem;
	background: var(--bg-app);
	border: 1px solid var(--border-color);
	border-radius: var(--radius-md);
}

.sec-title {
	font-size: var(--font-base);
	color: var(--text-primary);
}

.status-badge {
	display: inline-flex;
	align-items: center;
	gap: 0.35rem;
	font-size: var(--font-xs);
	padding: 0.125rem 0.55rem;
	border-radius: var(--radius-full);

	.dot {
		width: 6px;
		height: 6px;
		border-radius: 50%;
	}

	&.ok {
		color: var(--ok);
		background: var(--ok-bg);
		border: 1px solid var(--ok-border);

		.dot {
			background: var(--ok);
		}
	}

	&.err {
		color: var(--err);
		background: var(--err-bg);
		border: 1px solid var(--err-border);

		.dot {
			background: var(--err);
		}
	}
}

.gateway-err {
	padding: 0.625rem 0.875rem;
	background: var(--err-bg);
	border: 1px solid var(--err-border);
	border-radius: var(--radius-md);
	color: var(--err);
	font-size: var(--font-sm);
}

.folder-row {
	display: flex;
	align-items: center;
	gap: 0.5rem;
}

.flex-input {
	flex: 1;
}

.port-input {
	width: 10rem;
}

.field-desc {
	font-size: var(--font-sm);
	color: var(--text-secondary);
}

.toggle-row {
	display: flex;
	align-items: center;
	justify-content: space-between;
	padding-top: 0.5rem;
	border-top: 1px solid var(--border-color);
	gap: 1rem;
}

.toggle-info {
	display: flex;
	flex-direction: column;
	gap: 0.2rem;
}

.theme-options {
	display: flex;
	gap: 0.625rem;
}

.theme-btn {
	display: inline-flex;
	align-items: center;
	gap: 0.5rem;
	padding: 0.5rem 1rem;
	border: 1px solid var(--border-color);
	border-radius: var(--radius-md);
	background: var(--bg-app);
	color: var(--text-secondary);
	font-size: var(--font-sm);
	cursor: pointer;

	&:hover {
		border-color: var(--border-hover);
		color: var(--text-primary);
	}

	&.active {
		background: var(--p-primary-color);
		color: var(--on-primary);
		border-color: var(--p-primary-color);
	}

	.theme-icon {
		font-size: var(--icon-sm);
	}
}

.color-label-row {
	display: flex;
	align-items: center;
	gap: 0.5rem;
}

.color-tag {
	font-size: var(--font-xs);
	color: var(--on-primary);
	padding: 0.1rem 0.45rem;
	border-radius: var(--radius-sm);
}

.color-options {
	display: flex;
	align-items: center;
	gap: 0.625rem;
	flex-wrap: wrap;
}

.color-wrap {
	display: flex;
	align-items: center;
	justify-content: center;
	width: 2rem;
	height: 2rem;
	border-radius: 50%;
	cursor: pointer;
	border: 2px solid transparent;
	background: transparent;

	&.active {
		border-color: var(--text-primary);
	}

	.color-dot {
		width: 1.375rem;
		height: 1.375rem;
		border-radius: 50%;
		display: flex;
		align-items: center;
		justify-content: center;
	}

	.check-icon {
		font-size: var(--icon-xs);
		color: var(--on-primary);
	}
}

.color-picker-btn {
	width: 1.875rem;
	height: 1.875rem;
	border-radius: 50%;
	border: 1px dashed var(--border-color);
	display: flex;
	align-items: center;
	justify-content: center;
	cursor: pointer;
	position: relative;

	.color-input {
		position: absolute;
		opacity: 0;
		width: 100%;
		height: 100%;
		cursor: pointer;
	}

	.plus-icon {
		font-size: var(--icon-xs);
		color: var(--text-secondary);
	}
}
</style>
