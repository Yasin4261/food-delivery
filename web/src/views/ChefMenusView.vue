<script setup>
import { onMounted, ref } from 'vue'
import { RouterLink } from 'vue-router'
import { api, page, ApiError } from '@/api/client'

// Each entry: { menu, items, form } — form is the inline "add dish" state.
const entries = ref([])
const chef = ref(null)
const needsProfile = ref(false)
const loading = ref(true)
const error = ref('')
const newMenuName = ref('')
const creatingMenu = ref(false)

function blankItemForm() {
  return { name: '', price: '', description: '', available_quantity: '', is_unlimited: false, saving: false }
}

async function loadItems(entry) {
  entry.items = page(await api.get(`/menus/${entry.menu.id}/items`)).items
}

async function load() {
  loading.value = true
  error.value = ''
  try {
    try {
      chef.value = await api.get('/chefs/me')
    } catch (e) {
      if (e instanceof ApiError && e.status === 404) {
        needsProfile.value = true
        return
      }
      throw e
    }
    const menus = page(await api.get(`/chefs/${chef.value.id}/menus?limit=100`)).items
    const next = menus.map((menu) => ({ menu, items: [], form: blankItemForm() }))
    await Promise.all(next.map(loadItems))
    entries.value = next
  } catch (e) {
    error.value = e.message
  } finally {
    loading.value = false
  }
}

async function createMenu() {
  if (!newMenuName.value.trim()) return
  creatingMenu.value = true
  error.value = ''
  try {
    await api.post('/menus', { name: newMenuName.value })
    newMenuName.value = ''
    await load()
  } catch (e) {
    error.value = e.message
  } finally {
    creatingMenu.value = false
  }
}

async function deleteMenu(entry) {
  error.value = ''
  try {
    await api.del(`/menus/${entry.menu.id}`)
    await load()
  } catch (e) {
    error.value = e.message
  }
}

async function addItem(entry) {
  const f = entry.form
  f.saving = true
  error.value = ''
  try {
    const payload = {
      menu_id: entry.menu.id,
      name: f.name,
      description: f.description,
      price: Number(f.price),
      is_unlimited: f.is_unlimited,
    }
    if (!f.is_unlimited && f.available_quantity !== '') {
      payload.available_quantity = Number(f.available_quantity)
    }
    await api.post('/menu-items', payload)
    entry.form = blankItemForm()
    await loadItems(entry)
  } catch (e) {
    error.value = e.message
    f.saving = false
  }
}

async function deleteItem(entry, item) {
  error.value = ''
  try {
    await api.del(`/menu-items/${item.id}`)
    await loadItems(entry)
  } catch (e) {
    error.value = e.message
  }
}

onMounted(load)
</script>

<template>
  <div class="space-y-6">
    <div class="flex items-center justify-between">
      <h1 class="text-2xl font-bold">My menus</h1>
      <RouterLink to="/chef" class="btn-ghost">{{ $t('menus.back') }}</RouterLink>
    </div>

    <p v-if="loading" class="text-gray-500">Loading…</p>

    <div v-else-if="needsProfile" class="card">
      <p class="text-gray-600">
        {{ $t('menus.needProfile') }}
        <RouterLink to="/chef" class="text-brand-600 hover:underline">{{ $t('menus.needProfileLink') }}</RouterLink>.
      </p>
    </div>

    <template v-else>
      <p v-if="error" class="text-sm text-red-600">{{ error }}</p>

      <form class="card flex items-end gap-3" @submit.prevent="createMenu">
        <div class="grow">
          <label class="label">{{ $t('menus.newMenu') }}</label>
          <input v-model="newMenuName" class="input" :placeholder="$t('menus.newMenuPlaceholder')" required />
        </div>
        <button class="btn-primary" :disabled="creatingMenu">{{ $t('menus.createMenu') }}</button>
      </form>

      <p v-if="!entries.length" class="text-gray-500">{{ $t('menus.empty') }}</p>

      <div v-for="entry in entries" :key="entry.menu.id" class="card space-y-3">
        <div class="flex items-center justify-between">
          <h2 class="font-semibold">{{ entry.menu.name }}</h2>
          <button class="text-sm text-red-600 hover:underline" @click="deleteMenu(entry)">{{ $t('menus.deleteMenu') }}</button>
        </div>

        <p v-if="!entry.items.length" class="text-sm text-gray-500">{{ $t('menus.noDishesYet') }}</p>
        <div v-for="item in entry.items" :key="item.id" class="flex items-center justify-between border-t border-gray-100 pt-2 text-sm">
          <div>
            <span class="font-medium">{{ item.name }}</span>
            <span class="ml-2 text-gray-500">${{ item.price?.toFixed(2) }}</span>
            <span class="ml-2 text-gray-400">
              {{ item.is_unlimited ? $t('menus.unlimited') : $t('menus.stockN', { n: item.available_quantity ?? 0 }) }}
            </span>
          </div>
          <button class="text-red-600 hover:underline" @click="deleteItem(entry, item)">{{ $t('menus.remove') }}</button>
        </div>

        <form class="grid grid-cols-2 items-end gap-2 border-t border-gray-100 pt-3 sm:grid-cols-6" @submit.prevent="addItem(entry)">
          <div class="col-span-2">
            <label class="label">{{ $t('menus.dishName') }}</label>
            <input v-model="entry.form.name" class="input" required :placeholder="$t('menus.dishPlaceholder')" />
          </div>
          <div>
            <label class="label">{{ $t('menus.price') }}</label>
            <input v-model="entry.form.price" class="input" required type="number" step="0.01" min="0.01" />
          </div>
          <div>
            <label class="label">{{ $t('menus.stock') }}</label>
            <input
              v-model="entry.form.available_quantity"
              class="input"
              type="number"
              min="0"
              :disabled="entry.form.is_unlimited"
            />
          </div>
          <label class="flex items-center gap-1 pb-2 text-sm text-gray-600">
            <input v-model="entry.form.is_unlimited" type="checkbox" /> {{ $t('menus.unlimited') }}
          </label>
          <button class="btn-primary" :disabled="entry.form.saving">{{ $t('menus.addDish') }}</button>
        </form>
      </div>
    </template>
  </div>
</template>
