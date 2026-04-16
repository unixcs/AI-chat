<template>
  <section class="landingShell pageWrap">
    <div class="landingBackdrop"></div>

    <div class="landingWrap contentContainer">
      <article class="landingPanel card panelShell">
        <div class="landingActions">
          <router-link class="primaryBtn landingCta" to="/login">开启灵感</router-link>
        </div>
      </article>
    </div>
  </section>
</template>

<style scoped>
.landingShell {
  position: relative;
  display: grid;
  place-items: center;
  padding: 24px;
  overflow: hidden;
}

.landingBackdrop {
  position: absolute;
  inset: 0;
  background:
    linear-gradient(120deg, rgba(34, 40, 49, 0.62), rgba(75, 86, 100, 0.36)),
    url('/assets/login-bg.png') center/cover;
  filter: saturate(0.88);
}

.landingBackdrop::after {
  content: '';
  position: absolute;
  inset: 0;
  background:
    radial-gradient(circle at top left, rgba(255, 255, 255, 0.28), transparent 28%),
    rgba(18, 23, 29, 0.18);
  backdrop-filter: blur(8px);
}

.landingWrap {
  position: relative;
  z-index: 1;
  width: min(1160px, 100%);
}

.landingPanel {
  width: min(520px, 100%);
  margin: 0 auto;
  padding: 32px 24px;
  color: #f8f7f3;
  background: linear-gradient(180deg, rgba(255, 255, 255, 0.08) 0%, rgba(255, 255, 255, 0.03) 100%);
  border-color: rgba(255, 255, 255, 0.14);
  box-shadow: var(--shadow-float);
  text-align: center;
}

.landingActions {
  display: flex;
  justify-content: center;
}

.landingCta {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 168px;
  text-align: center;
}

@media (max-width: 640px) {
  .landingShell {
    padding: 14px;
  }

  .landingPanel {
    padding: 26px 18px;
  }
}
</style>
