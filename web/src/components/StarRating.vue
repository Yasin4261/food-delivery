<script setup>
defineProps({ modelValue: { type: Number, default: 0 } })
const emit = defineEmits(['update:modelValue'])
</script>

<template>
  <div class="flex gap-0.5">
    <button
      v-for="n in 5"
      :key="n"
      type="button"
      class="text-xl leading-none transition hover:scale-110"
      :class="n <= modelValue ? 'text-amber-400' : 'text-gray-300'"
      :title="`${n} star${n > 1 ? 's' : ''}`"
      @click="emit('update:modelValue', n)"
    >
      ★
    </button>
  </div>
</template>
