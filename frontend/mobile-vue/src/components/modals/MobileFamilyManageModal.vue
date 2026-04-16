<template>
  <div v-if="state.visible" class="fixed inset-0 z-[1000] flex items-end bg-slate-900/40 backdrop-blur-sm transition-all duration-300">
    <!-- Bottom Sheet -->
    <div class="bg-[#F8F9FA] w-full h-[92vh] rounded-t-[32px] overflow-hidden flex flex-col animate-slide-up shadow-[0_-10px_40px_rgba(0,0,0,0.1)] relative">
      <!-- Handle -->
      <div class="absolute top-0 left-0 right-0 flex justify-center pt-3 pb-1 z-20">
        <div class="w-12 h-1.5 bg-slate-200/80 rounded-full"></div>
      </div>

      <!-- Header -->
      <div class="pt-6 pb-4 px-6 flex justify-between items-center sticky top-0 bg-[#F8F9FA]/90 backdrop-blur-md z-10">
        <h3 class="font-bold text-[18px] text-slate-900 tracking-tight">家庭管理</h3>
        <button @click="state.close" class="text-slate-400 hover:text-slate-600 bg-slate-100/50 hover:bg-slate-200 p-2 rounded-full transition-colors">
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"></path></svg>
        </button>
      </div>

      <div class="flex-1 overflow-y-auto px-5 pt-2 pb-24 pb-safe space-y-6" style="-webkit-overflow-scrolling: touch; overscroll-behavior: contain;">
        <!-- 邀请成员 Section -->
        <section>
          <div class="mb-4">
            <h4 class="font-bold text-[16px] text-slate-900 tracking-tight">邀请成员</h4>
          </div>
          <div class="bg-white rounded-[24px] p-5 shadow-[0_4px_20px_rgba(0,0,0,0.02)] border-0 space-y-4">
            <div>
              <input v-model="state.familyInviteForm.invitee_email" type="email" placeholder="邮箱 (选填)" class="w-full bg-[#F8F9FA] border-0 rounded-[16px] px-5 py-4 text-slate-900 text-[14px] placeholder:text-slate-400 focus:ring-2 focus:ring-slate-900/5 focus:bg-white transition-all outline-none">
            </div>
            <div>
              <input v-model="state.familyInviteForm.invitee_phone" type="tel" placeholder="手机号 (选填)" class="w-full bg-[#F8F9FA] border-0 rounded-[16px] px-5 py-4 text-slate-900 text-[14px] placeholder:text-slate-400 focus:ring-2 focus:ring-slate-900/5 focus:bg-white transition-all outline-none">
            </div>
            
            <div class="m-dropdown" data-custom-dropdown>
              <button type="button" @click="state.toggleDropdown('family-invite-role')" :class="['w-full flex items-center justify-between bg-[#F8F9FA] border-0 rounded-[16px] px-5 py-4 text-left transition-all outline-none focus:ring-2 focus:ring-slate-900/5', state.openDropdownKey === 'family-invite-role' ? 'bg-white shadow-[0_2px_10px_rgba(0,0,0,0.04)] ring-2 ring-slate-900/5' : '']">
                <div class="min-w-0 flex-1">
                  <div class="text-[14px] font-bold text-slate-900">{{ state.getSelectedOptionLabel(state.familyRoleSelectOptions, state.familyInviteForm.role, '选择角色') }}</div>
                  <div class="text-[11px] text-slate-400 mt-1">{{ state.getSelectedOptionHint(state.familyRoleSelectOptions, state.familyInviteForm.role, '设置成员在家庭中的职责') }}</div>
                </div>
                <svg :class="['w-5 h-5 text-slate-400 transition-transform duration-300 ml-3', state.openDropdownKey === 'family-invite-role' ? 'rotate-180 text-slate-800' : '']" fill="none" stroke="currentColor" stroke-width="1.5" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" d="M19 9l-7 7-7-7"></path></svg>
              </button>
              <transition name="fade">
                <div v-if="state.openDropdownKey === 'family-invite-role'" class="absolute left-0 right-0 mt-2 bg-white rounded-[20px] shadow-[0_10px_30px_rgba(0,0,0,0.08)] border-0 p-2 z-50">
                  <button v-for="option in state.familyRoleSelectOptions" :key="`family-role-${option.value}`" type="button" @click="state.selectDropdownValue('family-invite-role', state.familyInviteForm, 'role', option.value)" :class="['w-full flex items-center justify-between px-4 py-3 rounded-[12px] transition-colors text-left', String(state.familyInviteForm.role) === String(option.value) ? 'bg-slate-50' : 'hover:bg-slate-50/50']">
                    <div class="min-w-0">
                      <div :class="['text-[14px] font-bold', String(state.familyInviteForm.role) === String(option.value) ? 'text-slate-900' : 'text-slate-700']">{{ option.label }}</div>
                      <div class="text-[11px] mt-0.5 text-slate-400">{{ option.hint }}</div>
                    </div>
                    <svg v-if="String(state.familyInviteForm.role) === String(option.value)" class="w-5 h-5 text-slate-800 shrink-0 ml-3" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" d="M5 13l4 4L19 7"></path></svg>
                  </button>
                </div>
              </transition>
            </div>
            
            <div>
              <input v-model="state.familyInviteForm.relation" type="text" placeholder="关系 (如: 父亲、配偶)" class="w-full bg-[#F8F9FA] border-0 rounded-[16px] px-5 py-4 text-slate-900 text-[14px] placeholder:text-slate-400 focus:ring-2 focus:ring-slate-900/5 focus:bg-white transition-all outline-none">
            </div>
            
            <button @click="state.createFamilyInvitation" class="w-full bg-slate-900 active:scale-[0.98] text-white rounded-[16px] py-4 font-bold text-[14px] shadow-[0_4px_12px_rgba(0,0,0,0.1)] transition-all mt-2">发送邀请</button>
          </div>
        </section>

        <!-- 配置守护 Section -->
        <section>
          <div class="mb-4">
            <h4 class="font-bold text-[16px] text-slate-900 tracking-tight">配置守护</h4>
          </div>
          <div class="bg-white rounded-[24px] p-5 shadow-[0_4px_20px_rgba(0,0,0,0.02)] border-0 space-y-4">
            <div class="m-dropdown" data-custom-dropdown>
              <button type="button" @click="state.toggleDropdown('guardian-user')" :class="['w-full flex items-center justify-between bg-[#F8F9FA] border-0 rounded-[16px] px-5 py-4 text-left transition-all outline-none focus:ring-2 focus:ring-slate-900/5', state.openDropdownKey === 'guardian-user' ? 'bg-white shadow-[0_2px_10px_rgba(0,0,0,0.04)] ring-2 ring-slate-900/5' : '']">
                <div class="min-w-0 flex-1">
                  <div class="text-[14px] font-bold text-slate-900">{{ state.getSelectedOptionLabel(state.familyGuardianSelectOptions, state.familyGuardianForm.guardian_user_id, '选择守护人') }}</div>
                  <div class="text-[11px] text-slate-400 mt-1">{{ state.getSelectedOptionHint(state.familyGuardianSelectOptions, state.familyGuardianForm.guardian_user_id, '守护人会收到高风险提醒') }}</div>
                </div>
                <svg :class="['w-5 h-5 text-slate-400 transition-transform duration-300 ml-3', state.openDropdownKey === 'guardian-user' ? 'rotate-180 text-slate-800' : '']" fill="none" stroke="currentColor" stroke-width="1.5" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" d="M19 9l-7 7-7-7"></path></svg>
              </button>
              <transition name="fade">
                <div v-if="state.openDropdownKey === 'guardian-user'" class="absolute left-0 right-0 mt-2 bg-white rounded-[20px] shadow-[0_10px_30px_rgba(0,0,0,0.08)] border-0 p-2 z-50">
                  <button v-for="option in state.familyGuardianSelectOptions" :key="`guardian-option-${option.value}`" type="button" @click="state.selectDropdownValue('guardian-user', state.familyGuardianForm, 'guardian_user_id', option.value)" :class="['w-full flex items-center justify-between px-4 py-3 rounded-[12px] transition-colors text-left', String(state.familyGuardianForm.guardian_user_id) === String(option.value) ? 'bg-slate-50' : 'hover:bg-slate-50/50']">
                    <div class="min-w-0">
                      <div :class="['text-[14px] font-bold', String(state.familyGuardianForm.guardian_user_id) === String(option.value) ? 'text-slate-900' : 'text-slate-700']">{{ option.label }}</div>
                      <div class="text-[11px] mt-0.5 text-slate-400">{{ option.hint }}</div>
                    </div>
                    <svg v-if="String(state.familyGuardianForm.guardian_user_id) === String(option.value)" class="w-5 h-5 text-slate-800 shrink-0 ml-3" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" d="M5 13l4 4L19 7"></path></svg>
                  </button>
                </div>
              </transition>
            </div>
            <div class="m-dropdown" data-custom-dropdown>
              <button type="button" @click="state.toggleDropdown('protected-user')" :class="['w-full flex items-center justify-between bg-[#F8F9FA] border-0 rounded-[16px] px-5 py-4 text-left transition-all outline-none focus:ring-2 focus:ring-slate-900/5', state.openDropdownKey === 'protected-user' ? 'bg-white shadow-[0_2px_10px_rgba(0,0,0,0.04)] ring-2 ring-slate-900/5' : '']">
                <div class="min-w-0 flex-1">
                  <div class="text-[14px] font-bold text-slate-900">{{ state.getSelectedOptionLabel(state.familyProtectedSelectOptions, state.familyGuardianForm.member_user_id, '选择被守护人') }}</div>
                  <div class="text-[11px] text-slate-400 mt-1">{{ state.getSelectedOptionHint(state.familyProtectedSelectOptions, state.familyGuardianForm.member_user_id, '出现高风险时会触发通知') }}</div>
                </div>
                <svg :class="['w-5 h-5 text-slate-400 transition-transform duration-300 ml-3', state.openDropdownKey === 'protected-user' ? 'rotate-180 text-slate-800' : '']" fill="none" stroke="currentColor" stroke-width="1.5" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" d="M19 9l-7 7-7-7"></path></svg>
              </button>
              <transition name="fade">
                <div v-if="state.openDropdownKey === 'protected-user'" class="absolute left-0 right-0 mt-2 bg-white rounded-[20px] shadow-[0_10px_30px_rgba(0,0,0,0.08)] border-0 p-2 z-50">
                  <button v-for="option in state.familyProtectedSelectOptions" :key="`protected-option-${option.value}`" type="button" @click="state.selectDropdownValue('protected-user', state.familyGuardianForm, 'member_user_id', option.value)" :class="['w-full flex items-center justify-between px-4 py-3 rounded-[12px] transition-colors text-left', String(state.familyGuardianForm.member_user_id) === String(option.value) ? 'bg-slate-50' : 'hover:bg-slate-50/50']">
                    <div class="min-w-0">
                      <div :class="['text-[14px] font-bold', String(state.familyGuardianForm.member_user_id) === String(option.value) ? 'text-slate-900' : 'text-slate-700']">{{ option.label }}</div>
                      <div class="text-[11px] mt-0.5 text-slate-400">{{ option.hint }}</div>
                    </div>
                    <svg v-if="String(state.familyGuardianForm.member_user_id) === String(option.value)" class="w-5 h-5 text-slate-800 shrink-0 ml-3" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" d="M5 13l4 4L19 7"></path></svg>
                  </button>
                </div>
              </transition>
            </div>
            <button @click="state.createGuardianLink" class="w-full bg-slate-900 active:scale-[0.98] text-white rounded-[16px] py-4 font-bold text-[14px] shadow-[0_4px_12px_rgba(0,0,0,0.1)] transition-all mt-2">保存关系</button>
          </div>
        </section>

        <!-- 家庭成员 Section -->
        <section>
          <div class="flex items-center justify-between mb-4">
            <h4 class="font-bold text-[16px] text-slate-900 tracking-tight">成员列表</h4>
            <span class="bg-slate-200/50 text-slate-500 text-[11px] px-2.5 py-1 rounded-lg font-medium">{{ state.familyMembers.length }} 人</span>
          </div>
          <div class="bg-white rounded-[24px] p-2 shadow-[0_4px_20px_rgba(0,0,0,0.02)] border-0">
            <div class="space-y-1">
              <div v-for="member in state.familyMembers" :key="member.member_id || member.user_id" class="p-3 rounded-[16px] flex justify-between items-center gap-4 hover:bg-[#F8F9FA] transition-colors">
                <div class="flex items-center gap-3">
                  <div class="w-10 h-10 rounded-full bg-[#F8F9FA] flex items-center justify-center shrink-0">
                    <span class="text-slate-700 font-bold text-[14px]">{{ member.username ? member.username.substring(0, 1).toUpperCase() : 'M' }}</span>
                  </div>
                  <div class="min-w-0">
                    <div class="font-bold text-[14px] text-slate-900 truncate">{{ member.username || '未命名' }}</div>
                    <div class="text-[11px] text-slate-500 truncate mt-1 flex items-center gap-2">
                      <span class="bg-slate-100 px-1.5 py-0.5 rounded">{{ member.relation || '未知关系' }}</span>
                      <span>{{ member.email || member.phone || '-' }}</span>
                    </div>
                  </div>
                </div>
                <button v-if="state.familyOverview.current_member && state.familyOverview.current_member.role === 'owner' && member.role !== 'owner'" @click="state.deleteFamilyMember(member)" :disabled="state.familyDeletingMembers[member.member_id]" class="text-[11px] font-bold text-rose-500 bg-rose-50 px-3 py-1.5 rounded-xl disabled:opacity-50 transition-colors whitespace-nowrap active:scale-[0.96]">
                  移除
                </button>
                <div v-else-if="member.role === 'owner'" class="text-[10px] font-medium text-slate-500 bg-slate-100 px-2.5 py-1 rounded-lg whitespace-nowrap">
                  创建者
                </div>
              </div>
              
              <div v-if="state.familyMembers.length === 0" class="p-8 text-center">
                <p class="text-slate-400 text-[12px]">暂无家庭成员</p>
              </div>
            </div>
          </div>
        </section>

        <!-- 邀请记录 Section -->
        <section>
          <div class="mb-4">
            <h4 class="font-bold text-[16px] text-slate-900 tracking-tight">邀请记录</h4>
          </div>
          <div class="bg-white rounded-[24px] p-2 shadow-[0_4px_20px_rgba(0,0,0,0.02)] border-0">
            <div class="space-y-1">
              <div v-for="invitation in state.familyInvitations" :key="invitation.id" class="p-4 rounded-[16px] hover:bg-[#F8F9FA] transition-colors">
                <div class="flex justify-between items-start mb-2">
                  <div class="font-bold text-[14px] text-slate-900">{{ invitation.invitee_email || invitation.invitee_phone || '未指定目标' }}</div>
                  <span :class="['text-[10px] font-medium px-2 py-0.5 rounded-lg', invitation.status === 'pending' ? 'bg-slate-800 text-white' : (invitation.status === 'accepted' ? 'bg-slate-100 text-slate-800' : 'bg-slate-50 text-slate-400')]">
                    {{ invitation.status === 'pending' ? '等待中' : (invitation.status === 'accepted' ? '已接受' : '已过期/拒绝') }}
                  </span>
                </div>
                <div class="flex items-center gap-2 text-[11px] text-slate-500 mb-3">
                  <span class="bg-white border border-slate-100 px-2 py-0.5 rounded shadow-sm">{{ invitation.role }}</span>
                  <span class="bg-white border border-slate-100 px-2 py-0.5 rounded shadow-sm">{{ invitation.relation || '未填写关系' }}</span>
                </div>
                <div class="bg-[#F8F9FA] rounded-xl p-2.5 flex items-center justify-between">
                  <div class="text-[11px] text-slate-400">邀请码: <span class="text-slate-800 font-mono font-bold tracking-widest ml-1">{{ invitation.invite_code }}</span></div>
                  <div class="text-[10px] text-slate-400">截止: {{ state.formatTime(invitation.expires_at) }}</div>
                </div>
              </div>
              
              <div v-if="state.familyInvitations.length === 0" class="py-8 text-center">
                <p class="text-slate-400 text-[12px]">暂无邀请记录</p>
              </div>
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
