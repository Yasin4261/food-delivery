import { defineConfig } from '@playwright/test'

// E2E smoke tests drive the real SPA (Vite dev server) against the real API +
// Postgres. Locally: have the backend running (docker compose up) — the dev
// server is started (or reused) automatically. In CI the workflow boots the
// API against a Postgres service first.
export default defineConfig({
  testDir: './e2e',
  timeout: 60_000,
  retries: process.env.CI ? 1 : 0,
  reporter: process.env.CI ? [['list'], ['html', { open: 'never' }]] : 'list',
  use: {
    baseURL: 'http://localhost:5173',
    trace: 'retain-on-failure',
  },
  webServer: {
    command: 'npm run dev',
    url: 'http://localhost:5173',
    reuseExistingServer: true,
    timeout: 60_000,
  },
  projects: [{ name: 'chromium', use: { browserName: 'chromium' } }],
})
