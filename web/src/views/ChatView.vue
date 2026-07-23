<script setup>
import { computed, nextTick, onBeforeUnmount, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { api, page } from '@/api/client'
import { useAuthStore } from '@/stores/auth'
import { i18n } from '@/i18n'

const auth = useAuthStore()
const route = useRoute()
const router = useRouter()

const conversations = ref([])
const active = ref(null)
const messages = ref([])
const draft = ref('')
const loading = ref(true)
const error = ref('')
const live = ref(false)
const listEl = ref(null)

let socket = null
const myId = auth.user?.id

const mine = (m) => m.sender_id === myId
const when = (iso) => new Date(iso).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })

const isAdmin = auth.user?.role === 'admin'

async function loadConversations() {
  const items = page(await api.get('/chat/conversations')).items
  // Admins also handle the support inbox (every support thread). It reuses the
  // same messaging UI — the shared chat endpoints admit an admin as a
  // participant of a support thread (#120).
  if (isAdmin) {
    const support = page(await api.get('/admin/support/conversations')).items
    const seen = new Set(items.map((c) => c.id))
    for (const c of support) if (!seen.has(c.id)) items.push(c)
  }
  // Label each thread from the *other* participant's perspective.
  await Promise.all(
    items.map(async (c) => {
      if (c.kind === 'support') {
        // For the target user this is "Support"; for an admin it's the customer.
        c._label =
          c.user_id === myId
            ? `🛟 ${i18n.global.t('chat.support')}`
            : `🛟 ${i18n.global.t('chat.customerNum', { id: c.user_id })}`
      } else if (c.user_id === myId) {
        try {
          c._label = `👨‍🍳 ${(await api.get(`/chefs/${c.chef_id}`)).business_name}`
        } catch {
          c._label = `👨‍🍳 Chef #${c.chef_id}`
        }
      } else {
        c._label = i18n.global.t('chat.customerNum', { id: c.user_id })
      }
    }),
  )
  conversations.value = items
}

// A regular user opens (or reuses) their own support thread and selects it.
async function contactSupport() {
  try {
    const conv = await api.post('/support/conversations')
    await loadConversations()
    const opened = conversations.value.find((c) => c.id === conv.id)
    if (opened) open(opened)
  } catch (e) {
    error.value = e.message
  }
}

function appendUnique(msg) {
  if (!messages.value.some((m) => m.id === msg.id)) {
    messages.value.push(msg)
    scrollDown()
  }
}

function connect(conv) {
  disconnect()
  // Browsers can't set headers on WS handshakes; the token goes in the query
  // (accepted by the API for upgrade requests only).
  const proto = location.protocol === 'https:' ? 'wss:' : 'ws:'
  socket = new WebSocket(
    `${proto}//${location.host}/api/v2/chat/conversations/${conv.id}/ws?access_token=${encodeURIComponent(auth.token)}`,
  )
  socket.onopen = () => (live.value = true)
  socket.onclose = () => (live.value = false)
  socket.onmessage = (ev) => {
    try {
      const frame = JSON.parse(ev.data)
      if (frame.type === 'read') {
        // The other party opened the thread — mark my delivered messages seen.
        if (frame.read && frame.read.reader_id !== myId) {
          messages.value.forEach((m) => {
            if (mine(m) && !m.read_at) m.read_at = frame.read.read_at
          })
        }
        return
      }
      // A message frame (envelope, or defensively a bare message).
      appendUnique(frame.message || frame)
    } catch {
      // ignore malformed frames
    }
  }
}

// Id of my most recent message the other party has read — drives the single
// "seen" marker under the last read outgoing bubble.
const lastSeenMineId = computed(() => {
  for (let i = messages.value.length - 1; i >= 0; i--) {
    const m = messages.value[i]
    if (mine(m) && m.read_at) return m.id
  }
  return null
})

function disconnect() {
  socket?.close()
  socket = null
  live.value = false
}

async function open(conv) {
  active.value = conv
  router.replace({ query: { c: conv.id } })
  error.value = ''
  try {
    messages.value = page(await api.get(`/chat/conversations/${conv.id}/messages?limit=100`)).items
    scrollDown()
    connect(conv)
    // Opening the thread marks the other party's messages read; clear the
    // badge optimistically.
    if (conv.unread_count > 0) {
      conv.unread_count = 0
      api.post(`/chat/conversations/${conv.id}/read`).catch(() => {})
    }
  } catch (e) {
    error.value = e.message
  }
}

async function send() {
  const body = draft.value.trim()
  if (!body || !active.value) return
  draft.value = ''
  try {
    if (socket && socket.readyState === WebSocket.OPEN) {
      // The server persists and broadcasts back to everyone in the room,
      // including us — appendUnique picks it up.
      socket.send(JSON.stringify({ body }))
    } else {
      appendUnique(await api.post(`/chat/conversations/${active.value.id}/messages`, { body }))
    }
  } catch (e) {
    error.value = e.message
  }
}

async function scrollDown() {
  await nextTick()
  if (listEl.value) listEl.value.scrollTop = listEl.value.scrollHeight
}

onMounted(async () => {
  try {
    await loadConversations()
    const wanted = Number(route.query.c)
    const target = conversations.value.find((c) => c.id === wanted)
    if (target) await open(target)
    else if (conversations.value.length === 1) await open(conversations.value[0])
  } catch (e) {
    error.value = e.message
  } finally {
    loading.value = false
  }
})

onBeforeUnmount(disconnect)
</script>

<template>
  <div class="space-y-4">
    <div class="flex items-start justify-between gap-3">
      <div>
        <h1 class="page-title">{{ $t('chat.title') }}</h1>
        <p class="page-subtitle">{{ auth.isChef ? $t('chat.subtitleChef') : $t('chat.subtitleCustomer') }}</p>
      </div>
      <button v-if="!isAdmin" class="btn-ghost shrink-0" @click="contactSupport">🛟 {{ $t('chat.contactSupport') }}</button>
    </div>

    <p v-if="error" class="rounded-lg bg-red-50 px-3 py-2 text-sm text-red-700">{{ error }}</p>
    <div v-if="loading" class="skeleton h-64"></div>

    <div v-else-if="!conversations.length" class="empty-state">
      <span class="empty-state-emoji">💬</span>
      <p class="font-medium text-gray-600">{{ $t('chat.empty') }}</p>
      <p class="text-sm">
        {{ auth.isChef ? $t('chat.emptyHintChef') : $t('chat.emptyHintCustomer') }}
      </p>
    </div>

    <div v-else class="grid gap-4 md:grid-cols-3">
      <!-- Thread list -->
      <div class="card space-y-1 p-2 md:col-span-1">
        <button
          v-for="c in conversations"
          :key="c.id"
          class="flex w-full items-center justify-between gap-2 rounded-lg px-3 py-2 text-left text-sm transition"
          :class="active?.id === c.id ? 'bg-brand-50 font-semibold text-brand-700' : 'hover:bg-gray-50'"
          @click="open(c)"
        >
          <span class="min-w-0 truncate">{{ c._label }}</span>
          <span
            v-if="c.unread_count > 0"
            class="flex h-5 min-w-5 shrink-0 items-center justify-center rounded-full bg-red-500 px-1 text-xs font-bold text-white"
          >{{ c.unread_count }}</span>
        </button>
      </div>

      <!-- Active thread -->
      <div class="card flex h-[28rem] flex-col p-0 md:col-span-2">
        <template v-if="active">
          <div class="flex items-center justify-between border-b border-gray-100 px-4 py-2.5">
            <span class="font-semibold">{{ active._label }}</span>
            <span class="badge" :class="live ? 'bg-green-100 text-green-700' : 'bg-gray-100 text-gray-500'">
              {{ live ? $t('chat.live') : $t('chat.offline') }}
            </span>
          </div>

          <div ref="listEl" class="grow space-y-2 overflow-y-auto px-4 py-3">
            <p v-if="!messages.length" class="pt-10 text-center text-sm text-gray-400">
              {{ $t('chat.sayHello') }}
            </p>
            <div v-for="m in messages" :key="m.id" class="flex flex-col" :class="mine(m) ? 'items-end' : 'items-start'">
              <div
                class="max-w-[75%] rounded-2xl px-3 py-1.5 text-sm"
                :class="mine(m) ? 'rounded-br-sm bg-brand-600 text-white' : 'rounded-bl-sm bg-gray-100 text-gray-800'"
              >
                <p class="whitespace-pre-wrap break-words">{{ m.body }}</p>
                <p class="mt-0.5 text-right text-[10px] opacity-60">{{ when(m.created_at) }}</p>
              </div>
              <span v-if="m.id === lastSeenMineId" class="mt-0.5 pr-1 text-[10px] text-gray-400">
                {{ $t('chat.seen') }}
              </span>
            </div>
          </div>

          <form class="flex gap-2 border-t border-gray-100 p-3" @submit.prevent="send">
            <input v-model="draft" class="input" :placeholder="$t('chat.typePlaceholder')" />
            <button class="btn-primary shrink-0" :disabled="!draft.trim()">{{ $t('chat.send') }}</button>
          </form>
        </template>
        <div v-else class="flex grow items-center justify-center text-sm text-gray-400">
          {{ $t('chat.pick') }}
        </div>
      </div>
    </div>
  </div>
</template>
