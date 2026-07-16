<script setup>
import NavBar from '@/components/NavBar.vue'
import VerifyBanner from '@/components/VerifyBanner.vue'
</script>

<template>
  <div class="flex min-h-screen flex-col">
    <NavBar />
    <VerifyBanner />
    <main class="mx-auto w-full max-w-5xl grow px-4 py-8">
      <RouterView />
    </main>
    <footer class="border-t border-gray-200 bg-white py-6">
      <div class="mx-auto max-w-5xl px-4 text-center text-sm text-gray-400">
        {{ $t('footer.tagline') }}
      </div>
    </footer>
  </div>
</template>
