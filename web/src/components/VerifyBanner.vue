<script setup>
import { computed, ref } from 'vue'
import { api } from '@/api/client'
import { useAuthStore } from '@/stores/auth'

const auth = useAuthStore()

// Shown only to a logged-in user whose email is not yet verified. The flag is
// explicitly compared to false so a legacy cached user missing the field (older
// login payload) does not trigger a false banner.
const show = computed(() => auth.isAuthenticated && auth.user && auth.user.is_verified === false)

const sending = ref(false)
const sent = ref(false)

async function resend() {
  sending.value = true
  try {
    await api.post('/auth/resend-verification')
    sent.value = true
  } catch {
    /* already verified elsewhere or throttled — stay quiet */
  } finally {
    sending.value = false
  }
}
</script>

<template>
  <div v-if="show" class="border-b border-amber-200 bg-amber-50">
    <div class="mx-auto flex max-w-5xl flex-wrap items-center justify-between gap-2 px-4 py-2 text-sm text-amber-800">
      <span>✉️ {{ sent ? $t('verify.resent') : $t('verify.banner') }}</span>
      <button v-if="!sent" class="font-medium underline hover:no-underline" :disabled="sending" @click="resend">
        {{ $t('verify.resend') }}
      </button>
    </div>
  </div>
</template>
