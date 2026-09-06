<script setup>
import { onMounted, reactive, ref } from 'vue'
import { getAdminRoles, updateAdminPassword } from '../../api/admin'
import Card from '@/components/ui/card/Card.vue'
import CardContent from '@/components/ui/card/CardContent.vue'
import Input from '@/components/ui/input/Input.vue'
import Label from '@/components/ui/label/Label.vue'
import Button from '@/components/ui/button/Button.vue'
import Alert from '@/components/ui/alert/Alert.vue'
import Table from '@/components/ui/table/Table.vue'
import TableHeader from '@/components/ui/table/TableHeader.vue'
import TableBody from '@/components/ui/table/TableBody.vue'
import TableRow from '@/components/ui/table/TableRow.vue'
import TableHead from '@/components/ui/table/TableHead.vue'
import TableCell from '@/components/ui/table/TableCell.vue'
import { toast } from 'vue-sonner'
import { KeyRound, ShieldCheck } from 'lucide-vue-next'

const roles = ref([])
const noticeText = ref('')
const errorText = ref('')

const passwordForm = reactive({
  newPassword: '',
  confirmPassword: ''
})

onMounted(async () => {
  const { data } = await getAdminRoles()
  roles.value = data.data
})

const submitAdminPassword = async () => {
  noticeText.value = ''
  errorText.value = ''
  if (!passwordForm.newPassword || !passwordForm.confirmPassword) {
    errorText.value = '请完整填写密码'
    return
  }
  if (passwordForm.newPassword.length < 6) {
    errorText.value = '新密码至少 6 位'
    return
  }
  if (passwordForm.newPassword !== passwordForm.confirmPassword) {
    errorText.value = '两次密码不一致'
    return
  }

  await updateAdminPassword(passwordForm.newPassword)
  noticeText.value = '管理员密码修改成功'
  toast.success('管理员密码修改成功')
  passwordForm.newPassword = ''
  passwordForm.confirmPassword = ''
}
</script>

<template>
  <section class="mx-auto grid max-w-6xl gap-4 lg:grid-cols-[1fr_380px]">
    <Card>
      <CardContent class="p-5 sm:p-6">
        <h2 class="flex items-center gap-2 text-lg font-semibold text-card-foreground">
          <ShieldCheck class="size-5 text-primary" />
          角色管理
        </h2>

        <div class="mt-4 overflow-hidden rounded-xl border border-border">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>角色名称</TableHead>
                <TableHead>描述</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              <TableRow v-for="item in roles" :key="item.id">
                <TableCell class="font-medium">{{ item.name }}</TableCell>
                <TableCell class="whitespace-normal text-muted-foreground">{{ item.description }}</TableCell>
              </TableRow>
            </TableBody>
          </Table>
        </div>
      </CardContent>
    </Card>

    <Card>
      <CardContent class="p-5 sm:p-6">
        <h2 class="flex items-center gap-2 text-lg font-semibold text-card-foreground">
          <KeyRound class="size-5 text-primary" />
          管理员密码
        </h2>

        <form class="mt-4 space-y-4" @submit.prevent="submitAdminPassword">
          <div class="space-y-2">
            <Label for="role-new-password">新密码</Label>
            <Input id="role-new-password" v-model="passwordForm.newPassword" type="password" placeholder="请输入新密码" />
          </div>
          <div class="space-y-2">
            <Label for="role-confirm-password">确认密码</Label>
            <Input id="role-confirm-password" v-model="passwordForm.confirmPassword" type="password" placeholder="请再次输入新密码" />
          </div>

          <Alert v-if="errorText" variant="destructive">{{ errorText }}</Alert>
          <p v-if="noticeText" class="m-0 text-[13px] text-primary">{{ noticeText }}</p>

          <Button type="submit">修改密码</Button>
        </form>
      </CardContent>
    </Card>
  </section>
</template>
