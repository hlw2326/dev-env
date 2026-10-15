<template>
	<!-- 已安装软件分组视图 -->
	<div class="installed-section">
		<!-- 未安装任何软件空状态 -->
		<div v-if="installedGroups.length === 0" class="surface-card">
			<h-empty
				icon="pi pi-shopping-bag"
				title="还没有安装任何软件"
				description="切换到「全部软件」页签选择需要的环境或数据库一键安装"
			>
				<p-button label="浏览全部软件" @click="tab = 'market'" />
			</h-empty>
		</div>

		<!-- 按软件类型分组展示 -->
		<div v-else class="group-list">
			<div
				v-for="group in installedGroups"
				:key="group.kind"
				class="software-group"
			>
				<span class="group-title">{{ group.name }}</span>
				<div class="group-rows">
					<comp-installed-row
						v-for="item in group.items"
						:key="item.id"
						:item="item"
						:progress-map="progressMap"
						@start="startItem"
						@stop="stopItem"
						@main="markMain"
						@config="openConfig"
						@remove="removeItem"
						@cancel="stopInstall"
						@retry="startInstall"
					/>
				</div>
			</div>
		</div>
	</div>
</template>

<script setup lang="ts">
import { useSoftware } from '../use-software'
import CompInstalledRow from './installed-row.vue'

// 已安装分组状态与操作方法
const {
	tab,
	installedGroups,
	progressMap,
	startInstall,
	stopInstall,
	removeItem,
	markMain,
	startItem,
	stopItem,
	openConfig,
} = useSoftware()
</script>

<style scoped lang="scss">
/* 已安装分组列表 */
.group-list {
	display: flex;
	flex-direction: column;
	gap: 1rem;
}

/* 单个软件分组 */
.software-group {
	display: flex;
	flex-direction: column;
	gap: 0.5rem;
}

/* 分组小标题 */
.group-title {
	font-size: var(--font-base);
	color: var(--text-secondary);
}

/* 分组内版本行列表 */
.group-rows {
	display: flex;
	flex-direction: column;
	gap: 0.5rem;
}
</style>
