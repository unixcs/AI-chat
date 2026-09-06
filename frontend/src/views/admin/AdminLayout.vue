<script setup>
import { computed, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '../../stores/auth'
import { getCurrentTheme, toggleTheme } from '../../utils/theme'
import Button from '@/components/ui/button/Button.vue'
import Sheet from '@/components/ui/sheet/Sheet.vue'
import SheetTrigger from '@/components/ui/sheet/SheetTrigger.vue'
import SheetContent from '@/components/ui/sheet/SheetContent.vue'
import SheetHeader from '@/components/ui/sheet/SheetHeader.vue'
import SheetTitle from '@/components/ui/sheet/SheetTitle.vue'
import AlertDialog from '@/components/ui/alert-dialog/AlertDialog.vue'
import AlertDialogContent from '@/components/ui/alert-dialog/AlertDialogContent.vue'
import AlertDialogHeader from '@/components/ui/alert-dialog/AlertDialogHeader.vue'
import AlertDialogTitle from '@/components/ui/alert-dialog/AlertDialogTitle.vue'
import AlertDialogDescription from '@/components/ui/alert-dialog/AlertDialogDescription.vue'
import AlertDialogFooter from '@/components/ui/alert-dialog/AlertDialogFooter.vue'
import AlertDialogAction from '@/components/ui/alert-dialog/AlertDialogAction.vue'
import AlertDialogCancel from '@/components/ui/alert-dialog/AlertDialogCancel.vue'
import {
  LayoutDashboard, Users, ShieldCheck, ListTree, Activity, FileText,
  Crown, Ticket, ReceiptText, MessagesSquare, Megaphone,
  LogOut, Sun, Moon, Menu, Sparkles
} from 'lucide-vue-next'

const authStore = useAuthStore()
const route = useRoute()
const router = useRouter()
const collapse = ref(false)
const currentTheme = ref(getCurrentTheme())
const showLogoutConfirm = ref(false)

const menuGroups = [
  {
    title: '首页',
    items: [{ label: '控制台', path: '/admin', icon: LayoutDashboard }]
  },
  {
    title: '系统管理',
    items: [
      { label: '用户管理', path: '/admin/users', icon: Users },
      { label: '角色管理', path: '/admin/roles', icon: ShieldCheck },
      { label: '菜单管理', path: '/admin/menus', icon: ListTree },
      { label: 'AI 状态', path: '/admin/ai', icon: Activity },
      { label: '提示词管理', path: '/admin/prompt', icon: FileText }
    ]
  },
  {
    title: '业务管理',
    items: [
      { label: '会员管理', path: '/admin/members', icon: Crown },
      { label: '兑换码管理', path: '/admin/redeem-codes', icon: Ticket },
      { label: '兑换记录', path: '/admin/redeem-records', icon: ReceiptText },
      { label: '会话管理', path: '/admin/conversations', icon: MessagesSquare },
      { label: '公告管理', path: '/admin/announcements', icon: Megaphone }
    ]
  }
]

const activePath = computed(() => route.path)

const toggleMenu = () => {
  collapse.value = !collapse.value
}

const navigateTo = (path) => {
  router.push(path)
  if (window.innerWidth <= 960) {
    collapse.value = false
  }
}

const switchTheme = () => {
  currentTheme.value = toggleTheme()
}

const logout = () => {
  showLogoutConfirm.value = false
  authStore.logoutAdmin()
  router.push('/admin/login')
}
</script>

<template>
  <section class="flex min-h-dvh">
    <!-- 桌面侧栏 -->
    <aside class="sticky top-0 hidden h-dvh w-64 shrink-0 flex-col border-r border-border bg-card/70 backdrop-blur-xl md:flex">
      <div class="flex items-center gap-2.5 px-5 py-5">
        <span class="inline-flex size-9 items-center justify-center rounded-xl bg-primary/15 text-primary">
          <Sparkles class="size-5" />
        </span>
        <div>
          <p class="text-xs font-semibold tracking-widest text-faint uppercase">Admin</p>
          <p class="text-[15px] leading-tight font-bold text-foreground">管理系统</p>
        </div>
      </div>

      <nav class="min-h-0 flex-1 overflow-y-auto px-3 pb-4">
        <section v-for="group in menuGroups" :key="group.title" class="mb-4">
          <p class="px-2.5 pb-1.5 text-xs font-bold tracking-wider text-faint">{{ group.title }}</p>
          <button
            v-for="item in group.items"
            :key="item.path"
            class="mb-0.5 flex min-h-10 w-full cursor-pointer items-center gap-3 rounded-lg px-2.5 text-sm font-medium transition-colors duration-150 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
            :class="activePath === item.path
              ? 'bg-accent font-semibold text-accent-foreground'
              : 'text-muted-foreground hover:bg-muted hover:text-foreground'"
            @click="navigateTo(item.path)"
          >
            <component :is="item.icon" class="size-4 shrink-0" />
            {{ item.label }}
          </button>
        </section>
      </nav>

      <div class="flex items-center gap-2 border-t border-border p-4">
        <Button variant="ghost" size="icon" :title="currentTheme === 'dark' ? '切换浅色' : '切换深色'" class="text-muted-foreground" @click="switchTheme">
          <Sun v-if="currentTheme === 'dark'" class="size-4 text-warning" />
          <Moon v-else class="size-4 text-info" />
        </Button>
        <Button variant="ghost" class="flex-1 justify-start gap-2.5 text-muted-foreground hover:text-destructive" @click="showLogoutConfirm = true">
          <LogOut class="size-4" />
          退出后台
        </Button>
      </div>
    </aside>

    <!-- 主区 -->
    <main class="flex min-w-0 flex-1 flex-col">
      <header class="sticky top-0 z-40 flex items-center justify-between gap-3 border-b border-border bg-card/85 px-4 py-2.5 backdrop-blur-xl">
        <div class="flex items-center gap-2">
          <!-- 移动菜单 -->
          <Sheet>
            <SheetTrigger as-child>
              <Button variant="ghost" size="icon" aria-label="打开菜单" class="md:hidden">
                <Menu class="size-5" />
              </Button>
            </SheetTrigger>
            <SheetContent side="left" class="w-72 p-0">
              <SheetHeader class="border-b border-border">
                <SheetTitle class="flex items-center gap-2">
                  <span class="inline-flex size-8 items-center justify-center rounded-lg bg-primary/15 text-primary">
                    <Sparkles class="size-4" />
                  </span>
                  管理系统
                </SheetTitle>
              </SheetHeader>
              <nav class="min-h-0 flex-1 overflow-y-auto p-3">
                <section v-for="group in menuGroups" :key="group.title" class="mb-3">
                  <p class="px-2.5 pb-1.5 text-xs font-bold tracking-wider text-faint">{{ group.title }}</p>
                  <button
                    v-for="item in group.items"
                    :key="item.path"
                    class="mb-0.5 flex min-h-11 w-full cursor-pointer items-center gap-3 rounded-lg px-2.5 text-sm font-medium transition-colors"
                    :class="activePath === item.path ? 'bg-accent font-semibold text-accent-foreground' : 'text-muted-foreground hover:bg-muted'"
                    @click="navigateTo(item.path)"
                  >
                    <component :is="item.icon" class="size-4 shrink-0" />
                    {{ item.label }}
                  </button>
                </section>
              </nav>
              <div class="border-t border-border p-4">
                <Button variant="ghost" class="w-full justify-start gap-2.5 text-muted-foreground hover:text-destructive" @click="showLogoutConfirm = true">
                  <LogOut class="size-4" />
                  退出后台
                </Button>
              </div>
            </SheetContent>
          </Sheet>
          <strong class="text-[15px] text-foreground">欢迎使用管理系统</strong>
        </div>
        <Button variant="ghost" size="icon" :title="currentTheme === 'dark' ? '切换浅色' : '切换深色'" class="text-muted-foreground md:hidden" @click="switchTheme">
          <Sun v-if="currentTheme === 'dark'" class="size-4 text-warning" />
          <Moon v-else class="size-4 text-info" />
        </Button>
      </header>

      <section class="min-h-0 flex-1 overflow-y-auto p-4 md:p-6">
        <router-view />
      </section>
    </main>

    <!-- 登出确认 -->
    <AlertDialog :open="showLogoutConfirm" @update:open="showLogoutConfirm = $event">
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>退出管理后台？</AlertDialogTitle>
          <AlertDialogDescription>退出后需要重新输入管理员账号与密码。</AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel>取消</AlertDialogCancel>
          <AlertDialogAction class="bg-destructive text-destructive-foreground hover:opacity-90" @click="logout">退出</AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  </section>
</template>
