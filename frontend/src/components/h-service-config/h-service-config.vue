<template>
	<h-modal
		v-if="item !== null"
		v-model:visible="visible"
		:title="`软件设置 · ${item.name} ${item.version}`"
		width="40rem"
		:modal="true"
		:closable="true"
		:mask-close="true"
		:footer="false"
		ok-text="确定"
		ok-icon="pi pi-check"
		cancel-text="取消"
		:loading="false"
	>
		<comp-config-php
			v-if="item.kind === 'php'"
			:item="item"
			@main="emit('main', $event)"
			@folder="emit('folder', $event)"
		/>

		<comp-config-daemon
			v-else-if="item.daemon"
			:item="item"
			:busy="busy"
			@start="emit('start', $event)"
			@stop="emit('stop', $event)"
			@main="emit('main', $event)"
			@autostart="(target, on) => emit('autostart', target, on)"
			@folder="emit('folder', $event)"
		/>

		<comp-config-overview
			v-else
			:item="item"
			:busy="busy"
			@main="emit('main', $event)"
			@folder="emit('folder', $event)"
		/>
	</h-modal>
</template>

<script setup lang="ts">
import CompConfigPhp from '@/views/main/software/comps/config-php.vue'
import CompConfigDaemon from '@/views/main/software/comps/config-daemon.vue'
import CompConfigOverview from '@/views/main/software/comps/config-overview.vue'
import type { HServiceConfigEmits, HServiceConfigProps } from './types.d'

defineOptions({ name: 'HServiceConfig' })

defineProps<HServiceConfigProps>()
const visible = defineModel<boolean>('visible', { required: true })
const emit = defineEmits<HServiceConfigEmits>()
</script>
