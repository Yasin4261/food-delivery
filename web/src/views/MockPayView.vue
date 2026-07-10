<script setup>
import { useRoute } from 'vue-router'

// Development-only stand-in for the iyzico hosted payment page. The dev mock
// gateway sends the browser here; these forms POST the token back to the real
// callback endpoint exactly like iyzico would, and the API redirects to
// /orders with the outcome.
const route = useRoute()
const token = typeof route.query.token === 'string' ? route.query.token : ''
</script>

<template>
  <div class="mx-auto mt-10 max-w-sm text-center">
    <div class="card space-y-4 border-dashed shadow-md">
      <div class="text-4xl">🏦</div>
      <h1 class="text-xl font-bold">Sandbox payment</h1>
      <p class="text-sm text-gray-500">
        This simulates the iyzico hosted checkout page (development only — no real charge).
      </p>
      <p v-if="!token" class="rounded-lg bg-red-50 px-3 py-2 text-sm text-red-700">Missing checkout token.</p>
      <template v-else>
        <form action="/api/v2/payments/callback" method="post">
          <input type="hidden" name="token" :value="token" />
          <button class="btn-primary w-full">✅ Complete payment</button>
        </form>
        <form action="/api/v2/payments/callback" method="post">
          <input type="hidden" name="token" :value="token + ':fail'" />
          <button class="btn-ghost w-full">❌ Simulate failure</button>
        </form>
      </template>
    </div>
  </div>
</template>
