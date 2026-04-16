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
    return `已过期 · ${dayjs(profile.memberExpireAt).format('YYYY-MM-DD')}`
  }
  return `有效期至 ${dayjs(profile.memberExpireAt).format('YYYY-MM-DD HH:mm')}`
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
  <section class="userWorkspaceShell pageWrap">
    <aside class="userSidebar card" :class="['panelShell', { open: sidebarOpen }]">
      <div class="sidebarGlow"></div>
      <div class="sidebarHead">
        <span class="sectionLabel">Workspace</span>
        <div class="profileMeta">
          <strong>{{ authStore.profile?.nickname || '用户' }}</strong>
          <p class="mutedText">{{ memberTag }}</p>
        </div>
      </div>

      <nav class="sidebarNav">
        <button
          v-for="item in menuItems"
          :key="item.key"
          class="menuBtn"
          :class="{ active: activePath === item.path }"
          @click="jump(item.path)"
        >
          <span class="menuBtnTitle">{{ item.label }}</span>
        </button>
      </nav>

      <div class="sidebarFooter">
        <button
          class="themeRoundBtn"
          :title="currentTheme === 'dark' ? '切换浅色' : '切换深色'"
          @click="switchTheme"
        >
          <span v-if="currentTheme === 'dark'" class="themeIcon sun"></span>
          <span v-else class="themeIcon moon"></span>
        </button>
        <button class="ghostBtn logoutBtn" @click="logout">退出<br />登录</button>
      </div>
    </aside>

    <section class="workspaceMain contentContainer">
      <header class="mobileWorkspaceBar card" :class="'panelShell'">
        <button class="ghostBtn" @click="sidebarOpen = !sidebarOpen">菜单</button>
        <div class="mobileBarMeta">
          <strong>{{ authStore.profile?.nickname || '用户' }}</strong>
          <small class="mutedText">{{ memberTag }}</small>
        </div>
      </header>

      <section class="workspaceView">
        <router-view />
      </section>
    </section>

    <div v-if="sidebarOpen" class="sidebarMask" @click.self="closeSidebar"></div>
  </section>
</template>

<style scoped>
.userWorkspaceShell {
  display: grid;
  grid-template-columns: 310px 1fr;
  gap: 22px;
  padding: 22px;
  min-height: 100dvh;
}

.userSidebar {
  position: sticky;
  top: 22px;
  height: calc(100dvh - 44px);
  display: flex;
  flex-direction: column;
  padding: 22px;
  overflow: hidden;
}

.sidebarGlow {
  position: absolute;
  inset: auto auto -80px -40px;
  width: 220px;
  height: 220px;
  border-radius: 999px;
  background: radial-gradient(circle, rgba(193, 201, 211, 0.42) 0%, transparent 70%);
  pointer-events: none;
}

.sidebarHead {
  position: relative;
  z-index: 1;
  display: grid;
  gap: 18px;
  padding-bottom: 22px;
  border-bottom: 1px solid var(--line-soft);
}

.profileMeta {
  display: grid;
  gap: 8px;
}

.profileMeta strong {
  font-size: 28px;
  color: var(--text-title);
  line-height: 1.05;
}

.profileMeta p {
  margin: 0;
  line-height: 1.55;
}

.sidebarNav {
  position: relative;
  z-index: 1;
  display: grid;
  gap: 10px;
  margin-top: 18px;
}

.menuBtn {
  width: 100%;
  border: 1px solid transparent;
  background: rgba(255, 255, 255, 0.3);
  text-align: left;
  padding: 16px 18px;
  border-radius: 20px;
  cursor: pointer;
  transition: transform 0.18s ease, border-color 0.18s ease, background 0.18s ease, box-shadow 0.18s ease;
}

[data-theme='dark'] .menuBtn {
  background: rgba(255, 255, 255, 0.04);
  border-color: rgba(255, 255, 255, 0.02);
}

.menuBtn:hover {
  transform: translateY(-1px);
  border-color: var(--line-strong);
  background: rgba(255, 255, 255, 0.54);
}

[data-theme='dark'] .menuBtn:hover {
  background: rgba(255, 255, 255, 0.08);
}

.menuBtn.active {
  background: linear-gradient(135deg, rgba(255, 255, 255, 0.74) 0%, rgba(238, 235, 231, 0.86) 100%);
  border-color: rgba(92, 102, 118, 0.18);
  box-shadow: var(--shadow-soft);
}

[data-theme='dark'] .menuBtn.active {
  background: linear-gradient(135deg, rgba(255, 255, 255, 0.09) 0%, rgba(163, 178, 198, 0.12) 100%);
  border-color: rgba(163, 178, 198, 0.16);
}

.menuBtnTitle {
  font-size: 16px;
  color: var(--text-title);
  font-weight: 700;
}

.sidebarFooter {
  position: relative;
  z-index: 1;
  margin-top: auto;
  padding-top: 20px;
  border-top: 1px solid var(--line-soft);
  display: grid;
  gap: 12px;
}

.logoutBtn {
  width: 100%;
}

.themeRoundBtn {
  width: 46px;
  height: 46px;
  border-radius: 16px;
  border: 1px solid var(--line-strong);
  background: rgba(255, 255, 255, 0.46);
  color: var(--text-main);
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.4);
}

[data-theme='dark'] .themeRoundBtn {
  background: rgba(255, 255, 255, 0.04);
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.04);
}

.themeIcon {
  width: 16px;
  height: 16px;
  display: inline-block;
}

.themeIcon.sun {
  border-radius: 50%;
  background: #f3c979;
  box-shadow: 0 0 0 4px rgba(243, 201, 121, 0.2);
}

.themeIcon.moon {
  border-radius: 50%;
  background: #cad5e3;
  position: relative;
}

.themeIcon.moon::after {
  content: '';
  position: absolute;
  right: -2px;
  top: -2px;
  width: 11px;
  height: 11px;
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.8);
}

.workspaceMain {
  display: flex;
  flex-direction: column;
  width: 100%;
  min-width: 0;
}

.workspaceView {
  flex: 1;
  min-height: 0;
  width: 100%;
}

.mobileWorkspaceBar {
  display: none;
  align-items: center;
  justify-content: space-between;
  gap: 14px;
  padding: 14px 16px;
  margin-bottom: 14px;
}

.mobileBarMeta {
  min-width: 0;
  display: grid;
  text-align: right;
}

.mobileBarMeta strong,
.mobileBarMeta small {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

@media (max-width: 960px) {
  .userWorkspaceShell {
    grid-template-columns: 1fr;
    gap: 14px;
    padding: 14px;
  }

  .userSidebar {
    position: fixed;
    left: 14px;
    top: 14px;
    width: clamp(96px, 33.333vw, 140px);
    height: calc(100dvh - 28px);
    transform: translateX(calc(-100% - 20px));
    transition: transform 0.22s ease;
    z-index: 50;
  }

  .sidebarHead {
    gap: 12px;
    padding-bottom: 16px;
  }

  .profileMeta {
    gap: 6px;
  }

  .profileMeta strong {
    font-size: 20px;
  }

  .profileMeta p {
    font-size: 12px;
    line-height: 1.4;
  }

  .sidebarNav {
    gap: 8px;
    margin-top: 14px;
  }

  .menuBtn {
    padding: 12px 10px;
    border-radius: 16px;
  }

  .menuBtnTitle {
    font-size: 14px;
    line-height: 1.2;
  }

  .sidebarFooter {
    gap: 10px;
    padding-top: 16px;
  }

  .userSidebar.open {
    transform: translateX(0);
  }

  .sidebarMask {
    position: fixed;
    inset: 0;
    background: var(--overlay-bg);
    z-index: 40;
  }

  .mobileWorkspaceBar {
    display: flex;
  }
}
</style>
