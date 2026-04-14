const test = require('node:test')
const assert = require('node:assert/strict')

const {
  normalizeAdminMemberExpireAtInput
} = require('../admin-member-expire')

test('normalizeAdminMemberExpireAtInput converts datetime-local strings with seconds into ISO strings', () => {
  const result = normalizeAdminMemberExpireAtInput('2026-04-14T21:30:45')

  assert.match(result, /^2026-04-14T\d{2}:30:45\.000Z$/)
})

test('normalizeAdminMemberExpireAtInput returns null for empty values', () => {
  assert.equal(normalizeAdminMemberExpireAtInput(''), null)
  assert.equal(normalizeAdminMemberExpireAtInput(null), null)
})

test('normalizeAdminMemberExpireAtInput rejects invalid datetime values', () => {
  assert.throws(
    () => normalizeAdminMemberExpireAtInput('not-a-date'),
    /会员到期时间格式不正确/
  )
})
