import { defineConfig } from '@playwright/test'

// E2E smoke tests drive the real SPA against the real API + Postgres. Locally
// and in CI they hit the Vite dev server (started/reused automatically). Set
// E2E_BASE_URL to run the same smoke against a deployed environment — e.g.
// staging (#75): `E2E_BASE_URL=http://localhost:8090 npm run test:e2e`. When
// it is set, Playwright does not start a dev server.
const baseURL = process.env.E2E_BASE_URL || 'http://localhost:5173'
const external = Boolean(process.env.E2E_BASE_URL)

export default defineConfig({
  testDir: './e2e',
  timeout: 60_000,
  retries: process.env.CI ? 1 : 0,
  reporter: process.env.CI ? [['list'], ['html', { open: 'never' }]] : 'list',
  use: {
    baseURL,
    // Pin the browser locale: the app language-detects (TR browsers get
    // Turkish), and the spec selects by the English catalogue.
    locale: 'en-US',
    trace: 'retain-on-failure',
  },
  // Against an external target the SPA is already served — no dev server.
  webServer: external
    ? undefined
    : {
        command: 'npm run dev',
        url: 'http://localhost:5173',
        reuseExistingServer: true,
        timeout: 60_000,
      },
  projects: [{ name: 'chromium', use: { browserName: 'chromium' } }],
})
