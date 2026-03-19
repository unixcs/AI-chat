import http from './http'

export const registerUser = (payload) => {
  return http.post('/auth/register', payload)
}

export const loginUser = (payload) => {
  return http.post('/auth/login', payload)
}

export const loginAdmin = (payload) => {
  return http.post('/admin/auth/login', payload)
}

export const getUserProfile = () => {
  return http.get('/user/profile', { headers: { needUserAuth: true } })
}

export const updateProfile = (payload) => {
  return http.put('/user/profile', payload, { headers: { needUserAuth: true } })
}

export const changePassword = (payload) => {
  return http.put('/user/password', payload, { headers: { needUserAuth: true } })
}
