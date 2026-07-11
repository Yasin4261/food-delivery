import { test, expect } from '@playwright/test'

// Golden-path smoke: a chef sets up shop and a customer orders from them,
// through the real SPA + API + Postgres. Two isolated browser contexts play
// the two roles; unique names per run keep the test re-runnable on a dirty
// dev database.
const run = Date.now()
const kitchen = `E2E Kitchen ${run}`
const dish = `E2E Dish ${run}`
const chefEmail = `chef${run}@e2e.test`
const customerEmail = `cust${run}@e2e.test`
const password = 'secret123'

async function register(page, { username, email, chef = false }) {
  await page.goto('/register')
  await page.getByPlaceholder('yasin').fill(username)
  await page.getByPlaceholder('you@example.com').fill(email)
  await page.getByPlaceholder('min. 6 characters').fill(password)
  if (chef) await page.getByRole('button', { name: /Cook & sell/ }).click()
  await page.getByRole('button', { name: 'Create account' }).click()
}

test('chef sets up shop, customer orders, chef delivers, cash settles', async ({ browser }) => {
  // ---------- Chef: register -> onboarding -> menu + dish -> online ----------
  const chefCtx = await browser.newContext()
  const chef = await chefCtx.newPage()

  await register(chef, { username: `chef${run}`, email: chefEmail, chef: true })

  // No profile yet -> the dashboard shows onboarding.
  await expect(chef.getByRole('heading', { name: 'Set up your kitchen' })).toBeVisible()
  await chef.getByPlaceholder("Yasin's Kitchen").fill(kitchen)
  await chef.getByPlaceholder('123 Main St').fill('42 E2E Street')
  await chef.getByRole('button', { name: 'Create my kitchen' }).click()

  // Dashboard appears with the kitchen name; go online.
  await expect(chef.getByRole('heading', { name: kitchen })).toBeVisible()
  await chef.getByRole('button', { name: 'Go online' }).click()
  await expect(chef.getByText('● live').or(chef.getByText('online')).first()).toBeVisible()

  // Create a menu and an unlimited dish.
  await chef.getByRole('link', { name: 'My menus' }).first().click()
  await chef.getByPlaceholder('Dinner menu').fill(`Menu ${run}`)
  await chef.getByRole('button', { name: 'Create menu' }).click()
  await expect(chef.getByRole('heading', { name: `Menu ${run}` })).toBeVisible()

  await chef.getByPlaceholder('Lentil soup').fill(dish)
  await chef.locator('input[step="0.01"]').fill('9.5')
  await chef.getByRole('checkbox').check() // unlimited stock
  await chef.getByRole('button', { name: 'Add dish' }).click()
  await expect(chef.getByText(dish)).toBeVisible()

  // ---------- Customer: register -> browse -> cart -> order ----------
  const custCtx = await browser.newContext()
  const cust = await custCtx.newPage()

  await register(cust, { username: `cust${run}`, email: customerEmail })

  // Browse shows the new kitchen; open it and add the dish.
  await expect(cust.getByRole('heading', { name: 'Discover home chefs 🧑‍🍳' })).toBeVisible()
  await cust.getByText(kitchen).first().click()
  await expect(cust.getByRole('heading', { name: kitchen })).toBeVisible()
  await cust.getByRole('button', { name: '+ Add' }).first().click()

  // Checkout with cash.
  await cust.getByRole('link', { name: /Cart/ }).click()
  await expect(cust.getByText(dish)).toBeVisible()
  await cust
    .locator('form')
    .filter({ hasText: 'Delivery address' })
    .locator('input')
    .first()
    .fill('7 Customer Lane')
  await cust.getByRole('button', { name: /Place order/ }).click()

  // Order history shows the pending cash order.
  await expect(cust.getByRole('heading', { name: 'My orders' })).toBeVisible()
  await expect(cust.getByText('pending', { exact: true })).toBeVisible()
  await expect(cust.getByText('💵 pending')).toBeVisible()

  // ---------- Chef: accept -> ... -> delivered ----------
  await chef.goto('/chef')
  await expect(chef.getByText(`1× ${dish}`)).toBeVisible()
  for (const action of ['Accept', 'Start preparing', 'Mark ready', 'Out for delivery', 'Mark delivered']) {
    await chef.getByRole('button', { name: action }).click()
  }
  await expect(chef.getByText('delivered', { exact: true })).toBeVisible()

  // ---------- Customer: delivered + paid (cash settles on delivery) ----------
  await cust.reload()
  await expect(cust.getByText('delivered', { exact: true })).toBeVisible()
  await expect(cust.getByText('💵 paid')).toBeVisible()

  await chefCtx.close()
  await custCtx.close()
})
