import { reactive } from 'vue'
import { clearToken, fetchMe, getToken, setToken } from '../api/client'

const state = reactive({
  user: null,
  loading: false,
  error: null,
})

export const useAuth = () => state

export const hasToken = () => Boolean(getToken())

export const setAuthToken = (token) => {
  setToken(token)
}

export const loadCurrentUser = async () => {
  if (!hasToken()) {
    state.user = null
    return null
  }

  state.loading = true
  state.error = null
  try {
    const user = await fetchMe()
    state.user = user
    return user
  } catch (error) {
    state.user = null
    state.error = error.message
    clearToken()
    return null
  } finally {
    state.loading = false
  }
}

export const logout = () => {
  clearToken()
  state.user = null
}
