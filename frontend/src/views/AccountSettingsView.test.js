import { flushPromises, mount, RouterLinkStub } from '@vue/test-utils'
import { describe, expect, it, beforeEach, vi } from 'vitest'
import AccountSettingsView from './AccountSettingsView.vue'
import i18n from '../i18n'

const mockAccount = { id: 'user-1', login: 'demo', email: 'demo@example.com' }
const apiMocks = vi.hoisted(() => ({
  fetchAccount: vi.fn(),
  updateAccount: vi.fn(),
  listApiTokens: vi.fn(),
  createApiToken: vi.fn(),
  revokeApiToken: vi.fn(),
}))

vi.mock('../api/client', () => apiMocks)

vi.mock('../stores/auth', () => {
  const state = { user: null, loading: false, error: null, initialized: true }
  return {
    useAuth: () => state,
  }
})

const mountView = async () => {
  const wrapper = mount(AccountSettingsView, {
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
  apiMocks.fetchAccount.mockReset().mockResolvedValue(mockAccount)
  apiMocks.updateAccount.mockReset().mockResolvedValue(mockAccount)
  apiMocks.listApiTokens.mockReset().mockResolvedValue([])
  apiMocks.createApiToken.mockReset()
  apiMocks.revokeApiToken.mockReset()
})

describe('AccountSettingsView', () => {
  it('loads and displays profile info', async () => {
    const wrapper = await mountView()
    const inputs = wrapper.findAll('input')
    expect(inputs[0].element.value).toBe(mockAccount.login)
    expect(inputs[1].element.value).toBe(mockAccount.email)
  })

  it('creates and revokes a token', async () => {
    const created = {
      id: 'tok-2',
      name: 'cli',
      token: 'ds_api_token',
      created_at: new Date().toISOString(),
      last_used_at: null,
      revoked_at: null,
    }
    apiMocks.listApiTokens.mockResolvedValue([
      { id: 'tok-1', name: 'old', created_at: new Date().toISOString(), last_used_at: null, revoked_at: null },
    ])
    apiMocks.createApiToken.mockResolvedValue(created)
    apiMocks.revokeApiToken.mockResolvedValue()

    const wrapper = await mountView()

    await wrapper.find('[data-testid="token-name-input"]').setValue('cli')
    await wrapper.find('[data-testid="create-token-button"]').trigger('click')
    await flushPromises()

    expect(apiMocks.createApiToken).toHaveBeenCalledWith({ name: 'cli' })
    expect(wrapper.text()).toContain(created.token)

    await wrapper.find('[data-testid="revoke-token-tok-1"]').trigger('click')
    await flushPromises()

    expect(apiMocks.revokeApiToken).toHaveBeenCalledWith('tok-1')
    expect(wrapper.text()).toContain('Revoked')
  })
})
