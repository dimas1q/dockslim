const BASE_URL = import.meta.env.VITE_API_BASE_URL || ''

const buildURL = (path) => {
  if (path.startsWith('http://') || path.startsWith('https://')) {
    return path
  }
  return `${BASE_URL.replace(/\/$/, '')}${path}`
}

export const apiRequest = async (path, options = {}) => {
  const headers = new Headers(options.headers || {})
  if (!headers.has('Content-Type') && options.body) {
    headers.set('Content-Type', 'application/json')
  }

  const response = await fetch(buildURL(path), {
    ...options,
    headers,
    credentials: 'include',
  })

  if (response.status === 204) {
    return null
  }

  const contentType = response.headers.get('Content-Type') || ''
  const isJSON = contentType.includes('application/json')
  const payload = isJSON ? await response.json() : await response.text()

  if (!response.ok) {
    const message = payload?.error || 'Request failed'
    const error = new Error(message)
    error.status = response.status
    throw error
  }

  return payload
}

export const registerUser = (payload) =>
  apiRequest('/api/v1/auth/register', {
    method: 'POST',
    body: JSON.stringify(payload),
  })

export const loginUser = (payload) =>
  apiRequest('/api/v1/auth/login', {
    method: 'POST',
    body: JSON.stringify(payload),
  })

export const fetchMe = () => apiRequest('/api/v1/me')

export const logoutUser = () =>
  apiRequest('/api/v1/auth/logout', {
    method: 'POST',
  })

export const listProjects = () => apiRequest('/api/v1/projects')

export const createProject = (payload) =>
  apiRequest('/api/v1/projects', {
    method: 'POST',
    body: JSON.stringify(payload),
  })

export const getProject = (projectId) => apiRequest(`/api/v1/projects/${projectId}`)

export const deleteProject = (projectId) =>
  apiRequest(`/api/v1/projects/${projectId}`, {
    method: 'DELETE',
  })
