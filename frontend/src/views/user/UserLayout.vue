<script setup>
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import dayjs from 'dayjs'
import { useAuthStore } from '../../stores/auth'
import { useChatStore } from '../../stores/chat'
import { getUserProfile } from '../../api/auth'
import { getCurrentTheme, toggleTheme } from '../../utils/theme'
import Button from '@/components/ui/button/Button.vue'
import Sheet from '@/components/ui/sheet/Sheet.vue'
import SheetTrigger from '@/components/ui/sheet/SheetTrigger.vue'
import SheetContent from '@/components/ui/sheet/SheetContent.vue'
import SheetHeader from '@/components/ui/sheet/SheetHeader.vue'
import SheetTitle from '@/components/ui/sheet/SheetTitle.vue'
import { MessageCircle, UserRound, LogOut, Sun, Moon, Menu, Sparkles } from 'lucide-vue-next'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()
const chatStore = useChatStore()
const sidebarOpen = ref(false)
const currentTheme = ref(getCurrentTheme())

const menuItems = [
  { key: 'chat', label: '对话', path: '/app/chat', icon: MessageCircle },
  { key: 'profile', label: '我的', path: '/app/profile', icon: UserRound }
]

const activePath = computed(() => route.path)

const memberTag = computed(() => {
  const profile = authStore.profile
  if (!profile?.memberExpireAt) {
    return '未开通会员'
  }
  const isExpired = dayjs(profile.memberExpireAt).isBefore(dayjs())
  if (isExpired) {
    return `已过期 · ${dayjs(profile.memberExpireAt).format('YYYY-MM-DD')}`
  }
  return `有效期至 ${dayjs(profile.memberExpireAt).format('YYYY-MM-DD HH:mm')}`
})

const memberActive = computed(() => {
  const profile = authStore.profile
  return Boolean(profile?.memberExpireAt) && !dayjs(profile.memberExpireAt).isBefore(dayjs())
})

const jump = (path) => {
  router.push(path)
  sidebarOpen.value = false
}

const closeSidebar = () => {
  sidebarOpen.value = false
}

const switchTheme = () => {
  currentTheme.value = toggleTheme()
}

const logout = () => {
  chatStore.resetChatState()
  authStore.logoutUser()
  router.push('/login')
}

const fetchProfile = async () => {
  const { data } = await getUserProfile()
  authStore.setProfile(data.data)
}

onMounted(async () => {
  await fetchProfile()
})
</script>

<template>
  <section class="flex min-h-dvh">
    <!-- 桌面常驻侧栏 -->
    <aside class="sticky top-0 hidden h-dvh w-72 shrink-0 flex-col gap-5 border-r border-border bg-card/60 p-5 backdrop-blur-xl md:flex">
      <div class="flex items-center gap-3">
        <span class="inline-flex size-10 items-center justify-center rounded-xl bg-primary/15 text-primary">
          <Sparkles class="size-5" />
        </span>
        <div class="min-w-0">
          <p class="truncate text-[15px] font-bold text-foreground">{{ authStore.profile?.nickname || '用户' }}</p>
          <p class="truncate text-xs text-muted-foreground">{{ memberTag }}</p>
        </div>
      </div>

      <nav class="mt-2 flex flex-col gap-1.5">
        <button
          v-for="item in menuItems"
          :key="item.key"
          class="flex min-h-11 w-full cursor-pointer items-center gap-3 rounded-lg px-3.5 text-[15px] font-medium transition-colors duration-150 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
          :class="activePath === item.path
            ? 'bg-accent text-accent-foreground'
            : 'text-muted-foreground hover:bg-muted hover:text-foreground'"
          @click="jump(item.path)"
        >
          <component :is="item.icon" class="size-[18px] shrink-0" />
          {{ item.label }}
        </button>
      </nav>

      <div class="mt-auto flex flex-col gap-3">
        <Button variant="outline" class="justify-start gap-3 text-muted-foreground" @click="switchTheme">
          <Sun v-if="currentTheme === 'dark'" class="size-4 text-warning" />
          <Moon v-else class="size-4 text-info" />
          {{ currentTheme === 'dark' ? '切换浅色' : '切换深色' }}
        </Button>
        <Button variant="ghost" class="justify-start gap-3 text-muted-foreground hover:text-destructive" @click="logout">
          <LogOut class="size-4" />
          退出登录
        </Button>
      </div>
    </aside>

    <!-- 主内容区 -->
    <section class="flex min-w-0 flex-1 flex-col">
      <!-- 移动顶栏 -->
      <header class="sticky top-0 z-40 flex items-center justify-between gap-3 border-b border-border bg-card/85 px-4 py-2.5 backdrop-blur-xl md:hidden">
        <Sheet>
          <SheetTrigger as-child>
            <Button variant="ghost" size="icon" aria-label="打开菜单">
              <Menu class="size-5" />
            </Button>
          </SheetTrigger>
          <SheetContent side="left" class="w-72">
            <SheetHeader>
              <SheetTitle class="flex items-center gap-2">
                <span class="inline-flex size-8 items-center justify-center rounded-lg bg-primary/15 text-primary">
                  <Sparkles class="size-4" />
                </span>
                Thallo
              </SheetTitle>
            </SheetHeader>
            <nav class="flex flex-col gap-1.5 px-3">
              <button
                v-for="item in menuItems"
                :key="item.key"
                class="flex min-h-11 w-full cursor-pointer items-center gap-3 rounded-lg px-3.5 text-[15px] font-medium transition-colors"
                :class="activePath === item.path ? 'bg-accent text-accent-foreground' : 'text-muted-foreground hover:bg-muted'"
                @click="jump(item.path)"
              >
                <component :is="item.icon" class="size-[18px] shrink-0" />
                {{ item.label }}
              </button>
            </nav>
            <div class="mt-auto flex flex-col gap-2 p-4">
              <Button variant="outline" class="justify-start gap-3" @click="switchTheme">
                <Sun v-if="currentTheme === 'dark'" class="size-4 text-warning" />
                <Moon v-else class="size-4 text-info" />
                {{ currentTheme === 'dark' ? '切换浅色' : '切换深色' }}
              </Button>
              <Button variant="ghost" class="justify-start gap-3 text-muted-foreground hover:text-destructive" @click="logout">
                <LogOut class="size-4" />
                退出登录
              </Button>
            </div>
          </SheetContent>
        </Sheet>

        <div class="min-w-0 text-right">
          <p class="truncate text-sm font-bold text-foreground">{{ authStore.profile?.nickname || '用户' }}</p>
          <p class="truncate text-xs text-muted-foreground">{{ memberTag }}</p>
        </div>
      </header>

      <main class="min-h-0 flex-1">
        <router-view />
      </main>
    </section>
  </section>
</template>
