import { reactive } from 'vue'
import { fetchMe, logoutUser } from '../api/client'

const state = reactive({
  user: null,
  loading: false,
  error: null,
  initialized: false,
})

export const useAuth = () => state

const sleep = (ms) => new Promise((resolve) => setTimeout(resolve, ms))

export const loadCurrentUser = async () => {
  state.loading = true
  state.error = null
  const maxAttempts = 5
  const baseDelayMs = 400

  for (let attempt = 1; attempt <= maxAttempts; attempt += 1) {
    try {
      const user = await fetchMe()
      state.user = user
      state.initialized = true
      return user
    } catch (error) {
      const status = error?.status
      if (status === 401) {
        state.user = null
        state.error = null
        state.initialized = true
        return null
      }

      if (typeof status === 'undefined' || status === null) {
        if (attempt < maxAttempts) {
          await sleep(baseDelayMs * attempt)
          continue
        }
        state.user = null
        state.error = error.message
        state.initialized = true
        return null
      }

      state.user = null
      state.error = error.message
      state.initialized = true
      return null
    }
  }

  state.user = null
  state.initialized = true
  return null
}

export const logout = async () => {
  try {
    await logoutUser()
  } catch (error) {
    state.error = error.message
  } finally {
    state.user = null
    state.initialized = true
  }
}
