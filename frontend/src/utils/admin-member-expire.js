import dayjs from 'dayjs'

export const formatMemberExpireAtForInput = (value) => {
  if (!value) {
    return ''
  }

  const parsed = dayjs(value)
  if (!parsed.isValid()) {
    return ''
  }

  return parsed.format('YYYY-MM-DDTHH:mm:ss')
}

export const formatNowForMemberExpireInput = () => {
  return dayjs().format('YYYY-MM-DDTHH:mm:ss')
}

export const getMemberExpireDisplayMeta = (value) => {
  if (!value) {
    return {
      label: '未设置',
      tagClass: 'tagWarn',
      formattedTime: '-'
    }
  }

  const parsed = dayjs(value)
  if (!parsed.isValid()) {
    return {
      label: '未设置',
      tagClass: 'tagWarn',
      formattedTime: '-'
    }
  }

  return {
    label: parsed.isAfter(dayjs()) ? '有效' : '已过期',
    tagClass: parsed.isAfter(dayjs()) ? 'tagSuccess' : 'tagDanger',
    formattedTime: parsed.format('YYYY-MM-DD HH:mm:ss')
  }
}
