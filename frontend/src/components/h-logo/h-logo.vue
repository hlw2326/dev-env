<template>
	<!-- 软件品牌图标容器 -->
	<div class="h-logo" :class="[sizeClass, toneClass]">
		<!-- 矢量图标展示 -->
		<img
			v-if="hasImage"
			:src="meta.logo"
			:alt="kind"
			class="logo-img"
			@error="onError"
		/>
		<!-- 字符缩写兜底展示 -->
		<span v-else class="logo-text font-mono">
			{{ meta.short }}
		</span>
	</div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import type { HLogoProps, LogoMeta } from './types.d'

// 全局软件图标组件名称
defineOptions({
	name: 'HLogo',
})

// 接收图标属性
const props = defineProps<HLogoProps>()

// 软件图标元数据查表
const LOGO_MAP: Record<string, LogoMeta> = {
	php: { short: 'PHP', tone: 'php', logo: '' },
	nodejs: { short: 'Node', tone: 'node', logo: '' },
	go: { short: 'Go', tone: 'go', logo: '' },
	python: { short: 'Py', tone: 'python', logo: '' },
	composer: { short: 'Pkg', tone: 'composer', logo: '' },
	mysql: { short: 'SQL', tone: 'mysql', logo: '' },
	mariadb: { short: 'MDB', tone: 'mariadb', logo: '' },
	postgresql: { short: 'PG', tone: 'postgresql', logo: '' },
	mongodb: { short: 'MGO', tone: 'mongodb', logo: '' },
	redis: { short: 'Redis', tone: 'redis', logo: '' },
	memcached: { short: 'Mem', tone: 'memcached', logo: '' },
	java: { short: 'JDK', tone: 'java', logo: '' },
	dotnet: { short: '.NET', tone: 'dotnet', logo: '' },
	deno: { short: 'Deno', tone: 'deno', logo: '' },
	bun: { short: 'Bun', tone: 'bun', logo: '' },
	nginx: { short: 'NGX', tone: 'nginx', logo: '' },
	caddy: { short: 'Caddy', tone: 'caddy', logo: '' },
	sqlite: { short: 'Lite', tone: 'sqlite', logo: '' },
	gradle: { short: 'Gdl', tone: 'gradle', logo: '' },
	maven: { short: 'Mvn', tone: 'maven', logo: '' },
}

// 默认软件图标元数据
const DEFAULT_META: LogoMeta = {
	short: 'DEV',
	tone: 'default',
	logo: '',
}

// 图片加载错误标记
const hasError = ref(false)

// 当前软件图标元数据
const meta = computed(() => {
	const item = LOGO_MAP[props.kind]
	if (item) {
		return item
	}
	return DEFAULT_META
})

// 是否显示图片
const hasImage = computed(() => {
	if (hasError.value) {
		return false
	}
	return meta.value.logo.length > 0
})

// 尺寸样式类名
const sizeClass = computed(() => {
	if (props.size) {
		return `size-${props.size}`
	}
	return 'size-medium'
})

// 色调样式类名
const toneClass = computed(() => {
	if (hasImage.value) {
		return 'tone-image'
	}
	return `tone-${meta.value.tone}`
})

// 处理图片加载失败
function onError(): void {
	hasError.value = true
}

// 监听软件类型变化重置错误状态
watch(
	() => props.kind,
	() => {
		hasError.value = false
	},
)
</script>

<style scoped lang="scss">
/* 品牌图标外框 */
.h-logo {
	display: inline-flex;
	align-items: center;
	justify-content: center;
	border-radius: var(--h-radius);
	border: 1px solid var(--border-color);
	flex-shrink: 0;
	overflow: hidden;
	user-select: none;

	/* 小号尺寸 */
	&.size-small {
		width: 1.875rem;
		height: 1.875rem;
		font-size: var(--font-xs);
	}

	/* 中号尺寸 */
	&.size-medium {
		width: 2.5rem;
		height: 2.5rem;
		font-size: var(--font-xs);
	}

	/* 大号尺寸 */
	&.size-large {
		width: 2.75rem;
		height: 2.75rem;
		font-size: var(--font-sm);
	}

	/* 图片背景色块 */
	&.tone-image {
		background-color: var(--bg-card);
		border-color: var(--border-color);
	}

	/* PHP 品牌色调 */
	&.tone-php {
		color: var(--logo-php);
		background-color: color-mix(in srgb, var(--logo-php) 12%, transparent);
		border-color: color-mix(in srgb, var(--logo-php) 30%, transparent);
	}

	/* Node.js 品牌色调 */
	&.tone-node {
		color: var(--logo-node);
		background-color: color-mix(in srgb, var(--logo-node) 12%, transparent);
		border-color: color-mix(in srgb, var(--logo-node) 30%, transparent);
	}

	/* Go 品牌色调 */
	&.tone-go {
		color: var(--logo-go);
		background-color: color-mix(in srgb, var(--logo-go) 12%, transparent);
		border-color: color-mix(in srgb, var(--logo-go) 30%, transparent);
	}

	/* Python 品牌色调 */
	&.tone-python {
		color: var(--logo-python);
		background-color: color-mix(in srgb, var(--logo-python) 12%, transparent);
		border-color: color-mix(in srgb, var(--logo-python) 30%, transparent);
	}

	/* Composer 品牌色调 */
	&.tone-composer {
		color: var(--logo-composer);
		background-color: color-mix(in srgb, var(--logo-composer) 12%, transparent);
		border-color: color-mix(in srgb, var(--logo-composer) 30%, transparent);
	}

	/* MySQL 品牌色调 */
	&.tone-mysql {
		color: var(--logo-mysql);
		background-color: color-mix(in srgb, var(--logo-mysql) 12%, transparent);
		border-color: color-mix(in srgb, var(--logo-mysql) 30%, transparent);
	}

	/* MariaDB 品牌色调 */
	&.tone-mariadb {
		color: var(--logo-mariadb);
		background-color: color-mix(in srgb, var(--logo-mariadb) 12%, transparent);
		border-color: color-mix(in srgb, var(--logo-mariadb) 30%, transparent);
	}

	/* PostgreSQL 品牌色调 */
	&.tone-postgresql {
		color: var(--logo-postgresql);
		background-color: color-mix(in srgb, var(--logo-postgresql) 12%, transparent);
		border-color: color-mix(in srgb, var(--logo-postgresql) 30%, transparent);
	}

	/* MongoDB 品牌色调 */
	&.tone-mongodb {
		color: var(--logo-mongodb);
		background-color: color-mix(in srgb, var(--logo-mongodb) 12%, transparent);
		border-color: color-mix(in srgb, var(--logo-mongodb) 30%, transparent);
	}

	/* Redis 品牌色调 */
	&.tone-redis {
		color: var(--logo-redis);
		background-color: color-mix(in srgb, var(--logo-redis) 12%, transparent);
		border-color: color-mix(in srgb, var(--logo-redis) 30%, transparent);
	}

	/* Memcached 品牌色调 */
	&.tone-memcached {
		color: var(--logo-memcached);
		background-color: color-mix(in srgb, var(--logo-memcached) 12%, transparent);
		border-color: color-mix(in srgb, var(--logo-memcached) 30%, transparent);
	}

	&.tone-java,
	&.tone-dotnet,
	&.tone-deno,
	&.tone-bun,
	&.tone-nginx,
	&.tone-caddy,
	&.tone-sqlite,
	&.tone-gradle,
	&.tone-maven {
		color: var(--logo-extra);
		background-color: color-mix(in srgb, var(--logo-extra) 12%, transparent);
		border-color: color-mix(in srgb, var(--logo-extra) 30%, transparent);
	}

	/* 默认品牌色调 */
	&.tone-default {
		color: var(--logo-default);
		background-color: color-mix(in srgb, var(--logo-default) 12%, transparent);
		border-color: color-mix(in srgb, var(--logo-default) 30%, transparent);
	}

	/* 内部图片样式 */
	.logo-img {
		width: 100%;
		height: 100%;
		object-fit: contain;
		padding: 0.25rem;
	}
}
</style>
