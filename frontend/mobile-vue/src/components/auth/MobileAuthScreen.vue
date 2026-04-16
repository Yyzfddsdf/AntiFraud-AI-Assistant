<template>
  <div class="absolute inset-0 z-50 bg-slate-50 flex flex-col px-6 sm:px-10 pt-safe overflow-y-auto" style="-webkit-overflow-scrolling: touch;">
    <!-- Abstract Background Decor -->
    <div class="fixed top-0 left-0 right-0 h-96 bg-gradient-to-br from-emerald-500/20 via-teal-500/5 to-transparent pointer-events-none -z-10 rounded-b-[100px] blur-3xl opacity-60"></div>
    <div class="fixed -top-32 -right-32 w-96 h-96 bg-blue-500/10 rounded-full blur-3xl pointer-events-none -z-10"></div>
    
    <div class="flex-1 flex flex-col justify-center min-h-[650px] py-12">
      <!-- Logo & Welcome -->
      <div class="mb-12 text-center relative">
        <div class="relative w-24 h-24 mx-auto mb-6">
          <div class="absolute inset-0 bg-gradient-to-br from-emerald-400 to-teal-600 rounded-[28px] rotate-3 opacity-20 blur-md"></div>
          <div class="relative w-full h-full bg-gradient-to-br from-emerald-500 to-teal-600 rounded-[28px] flex items-center justify-center shadow-[0_8px_30px_rgba(16,185,129,0.3)] border border-white/20">
            <svg class="w-12 h-12 text-white drop-shadow-md" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z"></path></svg>
          </div>
        </div>
        <h1 class="text-[32px] font-[800] tracking-tight mb-2 text-slate-800">反诈卫士</h1>
        <p class="text-slate-500 text-[15px] font-medium tracking-wide">守护您和家人的财产安全</p>
      </div>

      <!-- Auth Card -->
      <div class="bg-white/80 backdrop-blur-2xl rounded-[32px] px-[24px] py-[24px] sm:px-[32px] sm:py-[32px] shadow-[0_8px_40px_rgba(0,0,0,0.04)] border border-white/60">
        <!-- Tabs -->
        <div class="flex gap-8 mb-8 border-b border-slate-100 pb-0 justify-center relative">
          <button @click="state.authMode = 'login'" :class="['pb-3 text-[17px] font-[800] transition-all relative', state.authMode === 'login' ? 'text-emerald-600' : 'text-slate-400 hover:text-slate-600']">
            <span>登录</span>
            <div v-if="state.authMode === 'login'" class="absolute bottom-0 left-1/2 -translate-x-1/2 w-6 h-1 bg-emerald-500 rounded-t-full"></div>
          </button>
          <button @click="state.authMode = 'register'" :class="['pb-3 text-[17px] font-[800] transition-all relative', state.authMode === 'register' ? 'text-emerald-600' : 'text-slate-400 hover:text-slate-600']">
            <span>注册</span>
            <div v-if="state.authMode === 'register'" class="absolute bottom-0 left-1/2 -translate-x-1/2 w-6 h-1 bg-emerald-500 rounded-t-full"></div>
          </button>
        </div>

        <form @submit.prevent="state.handleAuth" class="space-y-[16px]">
          
          <!-- Registration Fields -->
          <div v-if="state.authMode === 'register'" class="space-y-4">
            <div>
              <input v-model="state.form.username" type="text" class="w-full bg-slate-50/80 border border-slate-100 rounded-[16px] px-5 py-4 text-slate-800 text-[15px] placeholder:text-slate-400 focus:ring-2 focus:ring-emerald-500/20 focus:border-emerald-500/30 focus:bg-white transition-all outline-none" placeholder="设置用户名" required>
            </div>
            <div>
              <input v-model="state.form.email" type="email" class="w-full bg-slate-50/80 border border-slate-100 rounded-[16px] px-5 py-4 text-slate-800 text-[15px] placeholder:text-slate-400 focus:ring-2 focus:ring-emerald-500/20 focus:border-emerald-500/30 focus:bg-white transition-all outline-none" placeholder="邮箱地址" required>
            </div>
          </div>

          <!-- Login Methods Toggle -->
          <div v-if="state.authMode === 'login'" class="flex gap-2 p-1 bg-slate-100/80 rounded-[16px] mb-2">
            <button type="button" @click="state.loginMethod = 'password'" :class="['flex-1 text-[13px] font-bold py-2.5 rounded-[12px] transition-all', state.loginMethod === 'password' ? 'bg-white text-emerald-700 shadow-sm' : 'text-slate-500 hover:text-slate-700']">密码登录</button>
            <button type="button" @click="state.loginMethod = 'sms'" :class="['flex-1 text-[13px] font-bold py-2.5 rounded-[12px] transition-all', state.loginMethod === 'sms' ? 'bg-white text-emerald-700 shadow-sm' : 'text-slate-500 hover:text-slate-700']">验证码登录</button>
          </div>

          <!-- Password Login Field -->
          <div v-if="state.authMode === 'login' && state.loginMethod === 'password'">
            <input v-model="state.form.account" type="text" class="w-full bg-slate-50/80 border border-slate-100 rounded-[16px] px-5 py-4 text-slate-800 text-[15px] placeholder:text-slate-400 focus:ring-2 focus:ring-emerald-500/20 focus:border-emerald-500/30 focus:bg-white transition-all outline-none" placeholder="邮箱或手机号" required>
          </div>

          <!-- Phone Field -->
          <div v-if="state.authMode === 'register' || state.loginMethod === 'sms'">
            <input v-model="state.form.phone" type="tel" class="w-full bg-slate-50/80 border border-slate-100 rounded-[16px] px-5 py-4 text-slate-800 text-[15px] placeholder:text-slate-400 focus:ring-2 focus:ring-emerald-500/20 focus:border-emerald-500/30 focus:bg-white transition-all outline-none tracking-wide" placeholder="11位手机号" required>
          </div>

          <!-- Password Field -->
          <div v-if="state.authMode === 'register' || state.loginMethod === 'password'">
            <input v-model="state.form.password" type="password" class="w-full bg-slate-50/80 border border-slate-100 rounded-[16px] px-5 py-4 text-slate-800 text-[15px] placeholder:text-slate-400 focus:ring-2 focus:ring-emerald-500/20 focus:border-emerald-500/30 focus:bg-white transition-all outline-none tracking-widest" placeholder="输入密码" required>
          </div>

          <!-- SMS Code Field -->
          <div v-if="state.shouldShowSMSCodeSection" class="grid grid-cols-[minmax(0,1fr)_112px] gap-3 items-stretch">
            <input v-model="state.form.smsCode" type="text" class="min-w-0 bg-slate-50/80 border border-slate-100 rounded-[16px] px-5 py-4 text-slate-800 text-[15px] placeholder:text-slate-400 focus:ring-2 focus:ring-emerald-500/20 focus:border-emerald-500/30 focus:bg-white transition-all outline-none text-center tracking-[0.2em] font-mono" placeholder="短信验证码" required>
            <button type="button" @click="state.sendSMSCode" :disabled="!state.canSendSMSCode" class="h-[56px] min-w-[112px] px-3 rounded-[16px] bg-emerald-50 hover:bg-emerald-100 active:bg-emerald-200 text-[13px] leading-[1.2] font-bold text-emerald-700 disabled:opacity-50 disabled:bg-slate-50 disabled:text-slate-400 transition-colors text-center">
              {{ state.smsCodeButtonText }}
            </button>
          </div>

          <!-- Captcha Field -->
          <div v-if="state.requiresGraphCaptcha" class="grid grid-cols-[minmax(0,1fr)_144px] gap-3 items-center">
            <input v-model="state.form.captchaCode" type="text" class="min-w-0 bg-slate-50/80 border border-slate-100 rounded-[16px] px-5 py-4 text-slate-800 text-[15px] placeholder:text-slate-400 focus:ring-2 focus:ring-emerald-500/20 focus:border-emerald-500/30 focus:bg-white transition-all outline-none tracking-widest uppercase font-mono" placeholder="图形验证码" required>
            <div @click="state.fetchCaptcha" class="w-[144px] min-w-[144px] h-[48px] bg-white rounded-[16px] overflow-hidden shrink-0 border border-slate-100/50 cursor-pointer active:opacity-80 transition-opacity self-center">
              <img :src="state.captchaImage" class="w-full h-full object-cover mix-blend-multiply" v-if="state.captchaImage">
              <div v-else class="w-full h-full flex items-center justify-center text-[10px] text-slate-400 font-medium">获取中...</div>
            </div>
          </div>

          <!-- Submit Button -->
          <button type="submit" :disabled="state.loading" class="w-full mt-[24px] bg-gradient-to-r from-emerald-500 to-teal-500 active:from-emerald-600 active:to-teal-600 text-white rounded-[16px] py-[16px] font-[800] text-[16px] shadow-[0_8px_20px_rgba(16,185,129,0.25)] transition-all flex items-center justify-center disabled:opacity-70 disabled:shadow-none">
            <span v-if="state.loading" class="animate-spin mr-2 w-5 h-5 border-2 border-white/30 border-t-white rounded-full"></span>
            {{ state.authSubmitLabel }}
          </button>
        </form>
        
        <!-- Footer info -->
        <p class="mt-6 text-center text-[11px] text-slate-400">
          登录即代表同意 <a href="#" class="text-emerald-600 font-medium">服务协议</a> 与 <a href="#" class="text-emerald-600 font-medium">隐私政策</a>
        </p>
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
