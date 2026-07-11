import { createApp } from 'vue'
import { createPinia } from 'pinia'

import App from './App.vue'
import { router } from './router'
import { i18n } from './i18n'
import { useAuthStore } from './stores/auth'
import { setUnauthorizedHandler } from './api/client'
import './assets/main.css'

const app = createApp(App)
app.use(createPinia())
app.use(router)
app.use(i18n)

// On a 401 from the API, clear the session and bounce to login.
const auth = useAuthStore()
setUnauthorizedHandler(() => {
  auth.clear()
  if (router.currentRoute.value.name !== 'login') {
    router.push({ name: 'login' })
  }
})

app.mount('#app')
