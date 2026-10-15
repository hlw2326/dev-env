<template>
	<Drawer
		v-model:visible="visible"
		:header="title"
		:aria-label="title"
		:position="position"
		:modal="modal"
		:block-scroll="modal"
		:show-close-icon="closable && !loading"
		:close-on-escape="closeOnEscape && !loading"
		:dismissable="maskClose && !loading"
		close-icon="pi pi-times"
		:style="{ width, maxWidth: '100vw' }"
		:pt="drawerPassThrough"
		class="hlw-drawer"
		@show="emit('open')"
		@hide="emit('close')"
	>
		<template v-if="$slots.header" #header>
			<slot name="header"></slot>
		</template>

		<div v-if="$slots.top" class="top">
			<slot name="top"></slot>
		</div>

		<div v-if="$slots.toolbar" class="toolbar">
			<slot name="toolbar"></slot>
		</div>
		<div class="body">
			<div class="content" :class="{ padded: padding }">
				<slot></slot>
			</div>
		</div>

		<template #footer>
			<slot name="footer">
				<Button
					:label="cancelText"
					severity="secondary"
					variant="outlined"
					:disabled="loading"
					@click="cancel"
				/>
				<Button :label="okText" :icon="okIcon" :loading="loading" @click="confirm" />
			</slot>
		</template>
	</Drawer>
</template>

<script setup lang="ts">
import Drawer from 'primevue/drawer'
import Button from 'primevue/button'
import type { HlwDrawerProps, HlwDrawerEmits, HlwDrawerExpose } from './types'

defineOptions({ name: 'HlwDrawer' })
const props = withDefaults(defineProps<HlwDrawerProps>(), {
	width: '52rem',
	position: 'right',
	modal: true,
	closable: true,
	maskClose: false,
	closeOnEscape: true,
	padding: true,
	footer: false,
	okText: '确定',
	okIcon: 'pi pi-check',
	cancelText: '取消',
	loading: false,
})
const visible = defineModel<boolean>('visible', { required: true })
const emit = defineEmits<HlwDrawerEmits>()
const drawerPassThrough = {
	root: {
		style: {
			display: 'flex',
			flexDirection: 'column',
			height: '100%',
			maxHeight: '100%',
			minHeight: '0',
			boxShadow: 'none',
			background: 'var(--bg-card)',
			color: 'var(--text-primary)',
			border: 'none',
			overflow: 'hidden',
		},
	},
	header: {
		style: {
			flexShrink: 0,
			padding: '0.625rem var(--h-padding)',
			borderBottom: '1px solid var(--border-color)',
			backgroundColor: 'var(--bg-card)',
			color: 'var(--text-primary)',
		},
	},
	title: {
		style: {
			minWidth: 0,
			overflow: 'hidden',
			textOverflow: 'ellipsis',
			whiteSpace: 'nowrap',
			fontSize: 'var(--font-base)',
			fontWeight: 400,
			lineHeight: 'var(--leading-snug)',
		},
	},
	content: {
		style: {
			display: 'flex',
			flexDirection: 'column',
			flex: 1,
			minHeight: 0,
			padding: 0,
			overflow: 'hidden',
		},
	},
	footer: () => ({
		style: {
			display: props.footer ? 'flex' : 'none',
			justifyContent: 'flex-end',
			alignItems: 'center',
			gap: 'var(--h-padding)',
			padding: 'var(--h-padding)',
			borderTop: '1px solid var(--border-color)',
			backgroundColor: 'var(--bg-card)',
			flexShrink: 0,
		},
	}),
} 

function confirm(): void {
	if (!props.loading) emit('ok')
}
function cancel(): void {
	if (props.loading) return
	visible.value = false
	emit('cancel')
}
function open(): void { visible.value = true }
function close(): void { visible.value = false }
defineExpose<HlwDrawerExpose>({ open, close })
</script>

<style scoped lang="scss">
.toolbar {
	flex-shrink: 0;
	min-width: 0;
	padding: var(--h-padding);
	border-bottom: 1px solid var(--border-color);
}
.top {
	flex-shrink: 0;
	min-width: 0;
}
.body {
	flex: 1;
	min-height: 0;
	min-width: 0;
	overflow: auto;
	overscroll-behavior: contain;
	scrollbar-gutter: stable;
}
.content { min-width: 0; }
.padded { padding: var(--h-padding); }
</style>
