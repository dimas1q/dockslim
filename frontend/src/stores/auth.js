import { reactive } from 'vue'
import { fetchMe, logoutUser } from '../api/client'

const state = reactive({
  user: null,
  loading: false,
  error: null,
  initialized: false,
})

export const useAuth = () => state

export const loadCurrentUser = async () => {
  state.loading = true
  state.error = null
  try {
    const user = await fetchMe()
    state.user = user
    state.initialized = true
    return user
  } catch (error) {
    state.user = null
    state.error = error.status === 401 ? null : error.message
    state.initialized = true
    return null
  } finally {
    state.loading = false
  }
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
