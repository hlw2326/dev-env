<template>
	<!-- 软件版本选择与安装弹窗 -->
	<h-modal v-model:visible="visible" :title="modalTitle" width="640px">
		<div v-if="pkg" class="detail-box">
			<!-- 软件头部简介 -->
			<div class="pkg-head">
				<div class="pkg-icon" :style="{ backgroundColor: pkg.brandColor }">
					<i :class="pkg.iconClass"></i>
				</div>
				<div class="pkg-text">
					<div class="pkg-title">{{ pkg.displayName }}</div>
					<div class="pkg-desc">{{ pkg.description }}</div>
					<a v-if="pkg.officialWebsite" :href="pkg.officialWebsite" target="_blank" class="official-link">
						<i class="pi pi-external-link"></i> 官方网站 ({{ pkg.officialWebsite }})
					</a>
				</div>
			</div>

			<!-- 在线版本标题与刷新 -->
			<div class="ver-head">
				<div class="sec-title">
					<span>在线版本列表</span>
					<span v-if="versions.length > 0" class="ver-count font-mono">实时可用 {{ versions.length }} 个版本</span>
				</div>
				<p-button icon="pi pi-refresh" label="在线刷新版本" size="small" severity="secondary" text :loading="loadingVer" @click="emit('refresh')" />
			</div>

			<!-- 版本选择列表区 -->
			<div class="ver-box">
				<div v-if="loadingVer && versions.length === 0" class="ver-loading">
					<i class="pi pi-spin pi-spinner"></i>
					<span>正在从官方源与镜像动态检测最新发布版本...</span>
				</div>

				<div v-if="versions.length > 0" class="ver-list">
					<div
						v-for="ver in versions"
						:key="ver.version"
						class="ver-item"
						:class="{ selected: curVer?.version === ver.version }"
						@click="curVer = ver"
					>
						<div class="ver-left">
							<input type="radio" name="pkg-version" :checked="curVer?.version === ver.version" />
							<div class="ver-info">
								<div class="ver-name font-mono">
									v{{ ver.version }}
									<span v-if="ver.variant" class="ver-tag">{{ ver.variant.toUpperCase() }}</span>
									<span class="ver-tag">{{ ver.architecture }}</span>
								</div>
								<div class="ver-sub font-mono">
									大小: {{ formatSize(ver.fileSize) }}
									<span v-if="ver.releaseDate"> · 发布: {{ ver.releaseDate }}</span>
								</div>
							</div>
						</div>
						<div class="ver-tags">
							<p-tag v-if="ver.isRecommended" severity="success" value="推荐" />
							<p-tag v-if="ver.isLTS" severity="info" value="LTS" />
							<p-tag v-if="checkInstalled(ver.version)" severity="secondary" value="本地已存在" />
						</div>
					</div>
				</div>
			</div>

			<!-- 下载镜像源选择 -->
			<div class="sec-title">下载镜像源</div>
			<div class="mirror-row">
				<label
					v-for="mir in mirrors"
					:key="mir.key"
					class="mirror-pill"
					:class="{ active: curMirror === mir.key }"
				>
					<input v-model="curMirror" type="radio" name="mirror-source" :value="mir.key" class="hidden-radio" />
					<span>{{ mir.label }}</span>
				</label>
			</div>

			<!-- 安装目标路径选择 -->
			<div class="sec-title">安装目标路径（支持自定义目录）</div>
			<div class="path-row">
				<p-input-text v-model="customDir" class="font-mono flex-1" placeholder="如 D:\DevEnv\runtimes\php\8.3.12" />
				<p-button label="选择目录" icon="pi pi-folder-open" severity="secondary" outlined @click="emit('pick-dir')" />
			</div>

			<!-- 安装进度条 -->
			<div v-if="installing" class="progress-box">
				<div class="progress-labels">
					<span>正在下载并部署 {{ pkg.displayName }} v{{ curVer?.version }}...</span>
					<span class="font-mono">{{ progress }}%</span>
				</div>
				<p-progress-bar :value="progress" :showValue="false" style="height: 6px;" />
			</div>
		</div>

		<!-- 底部操作按钮 -->
		<template #footer>
			<p-button label="取消" severity="secondary" outlined :disabled="installing" @click="visible = false" />
			<p-button
				v-if="!isSelectedInstalled"
				label="开始安装"
				icon="pi pi-check"
				:loading="installing"
				:disabled="loadingVer || !curVer"
				@click="emit('install')"
			/>
			<p-button v-else label="前往运行环境" icon="pi pi-arrow-right" severity="info" @click="goRuntimes" />
		</template>
	</h-modal>
</template>

<script setup lang="ts">
	import { computed } from 'vue'
	import { useRouter } from 'vue-router'
	import type { PackageInfo, PackageVersion } from '@/types'
	import { mirrors } from '../use-store'

	// 双向绑定状态
	const visible = defineModel<boolean>('visible', { default: false })
	const curVer = defineModel<PackageVersion | null>('curVer', { default: null })
	const curMirror = defineModel<string>('curMirror', { default: 'aliyun' })
	const customDir = defineModel<string>('customDir', { default: '' })

	// 接收软件包与版本数据
	const props = defineProps<{
		pkg: PackageInfo | null
		versions: PackageVersion[]
		loadingVer: boolean
		installing: boolean
		progress: number
		checkInstalled: (ver: string) => boolean
	}>()

	// 定义操作事件
	const emit = defineEmits<{ refresh: []; 'pick-dir': []; install: [] }>()

	// 路由实例
	const router = useRouter()

	// 弹窗标题
	const modalTitle = computed(() => {
		if (props.pkg) {
			return `${props.pkg.displayName} - 版本选择与安装`
		}
		return '软件安装'
	})

	// 当前选中版本是否已安装
	const isSelectedInstalled = computed(() => {
		if (!curVer.value) {
			return false
		}
		return props.checkInstalled(curVer.value.version)
	})

	// 格式化文件体积
	function formatSize(bytes: number): string {
		if (!bytes) {
			return '未知'
		}
		return `${(bytes / (1024 * 1024)).toFixed(1)} MB`
	}

	// 跳转到运行环境页面
	function goRuntimes() {
		visible.value = false
		router.push('/runtimes')
	}
</script>

<style scoped lang="scss">
	/* 详情弹窗内容容器 */
	.detail-box { display: flex; flex-direction: column; gap: 14px; }

	/* 软件顶部卡片 */
	.pkg-head {
		display: flex;
		align-items: flex-start;
		gap: 14px;
		background: var(--bg-app);
		padding: 12px;
		border-radius: var(--h-radius);
		border: 1px solid var(--border-color);
	}

	/* 软件图标与文字 */
	.pkg-icon {
		width: 48px; height: 48px; border-radius: var(--radius-sm);
		display: flex; align-items: center; justify-content: center;
		color: var(--on-primary); font-size: 1.4rem; flex-shrink: 0;
	}
	.pkg-text { display: flex; flex-direction: column; gap: 4px; }
	.pkg-title { font-size: var(--font-lg); color: var(--text-primary); }
	.pkg-desc { font-size: var(--font-sm); color: var(--text-secondary); }
	.official-link { font-size: var(--font-xs); color: var(--p-primary-color); text-decoration: none; display: flex; align-items: center; gap: 4px; }

	/* 版本列表头部与徽章 */
	.ver-head { display: flex; justify-content: space-between; align-items: center; }
	.sec-title { display: flex; align-items: center; font-size: var(--font-sm); color: var(--text-muted); }
	.ver-count { font-size: var(--font-xs); color: var(--info); background: var(--info-bg); padding: 2px 6px; border-radius: var(--radius-sm); margin-left: 8px; }

	/* 版本列表容器 */
	.ver-box { min-height: 180px; }
	.ver-loading {
		display: flex; align-items: center; justify-content: center; gap: 8px;
		height: 180px; color: var(--text-muted); font-size: var(--font-sm);
		background: var(--bg-app); border: 1px dashed var(--border-color); border-radius: var(--radius-sm);
	}
	.ver-list { display: flex; flex-direction: column; gap: 8px; max-height: 220px; overflow-y: auto; }
	.ver-item {
		display: flex; justify-content: space-between; align-items: center;
		padding: 10px 12px; border: 1px solid var(--border-color); border-radius: var(--radius-sm); cursor: pointer;
		&.selected { border-color: var(--p-primary-color); background: var(--bg-app); }
	}
	.ver-left { display: flex; align-items: center; gap: 10px; }
	.ver-name { font-size: var(--font-base); color: var(--text-primary); display: flex; align-items: center; gap: 6px; }
	.ver-tag { font-size: var(--font-xs); padding: 1px 5px; background: var(--bg-card); border: 1px solid var(--border-color); border-radius: var(--radius-sm); color: var(--text-secondary); }
	.ver-sub { font-size: var(--font-xs); color: var(--text-muted); }
	.ver-tags { display: flex; gap: 6px; }

	/* 镜像源选择行 */
	.mirror-row { display: flex; gap: 8px; }
	.mirror-pill {
		flex: 1; text-align: center; padding: 8px; border: 1px solid var(--border-color);
		border-radius: var(--radius-sm); font-size: var(--font-sm); color: var(--text-secondary); cursor: pointer;
		&.active { border-color: var(--p-primary-color); color: var(--p-primary-color); background: var(--bg-app); }
	}
	.hidden-radio { display: none; }

	/* 路径与进度 */
	.path-row { display: flex; gap: 8px; align-items: center; }
	.flex-1 { flex: 1; }
	.progress-box { display: flex; flex-direction: column; gap: 6px; }
	.progress-labels { display: flex; justify-content: space-between; font-size: var(--font-xs); color: var(--text-secondary); }
</style>
