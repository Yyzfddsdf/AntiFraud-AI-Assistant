<template>
  <div v-show="state.activeTab === 'chat'" class="fixed inset-0 z-[1000] flex flex-col bg-white">
    <div class="h-14 px-4 flex items-center justify-between border-b border-gray-100 shrink-0 bg-white/90 backdrop-blur-md pt-safe z-30">
      <div class="flex items-center gap-2">
        <button @click="state.activeTab = 'tasks'" class="text-slate-500 hover:bg-slate-100 p-1.5 rounded-lg transition-colors">
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7"></path></svg>
        </button>
        <div class="flex flex-col">
          <h2 class="text-sm font-bold text-slate-800">Sentinel AI</h2>
          <span class="text-[10px] text-green-500 font-medium flex items-center gap-1">
            <span class="w-1.5 h-1.5 bg-green-500 rounded-full animate-pulse"></span>
            Online
          </span>
        </div>
      </div>
      <button @click="state.clearChatHistory" class="text-slate-400 hover:text-red-500 p-2 rounded-lg hover:bg-red-50 transition-all" title="清空对话">
        <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"></path></svg>
      </button>
    </div>

    <div id="chat-container" class="flex-1 overflow-y-auto overflow-x-hidden min-h-0 p-3 space-y-4 bg-white scroll-smooth pb-4">
      <div v-if="state.chatMessages.length === 0" class="flex flex-col items-center justify-center h-[56vh] text-center px-5 opacity-0 animate-[fadeIn_0.5s_ease-out_forwards]">
        <div class="w-16 h-16 bg-gradient-to-br from-emerald-400 to-teal-500 rounded-full flex items-center justify-center mb-4 shadow-lg shadow-emerald-500/30">
          <svg class="w-8 h-8 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z"></path></svg>
        </div>
        <h3 class="text-lg font-bold text-slate-800 mb-1">我是您的反诈助手</h3>
        <p class="text-sm text-slate-500 max-w-[260px] leading-relaxed">可帮您识别诈骗信息、分析风险案例或提供安全建议。直接发送文字或图片即可。</p>
      </div>

      <div v-for="(msg, idx) in state.chatMessages" :key="idx" :class="['flex gap-3 group', msg.type === 'user' ? 'flex-row-reverse' : 'flex-row']">
        <div class="shrink-0 flex flex-col items-center">
          <div :class="['w-8 h-8 rounded-full flex items-center justify-center text-xs font-bold shadow-sm border border-white/20', msg.type === 'user' ? 'bg-slate-200 text-slate-600' : 'bg-gradient-to-br from-emerald-400 to-teal-500 text-white']">
            <span v-if="msg.type === 'user'">你</span>
            <svg v-else class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z"></path></svg>
          </div>
        </div>

        <div :class="['flex flex-col max-w-[86%] min-w-0', msg.type === 'user' ? 'items-end' : 'items-start']">
          <div v-if="msg.type !== 'tool'" class="text-xs font-bold text-slate-400 mb-1 px-1 opacity-0 group-hover:opacity-100 transition-opacity">
            {{ msg.type === 'user' ? '你' : '反诈助手' }}
          </div>

          <div v-if="msg.type === 'tool'" class="flex items-center gap-2 text-[11px] font-mono text-slate-500 my-1 px-2.5 py-1.5 bg-slate-50 rounded-lg border border-slate-200/60 w-full">
            <div class="w-2 h-2 bg-amber-400 rounded-full animate-pulse"></div>
            <span class="truncate">{{ msg.content }}</span>
          </div>

          <div v-else :class="['relative px-3 py-2.5 text-[14px] leading-6 shadow-sm max-w-full overflow-hidden', msg.type === 'user' ? 'bg-[#f4f4f4] text-slate-800 rounded-2xl rounded-tr-sm' : msg.type === 'error' ? 'bg-red-50 text-red-700 border border-red-100 rounded-2xl rounded-tl-sm' : 'bg-transparent text-slate-800 p-0 shadow-none']">
            <div v-if="msg.type === 'ai' && msg.rendered_content" class="break-words break-all max-w-full overflow-hidden markdown-body" v-html="msg.rendered_content"></div>
            <div v-else-if="msg.content" class="whitespace-pre-wrap break-words break-all max-w-full overflow-hidden">{{ msg.content }}</div>
            <div v-if="msg.images && msg.images.length" class="mt-2 grid grid-cols-2 gap-2">
              <button v-for="(img, imgIdx) in msg.images" :key="`${idx}-${imgIdx}`" type="button" class="relative group/img rounded-xl overflow-hidden border border-black/5 shadow-sm transition-transform hover:scale-[1.02] active:scale-95" @click="state.openImage(img)">
                <img :src="img" class="w-full h-28 object-cover bg-white">
                <div class="absolute inset-0 bg-black/0 group-hover/img:bg-black/10 transition-colors"></div>
              </button>
            </div>
          </div>
        </div>
      </div>

      <div v-if="state.isChatting" class="flex gap-3">
        <div class="w-8 h-8 rounded-full bg-gradient-to-br from-emerald-400 to-teal-500 flex items-center justify-center text-white shadow-sm shrink-0">
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z"></path></svg>
        </div>
        <div class="flex items-center h-7">
          <div class="flex gap-1">
            <span class="w-2.5 h-2.5 bg-emerald-400 rounded-full animate-[bounce_1.4s_infinite_ease-in-out_both] delay-0"></span>
            <span class="w-2.5 h-2.5 bg-emerald-400 rounded-full animate-[bounce_1.4s_infinite_ease-in-out_both] delay-150"></span>
            <span class="w-2.5 h-2.5 bg-emerald-400 rounded-full animate-[bounce_1.4s_infinite_ease-in-out_both] delay-300"></span>
          </div>
        </div>
      </div>
    </div>

    <div class="shrink-0 bg-white border-t border-gray-100 p-3 z-40" style="padding-bottom: max(0.75rem, env(safe-area-inset-bottom));">
      <div v-if="state.chatImages.length" class="flex gap-2 overflow-x-auto pb-2 mb-1 px-1">
        <div v-for="(img, idx) in state.chatImages" :key="idx" class="relative w-14 h-14 shrink-0 rounded-xl overflow-hidden border border-gray-200 shadow-sm group">
          <img :src="img" class="w-full h-full object-cover">
          <button @click="state.removeChatImage(idx)" class="absolute top-1 right-1 w-5 h-5 bg-black/60 text-white rounded-full flex items-center justify-center backdrop-blur-sm hover:bg-red-500 transition-colors"><svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"></path></svg></button>
        </div>
      </div>

      <form @submit.prevent="state.sendChatMessage" class="relative max-w-3xl mx-auto">
        <div class="relative flex items-end gap-2 bg-gray-50 border border-gray-200 rounded-[22px] p-2 shadow-sm focus-within:ring-2 focus-within:ring-emerald-500/20 focus-within:border-emerald-500 transition-all">
          <button type="button" @click="state.triggerChatImagePicker" :disabled="state.isChatting" class="p-2 rounded-full text-slate-400 hover:text-slate-600 hover:bg-gray-200/50 transition-colors shrink-0">
            <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4"></path></svg>
          </button>
          <input id="chat-image-input" type="file" accept="image/*" multiple class="hidden" @change="state.handleChatImageSelect">
          <textarea v-model="state.chatInput" placeholder="发送消息..." rows="1" class="w-full bg-transparent border-none outline-none text-[14px] text-slate-800 placeholder:text-slate-400 py-2 max-h-28 resize-none leading-6" @keydown.enter.prevent="!state.isChatting && (state.chatInput.trim() || state.chatImages.length) ? state.sendChatMessage() : null"></textarea>
          <button type="submit" :disabled="(!state.chatInput.trim() && !state.chatImages.length) || state.isChatting" class="p-2.5 rounded-full bg-emerald-600 text-white disabled:opacity-20 disabled:bg-slate-300 transition-all shrink-0 hover:bg-emerald-700 shadow-sm">
            <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 12h14M12 5l7 7-7 7"></path></svg>
          </button>
        </div>
        <div class="text-center mt-2">
          <p class="text-[10px] text-slate-400">Sentinel AI 可能会产生错误信息，请核实重要信息。</p>
        </div>
      </form>
    </div>
  </div>

  <div v-show="state.activeTab === 'alerts'" class="bg-slate-50 pt-3 pb-safe">
    <div class="px-5 mb-4">
      <h2 class="text-2xl font-black text-slate-900 tracking-tight">消息中心</h2>
      <p class="text-xs font-bold text-slate-500 mt-1">您有 {{ state.alertUnreadCount }} 条未读风险预警</p>
    </div>
    <div class="px-4 space-y-3">
      <div v-if="state.recentRiskAlerts.length === 0" class="flex flex-col items-center justify-center py-24 text-slate-400 bg-white rounded-[24px] border border-slate-100 shadow-sm">
        <div class="w-16 h-16 rounded-full bg-slate-50 flex items-center justify-center mb-3">
          <svg class="w-8 h-8 text-slate-300" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4 0v.341C7.67 6.165 6 8.388 6 11v3.159c0 .538-.214 1.055-.595 1.436L4 17h5m6 0v1a3 3 0 11-6 0v-1m6 0H9"></path></svg>
        </div>
        <p class="text-sm font-bold text-slate-500">暂无风险预警</p>
        <p class="text-xs text-slate-400 mt-1">您的环境非常安全</p>
      </div>
      <div v-for="item in state.recentRiskAlerts" :key="item.record_id" @click="state.openAlertCaseDetail(item)" class="bg-white rounded-3xl p-4 shadow-sm border border-slate-100 active:scale-[0.98] transition-transform relative overflow-hidden group">
        <div v-if="item.unread" :class="['absolute top-0 left-0 w-1.5 h-full', state.getAlertSeverityTheme(item.risk_level).barClass]"></div>
        <div class="flex justify-between items-start mb-2 pl-1">
          <span :class="['inline-block px-2.5 py-1 text-[10px] font-black uppercase tracking-widest rounded-lg', state.getAlertSeverityTheme(item.risk_level).badgeClass]">{{ item.risk_level }} Risk</span>
          <span class="text-[11px] font-medium text-slate-400">{{ state.formatTime(item.created_at || item.sent_at) }}</span>
        </div>
        <h3 class="font-black text-[15px] mb-1.5 truncate text-slate-900 pl-1">{{ item.title }}</h3>
        <p class="text-sm text-slate-500 line-clamp-2 leading-relaxed pl-1">{{ item.case_summary }}</p>
      </div>
    </div>
  </div>

  <div v-show="state.activeTab === 'family'" class="bg-slate-50 pb-24">
    <!-- Header -->
    <div class="sticky top-0 z-50 bg-white border-b border-slate-100 pt-safe">
      <div class="flex items-center justify-center px-4 h-14 relative">
        <h2 class="text-[17px] font-bold text-slate-900 tracking-tight">家庭守护</h2>
      </div>
    </div>

    <div class="px-4 py-5 space-y-4">
      <div v-if="state.familyLoading" class="bg-white rounded-[24px] p-6 text-center shadow-sm border border-slate-100/60">
        <div class="flex flex-col items-center justify-center">
          <svg class="animate-spin w-6 h-6 text-emerald-500 mb-3" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path></svg>
          <span class="text-[13px] font-bold text-slate-500">正在加载家庭数据...</span>
        </div>
      </div>

      <div v-else-if="!state.familyHasGroup" class="space-y-4">
        <div class="w-16 h-16 bg-white shadow-sm border border-slate-100/60 rounded-[20px] flex items-center justify-center mx-auto mb-2 text-emerald-500">
          <svg class="w-8 h-8" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197M13 7a4 4 0 11-8 0 4 4 0 018 0z"></path></svg>
        </div>

        <div class="bg-white rounded-[24px] p-5 shadow-sm border border-slate-100/60">
          <div class="flex items-center justify-between mb-4">
            <div class="flex items-center gap-2">
              <div class="w-1.5 h-1.5 rounded-full bg-emerald-500"></div>
              <h3 class="font-extrabold text-[15px] text-slate-900">收到的邀请</h3>
            </div>
            <span class="text-[10px] font-bold text-emerald-600 bg-emerald-50 px-2 py-0.5 rounded-md">{{ state.familyReceivedInvitations.length }} 条</span>
          </div>
          <div v-if="state.familyReceivedLoading && state.familyReceivedInvitations.length === 0" class="text-xs font-bold text-slate-400 py-3 text-center">正在加载邀请...</div>
          <div v-else-if="state.familyReceivedInvitations.length === 0" class="text-[11px] font-bold text-slate-400 py-6 text-center bg-slate-50/50 rounded-[16px] border border-slate-100 border-dashed">当前没有收到新的家庭邀请</div>
          <div v-else class="space-y-3">
            <div v-for="invitation in state.familyReceivedInvitations" :key="`received-${invitation.id}`" class="rounded-[20px] border border-emerald-100 bg-emerald-50/30 p-4 relative overflow-hidden group">
              <div class="absolute top-0 right-0 w-24 h-24 bg-emerald-100/30 rounded-bl-full opacity-50"></div>
              <div class="flex items-start justify-between gap-3 relative z-10">
                <div class="min-w-0">
                  <div class="text-[15px] font-black text-slate-900 tracking-tight">{{ invitation.family_name || '家庭邀请' }}</div>
                  <div class="text-[11px] font-bold text-slate-500 mt-1.5 flex items-center gap-1.5"><svg class="w-3.5 h-3.5 text-slate-400" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z"></path></svg> {{ invitation.inviter_name || invitation.inviter_email || invitation.inviter_phone || '未知' }}</div>
                  <div class="text-[11px] font-bold text-slate-500 mt-1 flex items-center gap-1.5">
                    <span class="bg-slate-100 px-1.5 py-0.5 rounded text-slate-600">角色: {{ invitation.role }}</span>
                    <span class="bg-slate-100 px-1.5 py-0.5 rounded text-slate-600">关系: {{ invitation.relation || '未填写' }}</span>
                  </div>
                </div>
                <span :class="['shrink-0 text-[10px] font-black px-2 py-1 rounded-md tracking-widest', invitation.status === 'pending' ? 'bg-emerald-100 text-emerald-700 shadow-sm' : 'bg-slate-100 text-slate-500']">{{ invitation.status === 'pending' ? '待处理' : invitation.status }}</span>
              </div>
              <div class="mt-4 bg-white rounded-xl border border-slate-100/80 p-3 flex items-center justify-between relative z-10 shadow-sm">
                <div class="text-[10px] uppercase tracking-[0.2em] text-slate-400 font-bold">邀请码</div>
                <div class="font-mono text-[13px] font-black text-emerald-700 tracking-[0.1em]">{{ invitation.invite_code }}</div>
              </div>
              <button @click="state.acceptFamilyInvitation(invitation.invite_code, invitation.id)" :disabled="invitation.status !== 'pending' || state.familyAcceptingInvitations[invitation.id]" class="w-full mt-3 h-12 rounded-xl bg-slate-900 text-white text-[13px] font-bold shadow-md active:scale-[0.98] transition-all disabled:opacity-50 relative z-10">
                {{ state.familyAcceptingInvitations[invitation.id] ? '加入中...' : '接受邀请' }}
              </button>
            </div>
          </div>
        </div>

        <div class="bg-white rounded-[24px] p-5 shadow-sm border border-slate-100/60">
          <div class="flex items-center gap-2 mb-4">
            <div class="w-1.5 h-1.5 rounded-full bg-blue-500"></div>
            <h3 class="font-extrabold text-[15px] text-slate-900">创建新家庭</h3>
          </div>
          <div class="relative mb-4">
            <input v-model="state.familyCreateForm.name" type="text" placeholder="给家庭起个名字" class="w-full h-12 pl-4 pr-4 rounded-[16px] bg-slate-50 border border-slate-100 focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500 text-[13px] font-bold text-slate-900 placeholder-slate-400 outline-none transition-all">
          </div>
          <button @click="state.createFamily" class="w-full h-12 rounded-[16px] bg-blue-600 text-white text-[14px] font-bold shadow-md active:scale-[0.98] transition-all">创建</button>
        </div>

        <div class="bg-white rounded-[24px] p-5 shadow-sm border border-slate-100/60">
          <div class="flex items-center gap-2 mb-1.5">
            <div class="w-1.5 h-1.5 rounded-full bg-fuchsia-500"></div>
            <h3 class="font-extrabold text-[15px] text-slate-900">邀请码加入</h3>
          </div>
          <p class="text-[11px] font-bold text-slate-400 mb-4">输入家人发来的邀请码，快速加入已有家庭</p>
          <div class="relative mb-4">
            <input v-model="state.familyAcceptForm.invite_code" type="text" placeholder="输入家庭邀请码" class="w-full h-12 pl-4 pr-4 rounded-[16px] bg-slate-50 border border-slate-100 focus:ring-2 focus:ring-fuchsia-500/20 focus:border-fuchsia-500 text-[13px] font-bold text-slate-900 placeholder-slate-400 uppercase tracking-[0.16em] outline-none transition-all">
          </div>
          <button @click="state.acceptFamilyInvitation" class="w-full h-12 rounded-[16px] bg-slate-900 text-white text-[14px] font-bold shadow-md active:scale-[0.98] transition-all">加入家庭</button>
        </div>
      </div>

      <div v-else class="space-y-4">
        <!-- Family Overview Card -->
        <div class="rounded-[28px] overflow-hidden bg-gradient-to-br from-emerald-500 via-teal-600 to-cyan-700 text-white p-6 shadow-xl shadow-emerald-500/20 relative active:scale-[0.98] transition-transform">
          <div class="absolute top-0 right-0 w-32 h-32 bg-white/10 rounded-bl-full blur-2xl pointer-events-none"></div>
          <div class="absolute -bottom-8 -left-8 w-32 h-32 bg-emerald-400/20 rounded-tr-full blur-xl pointer-events-none"></div>
          
          <div class="flex justify-between items-start gap-3 mb-6 relative z-10">
            <div class="min-w-0">
              <h3 class="font-black text-[22px] tracking-tight truncate drop-shadow-sm">{{ state.familyOverview.family.name }}</h3>
              <p class="mt-2 text-[11px] font-bold text-emerald-50/90 flex items-center gap-2 tracking-wide uppercase">
                <span class="bg-black/10 px-2 py-0.5 rounded-md backdrop-blur-sm border border-white/5">成员 {{ state.familyMembers.length }} 人</span>
                <span class="bg-black/10 px-2 py-0.5 rounded-md backdrop-blur-sm border border-white/5" v-if="state.familyUnreadCount > 0">未读 {{ state.familyUnreadCount }}</span>
              </p>
            </div>
            <div class="w-10 h-10 rounded-full bg-white/10 flex items-center justify-center backdrop-blur-md border border-white/20 shadow-sm shrink-0">
              <svg class="w-5 h-5 text-white" fill="none" stroke="currentColor" stroke-width="2.5" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z"></path></svg>
            </div>
          </div>
          
          <div class="rounded-[16px] border border-white/20 bg-black/15 backdrop-blur-md px-4 py-3 mb-5 relative z-10 flex items-center justify-between">
            <div>
              <div class="text-[10px] uppercase tracking-[0.2em] text-white/70 font-black mb-1">家庭专属邀请码</div>
              <p class="font-mono text-[15px] font-black tracking-widest text-white drop-shadow-sm">{{ state.familyOverview.family.invite_code || '暂无邀请码' }}</p>
            </div>
            <button @click="state.activeTab = 'family_invite'" class="h-8 px-3 rounded-xl bg-white/20 text-[11px] font-bold flex items-center justify-center backdrop-blur-md border border-white/10 active:bg-white/30 transition-colors">
              管理
            </button>
          </div>
          
          <div class="flex -space-x-3 overflow-hidden py-1 relative z-10">
            <div v-for="member in state.familyMembers" :key="member.user_id" class="w-10 h-10 rounded-full border-[2.5px] border-emerald-600 bg-white flex items-center justify-center text-teal-700 font-black text-[13px] relative shadow-sm">
              {{ member.username ? member.username.substring(0,1).toUpperCase() : 'U' }}
              <div v-if="member.risk_status === 'high'" class="absolute -bottom-0.5 -right-0.5 w-3.5 h-3.5 bg-rose-500 border-2 border-white rounded-full shadow-sm"></div>
            </div>
          </div>
        </div>

        <div class="bg-white rounded-[24px] p-5 shadow-sm border border-slate-100/60">
          <div class="flex items-center justify-between mb-4">
            <div class="flex items-center gap-2">
              <div class="w-1.5 h-1.5 rounded-full bg-emerald-500"></div>
              <h3 class="font-extrabold text-[15px] text-slate-900">家庭成员</h3>
            </div>
            <span class="text-[10px] font-bold text-slate-500 bg-slate-50 px-2 py-0.5 rounded-md border border-slate-100">{{ state.familyMembers.length }} 人</span>
          </div>
          <div class="space-y-3">
            <div v-for="member in state.familyMembers" :key="member.member_id || member.user_id" class="flex items-center gap-3 p-3 rounded-[16px] bg-slate-50/50 border border-slate-100/50 active:bg-slate-50 transition-colors group">
              <div class="w-11 h-11 rounded-full bg-white shadow-sm border border-slate-100 flex items-center justify-center font-black text-[15px] text-slate-700 shrink-0">
                {{ member.username ? member.username.substring(0,1).toUpperCase() : 'U' }}
              </div>
              <div class="min-w-0 flex-1">
                <div class="flex items-center gap-2 mb-0.5">
                  <p class="text-[14px] font-extrabold text-slate-900 truncate">{{ member.username }}</p>
                  <span class="text-[9px] font-black px-1.5 py-0.5 rounded uppercase tracking-wider bg-emerald-100/80 text-emerald-700">{{ member.role }}</span>
                </div>
                <p class="text-[11px] font-bold text-slate-400">{{ member.relation || '未设置关系' }} <span class="mx-1 opacity-50">·</span> {{ member.email || member.phone || '无联系方式' }}</p>
              </div>
            </div>
            <div v-if="state.familyMembers.length === 0" class="text-center py-8 text-slate-400 text-[12px] font-bold bg-slate-50/50 rounded-[16px] border border-slate-100 border-dashed">暂无家庭成员</div>
          </div>
        </div>

        <div class="bg-white rounded-[24px] p-5 shadow-sm border border-slate-100/60">
          <div class="flex items-center justify-between mb-4">
            <div class="flex items-center gap-2">
              <div class="w-1.5 h-1.5 rounded-full bg-cyan-500"></div>
              <h3 class="font-extrabold text-[15px] text-slate-900">守护关系</h3>
            </div>
            <span class="text-[10px] font-bold text-slate-500 bg-slate-50 px-2 py-0.5 rounded-md border border-slate-100">{{ state.familyGuardianLinks.length }} 条</span>
          </div>
          <div class="space-y-3">
            <div v-for="link in state.familyGuardianLinks" :key="link.id" class="p-3.5 rounded-[16px] bg-slate-50/50 border border-slate-100/50 flex items-center justify-between gap-3">
              <div>
                <div class="flex items-center gap-2 mb-1">
                  <span class="text-[13px] font-extrabold text-slate-800">{{ link.guardian_name }}</span>
                  <svg class="w-3.5 h-3.5 text-slate-300" fill="none" stroke="currentColor" stroke-width="3" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" d="M17 8l4 4m0 0l-4 4m4-4H3"></path></svg>
                  <span class="text-[13px] font-extrabold text-slate-800">{{ link.member_name }}</span>
                </div>
                <p class="text-[10px] font-bold text-slate-400">{{ link.guardian_email }} / {{ link.member_email }}</p>
              </div>
              <span class="text-[9px] font-black px-2 py-1 rounded bg-cyan-100 text-cyan-700 tracking-widest uppercase">守护中</span>
            </div>
            <div v-if="state.familyGuardianLinks.length === 0" class="text-center py-8 text-slate-400 text-[12px] font-bold bg-slate-50/50 rounded-[16px] border border-slate-100 border-dashed">当前还没有守护关系</div>
          </div>
        </div>

        <div class="bg-white rounded-[24px] p-5 shadow-sm border border-slate-100/60">
          <div class="flex items-center gap-2 mb-4">
            <div class="w-1.5 h-1.5 rounded-full bg-rose-500"></div>
            <h3 class="font-extrabold text-[15px] text-slate-900">最新动态</h3>
          </div>
          <div class="space-y-3">
            <div v-if="state.familyNotifications.length === 0" class="text-center py-8 text-slate-400 text-[12px] font-bold bg-slate-50/50 rounded-[16px] border border-slate-100 border-dashed">无家庭动态</div>
            <div v-for="note in state.familyNotifications" :key="note.id" class="p-3.5 rounded-[16px] bg-slate-50/50 border border-slate-100/50 relative overflow-hidden group">
              <div class="absolute top-0 left-0 w-1.5 h-full bg-rose-400"></div>
              <div class="flex items-start gap-3 pl-2">
                <div class="w-10 h-10 rounded-full bg-white shadow-sm border border-slate-100 flex items-center justify-center shrink-0 text-rose-500">
                  <svg class="w-5 h-5" fill="none" stroke="currentColor" stroke-width="2.5" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"></path></svg>
                </div>
                <div class="min-w-0 flex-1">
                  <div class="flex items-center gap-2 flex-wrap mb-1.5">
                    <span class="px-2 py-0.5 bg-rose-100 text-rose-700 text-[10px] font-black rounded tracking-widest uppercase">{{ note.risk_level || '高危' }}</span>
                    <span class="text-[10px] font-bold text-slate-400">{{ state.formatTime(note.event_at) }}</span>
                  </div>
                  <p class="text-[14px] font-extrabold text-slate-900 leading-snug">{{ note.title || '高风险案件预警' }}</p>
                  <p class="text-[12px] font-medium text-slate-500 mt-1.5 leading-relaxed line-clamp-2">{{ note.case_summary || note.summary }}</p>
                  <div class="flex items-center gap-2 mt-2.5 text-[10px] font-bold text-slate-500 bg-white px-2 py-1.5 rounded-lg border border-slate-100/60 shadow-sm">
                    <svg class="w-3 h-3 text-slate-400" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z"></path></svg>
                    <span>成员: {{ note.target_name }}</span>
                    <div v-if="note.scam_type" class="w-1 h-1 rounded-full bg-slate-300"></div>
                    <span v-if="note.scam_type" class="truncate">类型: {{ note.scam_type }}</span>
                  </div>
                </div>
              </div>
            </div>
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
