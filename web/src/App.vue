<script setup>
import NavBar from '@/components/NavBar.vue'
</script>

<template>
  <div class="min-h-screen">
    <NavBar />
    <main class="mx-auto max-w-5xl px-4 py-6">
      <RouterView />
    </main>
  </div>
</template>
