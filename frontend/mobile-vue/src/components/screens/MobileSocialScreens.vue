<template>
  <div v-show="state.activeTab === 'chat'" class="fixed inset-0 z-[1000] bg-white">
    <div class="chat-shell animate-fade-in">
      <div class="chat-frame">
        <div class="chat-topbar pt-safe px-4">
          <div class="flex items-center min-w-0">
            <button @click="state.activeTab = 'tasks'" class="chat-topbar-action chat-topbar-icon" title="返回首页" aria-label="返回首页">
              <i data-lucide="arrow-left" size="18"></i>
            </button>
          </div>
          <div class="chat-topbar-title">
            <div class="chat-topbar-eyebrow">用户问题助手回应</div>
            <div class="chat-topbar-name">Sentinel AI</div>
          </div>
          <div class="flex items-center justify-end">
            <button @click="state.clearChatHistory" class="chat-topbar-action chat-topbar-icon" title="清空对话" aria-label="清空对话">
              <i data-lucide="trash-2" size="18"></i>
            </button>
          </div>
        </div>

        <div id="chat-container" class="chat-stage hide-scrollbar">
          <div class="chat-thread">
            <div
              v-for="(msg, idx) in state.chatMessages"
              :key="idx"
              :class="['chat-row', msg.type === 'user' ? 'chat-row--user' : msg.type === 'tool' ? 'chat-row--tool' : 'chat-row--ai']"
            >
              <div v-if="msg.type === 'tool'" class="chat-tool-note">
                <span class="animate-pulse mr-1">*</span> {{ msg.content }}
              </div>
              <div
                v-else
                :class="[
                  'chat-message',
                  msg.type === 'error'
                    ? 'chat-message--error'
                    : msg.type === 'user'
                      ? 'chat-message--user'
                      : 'chat-message--ai'
                ]"
              >
                <div v-if="msg.type === 'ai' && msg.rendered_content" class="chat-markdown" v-html="msg.rendered_content"></div>
                <div v-else-if="msg.content" class="chat-markdown whitespace-pre-wrap">{{ msg.content }}</div>
                <div v-if="msg.images && msg.images.length" :class="[msg.content ? 'mt-3' : '', 'grid grid-cols-2 gap-3']">
                  <button
                    v-for="(image, imageIdx) in msg.images"
                    :key="`${idx}-${imageIdx}`"
                    type="button"
                    @click="state.openImage(image)"
                    class="chat-inline-image"
                  >
                    <img :src="image" alt="chat image" class="w-full h-28 object-cover block">
                  </button>
                </div>
              </div>
            </div>

            <div v-if="state.isChatting" class="chat-row chat-row--ai">
              <div class="chat-typing">
                <span class="w-2 h-2 bg-slate-400 rounded-full animate-bounce"></span>
                <span class="w-2 h-2 bg-slate-400 rounded-full animate-bounce delay-75"></span>
                <span class="w-2 h-2 bg-slate-400 rounded-full animate-bounce delay-150"></span>
              </div>
            </div>
          </div>
        </div>

        <div class="chat-composer-shell px-4" :style="{ paddingBottom: 'max(0.35rem, env(safe-area-inset-bottom))' }">
          <input id="chat-image-input" type="file" accept="image/*" multiple class="hidden" @change="state.handleChatImageSelect">
          <div v-if="state.chatImages.length" class="chat-preview-strip hide-scrollbar">
            <div v-for="(image, idx) in state.chatImages" :key="`chat-image-${idx}`" class="chat-preview-card">
              <img :src="image" alt="selected chat image" class="w-full h-full object-cover">
              <button type="button" @click="state.removeChatImage(idx)" class="chat-preview-remove">×</button>
            </div>
          </div>
          <form
            @submit.prevent="state.sendChatMessage"
            :class="['container-ia-chat', { 'composer-ready': state.chatInput.trim() || state.chatImages.length > 0, 'composer-busy': state.isChatting }]"
          >
            <div class="container-upload-files">
              <label
                :for="state.isChatting ? null : 'chat-image-input'"
                :class="['upload-file', { 'upload-file--disabled': state.isChatting }]"
                title="添加图片"
              >
                <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24">
                  <g fill="none" stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2">
                    <rect width="18" height="18" x="3" y="3" rx="2" ry="2"></rect>
                    <circle cx="9" cy="9" r="2"></circle>
                    <path d="m21 15-3.086-3.086a2 2 0 0 0-2.828 0L6 21"></path>
                  </g>
                </svg>
              </label>
            </div>
            <input
              v-model="state.chatInput"
              type="text"
              placeholder="Ask Anything..."
              class="input-text"
              :disabled="state.isChatting"
              :required="state.chatImages.length === 0"
              @keydown.enter.prevent="!state.isChatting && (state.chatInput.trim() || state.chatImages.length) ? state.sendChatMessage() : null"
            >
            <button
              type="submit"
              :disabled="(!state.chatInput.trim() && state.chatImages.length === 0) || state.isChatting"
              class="label-text"
              title="发送"
            >
              <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 11.5L19 4l-4.5 16l-2.5-6l-6-2.5z"></path>
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11.5 14L19 4"></path>
              </svg>
            </button>
          </form>
          <p class="chat-footnote">内容由 AI 生成，请仔细甄别</p>
        </div>
      </div>
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
import { nextTick, onMounted, onUpdated } from 'vue';

defineProps({
  state: {
    type: Object,
    required: true
  }
});

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

<style scoped>
.chat-markdown :deep(p) {
  margin: 0;
}

.chat-markdown :deep(p + p),
.chat-markdown :deep(ul),
.chat-markdown :deep(ol),
.chat-markdown :deep(pre),
.chat-markdown :deep(blockquote),
.chat-markdown :deep(h1),
.chat-markdown :deep(h2),
.chat-markdown :deep(h3),
.chat-markdown :deep(h4) {
  margin-top: 0.6rem;
}

.chat-markdown :deep(h1),
.chat-markdown :deep(h2),
.chat-markdown :deep(h3),
.chat-markdown :deep(h4) {
  font-weight: 800;
  line-height: 1.35;
}

.chat-markdown :deep(h1) {
  font-size: 1.05rem;
}

.chat-markdown :deep(h2) {
  font-size: 1rem;
}

.chat-markdown :deep(h3),
.chat-markdown :deep(h4) {
  font-size: 0.95rem;
}

.chat-markdown :deep(ul),
.chat-markdown :deep(ol) {
  padding-left: 1.2rem;
}

.chat-markdown :deep(li + li) {
  margin-top: 0.2rem;
}

.chat-markdown :deep(blockquote) {
  border-left: 3px solid rgba(100, 116, 139, 0.35);
  padding-left: 0.75rem;
  color: inherit;
  opacity: 0.9;
}

.chat-markdown :deep(code) {
  font-family: Consolas, 'Courier New', monospace;
  font-size: 0.92em;
  padding: 0.1rem 0.35rem;
  border-radius: 0.25rem;
  background: rgba(15, 23, 42, 0.08);
}

.chat-markdown :deep(pre) {
  overflow-x: auto;
  padding: 0.85rem 0.95rem;
  border-radius: 0.6rem;
  background: rgba(15, 23, 42, 0.92);
  color: #e2e8f0;
}

.chat-markdown :deep(pre code) {
  background: transparent;
  padding: 0;
  color: inherit;
}

.chat-markdown :deep(a) {
  text-decoration: underline;
  text-underline-offset: 2px;
}

.chat-markdown :deep(hr) {
  border: 0;
  border-top: 1px solid rgba(148, 163, 184, 0.35);
  margin-top: 0.75rem;
}

.chat-shell {
  width: 100%;
  height: 100%;
  min-height: 0;
  background: #ffffff;
}

.chat-frame {
  width: 100%;
  height: 100%;
  display: flex;
  flex-direction: column;
  background: transparent;
}

.chat-topbar {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto minmax(0, 1fr);
  align-items: center;
  gap: 1rem;
  padding-top: 0.75rem;
  padding-bottom: 1rem;
  background: transparent;
  flex-shrink: 0;
}

.chat-topbar-title {
  text-align: center;
  min-width: 0;
}

.chat-topbar-eyebrow {
  font-size: 0.65rem;
  font-weight: 700;
  letter-spacing: 0.18em;
  text-transform: uppercase;
  color: rgba(148, 163, 184, 0.92);
}

.chat-topbar-name {
  margin-top: 0.2rem;
  font-size: 0.85rem;
  font-weight: 700;
  color: #1e293b;
}

.chat-topbar-action {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-height: auto;
  padding: 0.2rem 0;
  border: none;
  border-radius: 0;
  background: transparent;
  color: #64748b;
  font-size: 0.7rem;
  font-weight: 800;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  transition: all 0.25s ease;
}

.chat-topbar-icon {
  width: 2rem;
  height: 2rem;
  border-radius: 999px;
}

.chat-topbar-action:hover,
.chat-topbar-action:focus-visible {
  color: #0f172a;
  background: transparent;
}

.chat-stage {
  flex: 1;
  overflow-y: auto;
  min-height: 0;
  padding: 1.2rem 1rem 1rem;
  background: #ffffff;
}

.chat-thread {
  width: 100%;
  margin: 0 auto;
  display: flex;
  flex-direction: column;
  gap: 1.4rem;
}

.chat-row {
  display: flex;
}

.chat-row--user {
  justify-content: flex-end;
}

.chat-row--ai {
  justify-content: flex-start;
}

.chat-row--tool {
  justify-content: center;
}

.chat-tool-note {
  font-size: 0.72rem;
  font-family: Consolas, 'Courier New', monospace;
  color: #94a3b8;
  padding: 0.3rem 0.5rem;
}

.chat-message {
  max-width: 100%;
  line-height: 1.9;
  color: #0f172a;
}

.chat-message--ai {
  padding: 0;
  background: transparent;
  border: none;
  box-shadow: none;
  border-radius: 0;
}

.chat-message--user {
  max-width: min(76%, 24rem);
  padding: 0;
  background: transparent;
  border: none;
  box-shadow: none;
  border-radius: 0;
  color: #334155;
  text-align: right;
}

.chat-message--error {
  max-width: min(84%, 30rem);
  padding: 0.8rem 1rem;
  border-radius: 1rem;
  background: rgba(254, 242, 242, 0.95);
  border: 1px solid rgba(254, 202, 202, 0.95);
  color: #b91c1c;
}

.chat-inline-image {
  overflow: hidden;
  border-radius: 1rem;
  border: 1px solid rgba(226, 232, 240, 0.95);
  background: #fff;
  box-shadow: 0 8px 18px rgba(148, 163, 184, 0.12);
  transition: transform 0.22s ease, box-shadow 0.22s ease;
}

.chat-inline-image:hover {
  transform: translateY(-1px);
  box-shadow: 0 14px 26px rgba(148, 163, 184, 0.18);
}

.chat-typing {
  display: inline-flex;
  align-items: center;
  gap: 0.45rem;
  padding: 0.25rem 0;
  color: #94a3b8;
}

.chat-composer-shell {
  flex-shrink: 0;
  padding-top: 1rem;
  background: #ffffff;
}

.chat-preview-strip {
  width: 100%;
  margin: 0 auto 0.65rem;
  display: flex;
  flex-wrap: wrap;
  gap: 0.65rem;
  overflow-x: auto;
}

.chat-preview-card {
  position: relative;
  width: 64px;
  height: 64px;
  overflow: hidden;
  border-radius: 1rem;
  border: 1px solid rgba(226, 232, 240, 0.95);
  background: #fff;
  box-shadow: 0 8px 20px rgba(148, 163, 184, 0.12);
}

.chat-preview-remove {
  position: absolute;
  top: 0.25rem;
  right: 0.25rem;
  width: 1.2rem;
  height: 1.2rem;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border: none;
  border-radius: 999px;
  background: rgba(15, 23, 42, 0.6);
  color: #fff;
  font-size: 0.72rem;
  transition: background-color 0.2s ease;
}

.chat-preview-remove:hover {
  background: #ef4444;
}

.container-ia-chat {
  position: relative;
  display: flex;
  align-items: center;
  justify-content: end;
  width: 100%;
  margin: 0 auto;
}

.container-upload-files {
  position: absolute;
  left: 0;
  display: flex;
  color: #aaaaaa;
  transition: all 0.5s;
}

.upload-file {
  margin: 5px;
  padding: 2px;
  cursor: pointer;
  transition: all 0.5s;
  border: none;
  background: transparent;
  display: inline-flex;
  align-items: center;
  justify-content: center;
}

.upload-file:hover,
.upload-file:focus-visible {
  color: #4c4c4c;
  transform: scale(1.1);
}

.upload-file--disabled {
  opacity: 0.45;
  cursor: not-allowed;
  transform: none;
  pointer-events: none;
}

.input-text {
  max-width: calc(100% - 72px);
  width: 100%;
  margin-left: 72px;
  padding: 0.75rem 1rem;
  padding-right: 46px;
  border-radius: 50px;
  border: none;
  outline: none;
  background-color: #e9e9e9;
  color: #4c4c4c;
  font-size: 14px;
  line-height: 18px;
  font-family: "Segoe UI", Tahoma, Geneva, Verdana, sans-serif;
  font-weight: 500;
  transition: all 0.5s cubic-bezier(0.175, 0.885, 0.32, 1.05);
  z-index: 3;
}

.input-text::placeholder {
  color: #959595;
}

.input-text::selection {
  background-color: #4c4c4c;
  color: #e9e9e9;
}

.container-ia-chat:focus-within .input-text,
.container-ia-chat.composer-ready .input-text {
  max-width: calc(100% - 42px);
  margin-left: 42px;
}

.container-ia-chat:focus-within .container-upload-files,
.container-ia-chat.composer-ready .container-upload-files {
  opacity: 0;
  visibility: hidden;
  pointer-events: none;
  filter: blur(5px);
}

.input-text:disabled {
  opacity: 0.75;
  cursor: not-allowed;
}

.label-text {
  position: absolute;
  top: 50%;
  right: 0.25rem;
  transform: translateY(-50%) scale(0.25);
  width: 36px;
  height: 36px;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 6px;
  border: none;
  outline: none;
  cursor: pointer;
  transition: all 0.5s cubic-bezier(0.175, 0.885, 0.32, 1.05);
  z-index: 4;
  opacity: 0;
  visibility: hidden;
  pointer-events: none;
  color: #e9e9e9;
  background: linear-gradient(to top right, #9147ff, #ff4141);
  box-shadow: inset 0 0 4px rgba(255, 255, 255, 0.5);
  border-radius: 50px;
}

.container-ia-chat:focus-within .label-text,
.container-ia-chat.composer-ready .label-text {
  transform: translateY(-50%) scale(1);
  opacity: 1;
  visibility: visible;
  pointer-events: all;
}

.label-text:hover,
.label-text:focus-visible {
  transform-origin: top center;
  box-shadow: inset 0 0 6px rgba(255, 255, 255, 1);
}

.label-text:active {
  transform: translateY(-50%) scale(0.9);
}

.label-text:disabled {
  opacity: 0;
  visibility: hidden;
  pointer-events: none;
}

.chat-footnote {
  width: 100%;
  margin: 0.45rem auto 0;
  text-align: center;
  font-size: 0.65rem;
  color: #c0c7d2;
  letter-spacing: 0.04em;
}
</style>
