<template>
	<!-- 全局消息提示 -->
	<p-toast position="top-center" :closable="false" />
	<!-- 全局气泡确认框 -->
	<p-confirm-popup />
	<!-- 路由出口 -->
	<router-view />
</template>
