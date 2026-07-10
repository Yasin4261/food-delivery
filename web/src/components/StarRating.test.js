import { describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'
import StarRating from './StarRating.vue'

describe('StarRating', () => {
  it('renders five stars and highlights up to the model value', () => {
    const wrapper = mount(StarRating, { props: { modelValue: 3 } })
    const stars = wrapper.findAll('button')
    expect(stars).toHaveLength(5)
    expect(stars[2].classes()).toContain('text-amber-400') // 3rd star lit
    expect(stars[3].classes()).toContain('text-gray-300') // 4th not
  })

  it('emits update:modelValue with the clicked star', async () => {
    const wrapper = mount(StarRating, { props: { modelValue: 0 } })
    await wrapper.findAll('button')[3].trigger('click') // 4th star

    expect(wrapper.emitted('update:modelValue')).toEqual([[4]])
  })
})
