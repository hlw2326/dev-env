<template>
	<h-modal
		v-if="item !== null"
		v-model:visible="visible"
		:title="`软件设置 · ${kindNames[item.kind]} ${item.version}`"
		width="40rem"
		:footer="false"
	>
		<cfg-php
			v-if="item.kind === 'php'"
			:item="item"
			@set-main="emit('setMain', $event)"
			@open-folder="emit('openFolder', $event)"
		/>

		<cfg-mysql
			v-else-if="item.kind === 'mysql'"
			:item="item"
			@start="emit('start', $event)"
			@stop="emit('stop', $event)"
			@autostart="(sw, on) => emit('autostart', sw, on)"
			@open-folder="emit('openFolder', $event)"
		/>

		<cfg-redis
			v-else-if="item.kind === 'redis'"
			:item="item"
			@start="emit('start', $event)"
			@stop="emit('stop', $event)"
			@autostart="(sw, on) => emit('autostart', sw, on)"
		/>

		<cfg-nginx
			v-else-if="item.kind === 'nginx'"
			:item="item"
			@start="emit('start', $event)"
			@stop="emit('stop', $event)"
			@autostart="(sw, on) => emit('autostart', sw, on)"
			@open-folder="emit('openFolder', $event)"
		/>

		<div v-else class="node-overview">
			<div class="kv-row">
				<span class="kv-label">软件版本</span>
				<span class="kv-val font-mono">Node.js {{ item.version }}</span>
			</div>
			<div class="kv-row">
				<span class="kv-label">安装路径</span>
				<div class="path-group">
					<span class="kv-val font-mono">{{ item.path }}</span>
					<p-button
						size="small"
						severity="secondary"
						variant="outlined"
						icon="pi pi-folder-open"
						label="打开"
						@click="emit('openFolder', item.id)"
					/>
				</div>
			</div>
			<div class="kv-row">
				<span class="kv-label">默认版本</span>
				<div>
					<span v-if="item.main" class="main-text">当前已设为默认 Node.js 版本</span>
					<p-button
						v-else
						size="small"
						label="设为默认版本"
						@click="emit('setMain', item)"
					/>
				</div>
			</div>
		</div>
	</h-modal>
</template>

<script setup lang="ts">
import CfgPhp from './cfg-php.vue'
import CfgMysql from './cfg-mysql.vue'
import CfgRedis from './cfg-redis.vue'
import CfgNginx from './cfg-nginx.vue'
import type { Software } from '@/types'

defineProps<{
	item: Software | null
}>()

const visible = defineModel<boolean>('visible', { default: false })

const emit = defineEmits<{
	start: [item: Software]
	stop: [item: Software]
	setMain: [item: Software]
	autostart: [item: Software, on: boolean]
	openFolder: [id: number]
}>()

const kindNames: Record<string, string> = {
	nginx: 'Nginx',
	php: 'PHP',
	nodejs: 'Node.js',
	mysql: 'MySQL',
	redis: 'Redis',
}
</script>

<style scoped lang="scss">
.node-overview {
	display: flex;
	flex-direction: column;
	gap: 0.75rem;
}

.kv-row {
	display: flex;
	align-items: center;
	justify-content: space-between;
	padding: 0.75rem 0.875rem;
	background: var(--bg-app);
	border: 1px solid var(--border-color);
	border-radius: var(--radius-md);
}

.kv-label {
	font-size: var(--font-sm);
	color: var(--text-secondary);
}

.kv-val {
	font-size: var(--font-xs);
	color: var(--text-primary);
}

.path-group {
	display: flex;
	align-items: center;
	gap: 0.5rem;
}

.main-text {
	font-size: var(--font-xs);
	color: var(--ok);
}
</style>
