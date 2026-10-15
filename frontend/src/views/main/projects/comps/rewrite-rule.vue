<template>
	<div class="rewrite-rule">
		<label for="nginx-rewrite-rule">伪静态</label>
		<p-select
			id="nginx-rewrite-rule"
			v-model="model"
			:options="options"
			option-label="label"
			option-value="value"
		/>
		<span class="hint">选择框架规则后由 DevEnv 生成；需要特殊规则时在下方高级配置中补充。</span>
	</div>
</template>

<script setup lang="ts">
const model = defineModel<string>({ default: 'default' })

const options = [
	{ label: '通用 PHP（默认）', value: 'default' },
	{ label: 'ThinkPHP', value: 'thinkphp' },
	{ label: 'Laravel', value: 'laravel' },
	{ label: 'WordPress', value: 'wordpress' },
	{ label: 'Typecho', value: 'typecho' },
]
</script>

<style scoped lang="scss">
.rewrite-rule {
	display: flex;
	flex-direction: column;
	gap: 0.35rem;
}

.hint {
	color: var(--text-secondary);
	font-size: var(--font-sm);
	line-height: 1.5;
}
</style>
