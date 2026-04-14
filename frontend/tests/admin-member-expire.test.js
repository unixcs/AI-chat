import test from 'node:test'
import assert from 'node:assert/strict'

import {
  getMemberExpireDisplayMeta,
  formatMemberExpireAtForInput,
  formatNowForMemberExpireInput
} from '../src/utils/admin-member-expire.js'

test('formatMemberExpireAtForInput converts ISO value into datetime-local with seconds', () => {
  const result = formatMemberExpireAtForInput('2026-04-14T13:30:45.000Z')

  assert.match(result, /^2026-04-14T\d{2}:30:45$/)
})

test('formatMemberExpireAtForInput returns empty string for blank values', () => {
  assert.equal(formatMemberExpireAtForInput(''), '')
  assert.equal(formatMemberExpireAtForInput(null), '')
})

test('formatNowForMemberExpireInput returns a datetime-local value with seconds', () => {
  const result = formatNowForMemberExpireInput()

  assert.match(result, /^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}$/)
})

test('getMemberExpireDisplayMeta marks blank values as unset', () => {
  const result = getMemberExpireDisplayMeta(null)

  assert.equal(result.label, '未设置')
  assert.equal(result.tagClass, 'tagWarn')
  assert.equal(result.formattedTime, '-')
})

test('getMemberExpireDisplayMeta marks future values as active', () => {
  const result = getMemberExpireDisplayMeta('2099-01-01T00:00:00.000Z')

  assert.equal(result.label, '有效')
  assert.equal(result.tagClass, 'tagSuccess')
  assert.match(result.formattedTime, /^2099-01-01 /)
})

test('getMemberExpireDisplayMeta marks past values as expired', () => {
  const result = getMemberExpireDisplayMeta('2000-01-01T00:00:00.000Z')

  assert.equal(result.label, '已过期')
  assert.equal(result.tagClass, 'tagDanger')
  assert.match(result.formattedTime, /^2000-01-01 /)
})
