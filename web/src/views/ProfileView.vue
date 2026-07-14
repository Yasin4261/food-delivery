<script setup>
import { onMounted, reactive, ref } from 'vue'
import { api, ApiError } from '@/api/client'
import { useAuthStore } from '@/stores/auth'

const auth = useAuthStore()

// --- account (contact + default location + notification preference) ---
const account = reactive({
  phone_number: '',
  address: '',
  city: '',
  latitude: '',
  longitude: '',
  email_notifications: true,
})
const accountMsg = ref('')
const accountErr = ref('')
const savingAccount = ref(false)

// --- password ---
const pw = reactive({ current: '', next: '' })
const pwMsg = ref('')
const pwErr = ref('')
const savingPw = ref(false)

// --- kitchen (chef only) ---
const kitchen = ref(null) // null = not a chef / no profile yet
const kitchenMsg = ref('')
const kitchenErr = ref('')
const savingKitchen = ref(false)

// --- address book ---
const addresses = ref([])
const newAddress = reactive({ label: '', address: '', city: '' })
const addressErr = ref('')
const savingAddress = ref(false)

async function loadAddresses() {
  try {
    addresses.value = await api.get('/addresses')
  } catch (e) {
    addressErr.value = e.message
  }
}

async function addAddress() {
  savingAddress.value = true
  addressErr.value = ''
  try {
    await api.post('/addresses', { label: newAddress.label, address: newAddress.address, city: newAddress.city })
    newAddress.label = newAddress.address = newAddress.city = ''
    await loadAddresses()
  } catch (e) {
    addressErr.value = e.message
  } finally {
    savingAddress.value = false
  }
}

async function makeDefault(a) {
  addressErr.value = ''
  try {
    await api.put(`/addresses/${a.id}`, { label: a.label, address: a.address, city: a.city || '', is_default: true })
    await loadAddresses()
  } catch (e) {
    addressErr.value = e.message
  }
}

async function removeAddress(a) {
  addressErr.value = ''
  try {
    await api.del(`/addresses/${a.id}`)
    await loadAddresses()
  } catch (e) {
    addressErr.value = e.message
  }
}

async function load() {
  try {
    const me = await api.get('/auth/me')
    account.phone_number = me.phone_number || ''
    account.address = me.address || ''
    account.city = me.city || ''
    account.latitude = me.latitude ?? ''
    account.longitude = me.longitude ?? ''
    account.email_notifications = me.email_notifications !== false
  } catch {
    /* form just starts empty */
  }
  if (auth.isChef) {
    try {
      const chef = await api.get('/chefs/me')
      kitchenImage.value = chef.image_url || ''
      kitchen.value = {
        business_name: chef.business_name || '',
        specialty: chef.specialty || '',
        bio: chef.bio || '',
        kitchen_address: chef.kitchen_address || '',
        kitchen_city: chef.kitchen_city || '',
        delivery_radius: chef.delivery_radius || 0,
        latitude: chef.kitchen_latitude ?? '',
        longitude: chef.kitchen_longitude ?? '',
      }
    } catch (e) {
      // 404 = no profile yet; the dashboard drives onboarding, not this page.
      if (!(e instanceof ApiError && e.status === 404)) kitchenErr.value = e.message
    }
  }
}

async function saveAccount() {
  savingAccount.value = true
  accountMsg.value = accountErr.value = ''
  try {
    await api.put('/users/me', {
      phone_number: account.phone_number,
      address: account.address,
      city: account.city,
      latitude: account.latitude === '' ? null : Number(account.latitude),
      longitude: account.longitude === '' ? null : Number(account.longitude),
      email_notifications: account.email_notifications,
    })
    accountMsg.value = 'saved'
  } catch (e) {
    accountErr.value = e.message
  } finally {
    savingAccount.value = false
  }
}

async function changePassword() {
  savingPw.value = true
  pwMsg.value = pwErr.value = ''
  try {
    await api.put('/auth/password', { current_password: pw.current, new_password: pw.next })
    pwMsg.value = 'changed'
    pw.current = pw.next = ''
  } catch (e) {
    pwErr.value = e.message
  } finally {
    savingPw.value = false
  }
}

// Kitchen photo upload (JPEG/PNG, max 5 MB).
const kitchenImage = ref('')
const uploadingKitchen = ref(false)
async function uploadKitchenPhoto(event) {
  const file = event.target.files?.[0]
  event.target.value = ''
  if (!file) return
  uploadingKitchen.value = true
  kitchenErr.value = ''
  try {
    const out = await api.upload('/chefs/me/image', file)
    kitchenImage.value = out.image_url
  } catch (e) {
    kitchenErr.value = e.message
  } finally {
    uploadingKitchen.value = false
  }
}

async function saveKitchen() {
  savingKitchen.value = true
  kitchenMsg.value = kitchenErr.value = ''
  try {
    const k = kitchen.value
    await api.put('/chefs/me', {
      business_name: k.business_name,
      specialty: k.specialty,
      bio: k.bio,
      kitchen_address: k.kitchen_address,
      kitchen_city: k.kitchen_city,
      delivery_radius: Number(k.delivery_radius) || 0,
      latitude: k.latitude === '' ? null : Number(k.latitude),
      longitude: k.longitude === '' ? null : Number(k.longitude),
    })
    kitchenMsg.value = 'saved'
  } catch (e) {
    kitchenErr.value = e.message
  } finally {
    savingKitchen.value = false
  }
}

onMounted(() => {
  load()
  loadAddresses()
})
</script>

<template>
  <div class="mx-auto max-w-2xl space-y-6">
    <div>
      <h1 class="page-title">{{ $t('profile.title') }}</h1>
      <p class="page-subtitle">{{ auth.user?.username }} · {{ auth.user?.email }}</p>
    </div>

    <!-- Contact & default location -->
    <form class="card space-y-4" @submit.prevent="saveAccount">
      <h2 class="font-semibold">{{ $t('profile.account') }}</h2>
      <div class="grid gap-4 sm:grid-cols-2">
        <div>
          <label class="label">{{ $t('profile.phone') }}</label>
          <input v-model="account.phone_number" class="input" maxlength="20" placeholder="+90 5xx xxx xx xx" />
        </div>
        <div>
          <label class="label">{{ $t('profile.city') }}</label>
          <input v-model="account.city" class="input" />
        </div>
      </div>
      <div>
        <label class="label">{{ $t('profile.address') }}</label>
        <input v-model="account.address" class="input" />
      </div>
      <div class="grid gap-4 sm:grid-cols-2">
        <div>
          <label class="label">{{ $t('browse.lat') }}</label>
          <input v-model="account.latitude" class="input" placeholder="41.0082" />
        </div>
        <div>
          <label class="label">{{ $t('browse.lng') }}</label>
          <input v-model="account.longitude" class="input" placeholder="28.9784" />
        </div>
      </div>
      <label class="flex items-center gap-2 text-sm text-gray-700">
        <input v-model="account.email_notifications" type="checkbox" class="h-4 w-4 rounded border-gray-300" />
        {{ $t('profile.emailNotifications') }}
      </label>
      <p class="text-xs text-gray-400">{{ $t('profile.emailNotificationsHint') }}</p>
      <p v-if="accountErr" class="text-sm text-red-600">{{ accountErr }}</p>
      <p v-if="accountMsg" class="text-sm text-green-700">{{ $t('profile.saved') }}</p>
      <button class="btn-primary" :disabled="savingAccount">{{ $t('profile.save') }}</button>
    </form>

    <!-- Password -->
    <form class="card space-y-4" @submit.prevent="changePassword">
      <h2 class="font-semibold">{{ $t('profile.password') }}</h2>
      <div class="grid gap-4 sm:grid-cols-2">
        <div>
          <label class="label">{{ $t('profile.currentPassword') }}</label>
          <input v-model="pw.current" type="password" class="input" required autocomplete="current-password" />
        </div>
        <div>
          <label class="label">{{ $t('profile.newPassword') }}</label>
          <input v-model="pw.next" type="password" class="input" required minlength="6" autocomplete="new-password" :placeholder="$t('auth.passwordMin')" />
        </div>
      </div>
      <p v-if="pwErr" class="text-sm text-red-600">{{ pwErr }}</p>
      <p v-if="pwMsg" class="text-sm text-green-700">{{ $t('profile.passwordChanged') }}</p>
      <button class="btn-primary" :disabled="savingPw">{{ $t('profile.changePassword') }}</button>
    </form>

    <!-- Address book -->
    <div class="card space-y-4">
      <h2 class="font-semibold">{{ $t('addresses.title') }}</h2>
      <p v-if="!addresses.length" class="text-sm text-gray-500">{{ $t('addresses.empty') }}</p>
      <ul v-else class="space-y-2">
        <li
          v-for="a in addresses"
          :key="a.id"
          class="flex flex-wrap items-center justify-between gap-2 rounded-lg border border-gray-100 px-3 py-2 text-sm"
        >
          <span class="min-w-0">
            <span class="font-medium">{{ a.label }}</span>
            <span v-if="a.is_default" class="badge ml-2 bg-brand-50 text-brand-700">{{ $t('addresses.default') }}</span>
            <span class="block truncate text-gray-500">{{ a.address }}<template v-if="a.city">, {{ a.city }}</template></span>
          </span>
          <span class="flex shrink-0 gap-2">
            <button v-if="!a.is_default" class="text-brand-600 hover:underline" @click="makeDefault(a)">{{ $t('addresses.makeDefault') }}</button>
            <button class="text-red-600 hover:underline" @click="removeAddress(a)">{{ $t('addresses.remove') }}</button>
          </span>
        </li>
      </ul>
      <form class="grid gap-3 sm:grid-cols-4" @submit.prevent="addAddress">
        <input v-model="newAddress.label" class="input" :placeholder="$t('addresses.labelPlaceholder')" required maxlength="50" />
        <input v-model="newAddress.address" class="input sm:col-span-2" :placeholder="$t('addresses.addressPlaceholder')" required />
        <div class="flex gap-2">
          <input v-model="newAddress.city" class="input min-w-0" :placeholder="$t('onboarding.cityPlaceholder')" />
          <button class="btn-primary shrink-0" :disabled="savingAddress">{{ $t('addresses.add') }}</button>
        </div>
      </form>
      <p v-if="addressErr" class="text-sm text-red-600">{{ addressErr }}</p>
    </div>

    <!-- Kitchen (chefs with a profile) -->
    <form v-if="kitchen" class="card space-y-4" @submit.prevent="saveKitchen">
      <h2 class="font-semibold">{{ $t('profile.kitchen') }}</h2>
      <div class="flex items-center gap-3">
        <img v-if="kitchenImage" :src="kitchenImage" :alt="kitchen.business_name" class="h-16 w-16 rounded-xl object-cover" />
        <div v-else class="flex h-16 w-16 items-center justify-center rounded-xl bg-gray-100 text-2xl">🍲</div>
        <label class="cursor-pointer text-sm text-brand-600 hover:underline">
          {{ uploadingKitchen ? $t('menus.uploading') : kitchenImage ? $t('menus.changePhoto') : $t('profile.addKitchenPhoto') }}
          <input type="file" accept="image/jpeg,image/png" class="hidden" :disabled="uploadingKitchen" @change="uploadKitchenPhoto" />
        </label>
      </div>
      <div class="grid gap-4 sm:grid-cols-2">
        <div>
          <label class="label">{{ $t('onboarding.businessName') }}</label>
          <input v-model="kitchen.business_name" class="input" required />
        </div>
        <div>
          <label class="label">{{ $t('onboarding.specialty') }}</label>
          <input v-model="kitchen.specialty" class="input" />
        </div>
      </div>
      <div>
        <label class="label">{{ $t('profile.bio') }}</label>
        <textarea v-model="kitchen.bio" class="input" rows="2"></textarea>
      </div>
      <div class="grid gap-4 sm:grid-cols-2">
        <div>
          <label class="label">{{ $t('onboarding.address') }}</label>
          <input v-model="kitchen.kitchen_address" class="input" required />
        </div>
        <div>
          <label class="label">{{ $t('onboarding.city') }}</label>
          <input v-model="kitchen.kitchen_city" class="input" />
        </div>
      </div>
      <div class="grid gap-4 sm:grid-cols-3">
        <div>
          <label class="label">{{ $t('onboarding.radius') }}</label>
          <input v-model="kitchen.delivery_radius" type="number" min="1" class="input" />
        </div>
        <div>
          <label class="label">{{ $t('browse.lat') }}</label>
          <input v-model="kitchen.latitude" class="input" />
        </div>
        <div>
          <label class="label">{{ $t('browse.lng') }}</label>
          <input v-model="kitchen.longitude" class="input" />
        </div>
      </div>
      <p v-if="kitchenErr" class="text-sm text-red-600">{{ kitchenErr }}</p>
      <p v-if="kitchenMsg" class="text-sm text-green-700">{{ $t('profile.saved') }}</p>
      <button class="btn-primary" :disabled="savingKitchen">{{ $t('profile.save') }}</button>
    </form>
  </div>
</template>
