<template>
	<!-- 项目详情域名管理页签 -->
	<div class="detail-domains">
		<!-- 新增域名输入行 -->
		<div class="add-row">
			<p-input-text
				v-model="newHost"
				placeholder="输入新域名，如 api.localhost"
				class="font-mono flex-input"
				@keyup.enter="addHost"
			/>
			<p-button icon="pi pi-plus" label="添加域名" @click="addHost" />
		</div>

		<!-- 已绑定域名列表 -->
		<div class="host-list">
			<div v-for="host in project.hosts" :key="host" class="host-item">
				<span class="host-name font-mono">{{ host }}</span>
				<p-button
					size="small"
					severity="danger"
					variant="outlined"
					icon="pi pi-trash"
					label="移除"
					:disabled="project.hosts.length <= 1"
					@click="removeHost(host)"
				/>
			</div>
		</div>
	</div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import type { Project, ProjectUpdateInput } from '@/types'

// 接收项目属性
const props = defineProps<{
	project: Project
}>()

// 保存更新事件
const emit = defineEmits<{
	save: [input: ProjectUpdateInput]
}>()

// 新增域名输入值
const newHost = ref('')

// 添加新域名并提交更新
function addHost(): void {
	const clean = newHost.value.trim().toLowerCase()
	if (clean === '') {
		return
	}
	if (props.project.hosts.includes(clean)) {
		newHost.value = ''
		return
	}
	const nextHosts = [...props.project.hosts, clean]
	newHost.value = ''
	emit('save', {
		id: props.project.id,
		hosts: nextHosts,
		root: props.project.root,
		run: props.project.run,
		software: props.project.software,
		command: props.project.command,
		nginx: props.project.nginx,
		rewriteRule: props.project.rewriteRule,
		environmentFile: props.project.environmentFile,
		autostart: props.project.autostart,
	})
}

// 移除指定域名并提交更新
function removeHost(target: string): void {
	if (props.project.hosts.length <= 1) {
		return
	}
	const nextHosts = props.project.hosts.filter((item) => item !== target)
	emit('save', {
		id: props.project.id,
		hosts: nextHosts,
		root: props.project.root,
		run: props.project.run,
		software: props.project.software,
		command: props.project.command,
		nginx: props.project.nginx,
		rewriteRule: props.project.rewriteRule,
		environmentFile: props.project.environmentFile,
		autostart: props.project.autostart,
	})
}
</script>

<style scoped lang="scss">
/* 域名管理容器 */
.detail-domains {
	display: flex;
	flex-direction: column;
	gap: 0.875rem;
}

/* 新增域名输入栏 */
.add-row {
	display: flex;
	align-items: center;
	gap: 0.5rem;
}

/* 自适应输入框 */
.flex-input {
	flex: 1;
}

/* 域名列表容器 */
.host-list {
	display: flex;
	flex-direction: column;
	gap: 0.5rem;
}

/* 单个域名条目 */
.host-item {
	display: flex;
	align-items: center;
	justify-content: space-between;
	padding: 0.625rem 0.875rem;
	background: var(--bg-app);
	border: 1px solid var(--border-color);
	border-radius: var(--h-radius);
}

/* 域名文本 */
.host-name {
	font-size: var(--font-base);
	color: var(--text-primary);
}
</style>
