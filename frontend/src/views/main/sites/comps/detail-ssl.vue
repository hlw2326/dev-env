<template>
	<!-- 详情页签4：SSL 证书管理与一键签发工具 -->
	<div v-if="curSite" class="detail-pane">
		<div class="ssl-grid">
			<!-- 左侧证书状态与文件路径卡片 -->
			<div class="info-card">
				<div class="sec-title">当前 SSL / HTTPS 状态</div>
				<div class="chk-line">
					<input id="detail_ssl_chk" v-model="editForm.tls" type="checkbox" />
					<label for="detail_ssl_chk">启用 HTTPS 加密访问 (监听 443 端口)</label>
				</div>

				<div class="form-field">
					<label>证书类型</label>
					<p-select
						v-model="editForm.issuer"
						:options="issuerOpts"
						optionLabel="label"
						optionValue="value"
					/>
				</div>

				<div class="form-field">
					<label>证书文件路径 (.crt / .pem)</label>
					<div class="input-row">
						<p-input-text
							v-model="editForm.cert"
							class="font-mono flex-1"
							:placeholder="`D:\\DevEnv\\certs\\${curSite.domains[0]}.pem`"
						/>
						<p-button
							label="选择文件"
							icon="pi pi-folder-open"
							severity="secondary"
							outlined
							@click="pickEditCert"
						/>
					</div>
				</div>

				<div class="form-field">
					<label>私钥文件路径 (.key)</label>
					<div class="input-row">
						<p-input-text
							v-model="editForm.key"
							class="font-mono flex-1"
							:placeholder="`D:\\DevEnv\\certs\\${curSite.domains[0]}.key`"
						/>
						<p-button
							label="选择文件"
							icon="pi pi-folder-open"
							severity="secondary"
							outlined
							@click="pickEditKey"
						/>
					</div>
				</div>
			</div>

			<!-- 右侧快捷证书生成工具卡片 -->
			<div class="info-card">
				<div class="sec-title">快捷证书工具</div>
				<p class="desc-text">
					本地开发环境可一键生成自签名证书；公网已解析域名可申请 Let's Encrypt 90 天证书并开启到期前 30 天自动续期。
				</p>
				<div class="quick-btns">
					<p-button
						label="一键生成本地自签名证书"
						icon="pi pi-shield"
						severity="secondary"
						outlined
						size="small"
						@click="genSelfCert"
					/>
					<p-button
						label="申请 Let's Encrypt 证书"
						icon="pi pi-globe"
						severity="secondary"
						outlined
						size="small"
						@click="reqLetsCert"
					/>
				</div>

				<div class="cert-preview font-mono">
					<div>绑定域名: {{ curSite.domains.join(', ') }}</div>
					<div>协议版本: TLSv1.2 / TLSv1.3 (HTTP/2)</div>
					<div>有效期至: 2027-10-14 (剩余 365 天 · 自动续期已开启)</div>
				</div>
			</div>
		</div>

		<div class="pane-footer">
			<p-button
				label="保存 SSL 配置"
				icon="pi pi-save"
				:loading="saving"
				@click="saveInfo"
			/>
		</div>
	</div>
</template>

<script setup lang="ts">
	import { useDetail } from '../use-detail'

	// 解构 SSL 编辑表单与文件选择方法
	const {
		curSite,
		saving,
		editForm,
		pickEditCert,
		pickEditKey,
		genSelfCert,
		reqLetsCert,
		saveInfo
	} = useDetail()

	// 证书签发类型选项
	const issuerOpts = [
		{ label: '自签名开发证书 (Self-Signed)', value: 'self-signed' },
		{ label: 'Let\'s Encrypt 免费自动证书', value: 'letsencrypt' },
		{ label: '自定义上传证书 (.pem / .key)', value: 'custom' }
	]
</script>

<style scoped lang="scss">
	/* 页签主体容器 */
	.detail-pane {
		display: flex;
		flex-direction: column;
		gap: 14px;
	}

	/* 双列卡片网格 */
	.ssl-grid {
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: 14px;
	}

	/* 分块信息卡片 */
	.info-card {
		background: var(--bg-app);
		border: 1px solid var(--border-color);
		border-radius: var(--radius-md);
		padding: 14px;
		display: flex;
		flex-direction: column;
		gap: 12px;
	}

	/* 小节标题 */
	.sec-title {
		font-size: var(--font-sm);
		font-weight: 500;
		color: var(--text-primary);
	}

	/* 复选框行 */
	.chk-line {
		display: inline-flex;
		align-items: center;
		gap: 8px;
		font-size: var(--font-sm);
		color: var(--text-secondary);
		cursor: pointer;
	}

	/* 输入与按钮同行 */
	.input-row {
		display: flex;
		align-items: center;
		gap: 8px;
	}

	/* 辅助说明文字 */
	.desc-text {
		font-size: var(--font-xs);
		color: var(--text-muted);
		margin: 0;
		line-height: 1.5;
	}

	/* 快捷证书按钮组 */
	.quick-btns {
		display: flex;
		gap: 8px;
		flex-wrap: wrap;
	}

	/* 证书信息预览框 */
	.cert-preview {
		background: var(--bg-card);
		border: 1px solid var(--border-color);
		border-radius: var(--radius-sm);
		padding: 10px 12px;
		font-size: var(--font-xs);
		color: var(--text-secondary);
		display: flex;
		flex-direction: column;
		gap: 4px;
	}

	/* 底部操作行 */
	.pane-footer {
		display: flex;
		justify-content: flex-end;
	}

	/* 自适应填充宽度 */
	.flex-1 {
		flex: 1;
	}
</style>
