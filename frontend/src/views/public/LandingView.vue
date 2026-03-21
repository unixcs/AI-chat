<template>
  <section class="landingPage pageWrap">
    <div class="contentWrap">
      <router-link class="primaryBtn ctaOnly" to="/login">开启灵感</router-link>
    </div>
  </section>
</template>

<style scoped>
.landingPage {
  background:
    linear-gradient(120deg, rgba(9, 16, 33, 0.62), rgba(31, 78, 147, 0.48)),
    url('/assets/login-bg.png') center/cover;
  display: grid;
  place-items: center;
  padding: 24px;
}

.contentWrap {
  width: min(520px, 100%);
  text-align: center;
  color: #fff;
  background: rgba(11, 20, 41, 0.38);
  border: 1px solid rgba(255, 255, 255, 0.18);
  border-radius: 18px;
  padding: 32px 24px;
  backdrop-filter: blur(6px);
  box-shadow: 0 14px 30px rgba(12, 20, 46, 0.35);
  position: relative;
}

.ctaOnly {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: 14px 34px;
  border-radius: 999px;
  font-size: 16px;
  letter-spacing: 1px;
  box-shadow: 0 10px 22px rgba(47, 124, 246, 0.38);
  transition: transform 0.2s ease, box-shadow 0.2s ease;
}

.ctaOnly:hover {
  transform: translateY(-2px);
  box-shadow: 0 16px 28px rgba(47, 124, 246, 0.35);
}

@media (max-width: 640px) {
  .contentWrap {
    padding: 26px 18px;
  }

  .ctaOnly {
    width: 100%;
  }
}
</style>
