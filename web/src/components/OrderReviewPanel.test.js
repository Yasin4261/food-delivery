import { describe, expect, it, vi, beforeEach } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'

const get = vi.fn()
const post = vi.fn()
vi.mock('@/api/client', () => ({ api: { get: (...a) => get(...a), post: (...a) => post(...a) } }))

import OrderReviewPanel from './OrderReviewPanel.vue'

// A delivered single-chef order, exactly as OrdersView passes it.
const order = {
  id: 49,
  status: 'delivered',
  items: [{ id: 1, chef_id: 43, menu_item_id: 34, item_name: 'Soup' }],
  sub_orders: [{ id: 1, chef_id: 43, chef_name: 'Kitchen', status: 'delivered' }],
}

function mountPanel() {
  return mount(OrderReviewPanel, {
    props: { order },
    global: { stubs: { StarRating: true }, mocks: { $t: (k) => k } },
  })
}

describe('OrderReviewPanel (#135)', () => {
  beforeEach(() => {
    get.mockReset()
    post.mockReset()
  })

  it('shows an already-given rating read-only, not an empty form', async () => {
    get.mockResolvedValue([{ id: 15, order_id: 49, chef_id: 43, rating: 5, comment: 'great' }])
    const w = mountPanel()
    await flushPromises()
    // The chef target is "done".
    expect(w.text()).toContain('review.yourRating')
  })

  it('surfaces a history-load failure instead of silently showing empty forms', async () => {
    get.mockRejectedValue(new Error('network'))
    const w = mountPanel()
    await flushPromises()

    // The failure is shown with a retry — NOT a set of empty rating forms
    // (which would look like the rating vanished).
    expect(w.text()).toContain('review.loadFailed')
    expect(w.find('button').text()).toContain('review.retry')
    // No submit forms are rendered while the load is errored.
    expect(w.findAll('button.btn-ghost')).toHaveLength(0)

    // Retrying re-fetches and recovers.
    get.mockResolvedValue([{ id: 15, order_id: 49, chef_id: 43, rating: 4 }])
    await w.find('button').trigger('click')
    await flushPromises()
    expect(w.text()).not.toContain('review.loadFailed')
    expect(w.text()).toContain('review.yourRating')
  })

  it('self-heals: submitting an already-reviewed target (409) flips it to done', async () => {
    // History loads empty (e.g. it had failed on a prior open), so the targets
    // render forms even though the server already has the review. Use the real
    // StarRating so a rating can be set by clicking a star.
    get.mockResolvedValue([])
    const w = mount(OrderReviewPanel, { props: { order }, global: { mocks: { $t: (k) => k } } })
    await flushPromises()

    // Before: a submit form is shown (no "your rating" yet).
    expect(w.text()).not.toContain('review.yourRating')

    // The server rejects the submit with 409 (the review already exists).
    const err = new Error('exists')
    err.status = 409
    post.mockRejectedValue(err)

    // Set a rating on the first target (click its 5th star) and submit.
    const firstRow = w.findAll('.rounded-lg.border')[0]
    const stars = firstRow.findAll('button:not(.btn-ghost)')
    await stars[4].trigger('click') // 5 stars
    await firstRow.find('button.btn-ghost').trigger('click') // Submit
    await flushPromises()

    // The target flipped to the read-only "your rating" state — no dead-end.
    expect(w.text()).toContain('review.yourRating')
  })
})
