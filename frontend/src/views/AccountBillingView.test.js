import { flushPromises, mount, RouterLinkStub } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import AccountBillingView from './AccountBillingView.vue'
import i18n from '../i18n'

const apiMocks = vi.hoisted(() => ({
  fetchSubscription: vi.fn(),
}))

vi.mock('../api/client', () => apiMocks)

const mountView = async () => {
  const wrapper = mount(AccountBillingView, {
    global: {
      plugins: [i18n],
      stubs: {
        RouterLink: RouterLinkStub,
      },
    },
  })
  await flushPromises()
  return wrapper
}

beforeEach(() => {
  i18n.global.locale.value = 'en'
  apiMocks.fetchSubscription.mockReset().mockResolvedValue({
    plan: {
      id: 'free',
      name: 'Free',
      status: 'active',
    },
    features: {
      advanced_insights: false,
      history_days_limit: 30,
    },
    limits: {
      history_days_limit: 30,
    },
  })
})

describe('AccountBillingView', () => {
  it('renders current plan from backend', async () => {
    const wrapper = await mountView()
    expect(apiMocks.fetchSubscription).toHaveBeenCalledTimes(1)
    expect(wrapper.text()).toContain('Free')
    expect(wrapper.text()).toContain('Current plan')
  })
})
