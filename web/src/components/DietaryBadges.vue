<script setup>
// Small dietary badges for a dish. Renders only the flags that are set.
defineProps({ item: { type: Object, required: true } })

const flags = [
  { key: 'is_vegetarian', label: 'vegetarian', icon: '🥗' },
  { key: 'is_vegan', label: 'vegan', icon: '🌱' },
  { key: 'is_gluten_free', label: 'glutenFree', icon: '🌾' },
  { key: 'is_halal', label: 'halal', icon: '☪️' },
]
</script>

<template>
  <span class="flex flex-wrap items-center gap-1">
    <span
      v-for="f in flags.filter((f) => item[f.key])"
      :key="f.key"
      class="badge bg-green-50 text-green-700"
      :title="$t(`dietary.${f.label}`)"
    >
      {{ f.icon }} {{ $t(`dietary.${f.label}`) }}
    </span>
  </span>
</template>
