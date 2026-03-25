<template>
  <div v-if="!state.authReady" class="fixed inset-0 z-[1200] flex items-center justify-center bg-slate-950 text-white">
    <div class="flex flex-col items-center gap-4 px-6 text-center">
      <div class="w-16 h-16 rounded-3xl bg-white/5 border border-white/10 flex items-center justify-center shadow-2xl">
        <div class="w-7 h-7 rounded-full border-2 border-white/20 border-t-white animate-spin"></div>
      </div>
      <div>
        <div class="text-sm font-bold tracking-[0.24em] uppercase text-white/70">Sentinel AI</div>
        <div class="mt-2 text-sm text-white/60">正在校验登录状态</div>
      </div>
    </div>
  </div>
</template>

<script setup>
defineProps({
  state: {
    type: Object,
    required: true
  }
});
</script>
