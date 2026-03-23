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
  <section class="adminShell pageWrap">
    <div v-if="collapse" class="asideMask" @click="collapse = false"></div>
    <aside class="adminAside" :class="{ collapse }">
      <div class="brandBar">
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

      <button class="themeRoundBtn" :title="currentTheme === 'dark' ? '切换浅色' : '切换深色'" @click="switchTheme">
        <span v-if="currentTheme === 'dark'" class="themeIcon sun"></span>
        <span v-else class="themeIcon moon"></span>
      </button>
      <button class="logoutBtn" @click="logout">退出后台</button>
    </aside>

    <main class="adminMain">
      <header class="adminHeader card contentContainer">
        <button class="ghostBtn" @click="toggleMenu">菜单</button>
        <span>欢迎使用管理系统</span>
      </header>
      <section class="contentContainer pageSection">
        <router-view />
      </section>
    </main>
  </section>
</template>

<style scoped>
.adminShell {
  display: grid;
  grid-template-columns: 236px 1fr;
  min-height: 100vh;
}

.adminAside {
  background: var(--bg-admin);
  color: #d1d5db;
  padding: 16px;
  display: flex;
  flex-direction: column;
  box-shadow: 4px 0 18px rgba(7, 15, 28, 0.2);
}

.brandBar {
  font-size: 18px;
  margin-bottom: 18px;
  color: #fff;
}

.menuArea {
  flex: 1;
  overflow: auto;
}

.menuGroup {
  margin-bottom: 16px;
}

.menuGroup p {
  margin: 0 0 8px;
  font-size: 12px;
  color: #9ca3af;
}

.menuGroup button {
  width: 100%;
  border: none;
  border-radius: 8px;
  background: transparent;
  color: #e5e7eb;
  text-align: left;
  padding: 9px 10px;
  cursor: pointer;
  margin-bottom: 6px;
}

.menuGroup button.active {
  background: rgba(47, 124, 246, 0.35);
}

[data-theme='dark'] .menuGroup button.active {
  background: rgba(94, 160, 255, 0.28);
}

.logoutBtn {
  margin-top: 10px;
  border: 1px solid rgba(255, 255, 255, 0.2);
  border-radius: 10px;
  background: transparent;
  color: #fff;
  padding: 8px;
  cursor: pointer;
}

.themeRoundBtn {
  margin-top: auto;
  width: 34px;
  height: 34px;
  border-radius: 999px;
  border: 1px solid rgba(255, 255, 255, 0.25);
  background: rgba(255, 255, 255, 0.08);
  color: #e5e7eb;
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
  background: rgba(255, 255, 255, 0.08);
}

.adminMain {
  padding: 14px;
}

.adminHeader {
  margin-bottom: 14px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 11px 14px;
}

.collapse {
  width: 74px;
}

.collapse .menuGroup p,
.collapse .menuGroup button,
.collapse .logoutBtn,
.collapse .brandBar {
  overflow: hidden;
  text-indent: -9999px;
  position: relative;
}

@media (max-width: 960px) {
  .adminShell {
    grid-template-columns: 1fr;
  }

  .asideMask {
    position: fixed;
    inset: 0;
    background: rgba(15, 23, 42, 0.4);
    z-index: 55;
  }

  .adminAside {
    position: fixed;
    left: 0;
    top: 0;
    width: 236px;
    height: 100vh;
    transform: translateX(-100%);
    transition: transform 0.2s ease;
    z-index: 60;
    display: flex;
  }

  .adminAside.collapse {
    transform: translateX(0);
    width: 236px;
  }

  .collapse .menuGroup p,
  .collapse .menuGroup button,
  .collapse .logoutBtn,
  .collapse .brandBar {
    overflow: visible;
    text-indent: 0;
    position: static;
  }

  .adminMain {
    padding: 10px;
  }
}
</style>
