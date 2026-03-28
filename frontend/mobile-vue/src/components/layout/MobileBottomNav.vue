<template>
  <nav
    class="absolute bottom-6 left-5 right-5 h-[76px] rounded-[16px] border shadow-2xl flex justify-between items-center px-[24px] z-30"
    style="background: rgba(255, 255, 255, 0.48); backdrop-filter: blur(26px) saturate(1.25); -webkit-backdrop-filter: blur(26px) saturate(1.25); border-color: rgba(255, 255, 255, 0.72); box-shadow: 0 22px 40px -18px rgba(15, 23, 42, 0.38), inset 0 1px 0 rgba(255, 255, 255, 0.45);"
  >
    <button type="button" class="flex flex-col items-center gap-0.5" :class="homeActive ? 'text-primary' : 'text-slate-400'" @click="state.activeTab = 'tasks'">
      <i data-lucide="home" size="20"></i>
      <span class="text-[9px] font-bold">首页</span>
    </button>

    <button type="button" class="flex flex-col items-center gap-0.5" :class="messageActive ? 'text-primary' : 'text-slate-400'" @click="state.activeTab = 'alerts'">
      <i data-lucide="message-square" size="20"></i>
      <span class="text-[9px] font-bold">消息</span>
    </button>

    <div class="relative">
      <button type="button" class="w-14 h-14 bg-slate-900 text-white rounded-[16px] flex items-center justify-center shadow-xl shadow-slate-900/30 active:scale-90 transition-transform" @click="state.activeTab = 'submit'">
        <i data-lucide="maximize" size="24"></i>
      </button>
    </div>

    <button type="button" class="flex flex-col items-center gap-0.5" :class="familyActive ? 'text-primary' : 'text-slate-400'" @click="state.activeTab = 'family'">
      <i data-lucide="users" size="20"></i>
      <span class="text-[9px] font-bold">家庭</span>
    </button>

    <button type="button" class="flex flex-col items-center gap-0.5" :class="profileActive ? 'text-primary' : 'text-slate-400'" @click="state.activeTab = 'profile'">
      <i data-lucide="circle-user-round" size="20"></i>
      <span class="text-[9px] font-bold">我的</span>
    </button>
  </nav>
</template>

<script setup>
import { computed, nextTick, onMounted, onUpdated } from 'vue';

const props = defineProps({
  state: {
    type: Object,
    required: true
  }
});

const homeActive = computed(() => ['tasks', 'history', 'risk_trend'].includes(props.state.activeTab));
const messageActive = computed(() => ['alerts'].includes(props.state.activeTab));
const familyActive = computed(() => ['family', 'family_invite'].includes(props.state.activeTab));
const profileActive = computed(() => ['profile', 'profile_privacy'].includes(props.state.activeTab));

const refreshLucideIcons = () => {
  nextTick(() => {
    if (window.lucide && typeof window.lucide.createIcons === 'function') {
      window.lucide.createIcons();
    }
  });
};

onMounted(refreshLucideIcons);
onUpdated(refreshLucideIcons);
</script>
