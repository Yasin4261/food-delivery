<script setup>
import { onMounted, ref } from 'vue'
import { formatMoney as money } from '@/lib/money'
import { RouterLink } from 'vue-router'
import DietaryBadges from '@/components/DietaryBadges.vue'
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
  return {
    name: '',
    price: '',
    description: '',
    available_quantity: '',
    is_unlimited: false,
    is_vegetarian: false,
    is_vegan: false,
    is_gluten_free: false,
    is_halal: false,
    saving: false,
  }
}

// Dietary flag keys shared by the editor + badges.
const dietary = [
  { key: 'is_vegetarian', label: 'vegetarian' },
  { key: 'is_vegan', label: 'vegan' },
  { key: 'is_gluten_free', label: 'glutenFree' },
  { key: 'is_halal', label: 'halal' },
]

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
      is_vegetarian: f.is_vegetarian,
      is_vegan: f.is_vegan,
      is_gluten_free: f.is_gluten_free,
      is_halal: f.is_halal,
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

// Gallery helpers (#93): item.images is a JSON array string.
function gallery(item) {
  if (!item.images) return []
  try {
    return JSON.parse(item.images)
  } catch {
    return []
  }
}
const uploadingGallery = ref(0)
async function addGalleryPhoto(item, event) {
  const file = event.target.files?.[0]
  event.target.value = ''
  if (!file) return
  uploadingGallery.value = item.id
  error.value = ''
  try {
    const out = await api.upload(`/menu-items/${item.id}/images`, file)
    item.images = JSON.stringify(out.images)
  } catch (e) {
    error.value = e.message
  } finally {
    uploadingGallery.value = 0
  }
}
async function removeGalleryPhoto(item, url) {
  error.value = ''
  try {
    const out = await api.del(`/menu-items/${item.id}/images?url=${encodeURIComponent(url)}`)
    item.images = JSON.stringify(out.images)
  } catch (e) {
    error.value = e.message
  }
}

// Dish photo upload (JPEG/PNG, max 5 MB — the API validates and re-encodes).
const uploadingItem = ref(0)
async function uploadPhoto(item, event) {
  const file = event.target.files?.[0]
  event.target.value = '' // allow re-picking the same file
  if (!file) return
  uploadingItem.value = item.id
  error.value = ''
  try {
    const out = await api.upload(`/menu-items/${item.id}/image`, file)
    item.image_url = out.image_url
  } catch (e) {
    error.value = e.message
  } finally {
    uploadingItem.value = 0
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
        <div v-for="item in entry.items" :key="item.id" class="space-y-2 border-t border-gray-100 pt-2 text-sm">
          <div class="flex items-center justify-between gap-2">
            <div class="flex min-w-0 items-center gap-2">
              <img v-if="item.image_url" :src="item.image_url" :alt="item.name" class="h-9 w-9 shrink-0 rounded-lg object-cover" />
              <div class="min-w-0">
                <span class="font-medium">{{ item.name }}</span>
                <span class="ml-2 text-gray-500">{{ money(item.price) }}</span>
                <span class="ml-2 text-gray-400">
                  {{ item.is_unlimited ? $t('menus.unlimited') : $t('menus.stockN', { n: item.available_quantity ?? 0 }) }}
                </span>
                <DietaryBadges :item="item" class="ml-1 inline-flex" />
              </div>
            </div>
            <div class="flex shrink-0 items-center gap-2">
              <label class="cursor-pointer text-brand-600 hover:underline">
                {{ uploadingItem === item.id ? $t('menus.uploading') : item.image_url ? $t('menus.changePhoto') : $t('menus.addPhoto') }}
                <input type="file" accept="image/jpeg,image/png" class="hidden" :disabled="uploadingItem === item.id" @change="uploadPhoto(item, $event)" />
              </label>
              <button class="text-red-600 hover:underline" @click="deleteItem(entry, item)">{{ $t('menus.remove') }}</button>
            </div>
          </div>
          <!-- Gallery: extra photos beyond the cover (#93). -->
          <div class="flex flex-wrap items-center gap-2 pl-11">
            <div v-for="url in gallery(item)" :key="url" class="group relative">
              <img :src="url" class="h-12 w-12 rounded-lg object-cover" alt="" />
              <button
                class="absolute -right-1 -top-1 hidden h-5 w-5 items-center justify-center rounded-full bg-red-600 text-xs text-white group-hover:flex"
                :title="$t('menus.removePhoto')"
                @click="removeGalleryPhoto(item, url)"
              >×</button>
            </div>
            <label v-if="gallery(item).length < 5" class="flex h-12 w-12 cursor-pointer items-center justify-center rounded-lg border border-dashed border-gray-300 text-lg text-gray-400 hover:border-brand-300">
              <span v-if="uploadingGallery === item.id" class="text-xs">…</span>
              <span v-else>＋</span>
              <input type="file" accept="image/jpeg,image/png" class="hidden" :disabled="uploadingGallery === item.id" @change="addGalleryPhoto(item, $event)" />
            </label>
          </div>
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
          <div class="col-span-2 flex flex-wrap gap-3 sm:col-span-6">
            <label v-for="d in dietary" :key="d.key" class="flex items-center gap-1 text-sm text-gray-600">
              <input v-model="entry.form[d.key]" type="checkbox" /> {{ $t(`dietary.${d.label}`) }}
            </label>
          </div>
        </form>
      </div>
    </template>
  </div>
</template>
