<script setup>
defineProps({
  modelValue: { type: Number, default: 0 },
  readonly: { type: Boolean, default: false },
})
const emit = defineEmits(['update:modelValue'])
</script>

<template>
  <div class="flex gap-0.5">
    <button
      v-for="n in 5"
      :key="n"
      type="button"
      class="text-xl leading-none"
      :class="[n <= modelValue ? 'text-amber-400' : 'text-gray-300', readonly ? 'cursor-default' : 'transition hover:scale-110']"
      :title="`${n} star${n > 1 ? 's' : ''}`"
      :disabled="readonly"
      @click="!readonly && emit('update:modelValue', n)"
    >
      ★
    </button>
  </div>
</template>
