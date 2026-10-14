<template>
  <h-page title="软件商店" desc="发现并安装 PHP、Nginx、MySQL 等主流本地开发环境">
    <template #extra>
      <div class="search-box">
        <i class="pi pi-search search-icon"></i>
        <p-input-text 
          v-model="searchQuery" 
          placeholder="搜索软件包名称或描述..." 
          class="search-input"
        />
      </div>
    </template>

    <!-- 分类导航 -->
    <div class="category-tabs">
      <button 
        v-for="cat in categories" 
        :key="cat.key"
        type="button"
        class="category-btn"
        :class="{ active: activeCategory === cat.key }"
        @click="activeCategory = cat.key"
      >
        {{ cat.label }}
      </button>
    </div>

    <!-- 软件包卡片网格 -->
    <div v-if="filteredPackages.length === 0" class="empty-state">
      <i class="pi pi-inbox empty-icon"></i>
      <p class="empty-text">未找到与 "{{ searchQuery }}" 匹配的软件包</p>
    </div>

    <div v-else class="packages-grid">
      <div 
        v-for="pkg in filteredPackages" 
        :key="pkg.key" 
        class="package-card"
        @click="openPackageDetail(pkg)"
      >
        <div class="pkg-top">
          <div class="pkg-icon-wrap" :style="{ backgroundColor: pkg.brandColor }">
            <i :class="pkg.iconClass"></i>
          </div>
          <div class="pkg-header-info">
            <div class="pkg-name-row">
              <span class="pkg-name">{{ pkg.displayName }}</span>
              <p-tag 
                v-if="pkg.isInstalled" 
                severity="success" 
                value="已安装" 
                icon="pi pi-check"
                class="installed-tag"
              />
            </div>
            <span class="pkg-ver font-mono">推荐: v{{ pkg.recommendedVersion }}</span>
          </div>
        </div>

        <p class="pkg-desc">{{ pkg.description }}</p>

        <div class="pkg-footer">
          <div class="pkg-meta">
            <span class="category-badge">{{ getCategoryLabel(pkg.category) }}</span>
          </div>
          <div class="pkg-actions" @click.stop>
            <p-button 
              v-if="pkg.isInstalled"
              label="版本管理" 
              icon="pi pi-list" 
              size="small" 
              severity="secondary" 
              outlined
              @click="openPackageDetail(pkg)" 
            />
            <p-button 
              v-else 
              label="安装" 
              icon="pi pi-download" 
              size="small" 
              @click="openPackageDetail(pkg)" 
            />
          </div>
        </div>
      </div>
    </div>

    <!-- 软件包详情与安装对话框 -->
    <p-dialog 
      v-model:visible="detailDialogVisible" 
      modal 
      :header="selectedPkg ? `${selectedPkg.displayName} - 版本选择与安装` : '软件安装'"
      :style="{ width: '600px' }"
    >
      <div v-if="selectedPkg" class="detail-container">
        <!-- 基础简介 -->
        <div class="dialog-pkg-info">
          <div class="dialog-pkg-icon" :style="{ backgroundColor: selectedPkg.brandColor }">
            <i :class="selectedPkg.iconClass"></i>
          </div>
          <div class="dialog-pkg-text">
            <div class="dialog-pkg-title">{{ selectedPkg.displayName }}</div>
            <div class="dialog-pkg-desc">{{ selectedPkg.description }}</div>
            <a 
              v-if="selectedPkg.officialWebsite" 
              :href="selectedPkg.officialWebsite" 
              target="_blank" 
              class="official-link"
            >
              <i class="pi pi-external-link"></i> 官方网站 ({{ selectedPkg.officialWebsite }})
            </a>
          </div>
        </div>

        <!-- 版本选择 -->
        <div class="section-title">可用版本</div>
        <div class="version-list">
          <div 
            v-for="ver in availableVersions" 
            :key="ver.version"
            class="version-item"
            :class="{ selected: selectedVersion?.version === ver.version }"
            @click="selectedVersion = ver"
          >
            <div class="version-left">
              <input 
                type="radio" 
                name="pkg-version" 
                :checked="selectedVersion?.version === ver.version"
                class="version-radio"
              />
              <div class="version-info">
                <div class="version-name font-mono">
                  v{{ ver.version }} 
                  <span v-if="ver.variant" class="variant-tag">{{ ver.variant.toUpperCase() }}</span>
                  <span class="arch-tag">{{ ver.architecture }}</span>
                </div>
                <div class="version-sub font-mono">
                  大小: {{ formatFileSize(ver.fileSize) }}
                  <span v-if="ver.releaseDate"> · 发布于: {{ ver.releaseDate }}</span>
                </div>
              </div>
            </div>
            <div class="version-tags">
              <p-tag v-if="ver.isRecommended" severity="success" value="推荐" />
              <p-tag v-if="ver.isLTS" severity="info" value="LTS" />
              <p-tag 
                v-if="isVersionInstalled(ver.version)" 
                severity="secondary" 
                value="本地已存在" 
              />
            </div>
          </div>
        </div>

        <!-- 镜像源选择 -->
        <div class="section-title">下载镜像源</div>
        <div class="mirror-row">
          <label 
            v-for="mir in mirrors" 
            :key="mir.key" 
            class="mirror-pill"
            :class="{ active: selectedMirror === mir.key }"
          >
            <input 
              type="radio" 
              name="mirror-source" 
              :value="mir.key" 
              v-model="selectedMirror" 
              class="hidden-radio"
            />
            <span>{{ mir.label }}</span>
          </label>
        </div>

        <!-- 安装路径预览 -->
        <div class="section-title">安装目标路径</div>
        <div class="path-preview font-mono">
          <i class="pi pi-folder"></i>
          <span>{{ targetInstallPath }}</span>
        </div>

        <!-- 安装中进度条 -->
        <div v-if="installing" class="install-progress">
          <div class="progress-labels">
            <span>正在下载并部署 {{ selectedPkg.displayName }} v{{ selectedVersion?.version }}...</span>
            <span class="font-mono">{{ installProgress }}%</span>
          </div>
          <p-progress-bar :value="installProgress" :showValue="false" style="height: 6px;" />
        </div>
      </div>

      <template #footer>
        <p-button 
          label="取消" 
          severity="secondary" 
          outlined 
          :disabled="installing"
          @click="detailDialogVisible = false" 
        />
        <p-button 
          v-if="!isVersionInstalled(selectedVersion?.version || '')"
          label="开始安装" 
          icon="pi pi-check" 
          :loading="installing" 
          @click="handleStartInstall" 
        />
        <p-button 
          v-else
          label="前往运行环境" 
          icon="pi pi-arrow-right" 
          severity="info"
          @click="goToRuntimes" 
        />
      </template>
    </p-dialog>
  </h-page>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import HPage from '@/components/h-page/h-page.vue'
import { listPackages, getPackageVersions, installPackage } from '@/api/store'
import type { PackageInfo, PackageVersion } from '@/types'

defineOptions({
  name: 'StoreView'
})

const router = useRouter()
const searchQuery = ref('')
const activeCategory = ref('all')
const packages = ref<PackageInfo[]>([])
const detailDialogVisible = ref(false)
const selectedPkg = ref<PackageInfo | null>(null)
const availableVersions = ref<PackageVersion[]>([])
const selectedVersion = ref<PackageVersion | null>(null)
const selectedMirror = ref('aliyun')
const installing = ref(false)
const installProgress = ref(0)

const categories = [
  { key: 'all', label: '全部' },
  { key: 'language', label: '语言运行时' },
  { key: 'webserver', label: 'Web 服务器' },
  { key: 'database', label: '数据库' },
  { key: 'cache', label: '缓存' },
]

const mirrors = [
  { key: 'aliyun', label: '阿里云镜像 (国内高速)' },
  { key: 'official', label: '官方源' },
  { key: 'tencent', label: '腾讯云镜像' },
]

const filteredPackages = computed(() => {
  let list = packages.value
  if (activeCategory.value !== 'all') {
    list = list.filter(p => p.category === activeCategory.value)
  }
  if (searchQuery.value.trim()) {
    const q = searchQuery.value.toLowerCase().trim()
    list = list.filter(p => 
      p.displayName.toLowerCase().includes(q) ||
      p.description.toLowerCase().includes(q)
    )
  }
  return list
})

const targetInstallPath = computed(() => {
  if (!selectedPkg.value || !selectedVersion.value) return 'runtimes/'
  const p = selectedPkg.value.key
  const v = selectedVersion.value.version
  const variant = selectedVersion.value.variant ? `-${selectedVersion.value.variant}` : ''
  const arch = selectedVersion.value.architecture
  return `runtimes\\${p}\\${v}${variant}-${arch}`
})

function getCategoryLabel(catKey: string): string {
  const cat = categories.find(c => c.key === catKey)
  return cat ? cat.label : catKey
}

function formatFileSize(bytes: number): string {
  if (!bytes) return '未知'
  const mb = (bytes / (1024 * 1024)).toFixed(1)
  return `${mb} MB`
}

function isVersionInstalled(version: string): boolean {
  if (!selectedPkg.value || !version) return false
  return selectedPkg.value.installedVersions.includes(version)
}

async function loadPackages() {
  try {
    packages.value = await listPackages()
  } catch (err: any) {
    toast.error('加载软件包失败: ' + (err?.message || '未知错误'))
  }
}

async function openPackageDetail(pkg: PackageInfo) {
  selectedPkg.value = pkg
  installing.value = false
  installProgress.value = 0
  const versions = await getPackageVersions(pkg.key)
  availableVersions.value = versions
  selectedVersion.value = versions.find(v => v.isRecommended) || versions[0] || null
  detailDialogVisible.value = true
}

async function handleStartInstall() {
  if (!selectedPkg.value || !selectedVersion.value) return

  installing.value = true
  installProgress.value = 10

  const timer = setInterval(() => {
    if (installProgress.value < 85) {
      installProgress.value += 15
    }
  }, 200)

  try {
    await installPackage({
      packageKey: selectedPkg.value.key,
      version: selectedVersion.value.version,
      variant: selectedVersion.value.variant,
      architecture: selectedVersion.value.architecture,
      downloadUrl: selectedVersion.value.downloadUrl,
      installPath: targetInstallPath.value,
      mirror: selectedMirror.value,
    })

    clearInterval(timer)
    installProgress.value = 100

    setTimeout(async () => {
      toast.success(`${selectedPkg.value?.displayName} v${selectedVersion.value?.version} 安装成功！`)
      installing.value = false
      detailDialogVisible.value = false
      await loadPackages()
    }, 400)
  } catch (err: any) {
    clearInterval(timer)
    installing.value = false
    toast.error('安装失败: ' + (err?.message || '未知错误'))
  }
}

function goToRuntimes() {
  detailDialogVisible.value = false
  router.push('/runtimes')
}

onMounted(() => {
  loadPackages()
})
</script>

<style scoped>
.search-box {
  position: relative;
  display: flex;
  align-items: center;
  width: 260px;
}

.search-icon {
  position: absolute;
  left: 10px;
  color: var(--text-muted);
  font-size: 0.85rem;
  pointer-events: none;
}

.search-input {
  width: 100%;
  padding-left: 32px;
  height: 32px;
  font-size: 0.825rem;
  border-radius: var(--h-radius);
}

.category-tabs {
  display: flex;
  gap: 8px;
  border-bottom: 1px solid var(--border-color);
  padding-bottom: 10px;
  margin-bottom: var(--h-padding);
}

.category-btn {
  background: transparent;
  border: none;
  padding: 6px 14px;
  font-size: 0.85rem;
  color: var(--text-secondary);
  border-radius: var(--h-radius);
  cursor: pointer;
  transition: none !important;
}

.category-btn:hover {
  background: var(--bg-hover);
  color: var(--text-primary);
}

.category-btn.active {
  background: var(--p-primary-color);
  color: #ffffff;
}

.packages-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: var(--h-padding);
}

.package-card {
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  border-radius: var(--h-radius);
  padding: var(--h-padding);
  display: flex;
  flex-direction: column;
  gap: 12px;
  cursor: pointer;
  transition: none !important;
}

.package-card:hover {
  background: #f8fafc;
  border-color: var(--border-hover);
}

.pkg-top {
  display: flex;
  align-items: center;
  gap: 12px;
}

.pkg-icon-wrap {
  width: 44px;
  height: 44px;
  border-radius: var(--radius-sm);
  display: flex;
  align-items: center;
  justify-content: center;
  color: #ffffff;
  font-size: 1.3rem;
  flex-shrink: 0;
}

.pkg-header-info {
  display: flex;
  flex-direction: column;
  gap: 2px;
  flex: 1;
}

.pkg-name-row {
  display: flex;
  align-items: center;
  gap: 8px;
}

.pkg-name {
  font-size: 1rem;
  color: var(--text-primary);
}

.installed-tag {
  font-size: 0.7rem;
  padding: 1px 6px;
}

.pkg-ver {
  font-size: 0.775rem;
  color: var(--text-muted);
}

.pkg-desc {
  font-size: 0.8rem;
  color: var(--text-secondary);
  line-height: 1.5;
  margin: 0;
  min-height: 38px;
}

.pkg-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-top: auto;
  padding-top: 8px;
  border-top: 1px solid var(--border-color);
}

.category-badge {
  font-size: 0.725rem;
  color: var(--text-muted);
  background: var(--bg-app);
  padding: 2px 8px;
  border-radius: var(--radius-sm);
  border: 1px solid var(--border-color);
}

.pkg-actions {
  display: flex;
  gap: 6px;
}

/* 详情弹窗样式 */
.detail-container {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.dialog-pkg-info {
  display: flex;
  align-items: flex-start;
  gap: 14px;
  background: var(--bg-app);
  padding: 12px;
  border-radius: var(--h-radius);
  border: 1px solid var(--border-color);
}

.dialog-pkg-icon {
  width: 48px;
  height: 48px;
  border-radius: var(--radius-sm);
  display: flex;
  align-items: center;
  justify-content: center;
  color: #ffffff;
  font-size: 1.4rem;
  flex-shrink: 0;
}

.dialog-pkg-text {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.dialog-pkg-title {
  font-size: 1rem;
  color: var(--text-primary);
}

.dialog-pkg-desc {
  font-size: 0.8rem;
  color: var(--text-secondary);
}

.official-link {
  font-size: 0.75rem;
  color: var(--p-primary-color);
  text-decoration: none;
  display: flex;
  align-items: center;
  gap: 4px;
}
.official-link:hover {
  text-decoration: underline;
}

.section-title {
  font-size: 0.825rem;
  color: var(--text-muted);
  margin-top: 4px;
}

.version-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
  max-height: 220px;
  overflow-y: auto;
}

.version-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 10px 12px;
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-sm);
  cursor: pointer;
}

.version-item:hover {
  background: #f8fafc;
  border-color: var(--border-hover);
}

.version-item.selected {
  border-color: var(--p-primary-color);
  background: #f0fdf4;
}

.version-left {
  display: flex;
  align-items: center;
  gap: 10px;
}

.version-radio {
  cursor: pointer;
}

.version-info {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.version-name {
  font-size: 0.875rem;
  color: var(--text-primary);
}

.variant-tag, .arch-tag {
  font-size: 0.7rem;
  background: #e2e8f0;
  padding: 1px 4px;
  border-radius: 3px;
  margin-left: 4px;
  color: var(--text-secondary);
}

.version-sub {
  font-size: 0.725rem;
  color: var(--text-muted);
}

.version-tags {
  display: flex;
  gap: 6px;
}

.mirror-row {
  display: flex;
  gap: 8px;
}

.mirror-pill {
  cursor: pointer;
  padding: 6px 12px;
  border: 1px solid var(--border-color);
  border-radius: var(--h-radius);
  font-size: 0.8rem;
  color: var(--text-secondary);
  background: var(--bg-card);
}

.mirror-pill.active {
  border-color: var(--p-primary-color);
  color: var(--p-primary-color);
  background: #f0fdf4;
}

.hidden-radio {
  display: none;
}

.path-preview {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 0.8rem;
  background: var(--bg-app);
  padding: 8px 12px;
  border-radius: var(--radius-sm);
  border: 1px solid var(--border-color);
  color: var(--text-secondary);
}

.install-progress {
  display: flex;
  flex-direction: column;
  gap: 6px;
  margin-top: 6px;
}

.progress-labels {
  display: flex;
  justify-content: space-between;
  font-size: 0.775rem;
  color: var(--text-secondary);
}

.empty-state {
  text-align: center;
  padding: 48px 0;
  color: var(--text-dim);
}

.empty-icon {
  font-size: 2.2rem;
  margin-bottom: 8px;
}

.empty-text {
  font-size: 0.875rem;
}
</style>
