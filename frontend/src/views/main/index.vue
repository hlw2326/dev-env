<template>
	<div class="main-shell">
		<!-- 顶部自定义标题栏 -->
		<main-titlebar />

		<!-- 主体区域：左侧导航 + 右侧内容 -->
		<div class="main-body">
			<main-sidebar />
			<main class="main-content">
				<router-view />
			</main>
		</div>
	</div>
</template>

<script setup lang="ts">
import MainTitlebar from './comps/titlebar.vue'
import MainSidebar from './comps/sidebar.vue'
</script>

<style scoped lang="scss">
/* 主界面外壳 */
.main-shell {
	display: flex;
	flex-direction: column;
	height: 100vh;
	width: 100vw;
	background-color: var(--bg-app);
	overflow: hidden;
}

/* 主体布局 */
.main-body {
	display: flex;
	flex: 1;
	overflow: hidden;
}

/* 页面内容滚动区 */
.main-content {
	flex: 1;
	overflow-y: auto;
}
</style>
