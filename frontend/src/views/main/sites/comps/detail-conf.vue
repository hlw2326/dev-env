<template>
	<!-- 详情页签3：Nginx 配置编辑、可视化规则与历史快照 -->
	<div v-if="curSite" class="detail-pane">
		<!-- 顶部子模式切换与模板插入栏 -->
		<div class="conf-toolbar">
			<div class="mode-switch">
				<button type="button" class="mode-btn" :class="{ active: nginxMode === 'editor' }" @click="nginxMode = 'editor'">
					在线配置编辑
				</button>
				<button type="button" class="mode-btn" :class="{ active: nginxMode === 'visual' }" @click="nginxMode = 'visual'">
					表单可视化配置
				</button>
				<button type="button" class="mode-btn" :class="{ active: nginxMode === 'history' }" @click="nginxMode = 'history'">
					配置版本历史 ({{ historyList.length }})
				</button>
			</div>

			<p-select
				v-model="selectedTpl"
				:options="nginxTpls"
				optionLabel="label"
				optionValue="key"
				placeholder="快速插入配置模板..."
				@change="insertTpl"
			/>
		</div>

		<!-- 子模式1：在线代码编辑器 -->
		<div v-if="nginxMode === 'editor'" class="editor-box">
			<div class="path-line font-mono">配置文件路径: {{ curSite.conf }}</div>
			<textarea v-model="confContent" rows="12" class="code-editor font-mono"></textarea>

			<div v-if="confCheck" class="test-box font-mono" :class="checkClass">
				<div>{{ checkTitle }}</div>
				<pre>{{ confCheck.message }}</pre>
			</div>

			<div class="pane-footer">
				<p-button label="恢复默认生成配置" icon="pi pi-replay" severity="secondary" outlined size="small" @click="resetConf($event)" />
				<p-button label="测试配置 (nginx -t)" icon="pi pi-check-square" severity="secondary" outlined size="small" @click="testConf" />
				<p-button label="保存并重载 Nginx" icon="pi pi-save" size="small" :loading="saving" @click="saveConf" />
			</div>
		</div>

		<!-- 子模式2：表单可视化规则 -->
		<div v-if="nginxMode === 'visual'" class="visual-grid">
			<div class="info-card">
				<div class="sec-title">性能优化策略</div>
				<label class="chk-line">
					<input v-model="visualForm.gzip" type="checkbox" />
					<span>启用 Gzip 压缩 (gzip on; 压缩级别 6)</span>
				</label>
				<label class="chk-line">
					<input v-model="visualForm.staticCache" type="checkbox" />
					<span>启用静态资源缓存 (图片/CSS/JS 缓存 30 天)</span>
				</label>
				<label class="chk-line">
					<input v-model="visualForm.cors" type="checkbox" />
					<span>启用跨域 CORS 响应头</span>
				</label>
				<label class="chk-line">
					<input v-model="visualForm.securityHeaders" type="checkbox" />
					<span>启用安全防护头 (X-Frame-Options / 隐藏版本号)</span>
				</label>
			</div>

			<div class="info-card">
				<div class="sec-title">安全与 IP 访问控制</div>
				<div class="form-field">
					<label>IP 访问策略</label>
					<p-select v-model="visualForm.ipPolicy" :options="ipOpts" optionLabel="label" optionValue="value" />
				</div>
				<div v-if="visualForm.ipPolicy !== 'allow_all'" class="form-field">
					<label>IP 列表（每行一个 IP 或网段）</label>
					<textarea v-model="visualForm.ipList" rows="3" class="code-editor font-mono" placeholder="192.168.1.100&#10;10.0.0.0/24"></textarea>
				</div>
				<p-button label="应用可视化规则到 Nginx 配置" icon="pi pi-check" size="small" @click="applyVisual" />
			</div>
		</div>

		<!-- 子模式3：历史版本列表 -->
		<div v-if="nginxMode === 'history'" class="history-list">
			<div v-if="historyList.length === 0" class="empty-small">暂无历史版本记录</div>
			<div v-for="(snap, idx) in historyList" :key="idx" class="history-item">
				<div class="hist-left">
					<span class="font-mono hist-time">{{ snap.timestamp }}</span>
					<span class="hist-sum">{{ snap.summary }}</span>
				</div>
				<p-button label="查看并恢复此版本" size="small" severity="secondary" outlined @click="restoreSnap($event, snap)" />
			</div>
		</div>
	</div>
</template>

<script setup lang="ts">
	import { computed } from 'vue'
	import type { SiteSnapshot } from '@/types'
	import { nginxTpls, useDetail } from '../use-detail'

	// 解构 Nginx 配置状态与方法
	const {
		curSite, saving, nginxMode, confContent, confCheck, selectedTpl,
		visualForm, insertTpl, applyVisual, resetConf, testConf, saveConf, restoreSnap
	} = useDetail()

	// IP 访问策略下拉选项
	const ipOpts = [
		{ label: '允许所有 IP 访问（默认）', value: 'allow_all' },
		{ label: '白名单模式（仅允许下方 IP）', value: 'whitelist' },
		{ label: '黑名单模式（禁止下方 IP）', value: 'blacklist' }
	]

	// 历史配置快照列表
	const historyList = computed<SiteSnapshot[]>(() => {
		if (curSite.value) {
			return curSite.value.meta.history
		}
		return []
	})

	// 语法检测样式类名
	const checkClass = computed(() => {
		if (confCheck.value && confCheck.value.valid) {
			return 'ok'
		}
		return 'err'
	})

	// 语法检测标题文案
	const checkTitle = computed(() => {
		if (confCheck.value && confCheck.value.valid) {
			return '配置文件语法正确'
		}
		return '配置文件存在语法错误'
	})
</script>

<style scoped lang="scss">
	/* 页签主体容器 */
	.detail-pane, .editor-box { display: flex; flex-direction: column; gap: 12px; }
	.conf-toolbar { display: flex; justify-content: space-between; align-items: center; }
	.mode-switch { display: flex; gap: 6px; }
	.mode-btn {
		padding: 6px 12px; border: 1px solid var(--border-color); border-radius: var(--radius-sm);
		background: var(--bg-app); color: var(--text-secondary); font-size: var(--font-xs); cursor: pointer;
		&.active { background: var(--p-primary-color); color: var(--on-primary); border-color: var(--p-primary-color); }
	}
	.path-line { font-size: var(--font-xs); color: var(--text-muted); }
	.code-editor {
		width: 100%; padding: 10px 12px; background: var(--code-bg); color: var(--code-text);
		border: 1px solid var(--border-color); border-radius: var(--radius-sm); font-size: var(--font-xs); resize: vertical;
	}
	.test-box {
		padding: 10px 12px; border-radius: var(--radius-sm); font-size: var(--font-xs);
		&.ok { background: var(--ok-bg); border: 1px solid var(--ok-border); color: var(--ok); }
		&.err { background: var(--err-bg); border: 1px solid var(--err-border); color: var(--err); }
		pre { margin: 4px 0 0; white-space: pre-wrap; }
	}
	.visual-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 14px; }
	.info-card {
		background: var(--bg-app); border: 1px solid var(--border-color);
		border-radius: var(--radius-md); padding: 14px; display: flex; flex-direction: column; gap: 10px;
	}
	.sec-title { font-size: var(--font-sm); font-weight: 500; color: var(--text-primary); }
	.chk-line { display: flex; align-items: center; gap: 8px; font-size: var(--font-xs); color: var(--text-secondary); cursor: pointer; }
	.history-list { display: flex; flex-direction: column; gap: 8px; }
	.history-item {
		display: flex; justify-content: space-between; align-items: center;
		padding: 10px 12px; background: var(--bg-app); border: 1px solid var(--border-color);
		border-radius: var(--radius-sm); font-size: var(--font-xs);
	}
	.hist-left { display: flex; gap: 12px; }
	.pane-footer { display: flex; justify-content: flex-end; gap: 8px; }
	.empty-small { padding: 30px 0; text-align: center; color: var(--text-muted); font-size: var(--font-xs); }
</style>
