<script setup>
import { onMounted, ref } from 'vue'
import { getAdminMenus } from '../../api/admin'
import Card from '@/components/ui/card/Card.vue'
import CardContent from '@/components/ui/card/CardContent.vue'
import Table from '@/components/ui/table/Table.vue'
import TableHeader from '@/components/ui/table/TableHeader.vue'
import TableBody from '@/components/ui/table/TableBody.vue'
import TableRow from '@/components/ui/table/TableRow.vue'
import TableHead from '@/components/ui/table/TableHead.vue'
import TableCell from '@/components/ui/table/TableCell.vue'
import { ListTree } from 'lucide-vue-next'

const menus = ref([])

onMounted(async () => {
  const { data } = await getAdminMenus()
  menus.value = data.data
})
</script>

<template>
  <section class="mx-auto max-w-5xl">
    <Card>
      <CardContent class="p-5 sm:p-6">
        <h2 class="flex items-center gap-2 text-lg font-semibold text-card-foreground">
          <ListTree class="size-5 text-primary" />
          菜单管理
        </h2>

        <div class="mt-4 overflow-hidden rounded-xl border border-border">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>菜单名称</TableHead>
                <TableHead>路径</TableHead>
                <TableHead>分组</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              <TableRow v-for="item in menus" :key="item.id">
                <TableCell class="font-medium">{{ item.name }}</TableCell>
                <TableCell class="text-muted-foreground">{{ item.path }}</TableCell>
                <TableCell>{{ item.group }}</TableCell>
              </TableRow>
            </TableBody>
          </Table>
        </div>
      </CardContent>
    </Card>
  </section>
</template>
