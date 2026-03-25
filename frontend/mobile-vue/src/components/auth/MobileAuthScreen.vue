<template>
  <div class="absolute inset-0 z-50 bg-white flex flex-col px-8 pt-safe overflow-y-auto">
    <div class="flex-1 flex flex-col justify-center min-h-[600px]">
      <div class="mb-12 text-center">
        <div class="w-24 h-24 bg-gradient-to-br from-emerald-500 to-teal-600 rounded-3xl mb-6 flex items-center justify-center mx-auto shadow-lg shadow-emerald-500/30">
          <svg class="w-14 h-14 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z"></path></svg>
        </div>
        <h1 class="text-3xl font-bold tracking-tight mb-2 text-slate-900">反诈卫士</h1>
        <p class="text-slate-500 text-base">守护您和家人的财产安全</p>
      </div>

      <div class="flex gap-8 mb-8 border-b border-gray-100 pb-1 justify-center">
        <button @click="state.authMode = 'login'" :class="['pb-2 text-lg font-bold transition-colors', state.authMode === 'login' ? 'text-emerald-600 border-b-2 border-emerald-600' : 'text-gray-400']">登录</button>
        <button @click="state.authMode = 'register'" :class="['pb-2 text-lg font-bold transition-colors', state.authMode === 'register' ? 'text-emerald-600 border-b-2 border-emerald-600' : 'text-gray-400']">注册</button>
      </div>

      <form @submit.prevent="state.handleAuth" class="space-y-3">
        <div v-if="state.authMode === 'register'" class="space-y-0.5">
          <label class="text-xs font-bold text-gray-400 uppercase tracking-wider">用户名</label>
          <input v-model="state.form.username" type="text" class="m-input !py-2" placeholder="设置用户名" required>
        </div>
        <div v-if="state.authMode === 'register'" class="space-y-0.5">
          <label class="text-xs font-bold text-gray-400 uppercase tracking-wider">邮箱</label>
          <input v-model="state.form.email" type="email" class="m-input !py-2" placeholder="name@example.com" required>
        </div>
        <div v-if="state.authMode === 'login'" class="flex gap-3 mb-1">
          <button type="button" @click="state.loginMethod = 'password'" :class="['text-sm font-bold py-2 px-5 rounded-full border transition-colors', state.loginMethod === 'password' ? 'bg-emerald-600 text-white border-emerald-600' : 'bg-white text-gray-400 border-gray-200']">密码登录</button>
          <button type="button" @click="state.loginMethod = 'sms'" :class="['text-sm font-bold py-2 px-5 rounded-full border transition-colors', state.loginMethod === 'sms' ? 'bg-emerald-600 text-white border-emerald-600' : 'bg-white text-gray-400 border-gray-200']">短信登录</button>
        </div>
        <div v-if="state.authMode === 'login' && state.loginMethod === 'password'" class="space-y-0.5">
          <label class="text-xs font-bold text-gray-400 uppercase tracking-wider">账号</label>
          <input v-model="state.form.account" type="text" class="m-input !py-2" placeholder="邮箱或手机号" required>
        </div>
        <div v-if="state.authMode === 'register' || state.loginMethod === 'sms'" class="space-y-0.5">
          <label class="text-xs font-bold text-gray-400 uppercase tracking-wider">手机号</label>
          <input v-model="state.form.phone" type="tel" class="m-input !py-2" placeholder="11位手机号" required>
        </div>
        <div v-if="state.authMode === 'register' || state.loginMethod === 'password'" class="space-y-0.5">
          <label class="text-xs font-bold text-gray-400 uppercase tracking-wider">密码</label>
          <input v-model="state.form.password" type="password" class="m-input !py-2" placeholder="••••••••" required>
        </div>
        <div v-if="state.shouldShowSMSCodeSection" class="space-y-0.5">
          <label class="text-xs font-bold text-gray-400 uppercase tracking-wider">验证码</label>
          <div class="flex items-end gap-3">
            <input v-model="state.form.smsCode" type="text" class="m-input !py-2 text-center tracking-widest" placeholder="000000" required>
            <button type="button" @click="state.sendSMSCode" :disabled="!state.canSendSMSCode" class="shrink-0 h-9 px-4 rounded-lg bg-gray-100 text-xs font-bold text-black disabled:opacity-50">
              {{ state.smsCodeButtonText }}
            </button>
          </div>
        </div>
        <div v-if="state.requiresGraphCaptcha" class="space-y-0.5">
          <label class="text-xs font-bold text-gray-400 uppercase tracking-wider">图形验证</label>
          <div class="flex items-end gap-3">
            <input v-model="state.form.captchaCode" type="text" class="m-input !py-2" placeholder="输入右侧字符" required>
            <div @click="state.fetchCaptcha" class="h-9 w-24 bg-gray-100 rounded-lg overflow-hidden shrink-0">
              <img :src="state.captchaImage" class="w-full h-full object-cover" v-if="state.captchaImage">
            </div>
          </div>
        </div>
        <button type="submit" :disabled="state.loading" class="m-btn-primary mt-6">
          <span v-if="state.loading" class="animate-spin mr-2 w-4 h-4 border-2 border-white/30 border-t-white rounded-full"></span>
          {{ state.authSubmitLabel }}
        </button>
      </form>
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
