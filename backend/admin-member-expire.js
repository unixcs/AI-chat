const dayjs = require('dayjs')

const normalizeAdminMemberExpireAtInput = (value) => {
  if (value === null || value === undefined) {
    return null
  }

  const text = String(value).trim()
  if (!text) {
    return null
  }

  const parsed = dayjs(text)
  if (!parsed.isValid()) {
    throw new Error('会员到期时间格式不正确')
  }

  return parsed.toDate().toISOString()
}

module.exports = {
  normalizeAdminMemberExpireAtInput
}
