<script setup>
import Card from '@/components/ui/card/Card.vue'
import CardContent from '@/components/ui/card/CardContent.vue'
import Button from '@/components/ui/button/Button.vue'
import { Sparkles } from 'lucide-vue-next'
</script>

<template>
  <section class="relative grid min-h-screen place-items-center overflow-hidden px-4 py-6">
    <div class="landingBackdrop absolute inset-0"></div>

    <Card class="relative z-10 w-full max-w-xl bg-white/10 text-center text-primary-foreground backdrop-blur-xl">
      <CardContent class="flex flex-col items-center gap-6 p-8 sm:p-10">
        <span class="inline-flex size-14 items-center justify-center rounded-2xl bg-primary/20 text-primary-foreground">
          <Sparkles class="size-7" />
        </span>
        <div class="space-y-3">
          <h1 class="text-3xl leading-tight font-bold sm:text-4xl">Thallo</h1>
          <p class="text-[15px] leading-relaxed text-white/80">
            轻量的 AI 对话与塔罗解读空间——开一张牌，聊一段心事，答案即刻呈现。
          </p>
        </div>
        <Button as-child size="lg" class="min-w-44 bg-primary text-primary-foreground shadow-lg hover:bg-primary-hover">
          <router-link to="/login" class="inline-flex items-center gap-2">
            <Sparkles class="size-4" />
            开启灵感
          </router-link>
        </Button>
      </CardContent>
    </Card>
  </section>
</template>

<style scoped>
.landingBackdrop {
  background:
    linear-gradient(120deg, rgba(28, 44, 32, 0.66), rgba(63, 125, 78, 0.3)),
    url('/assets/login-bg.png') center/cover;
  filter: saturate(0.88);
}

.landingBackdrop::after {
  content: '';
  position: absolute;
  inset: 0;
  background:
    radial-gradient(circle at top left, rgba(255, 255, 255, 0.2), transparent 28%),
    rgba(16, 26, 19, 0.2);
  backdrop-filter: blur(8px);
}
</style>
