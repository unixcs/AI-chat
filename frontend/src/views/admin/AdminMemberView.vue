<script setup>
import { onMounted, ref } from 'vue'
import { getAdminMembers } from '../../api/admin'
import { getMemberExpireDisplayMeta } from '../../utils/admin-member-expire'
import Card from '@/components/ui/card/Card.vue'
import CardContent from '@/components/ui/card/CardContent.vue'
import Badge from '@/components/ui/badge/Badge.vue'
import Table from '@/components/ui/table/Table.vue'
import TableHeader from '@/components/ui/table/TableHeader.vue'
import TableBody from '@/components/ui/table/TableBody.vue'
import TableRow from '@/components/ui/table/TableRow.vue'
import TableHead from '@/components/ui/table/TableHead.vue'
import TableCell from '@/components/ui/table/TableCell.vue'
import { Crown } from 'lucide-vue-next'

const members = ref([])

onMounted(async () => {
  const { data } = await getAdminMembers()
  members.value = data.data
})

const getMemberExpireMeta = (time) => {
  return getMemberExpireDisplayMeta(time)
}

// 旧工具函数返回 tagClass（tagSuccess/tagWarn/tagDanger），映射到 Badge variant
const memberBadgeVariant = (time) => {
  const tagClass = getMemberExpireMeta(time).tagClass
  if (tagClass === 'tagSuccess') return 'default'
  if (tagClass === 'tagDanger') return 'destructive'
  return 'secondary'
}
</script>

<template>
  <section class="mx-auto max-w-5xl">
    <Card>
      <CardContent class="p-5 sm:p-6">
        <h2 class="flex items-center gap-2 text-lg font-semibold text-card-foreground">
          <Crown class="size-5 text-primary" />
          会员管理
        </h2>

        <div class="mt-4 overflow-hidden rounded-xl border border-border">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>手机号</TableHead>
                <TableHead>昵称</TableHead>
                <TableHead>会员到期</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              <TableRow v-for="item in members" :key="item.id">
                <TableCell class="font-medium">{{ item.phone }}</TableCell>
                <TableCell>{{ item.nickname }}</TableCell>
                <TableCell>
                  <div class="flex flex-col items-start gap-1">
                    <Badge :variant="memberBadgeVariant(item.memberExpireAt)">
                      {{ getMemberExpireMeta(item.memberExpireAt).label }}
                    </Badge>
                    <small class="text-xs leading-tight text-muted-foreground">
                      {{ getMemberExpireMeta(item.memberExpireAt).formattedTime }}
                    </small>
                  </div>
                </TableCell>
              </TableRow>
            </TableBody>
          </Table>
        </div>
      </CardContent>
    </Card>
  </section>
</template>
