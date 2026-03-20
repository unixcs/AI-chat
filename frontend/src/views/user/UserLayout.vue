<script setup>
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import dayjs from 'dayjs'
import { useAuthStore } from '../../stores/auth'
import { useChatStore } from '../../stores/chat'
import { getUserProfile } from '../../api/auth'
import { getCurrentTheme, toggleTheme } from '../../utils/theme'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()
const chatStore = useChatStore()
const sidebarOpen = ref(false)
const currentTheme = ref(getCurrentTheme())

const menuItems = [
  { key: 'chat', label: '对话', path: '/app/chat' },
  { key: 'profile', label: '我的', path: '/app/profile' }
]

const activePath = computed(() => route.path)

const memberTag = computed(() => {
  const profile = authStore.profile
  if (!profile?.memberExpireAt) {
    return '未开通会员'
  }
  const isExpired = dayjs(profile.memberExpireAt).isBefore(dayjs())
  if (isExpired) {
    return `已过期（${dayjs(profile.memberExpireAt).format('YYYY-MM-DD')}）`
  }
  return `到期：${dayjs(profile.memberExpireAt).format('YYYY-MM-DD HH:mm')}`
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
  <section class="userShell pageWrap">
    <aside class="sidebar card" :class="{ open: sidebarOpen }">
      <div class="sidebarHead">
        <strong>{{ authStore.profile?.nickname || '用户' }}</strong>
        <p class="mutedText">{{ memberTag }}</p>
      </div>

      <nav>
        <button
          v-for="item in menuItems"
          :key="item.key"
          class="menuBtn"
          :class="{ active: activePath === item.path }"
          @click="jump(item.path)"
        >
          {{ item.label }}
        </button>
      </nav>

      <button class="themeRoundBtn" :title="currentTheme === 'dark' ? '切换浅色' : '切换深色'" @click="switchTheme">
        <span v-if="currentTheme === 'dark'" class="themeIcon sun"></span>
        <span v-else class="themeIcon moon"></span>
      </button>
      <button class="ghostBtn logoutBtn" @click="logout">退出登录</button>
    </aside>

    <section class="mainArea contentContainer">
      <header class="mobileBar card">
        <button class="ghostBtn" @click="sidebarOpen = !sidebarOpen">菜单</button>
        <span>{{ authStore.profile?.nickname || '用户' }}</span>
      </header>
      <section class="viewArea">
        <router-view />
      </section>
    </section>

    <div v-if="sidebarOpen" class="sidebarMask" @click.self="closeSidebar"></div>
  </section>
</template>

<style scoped>
.userShell {
  display: grid;
  grid-template-columns: 248px 1fr;
  gap: 14px;
  padding: 14px;
  min-height: 100dvh;
}

.sidebar {
  padding: 16px;
  position: sticky;
  top: 14px;
  height: calc(100dvh - 28px);
  display: flex;
  flex-direction: column;
  background: linear-gradient(180deg, var(--bg-panel) 0%, var(--bg-soft) 100%);
}

.sidebarHead {
  border-bottom: 1px solid var(--border-main);
  padding-bottom: 12px;
  margin-bottom: 10px;
}

.sidebarHead p {
  margin: 6px 0 0;
  font-size: 12px;
}

.menuBtn {
  width: 100%;
  border: none;
  background: transparent;
  text-align: left;
  padding: 10px;
  border-radius: 10px;
  margin-bottom: 8px;
  cursor: pointer;
}

.menuBtn.active {
  background: rgba(47, 124, 246, 0.14);
  color: var(--bg-accent);
  font-weight: 600;
}

.logoutBtn {
  margin-top: 10px;
}

.themeRoundBtn {
  margin-top: auto;
  width: 34px;
  height: 34px;
  border-radius: 999px;
  border: 1px solid var(--border-main);
  background: var(--bg-panel);
  color: var(--text-main);
  cursor: pointer;
  font-size: 12px;
  align-self: flex-start;
  display: inline-flex;
  align-items: center;
  justify-content: center;
}

.themeIcon {
  width: 14px;
  height: 14px;
  display: inline-block;
}

.themeIcon.sun {
  border-radius: 50%;
  background: #facc15;
  box-shadow: 0 0 0 2px rgba(250, 204, 21, 0.28);
}

.themeIcon.moon {
  border-radius: 50%;
  background: #bfdbfe;
  position: relative;
}

.themeIcon.moon::after {
  content: '';
  position: absolute;
  right: -1px;
  top: -1px;
  width: 10px;
  height: 10px;
  border-radius: 50%;
  background: var(--bg-panel);
}

.mainArea {
  display: flex;
  flex-direction: column;
  width: 100%;
  min-width: 0;
}

.viewArea {
  flex: 1;
  min-height: 0;
  width: 100%;
}

.mobileBar {
  display: none;
  align-items: center;
  justify-content: space-between;
  padding: 10px 14px;
  margin-bottom: 12px;
}

@media (max-width: 960px) {
  .userShell {
    grid-template-columns: 1fr;
    padding: 10px;
    min-height: 100dvh;
  }

  .sidebar {
    position: fixed;
    left: -94px;
    top: 10px;
    z-index: 50;
    width: 84px;
    height: calc(100dvh - 20px);
    padding: 10px 8px;
    transition: left 0.2s ease;
  }

  .sidebar.open {
    left: 10px;
  }

  .sidebarHead {
    padding-bottom: 8px;
    margin-bottom: 8px;
  }

  .sidebarHead strong {
    display: block;
    font-size: 12px;
    line-height: 1.3;
    word-break: break-all;
  }

  .sidebarHead p {
    margin-top: 4px;
    font-size: 10px;
    line-height: 1.35;
    word-break: break-all;
  }

  .menuBtn {
    text-align: center;
    padding: 8px 4px;
    margin-bottom: 6px;
    font-size: 12px;
  }

  .themeRoundBtn {
    align-self: center;
    width: 30px;
    height: 30px;
  }

  .logoutBtn {
    padding: 6px 4px;
    font-size: 11px;
  }

  .sidebarMask {
    position: fixed;
    inset: 0;
    background: rgba(15, 23, 42, 0.3);
    z-index: 40;
  }

  .mobileBar {
    display: flex;
    flex: 0 0 auto;
  }

  .mainArea {
    min-height: calc(100dvh - 20px);
  }
}
</style>
