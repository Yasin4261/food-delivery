import { fileURLToPath, URL } from 'node:url'
import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

// In dev, proxy API + WebSocket calls to the Go backend so the SPA and API
// share an origin (no CORS needed locally). Override the target with
// VITE_API_TARGET if the backend runs elsewhere.
const apiTarget = process.env.VITE_API_TARGET || 'http://localhost:8080'

export default defineConfig({
  plugins: [vue()],
  resolve: {
    alias: { '@': fileURLToPath(new URL('./src', import.meta.url)) },
  },
  server: {
    port: 5173,
    proxy: {
      '/api': { target: apiTarget, changeOrigin: true, ws: true },
      '/health': { target: apiTarget, changeOrigin: true },
      '/uploads': { target: apiTarget, changeOrigin: true },
    },
  },
  test: {
    environment: 'jsdom',
    // Unit tests only — Playwright owns e2e/*.spec.js.
    include: ['src/**/*.test.js'],
  },
})
