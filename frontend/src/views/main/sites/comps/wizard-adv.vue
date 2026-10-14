<template>
	<!-- 向导步骤4：高级配置与 SSL -->
	<div class="wizard-body">
		<!-- 默认入口文件与 HTTPS 开关 -->
		<div class="form-row">
			<div class="form-field flex-1">
				<label>默认入口文件</label>
				<p-input-text
					v-model="form.entry"
					class="font-mono"
					placeholder="index.php / index.html"
				/>
			</div>
			<div class="form-field flex-1">
				<label>HTTPS / SSL 证书</label>
				<div class="chk-line">
					<input id="ssl_enable_chk" v-model="form.tls" type="checkbox" />
					<label for="ssl_enable_chk">启用 HTTPS (443 端口)</label>
				</div>
			</div>
		</div>

		<!-- SSL 证书签发方式与文件路径 -->
		<div v-if="form.tls" class="ssl-box">
			<div class="form-field">
				<label>证书签发方式</label>
				<div class="radio-row">
					<label class="radio-opt">
						<input v-model="form.issuer" type="radio" value="self-signed" />
						<span>自签名证书（本地开发推荐）</span>
					</label>
					<label class="radio-opt">
						<input v-model="form.issuer" type="radio" value="letsencrypt" />
						<span>自动申请 Let's Encrypt</span>
					</label>
					<label class="radio-opt">
						<input v-model="form.issuer" type="radio" value="custom" />
						<span>自定义证书路径</span>
					</label>
				</div>
			</div>

			<div v-if="form.issuer === 'custom'" class="form-row">
				<div class="form-field flex-1">
					<label>证书文件 (.pem / .crt)</label>
					<div class="input-row">
						<p-input-text v-model="form.cert" class="font-mono flex-1" placeholder="D:\certs\site.pem" />
						<p-button
							label="选择"
							icon="pi pi-folder-open"
							severity="secondary"
							outlined
							@click="pickCreateCert"
						/>
					</div>
				</div>
				<div class="form-field flex-1">
					<label>私钥文件 (.key)</label>
					<div class="input-row">
						<p-input-text v-model="form.key" class="font-mono flex-1" placeholder="D:\certs\site.key" />
						<p-button
							label="选择"
							icon="pi pi-folder-open"
							severity="secondary"
							outlined
							@click="pickCreateKey"
						/>
					</div>
				</div>
			</div>
		</div>

		<!-- 启动策略勾选 -->
		<div class="form-field">
			<label>启动策略</label>
			<div class="chk-group">
				<label class="chk-line">
					<input v-model="form.autoStart" type="checkbox" />
					<span>随 DevEnv 启动时自动运行网站</span>
				</label>
				<label class="chk-line">
					<input v-model="form.startNow" type="checkbox" />
					<span>创建完成后立即启动网站</span>
				</label>
			</div>
		</div>

		<!-- 自定义 Nginx 配置片段 -->
		<div class="form-field">
			<label>自定义 Nginx 配置片段（可选）</label>
			<textarea
				v-model="form.custom"
				rows="4"
				class="code-area font-mono"
				placeholder="# 例如自定义重写或跨域头部:&#10;add_header X-Frame-Options SAMEORIGIN;"
			></textarea>
		</div>
	</div>
</template>

<script setup lang="ts">
	import { useWizard } from '../use-wizard'

	// 解构向导表单与证书文件选择方法
	const {
		form,
		pickCreateCert,
		pickCreateKey
	} = useWizard()
</script>

<style scoped lang="scss">
	/* 向导步骤内容容器 */
	.wizard-body {
		display: flex;
		flex-direction: column;
		gap: 14px;
	}

	/* 并排表单行 */
	.form-row {
		display: flex;
		gap: 12px;
	}

	/* 输入框与按钮同行 */
	.input-row {
		display: flex;
		align-items: center;
		gap: 8px;
	}

	/* 复选框组 */
	.chk-group,
	.radio-row {
		display: flex;
		gap: 16px;
		align-items: center;
	}

	/* 单行复选框或单选框 */
	.chk-line,
	.radio-opt {
		display: inline-flex;
		align-items: center;
		gap: 6px;
		font-size: var(--font-sm);
		color: var(--text-secondary);
		cursor: pointer;
	}

	/* SSL 配置附加框 */
	.ssl-box {
		background: var(--bg-app);
		border: 1px solid var(--border-color);
		border-radius: var(--radius-sm);
		padding: 12px;
		display: flex;
		flex-direction: column;
		gap: 10px;
	}

	/* 代码多行文本域 */
	.code-area {
		width: 100%;
		padding: 8px 10px;
		background: var(--bg-app);
		border: 1px solid var(--border-color);
		border-radius: var(--radius-sm);
		color: var(--text-primary);
		font-size: var(--font-xs);
		resize: vertical;
	}

	/* 自适应宽度 */
	.flex-1 {
		flex: 1;
	}
</style>
