import http from './http'

export const redeemCode = (code) => {
  return http.post('/user/redeem', { code }, { headers: { needUserAuth: true } })
}
