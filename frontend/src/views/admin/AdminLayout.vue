<script setup>
import { computed, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '../../stores/auth'
import { getCurrentTheme, toggleTheme } from '../../utils/theme'

const authStore = useAuthStore()
const route = useRoute()
const router = useRouter()
const collapse = ref(false)
const currentTheme = ref(getCurrentTheme())

const menuGroups = [
  {
    title: '首页',
    items: [{ label: '控制台', path: '/admin' }]
  },
  {
    title: '系统管理',
    items: [
      { label: '用户管理', path: '/admin/users' },
      { label: '角色管理', path: '/admin/roles' },
      { label: '菜单管理', path: '/admin/menus' }
    ]
  },
  {
    title: '业务管理',
    items: [
      { label: '会员管理', path: '/admin/members' },
      { label: '兑换码管理', path: '/admin/redeem-codes' },
      { label: '兑换记录', path: '/admin/redeem-records' },
      { label: '会话管理', path: '/admin/conversations' }
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
  authStore.logoutAdmin()
  router.push('/admin/login')
}
</script>

<template>
  <section class="adminWorkspaceShell pageWrap">
    <div v-if="collapse" class="asideMask" @click="collapse = false"></div>

    <aside class="adminSidebar card" :class="['panelShell', { collapse }]">
      <div class="adminSidebarGlow"></div>

      <div class="brandBar">
        <span class="sectionLabel">Admin</span>
        <strong>管理系统</strong>
      </div>

      <div class="menuArea">
        <section v-for="group in menuGroups" :key="group.title" class="menuGroup">
          <p>{{ group.title }}</p>
          <button
            v-for="item in group.items"
            :key="item.path"
            :class="{ active: activePath === item.path }"
            @click="navigateTo(item.path)"
          >
            {{ item.label }}
          </button>
        </section>
      </div>

      <div class="adminNavFooter">
        <button
          class="themeRoundBtn"
          :title="currentTheme === 'dark' ? '切换浅色' : '切换深色'"
          @click="switchTheme"
        >
          <span v-if="currentTheme === 'dark'" class="themeIcon sun"></span>
          <span v-else class="themeIcon moon"></span>
        </button>
        <button class="logoutBtn" @click="logout">退出后台</button>
      </div>
    </aside>

    <main class="adminMain">
      <header class="adminTopbar card" :class="['contentContainer', 'panelShell']">
        <button class="ghostBtn" @click="toggleMenu">菜单</button>
        <div class="topbarText">
          <strong>欢迎使用管理系统</strong>
          <small class="mutedText">信息结构与权限逻辑保持原样，统一为更简约的现代雾感风格。</small>
        </div>
      </header>

      <section class="contentContainer pageSection adminPageSection">
        <router-view />
      </section>
    </main>
  </section>
</template>

<style scoped>
.adminWorkspaceShell {
  display: grid;
  grid-template-columns: 292px 1fr;
  min-height: 100vh;
  gap: 20px;
  padding: 20px;
}

.adminSidebar {
  position: sticky;
  top: 20px;
  height: calc(100dvh - 40px);
  padding: 22px;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  background: linear-gradient(180deg, rgba(251, 248, 244, 0.92) 0%, rgba(238, 233, 225, 0.88) 100%);
}

[data-theme='dark'] .adminSidebar {
  background: linear-gradient(180deg, rgba(25, 30, 37, 0.96) 0%, rgba(18, 23, 29, 0.94) 100%);
}

.adminSidebarGlow {
  position: absolute;
  inset: -40px auto auto -20px;
  width: 180px;
  height: 180px;
  border-radius: 999px;
  background: radial-gradient(circle, rgba(219, 209, 193, 0.48) 0%, transparent 72%);
  pointer-events: none;
}

[data-theme='dark'] .adminSidebarGlow {
  background: radial-gradient(circle, rgba(116, 137, 164, 0.2) 0%, transparent 72%);
}

.brandBar {
  position: relative;
  z-index: 1;
  display: grid;
  gap: 14px;
  padding-bottom: 18px;
  border-bottom: 1px solid var(--line-soft);
}

.brandBar strong {
  font-size: 28px;
  color: var(--text-title);
  line-height: 1.08;
}

.menuArea {
  position: relative;
  z-index: 1;
  flex: 1;
  overflow: auto;
  margin-top: 18px;
}

.menuGroup {
  margin-bottom: 18px;
}

.menuGroup p {
  margin: 0 0 10px;
  font-size: 12px;
  color: var(--text-soft);
  letter-spacing: 0.06em;
  text-transform: uppercase;
}

.menuGroup button {
  width: 100%;
  border: 1px solid transparent;
  border-radius: 18px;
  background: rgba(255, 255, 255, 0.28);
  color: var(--text-main);
  text-align: left;
  padding: 12px 14px;
  cursor: pointer;
  margin-bottom: 8px;
  transition: transform 0.18s ease, border-color 0.18s ease, background 0.18s ease, box-shadow 0.18s ease;
}

[data-theme='dark'] .menuGroup button {
  background: rgba(255, 255, 255, 0.04);
  border-color: rgba(255, 255, 255, 0.02);
}

.menuGroup button:hover {
  transform: translateY(-1px);
  border-color: var(--line-strong);
  background: rgba(255, 255, 255, 0.52);
}

[data-theme='dark'] .menuGroup button:hover {
  background: rgba(255, 255, 255, 0.08);
}

.menuGroup button.active {
  background: rgba(95, 111, 133, 0.12);
  border-color: rgba(95, 111, 133, 0.22);
  box-shadow: var(--shadow-soft);
}

[data-theme='dark'] .menuGroup button.active {
  background: rgba(163, 178, 198, 0.12);
  border-color: rgba(163, 178, 198, 0.18);
}

.adminNavFooter {
  position: relative;
  z-index: 1;
  margin-top: auto;
  padding-top: 18px;
  border-top: 1px solid var(--line-soft);
  display: grid;
  gap: 12px;
}

.logoutBtn {
  min-height: 44px;
  border: 1px solid var(--line-strong);
  border-radius: 16px;
  background: rgba(255, 255, 255, 0.44);
  color: var(--text-main);
  cursor: pointer;
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
}

[data-theme='dark'] .themeRoundBtn,
[data-theme='dark'] .logoutBtn {
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

.adminMain {
  min-width: 0;
}

.adminTopbar {
  margin-bottom: 14px;
  display: flex;
  align-items: center;
  gap: 14px;
  padding: 16px 18px;
}

.topbarText {
  display: grid;
  gap: 4px;
}

.topbarText strong {
  color: var(--text-title);
}

.adminPageSection {
  padding: 0;
}

@media (max-width: 960px) {
  .adminWorkspaceShell {
    grid-template-columns: 1fr;
    gap: 14px;
    padding: 14px;
  }

  .asideMask {
    position: fixed;
    inset: 0;
    background: var(--overlay-bg);
    z-index: 55;
  }

  .adminSidebar {
    position: fixed;
    left: 14px;
    top: 14px;
    width: min(320px, calc(100vw - 28px));
    height: calc(100dvh - 28px);
    transform: translateX(calc(-100% - 18px));
    transition: transform 0.22s ease;
    z-index: 60;
  }

  .adminSidebar.collapse {
    transform: translateX(0);
    width: min(320px, calc(100vw - 28px));
  }

  .adminTopbar {
    align-items: flex-start;
    padding: 14px 16px;
    flex-direction: column;
  }

  .topbarText {
    width: 100%;
  }

  .topbarText strong,
  .topbarText small {
    word-break: break-word;
    overflow-wrap: anywhere;
  }
}
</style>
