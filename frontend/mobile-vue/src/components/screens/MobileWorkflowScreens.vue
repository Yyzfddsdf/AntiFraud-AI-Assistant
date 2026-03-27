<template>
  <div v-if="state.activeTab === 'submit'" data-mobile-scroll="submit-page" class="bg-slate-50 pt-3 pb-28">
    <div class="px-5 mb-4">
      <h2 class="text-2xl font-black text-slate-900 tracking-tight">智能检测</h2>
      <p class="text-xs font-bold text-slate-500 mt-1">提交可疑信息，AI 护航实时为您排查风险</p>
    </div>

    <div class="px-4 space-y-4">
      <div class="bg-white rounded-[24px] p-5 shadow-sm border border-slate-100">
        <div class="mb-3 flex items-center gap-2">
          <div class="w-1.5 h-1.5 rounded-full bg-emerald-500"></div>
          <span class="text-sm font-bold text-slate-800">可疑内容描述</span>
        </div>
        <textarea v-model="state.analyzeForm.text" class="w-full h-32 p-4 bg-slate-50 rounded-2xl border-none focus:ring-2 focus:ring-emerald-500 text-[15px] leading-relaxed resize-none text-slate-800 placeholder-slate-400 transition-all" placeholder="请粘贴可疑的聊天记录、短信、链接或描述遇到的情况..."></textarea>

        <div class="mt-4 pt-4 border-t border-slate-50">
          <div class="mb-3 flex items-center gap-2">
            <div class="w-1.5 h-1.5 rounded-full bg-emerald-500"></div>
            <span class="text-sm font-bold text-slate-800">上传附件 <span class="text-slate-400 font-normal">(选填)</span></span>
          </div>
          <div class="grid grid-cols-3 gap-3">
            <label class="mobile-upload-card" for="mobile-upload-image">
              <input id="mobile-upload-image" type="file" multiple accept="image/*" class="hidden" @change="state.handleFileSelect($event, 'images')">
              <div class="mobile-upload-card__icon">
                <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" aria-hidden="true">
                  <path d="M10 1C9.73478 1 9.48043 1.10536 9.29289 1.29289L3.29289 7.29289C3.10536 7.48043 3 7.73478 3 8V20C3 21.6569 4.34315 23 6 23H7C7.55228 23 8 22.5523 8 22C8 21.4477 7.55228 21 7 21H6C5.44772 21 5 20.5523 5 20V9H10C10.5523 9 11 8.55228 11 8V3H18C18.5523 3 19 3.44772 19 4V9C19 9.55228 19.4477 10 20 10C20.5523 10 21 9.55228 21 9V4C21 2.34315 19.6569 1 18 1H10ZM9 7H6.41421L9 4.41421V7ZM14 15.5C14 14.1193 15.1193 13 16.5 13C17.8807 13 19 14.1193 19 15.5V16V17H20C21.1046 17 22 17.8954 22 19C22 20.1046 21.1046 21 20 21H13C11.8954 21 11 20.1046 11 19C11 17.8954 11.8954 17 13 17H14V16V15.5ZM16.5 11C14.142 11 12.2076 12.8136 12.0156 15.122C10.2825 15.5606 9 17.1305 9 19C9 21.2091 10.7909 23 13 23H20C22.2091 23 24 21.2091 24 19C24 17.1305 22.7175 15.5606 20.9844 15.122C20.7924 12.8136 18.858 11 16.5 11Z" clip-rule="evenodd" fill-rule="evenodd"></path>
                </svg>
                <span class="mobile-upload-card__marker" aria-hidden="true">
                  <svg viewBox="0 0 24 24">
                    <circle cx="8" cy="8" r="1.6"></circle>
                    <path d="m4.5 18 5.2-5.1 3.4 3 2.8-2.7 3.6 4.8"></path>
                  </svg>
                </span>
              </div>
              <span class="mobile-upload-card__text">图片</span>
              <span v-if="state.analyzeForm.images.length" class="absolute -top-1.5 -right-1.5 w-5 h-5 bg-emerald-500 text-white text-[10px] font-bold flex items-center justify-center rounded-full shadow-sm ring-2 ring-white">{{ state.analyzeForm.images.length }}</span>
            </label>
            <label class="mobile-upload-card" for="mobile-upload-audio">
              <input id="mobile-upload-audio" type="file" multiple accept="audio/*" class="hidden" @change="state.handleFileSelect($event, 'audios')">
              <div class="mobile-upload-card__icon">
                <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" aria-hidden="true">
                  <path d="M10 1C9.73478 1 9.48043 1.10536 9.29289 1.29289L3.29289 7.29289C3.10536 7.48043 3 7.73478 3 8V20C3 21.6569 4.34315 23 6 23H7C7.55228 23 8 22.5523 8 22C8 21.4477 7.55228 21 7 21H6C5.44772 21 5 20.5523 5 20V9H10C10.5523 9 11 8.55228 11 8V3H18C18.5523 3 19 3.44772 19 4V9C19 9.55228 19.4477 10 20 10C20.5523 10 21 9.55228 21 9V4C21 2.34315 19.6569 1 18 1H10ZM9 7H6.41421L9 4.41421V7ZM14 15.5C14 14.1193 15.1193 13 16.5 13C17.8807 13 19 14.1193 19 15.5V16V17H20C21.1046 17 22 17.8954 22 19C22 20.1046 21.1046 21 20 21H13C11.8954 21 11 20.1046 11 19C11 17.8954 11.8954 17 13 17H14V16V15.5ZM16.5 11C14.142 11 12.2076 12.8136 12.0156 15.122C10.2825 15.5606 9 17.1305 9 19C9 21.2091 10.7909 23 13 23H20C22.2091 23 24 21.2091 24 19C24 17.1305 22.7175 15.5606 20.9844 15.122C20.7924 12.8136 18.858 11 16.5 11Z" clip-rule="evenodd" fill-rule="evenodd"></path>
                </svg>
                <span class="mobile-upload-card__marker" aria-hidden="true">
                  <svg viewBox="0 0 24 24">
                    <path d="M6.5 15.8v-3.6"></path>
                    <path d="M10 17.2V8.8"></path>
                    <path d="M13.5 15.8v-5.2"></path>
                    <path d="M17 16.6v-7"></path>
                  </svg>
                </span>
              </div>
              <span class="mobile-upload-card__text">音频</span>
              <span v-if="state.analyzeForm.audios.length" class="absolute -top-1.5 -right-1.5 w-5 h-5 bg-emerald-500 text-white text-[10px] font-bold flex items-center justify-center rounded-full shadow-sm ring-2 ring-white">{{ state.analyzeForm.audios.length }}</span>
            </label>
            <label class="mobile-upload-card" for="mobile-upload-video">
              <input id="mobile-upload-video" type="file" multiple accept="video/*" class="hidden" @change="state.handleFileSelect($event, 'videos')">
              <div class="mobile-upload-card__icon">
                <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" aria-hidden="true">
                  <path d="M10 1C9.73478 1 9.48043 1.10536 9.29289 1.29289L3.29289 7.29289C3.10536 7.48043 3 7.73478 3 8V20C3 21.6569 4.34315 23 6 23H7C7.55228 23 8 22.5523 8 22C8 21.4477 7.55228 21 7 21H6C5.44772 21 5 20.5523 5 20V9H10C10.5523 9 11 8.55228 11 8V3H18C18.5523 3 19 3.44772 19 4V9C19 9.55228 19.4477 10 20 10C20.5523 10 21 9.55228 21 9V4C21 2.34315 19.6569 1 18 1H10ZM9 7H6.41421L9 4.41421V7ZM14 15.5C14 14.1193 15.1193 13 16.5 13C17.8807 13 19 14.1193 19 15.5V16V17H20C21.1046 17 22 17.8954 22 19C22 20.1046 21.1046 21 20 21H13C11.8954 21 11 20.1046 11 19C11 17.8954 11.8954 17 13 17H14V16V15.5ZM16.5 11C14.142 11 12.2076 12.8136 12.0156 15.122C10.2825 15.5606 9 17.1305 9 19C9 21.2091 10.7909 23 13 23H20C22.2091 23 24 21.2091 24 19C24 17.1305 22.7175 15.5606 20.9844 15.122C20.7924 12.8136 18.858 11 16.5 11Z" clip-rule="evenodd" fill-rule="evenodd"></path>
                </svg>
                <span class="mobile-upload-card__marker" aria-hidden="true">
                  <svg viewBox="0 0 24 24">
                    <path d="m9 8 7 4-7 4z"></path>
                  </svg>
                </span>
              </div>
              <span class="mobile-upload-card__text">视频</span>
              <span v-if="state.analyzeForm.videos.length" class="absolute -top-1.5 -right-1.5 w-5 h-5 bg-emerald-500 text-white text-[10px] font-bold flex items-center justify-center rounded-full shadow-sm ring-2 ring-white">{{ state.analyzeForm.videos.length }}</span>
            </label>
          </div>
        </div>
      </div>

      <button @click="state.submitAnalysis" :disabled="state.analyzing" class="w-full h-14 rounded-2xl bg-slate-900 text-white text-[16px] font-bold shadow-lg shadow-slate-900/20 active:scale-[0.98] transition-all disabled:opacity-70 flex items-center justify-center gap-2">
        <svg v-if="state.analyzing" class="animate-spin w-5 h-5" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path></svg>
        <svg v-else class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z"></path></svg>
        <span>{{ state.analyzing ? '正在深度分析中...' : '开始全面检测' }}</span>
      </button>
    </div>
  </div>

  <div v-if="state.activeTab === 'simulation_quiz' && state.simulationViewMode === 'overview'" data-mobile-scroll="simulation-overview" class="fixed inset-x-0 overflow-y-auto overflow-x-hidden bg-slate-50 z-20" style="-webkit-overflow-scrolling: touch; overscroll-behavior: contain; top: 0; bottom: calc(4.5rem + env(safe-area-inset-bottom));">
    <div class="fixed inset-x-0 z-40 bg-white/90 backdrop-blur-lg px-4 pb-3 border-b border-slate-100 shadow-sm" style="top: 0; padding-top: calc(env(safe-area-inset-top) + 0.5rem);">
      <div class="flex items-center justify-between">
        <div class="flex items-center gap-3">
          <button @click="state.activeTab = 'tasks'" class="w-8 h-8 rounded-full bg-slate-100 text-slate-600 flex items-center justify-center active:scale-90 transition-transform" aria-label="返回">
            <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7"></path></svg>
          </button>
          <h2 class="text-xl font-black text-slate-900 tracking-tight">实景防骗演练</h2>
        </div>
      </div>
    </div>

    <div class="p-4 space-y-5" style="margin-top: calc(env(safe-area-inset-top) + 3.5rem);">
      <section>
        <div class="flex items-center justify-between mb-3">
          <h3 class="text-sm font-bold text-slate-800 ml-1">定制演练场景</h3>
        </div>
        <div class="bg-white rounded-3xl p-4 shadow-sm border border-slate-100 space-y-4">
          <div class="space-y-3">
            <div class="relative">
              <div class="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
                <svg class="w-5 h-5 text-slate-400" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"></path></svg>
              </div>
              <input v-model="state.simulationForm.caseType" type="text" placeholder="场景（如：冒充公检法、更新软件）" class="w-full h-11 pl-10 pr-3 rounded-xl bg-slate-50 border-none focus:ring-2 focus:ring-emerald-500 text-sm text-slate-700 placeholder-slate-400">
            </div>
            <div class="relative">
              <div class="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
                <svg class="w-5 h-5 text-slate-400" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z"></path></svg>
              </div>
              <input v-model="state.simulationForm.targetPersona" type="text" placeholder="目标身份（如：老人、学生）" class="w-full h-11 pl-10 pr-3 rounded-xl bg-slate-50 border-none focus:ring-2 focus:ring-emerald-500 text-sm text-slate-700 placeholder-slate-400">
            </div>
          </div>

          <div class="grid grid-cols-3 gap-2 bg-slate-50 p-1 rounded-xl">
            <button @click="state.simulationForm.difficulty = 'easy'" :class="['h-9 rounded-lg text-xs font-bold transition-all', state.simulationForm.difficulty === 'easy' ? 'bg-white text-emerald-600 shadow-sm' : 'text-slate-500']">简单</button>
            <button @click="state.simulationForm.difficulty = 'medium'" :class="['h-9 rounded-lg text-xs font-bold transition-all', state.simulationForm.difficulty === 'medium' ? 'bg-white text-amber-600 shadow-sm' : 'text-slate-500']">中等</button>
            <button @click="state.simulationForm.difficulty = 'hard'" :class="['h-9 rounded-lg text-xs font-bold transition-all', state.simulationForm.difficulty === 'hard' ? 'bg-white text-rose-600 shadow-sm' : 'text-slate-500']">困难</button>
          </div>

          <button @click="state.generateSimulationPack" :disabled="state.simulationGenerating" class="w-full h-12 rounded-xl bg-slate-900 text-white text-[15px] font-bold shadow-md active:scale-[0.98] transition-all disabled:opacity-70 flex items-center justify-center gap-2">
            <svg v-if="state.simulationGenerating" class="animate-spin w-5 h-5" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path></svg>
            <span>{{ state.simulationGenerating ? '正在智能生成...' : '生成专属演练' }}</span>
          </button>
        </div>
      </section>

      <section>
        <div class="flex items-center justify-between mb-3 px-1">
          <h3 class="text-sm font-bold text-slate-800">待挑战题库</h3>
          <button @click="state.fetchSimulationPacks" class="text-xs text-emerald-600 font-bold flex items-center gap-1 active:opacity-50">
            <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"></path></svg>
            刷新
          </button>
        </div>
        <div class="space-y-3">
          <div v-for="item in state.simulationPackList" :key="`m-pack-${item.pack_id}`" class="bg-white rounded-2xl p-4 shadow-sm border border-slate-100 flex flex-col gap-3">
            <div>
              <div class="flex items-center gap-2 mb-1">
                <span class="px-2 py-0.5 rounded-md bg-slate-100 text-slate-600 text-[10px] font-bold">{{ item.case_type }}</span>
                <span :class="['px-2 py-0.5 rounded-md text-[10px] font-bold', item.difficulty === 'hard' ? 'bg-rose-50 text-rose-600' : item.difficulty === 'medium' ? 'bg-amber-50 text-amber-600' : 'bg-emerald-50 text-emerald-600']">{{ item.difficulty === 'hard' ? '困难' : item.difficulty === 'medium' ? '中等' : '简单' }}</span>
              </div>
              <h4 class="text-[15px] font-bold text-slate-900 leading-snug">{{ item.title }}</h4>
            </div>
            <button @click="state.startSimulationSession(item.pack_id)" :disabled="state.simulationSubmitting" class="w-full h-10 rounded-xl bg-emerald-50 text-emerald-700 text-sm font-bold active:bg-emerald-100 transition-colors disabled:opacity-50 flex items-center justify-center gap-1">
              开始挑战
              <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M14 5l7 7m0 0l-7 7m7-7H3"></path></svg>
            </button>
          </div>
          <div v-if="!state.simulationPackList.length" class="text-xs text-slate-400 text-center py-6 bg-white rounded-2xl border border-slate-100 border-dashed">
            暂无待挑战的演练
          </div>
        </div>
      </section>

      <section>
        <div class="flex items-center justify-between mb-3 px-1">
          <h3 class="text-sm font-bold text-slate-800">演练记录</h3>
          <button @click="state.fetchSimulationSessions" class="text-xs text-emerald-600 font-bold active:opacity-50">刷新</button>
        </div>
        <div class="space-y-3">
          <div v-for="item in state.simulationSessionList" :key="`m-session-${item.pack_id}`" class="bg-white rounded-2xl p-4 shadow-sm border border-slate-100 flex items-center gap-3">
            <div class="w-12 h-12 shrink-0 rounded-full flex items-center justify-center font-black text-lg" :class="item.score >= 80 ? 'bg-emerald-100 text-emerald-600' : item.score >= 60 ? 'bg-amber-100 text-amber-600' : 'bg-rose-100 text-rose-600'">
              {{ item.score }}
            </div>
            <div class="flex-1 min-w-0">
              <h4 class="text-sm font-bold text-slate-900 truncate">{{ item.title || '未知演练' }}</h4>
              <div class="text-[11px] text-slate-500 mt-0.5 flex items-center gap-2">
                <span>评级: {{ item.level || '未评分' }}</span>
                <span class="w-1 h-1 rounded-full bg-slate-300"></span>
                <span :class="item.status === 'completed' ? 'text-emerald-500' : 'text-amber-500'">{{ item.status === 'completed' ? '已完成' : '未完成' }}</span>
              </div>
            </div>
            <button v-if="item.status !== 'completed'" @click="state.startSimulationSession(item.pack_id)" class="w-8 h-8 shrink-0 rounded-full bg-emerald-50 text-emerald-600 flex items-center justify-center active:scale-90 transition-transform">
              <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M14 5l7 7m0 0l-7 7m7-7H3"></path></svg>
            </button>
            <button v-else @click="state.deleteSimulationSession(item.pack_id)" class="w-8 h-8 shrink-0 rounded-full bg-slate-50 text-slate-400 flex items-center justify-center active:scale-90 transition-transform">
              <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"></path></svg>
            </button>
          </div>
          <div v-if="!state.simulationSessionList.length" class="text-xs text-slate-400 text-center py-6 bg-white rounded-2xl border border-slate-100 border-dashed">
            暂无演练记录
          </div>
        </div>
      </section>

      <div class="h-6 shrink-0"></div>
    </div>
  </div>

  <div v-if="state.activeTab === 'simulation_quiz' && state.simulationViewMode === 'exam'" class="fixed inset-0 z-[1000] flex flex-col bg-slate-50 animate-slide-up" style="padding-bottom: env(safe-area-inset-bottom);">
    <div class="shrink-0 bg-white/80 backdrop-blur-md z-10 px-4 pt-safe pb-3 flex flex-col gap-3 sticky top-0 border-b border-slate-100">
      <div class="flex items-center justify-between mt-2">
        <button @click="state.closeSimulationExamView" class="w-8 h-8 rounded-full bg-slate-100 text-slate-500 flex items-center justify-center active:scale-90 transition-transform" aria-label="退出">
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"></path></svg>
        </button>
        <div class="font-bold text-slate-800 text-base">模拟演练</div>
        <div class="text-sm font-bold text-emerald-600 flex items-center gap-1">
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z"></path></svg>
          {{ state.simulationCurrentScore }}
        </div>
      </div>
      <div v-if="state.simulationPack && state.simulationPack.steps" class="flex items-center gap-3">
        <div class="flex-1 h-2 rounded-full bg-slate-200 overflow-hidden">
          <div class="h-full rounded-full bg-emerald-500 transition-all duration-500 ease-out" :style="{ width: `${Math.max(5, (((state.simulationAnswers?.length || 0) + 1) / state.simulationPack.steps.length) * 100)}%` }"></div>
        </div>
        <span class="text-xs font-bold text-slate-400 shrink-0 w-8 text-right">
          {{ Math.min((state.simulationAnswers?.length || 0) + 1, state.simulationPack.steps.length) }}/{{ state.simulationPack.steps.length }}
        </span>
      </div>
    </div>

    <div data-mobile-scroll="simulation-exam" class="flex-1 overflow-y-auto p-4 flex flex-col" style="-webkit-overflow-scrolling: touch; overscroll-behavior: contain;">
      <div class="flex-1 flex flex-col justify-center max-w-md mx-auto w-full space-y-6 pb-6">
        <div v-if="state.simulationPack && (state.simulationAnswers?.length || 0) === 0 && state.simulationStatus === 'in_progress'" class="text-center space-y-2 mb-4 animate-fade-in">
          <div class="inline-block px-3 py-1 bg-emerald-100 text-emerald-800 text-xs font-bold rounded-full mb-2">任务目标</div>
          <h2 class="text-2xl font-black text-slate-900 leading-tight">{{ state.simulationPack.title }}</h2>
          <p class="text-sm text-slate-500 leading-relaxed">{{ state.simulationPack.intro }}</p>
        </div>

        <div v-if="state.simulationCurrentStep && state.simulationStatus === 'in_progress'" class="space-y-6 animate-slide-up" :key="state.simulationCurrentStep.step_id">
          <div class="flex flex-col items-start gap-1">
            <div class="text-[11px] font-bold text-slate-400 ml-1">场景提示 · {{ state.simulationCurrentStep.step_type }}</div>
            <div class="bg-white border border-slate-100 shadow-sm rounded-2xl rounded-tl-sm p-4 text-slate-700 text-[15px] leading-relaxed relative">
              {{ state.simulationCurrentStep.narrative }}
            </div>
          </div>

          <div class="text-[19px] font-black leading-snug text-slate-900 px-1">
            {{ state.simulationCurrentStep.question }}
          </div>

          <div class="space-y-3 pt-2">
            <button v-for="option in state.simulationCurrentStep.options" :key="`m-exam-${state.simulationCurrentStep.step_id}-${option.key}`" @click="state.submitSimulationAnswer(option.key)" :disabled="state.simulationSubmitting || state.simulationStatus !== 'in_progress'" class="w-full text-left rounded-2xl border-2 border-slate-100 bg-white p-4 active:scale-[0.98] disabled:opacity-50 transition-all duration-200 flex items-center gap-4 hover:border-emerald-500 hover:bg-emerald-50/30 group">
              <div class="w-10 h-10 shrink-0 rounded-full border-2 border-slate-100 text-slate-400 flex items-center justify-center text-sm font-black group-hover:border-emerald-500 group-hover:bg-emerald-500 group-hover:text-white transition-colors">
                {{ option.key }}
              </div>
              <div class="flex-1 text-[15px] font-bold text-slate-700 group-hover:text-slate-900 leading-snug">
                {{ option.text }}
              </div>
            </button>
          </div>
        </div>

        <div v-if="state.simulationStatus === 'completed' && state.simulationResult" class="flex flex-col items-center justify-center h-full space-y-6 animate-fade-in text-center mt-8">
          <div class="w-24 h-24 rounded-full bg-emerald-100 flex items-center justify-center mb-2">
            <svg class="w-12 h-12 text-emerald-600" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="3" d="M5 13l4 4L19 7"></path></svg>
          </div>
          <div>
            <h2 class="text-3xl font-black text-slate-900 mb-2">完成挑战</h2>
            <div class="text-lg font-bold text-emerald-600">评级：{{ state.simulationResult.level }} · 得分：{{ state.simulationResult.total_score }}</div>
          </div>

          <div class="w-full bg-white rounded-3xl border border-slate-100 shadow-sm p-5 text-left mt-4 space-y-3">
            <div class="text-sm font-black text-slate-900 border-b border-slate-100 pb-2">防骗建议</div>
            <div v-for="(advice, idx) in state.simulationResult.advice" :key="`m-sim-advice-${idx}`" class="flex gap-3 text-[14px] text-slate-600 leading-relaxed">
              <span class="text-emerald-500 font-black">{{ idx + 1 }}.</span>
              <span>{{ advice }}</span>
            </div>
          </div>

          <div class="w-full pt-4">
            <button @click="state.closeSimulationExamView" class="w-full h-14 rounded-2xl bg-slate-900 text-white text-[16px] font-bold shadow-lg active:scale-95 transition-transform">
              返回题包列表
            </button>
          </div>
        </div>
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

<style scoped>
.mobile-upload-card {
  position: relative;
  aspect-ratio: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  border: 2px dashed #cacaca;
  background: #ffffff;
  border-radius: 1.15rem;
  box-shadow: 0 32px 28px -32px rgba(0, 0, 0, 0.1);
  transition:
    transform 0.18s ease,
    border-color 0.18s ease,
    background-color 0.18s ease;
}

.mobile-upload-card:active {
  transform: scale(0.98);
  border-color: #bfc5ce;
  background-color: #fafafa;
}

.mobile-upload-card__icon {
  position: relative;
  display: flex;
  align-items: center;
  justify-content: center;
}

.mobile-upload-card__icon > svg {
  width: 2.8rem;
  height: 2.8rem;
  fill: rgba(75, 85, 99, 1);
}

.mobile-upload-card__marker {
  position: absolute;
  right: -0.2rem;
  bottom: 0.05rem;
  width: 1.2rem;
  height: 1.2rem;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 999px;
  background: #ffffff;
  border: 1px solid #d7dbe2;
  box-shadow: 0 6px 12px rgba(0, 0, 0, 0.08);
}

.mobile-upload-card__marker svg {
  width: 0.75rem;
  height: 0.75rem;
  fill: none;
  stroke: rgba(75, 85, 99, 1);
  stroke-width: 2;
  stroke-linecap: round;
  stroke-linejoin: round;
}

.mobile-upload-card__text {
  font-size: 0.68rem;
  font-weight: 600;
  color: rgba(75, 85, 99, 1);
  line-height: 1;
}
</style>
