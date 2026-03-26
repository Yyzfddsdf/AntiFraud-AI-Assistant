<template>
  <div v-if="state.visible" class="fixed inset-0 z-[1000] flex items-end bg-slate-900/40 backdrop-blur-sm transition-all duration-300">
    <!-- Bottom Sheet -->
    <div class="bg-white w-full h-[92vh] rounded-t-[32px] overflow-hidden flex flex-col animate-slide-up shadow-[0_-10px_40px_rgba(0,0,0,0.1)] relative">
      <!-- Handle -->
      <div class="absolute top-0 left-0 right-0 flex justify-center pt-3 pb-1 z-20">
        <div class="w-12 h-1.5 bg-gray-200/80 rounded-full"></div>
      </div>

      <!-- Header -->
      <div class="pt-6 pb-4 px-5 border-b border-gray-100 flex justify-between items-center sticky top-0 bg-white/95 backdrop-blur-md z-10">
        <h3 class="font-[800] text-xl text-slate-800 tracking-tight">家庭管理</h3>
        <button @click="state.close" class="text-slate-400 hover:text-slate-600 bg-slate-50 hover:bg-slate-100 p-2 rounded-full transition-colors">
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M6 18L18 6M6 6l12 12"></path></svg>
        </button>
      </div>

      <div class="flex-1 overflow-y-auto p-5 space-y-8 pb-24 pb-safe bg-slate-50/50" style="-webkit-overflow-scrolling: touch; overscroll-behavior: contain;">
        <!-- 邀请成员 Section -->
        <section>
          <div class="flex items-center mb-4 space-x-2">
            <div class="w-1 h-5 bg-blue-500 rounded-full"></div>
            <h4 class="font-bold text-lg text-slate-800 tracking-tight">邀请成员</h4>
          </div>
          <div class="bg-white rounded-3xl p-5 shadow-sm border border-slate-100/60 space-y-4">
            <div>
              <input v-model="state.familyInviteForm.invitee_email" type="email" placeholder="邮箱 (选填)" class="w-full bg-slate-50 border-0 rounded-2xl px-4 py-3.5 text-slate-800 text-sm placeholder:text-slate-400 focus:ring-2 focus:ring-blue-500/20 focus:bg-white transition-all">
            </div>
            <div>
              <input v-model="state.familyInviteForm.invitee_phone" type="tel" placeholder="手机号 (选填)" class="w-full bg-slate-50 border-0 rounded-2xl px-4 py-3.5 text-slate-800 text-sm placeholder:text-slate-400 focus:ring-2 focus:ring-blue-500/20 focus:bg-white transition-all">
            </div>
            
            <div class="m-dropdown" data-custom-dropdown>
              <button type="button" @click="state.toggleDropdown('family-invite-role')" :class="['w-full flex items-center justify-between bg-slate-50 border-0 rounded-2xl px-4 py-3.5 text-left transition-all focus:ring-2 focus:ring-blue-500/20', state.openDropdownKey === 'family-invite-role' ? 'bg-white shadow-sm ring-2 ring-blue-500/10' : '']">
                <div class="min-w-0 flex-1">
                  <div class="text-sm font-semibold text-slate-800">{{ state.getSelectedOptionLabel(state.familyRoleSelectOptions, state.familyInviteForm.role, '选择角色') }}</div>
                  <div class="text-xs text-slate-400 mt-0.5">{{ state.getSelectedOptionHint(state.familyRoleSelectOptions, state.familyInviteForm.role, '设置成员在家庭中的职责') }}</div>
                </div>
                <svg :class="['w-5 h-5 text-slate-400 transition-transform duration-300 ml-3', state.openDropdownKey === 'family-invite-role' ? 'rotate-180 text-blue-500' : '']" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7"></path></svg>
              </button>
              <transition name="fade">
                <div v-if="state.openDropdownKey === 'family-invite-role'" class="absolute left-0 right-0 mt-2 bg-white rounded-2xl shadow-lg border border-slate-100 p-2 z-50">
                  <button v-for="option in state.familyRoleSelectOptions" :key="`family-role-${option.value}`" type="button" @click="state.selectDropdownValue('family-invite-role', state.familyInviteForm, 'role', option.value)" :class="['w-full flex items-center justify-between px-4 py-3 rounded-xl transition-colors text-left', String(state.familyInviteForm.role) === String(option.value) ? 'bg-blue-50' : 'hover:bg-slate-50']">
                    <div class="min-w-0">
                      <div :class="['text-sm font-semibold', String(state.familyInviteForm.role) === String(option.value) ? 'text-blue-700' : 'text-slate-800']">{{ option.label }}</div>
                      <div :class="['text-xs mt-0.5', String(state.familyInviteForm.role) === String(option.value) ? 'text-blue-500/80' : 'text-slate-400']">{{ option.hint }}</div>
                    </div>
                    <svg v-if="String(state.familyInviteForm.role) === String(option.value)" class="w-5 h-5 text-blue-600 shrink-0 ml-3" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M5 13l4 4L19 7"></path></svg>
                  </button>
                </div>
              </transition>
            </div>
            
            <div>
              <input v-model="state.familyInviteForm.relation" type="text" placeholder="关系 (如: 父亲、配偶)" class="w-full bg-slate-50 border-0 rounded-2xl px-4 py-3.5 text-slate-800 text-sm placeholder:text-slate-400 focus:ring-2 focus:ring-blue-500/20 focus:bg-white transition-all">
            </div>
            
            <button @click="state.createFamilyInvitation" class="w-full bg-slate-900 active:bg-slate-800 text-white rounded-2xl py-3.5 font-bold text-[15px] shadow-md shadow-slate-900/20 transition-all mt-2">发送邀请</button>
          </div>
        </section>

        <!-- 配置守护 Section -->
        <section>
          <div class="flex items-center mb-4 space-x-2">
            <div class="w-1 h-5 bg-emerald-500 rounded-full"></div>
            <h4 class="font-bold text-lg text-slate-800 tracking-tight">配置守护</h4>
          </div>
          <div class="bg-white rounded-3xl p-5 shadow-sm border border-slate-100/60 space-y-4">
            <div class="m-dropdown" data-custom-dropdown>
              <button type="button" @click="state.toggleDropdown('guardian-user')" :class="['w-full flex items-center justify-between bg-slate-50 border-0 rounded-2xl px-4 py-3.5 text-left transition-all focus:ring-2 focus:ring-emerald-500/20', state.openDropdownKey === 'guardian-user' ? 'bg-white shadow-sm ring-2 ring-emerald-500/10' : '']">
                <div class="min-w-0 flex-1">
                  <div class="text-sm font-semibold text-slate-800">{{ state.getSelectedOptionLabel(state.familyGuardianSelectOptions, state.familyGuardianForm.guardian_user_id, '选择守护人') }}</div>
                  <div class="text-xs text-slate-400 mt-0.5">{{ state.getSelectedOptionHint(state.familyGuardianSelectOptions, state.familyGuardianForm.guardian_user_id, '守护人会收到高风险提醒') }}</div>
                </div>
                <svg :class="['w-5 h-5 text-slate-400 transition-transform duration-300 ml-3', state.openDropdownKey === 'guardian-user' ? 'rotate-180 text-emerald-500' : '']" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7"></path></svg>
              </button>
              <transition name="fade">
                <div v-if="state.openDropdownKey === 'guardian-user'" class="absolute left-0 right-0 mt-2 bg-white rounded-2xl shadow-lg border border-slate-100 p-2 z-50">
                  <button v-for="option in state.familyGuardianSelectOptions" :key="`guardian-option-${option.value}`" type="button" @click="state.selectDropdownValue('guardian-user', state.familyGuardianForm, 'guardian_user_id', option.value)" :class="['w-full flex items-center justify-between px-4 py-3 rounded-xl transition-colors text-left', String(state.familyGuardianForm.guardian_user_id) === String(option.value) ? 'bg-emerald-50' : 'hover:bg-slate-50']">
                    <div class="min-w-0">
                      <div :class="['text-sm font-semibold', String(state.familyGuardianForm.guardian_user_id) === String(option.value) ? 'text-emerald-700' : 'text-slate-800']">{{ option.label }}</div>
                      <div :class="['text-xs mt-0.5', String(state.familyGuardianForm.guardian_user_id) === String(option.value) ? 'text-emerald-500/80' : 'text-slate-400']">{{ option.hint }}</div>
                    </div>
                    <svg v-if="String(state.familyGuardianForm.guardian_user_id) === String(option.value)" class="w-5 h-5 text-emerald-600 shrink-0 ml-3" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M5 13l4 4L19 7"></path></svg>
                  </button>
                </div>
              </transition>
            </div>
            <div class="m-dropdown" data-custom-dropdown>
              <button type="button" @click="state.toggleDropdown('protected-user')" :class="['w-full flex items-center justify-between bg-slate-50 border-0 rounded-2xl px-4 py-3.5 text-left transition-all focus:ring-2 focus:ring-emerald-500/20', state.openDropdownKey === 'protected-user' ? 'bg-white shadow-sm ring-2 ring-emerald-500/10' : '']">
                <div class="min-w-0 flex-1">
                  <div class="text-sm font-semibold text-slate-800">{{ state.getSelectedOptionLabel(state.familyProtectedSelectOptions, state.familyGuardianForm.member_user_id, '选择被守护人') }}</div>
                  <div class="text-xs text-slate-400 mt-0.5">{{ state.getSelectedOptionHint(state.familyProtectedSelectOptions, state.familyGuardianForm.member_user_id, '出现高风险时会触发通知') }}</div>
                </div>
                <svg :class="['w-5 h-5 text-slate-400 transition-transform duration-300 ml-3', state.openDropdownKey === 'protected-user' ? 'rotate-180 text-emerald-500' : '']" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7"></path></svg>
              </button>
              <transition name="fade">
                <div v-if="state.openDropdownKey === 'protected-user'" class="absolute left-0 right-0 mt-2 bg-white rounded-2xl shadow-lg border border-slate-100 p-2 z-50">
                  <button v-for="option in state.familyProtectedSelectOptions" :key="`protected-option-${option.value}`" type="button" @click="state.selectDropdownValue('protected-user', state.familyGuardianForm, 'member_user_id', option.value)" :class="['w-full flex items-center justify-between px-4 py-3 rounded-xl transition-colors text-left', String(state.familyGuardianForm.member_user_id) === String(option.value) ? 'bg-emerald-50' : 'hover:bg-slate-50']">
                    <div class="min-w-0">
                      <div :class="['text-sm font-semibold', String(state.familyGuardianForm.member_user_id) === String(option.value) ? 'text-emerald-700' : 'text-slate-800']">{{ option.label }}</div>
                      <div :class="['text-xs mt-0.5', String(state.familyGuardianForm.member_user_id) === String(option.value) ? 'text-emerald-500/80' : 'text-slate-400']">{{ option.hint }}</div>
                    </div>
                    <svg v-if="String(state.familyGuardianForm.member_user_id) === String(option.value)" class="w-5 h-5 text-emerald-600 shrink-0 ml-3" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M5 13l4 4L19 7"></path></svg>
                  </button>
                </div>
              </transition>
            </div>
            <button @click="state.createGuardianLink" class="w-full bg-emerald-500 active:bg-emerald-600 text-white rounded-2xl py-3.5 font-bold text-[15px] shadow-md shadow-emerald-500/20 transition-all mt-2">保存关系</button>
          </div>
        </section>

        <!-- 家庭成员 Section -->
        <section>
          <div class="flex items-center mb-4 space-x-2">
            <div class="w-1 h-5 bg-violet-500 rounded-full"></div>
            <h4 class="font-bold text-lg text-slate-800 tracking-tight">成员列表</h4>
            <span class="bg-slate-100 text-slate-500 text-xs px-2 py-0.5 rounded-full font-bold ml-1">{{ state.familyMembers.length }}</span>
          </div>
          <div class="space-y-3">
            <div v-for="member in state.familyMembers" :key="member.member_id || member.user_id" class="bg-white rounded-3xl p-4 shadow-sm border border-slate-100/60 flex justify-between items-center gap-4">
              <div class="flex items-center gap-3">
                <div class="w-10 h-10 rounded-full bg-gradient-to-br from-indigo-100 to-purple-100 flex items-center justify-center shrink-0 border border-white shadow-sm">
                  <span class="text-indigo-600 font-bold text-sm">{{ member.username ? member.username.substring(0, 1).toUpperCase() : 'M' }}</span>
                </div>
                <div class="min-w-0">
                  <div class="font-bold text-[15px] text-slate-800 truncate">{{ member.username || '未命名' }}</div>
                  <div class="text-xs text-slate-500 truncate mt-0.5 flex items-center gap-2">
                    <span class="bg-slate-100 px-1.5 py-0.5 rounded text-[10px] font-medium">{{ member.relation || '未知关系' }}</span>
                    <span>{{ member.email || member.phone || '-' }}</span>
                  </div>
                </div>
              </div>
              <button v-if="state.familyOverview.current_member && state.familyOverview.current_member.role === 'owner' && member.role !== 'owner'" @click="state.deleteFamilyMember(member)" :disabled="state.familyDeletingMembers[member.member_id]" class="text-xs font-bold text-red-500 bg-red-50 hover:bg-red-100 px-3 py-1.5 rounded-xl disabled:opacity-50 transition-colors whitespace-nowrap">
                移除
              </button>
              <div v-else-if="member.role === 'owner'" class="text-[10px] font-bold text-amber-600 bg-amber-50 px-2.5 py-1 rounded-lg whitespace-nowrap">
                创建者
              </div>
            </div>
            
            <div v-if="state.familyMembers.length === 0" class="bg-white rounded-3xl p-8 border border-slate-100 border-dashed text-center">
              <div class="w-12 h-12 bg-slate-50 rounded-full flex items-center justify-center mx-auto mb-3">
                <svg class="w-6 h-6 text-slate-300" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4"></path></svg>
              </div>
              <p class="text-slate-400 text-sm font-medium">暂无家庭成员</p>
            </div>
          </div>
        </section>

        <!-- 邀请记录 Section -->
        <section>
          <div class="flex items-center mb-4 space-x-2">
            <div class="w-1 h-5 bg-slate-300 rounded-full"></div>
            <h4 class="font-bold text-lg text-slate-800 tracking-tight">邀请记录</h4>
          </div>
          <div class="space-y-3">
            <div v-for="invitation in state.familyInvitations" :key="invitation.id" class="bg-white rounded-3xl p-4 shadow-sm border border-slate-100/60">
              <div class="flex justify-between items-start mb-2">
                <div class="font-bold text-[15px] text-slate-800">{{ invitation.invitee_email || invitation.invitee_phone || '未指定目标' }}</div>
                <span :class="['text-[10px] font-bold px-2 py-0.5 rounded-lg', invitation.status === 'pending' ? 'bg-amber-50 text-amber-600' : (invitation.status === 'accepted' ? 'bg-emerald-50 text-emerald-600' : 'bg-slate-100 text-slate-500')]">
                  {{ invitation.status === 'pending' ? '等待中' : (invitation.status === 'accepted' ? '已接受' : '已过期/拒绝') }}
                </span>
              </div>
              <div class="flex items-center gap-2 text-xs text-slate-500 mb-2">
                <span class="bg-slate-50 px-2 py-0.5 rounded-md border border-slate-100">{{ invitation.role }}</span>
                <span class="bg-slate-50 px-2 py-0.5 rounded-md border border-slate-100">{{ invitation.relation || '未填写关系' }}</span>
              </div>
              <div class="bg-slate-50 rounded-xl p-2 mt-2 flex items-center justify-between border border-slate-100">
                <div class="text-xs text-slate-400 font-mono">邀请码: <span class="text-slate-700 font-bold ml-1">{{ invitation.invite_code }}</span></div>
              </div>
              <div class="text-[10px] text-slate-400 mt-2 text-right">截止时间: {{ state.formatTime(invitation.expires_at) }}</div>
            </div>
            
            <div v-if="state.familyInvitations.length === 0" class="text-center py-6">
              <p class="text-slate-400 text-sm font-medium">暂无邀请记录</p>
            </div>
          </div>
        </section>
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
