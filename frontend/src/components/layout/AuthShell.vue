<script setup>
defineProps({
  title: { type: String, default: '' },
  subtitle: { type: String, default: '' },
  maxWidth: { type: String, default: 'max-w-[440px]' },
})
</script>

<template>
  <section class="relative grid min-h-screen place-items-center overflow-hidden px-4 py-8">
    <div class="authBackdrop absolute inset-0"></div>

    <div class="relative z-10 w-full" :class="maxWidth">
      <slot name="brand" />
      <div class="rounded-xl border border-white/15 bg-white/10 p-6 shadow-[var(--shadow-overlay)] backdrop-blur-xl sm:p-8">
        <slot />
      </div>
    </div>
  </section>
</template>

<style scoped>
.authBackdrop {
  background:
    linear-gradient(135deg, rgba(24, 38, 30, 0.6), rgba(63, 125, 78, 0.32)),
    url('/assets/login-bg.png') center / cover no-repeat;
  filter: saturate(0.88);
  transform: scale(1.03);
}

.authBackdrop::after {
  content: '';
  position: absolute;
  inset: 0;
  background:
    radial-gradient(circle at top left, rgba(255, 255, 255, 0.24), transparent 30%),
    rgba(16, 26, 19, 0.2);
  backdrop-filter: blur(10px);
}
</style>
