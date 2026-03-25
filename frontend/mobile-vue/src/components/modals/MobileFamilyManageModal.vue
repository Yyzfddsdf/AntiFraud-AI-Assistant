<template>
  <div v-if="state.visible" class="fixed inset-0 z-[1000] flex items-end bg-black/50 backdrop-blur-sm">
    <div class="bg-white w-full h-[92vh] rounded-t-2xl overflow-hidden flex flex-col animate-slide-up">
      <div class="p-3 border-b border-gray-100 flex justify-between items-center sticky top-0 bg-white z-10">
        <h3 class="font-bold text-base">家庭管理</h3>
        <button @click="state.close" class="text-gray-400 p-1.5"><svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"></path></svg></button>
      </div>
      <div class="flex-1 overflow-y-auto p-4 space-y-4 pb-20 pb-safe" style="-webkit-overflow-scrolling: touch; overscroll-behavior: contain;">
        <div>
          <h4 class="font-bold mb-2">邀请成员</h4>
          <div class="space-y-2">
            <input v-model="state.familyInviteForm.invitee_email" type="email" placeholder="邮箱" class="m-input">
            <input v-model="state.familyInviteForm.invitee_phone" type="tel" placeholder="手机号" class="m-input">
            <div class="m-dropdown" data-custom-dropdown>
              <button type="button" @click="state.toggleDropdown('family-invite-role')" :class="['m-dropdown-trigger', state.openDropdownKey === 'family-invite-role' ? 'is-open' : '']">
                <div class="min-w-0">
                  <div class="text-sm font-semibold text-slate-900">{{ state.getSelectedOptionLabel(state.familyRoleSelectOptions, state.familyInviteForm.role, '选择角色') }}</div>
                  <div class="text-[11px] text-slate-400 mt-1">{{ state.getSelectedOptionHint(state.familyRoleSelectOptions, state.familyInviteForm.role, '设置成员在家庭中的职责') }}</div>
                </div>
                <svg :class="['w-4 h-4 text-slate-400 transition-transform', state.openDropdownKey === 'family-invite-role' ? 'rotate-180' : '']" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7"></path></svg>
              </button>
              <transition name="fade">
                <div v-if="state.openDropdownKey === 'family-invite-role'" class="m-dropdown-menu">
                  <button v-for="option in state.familyRoleSelectOptions" :key="`family-role-${option.value}`" type="button" @click="state.selectDropdownValue('family-invite-role', state.familyInviteForm, 'role', option.value)" :class="['m-dropdown-option', String(state.familyInviteForm.role) === String(option.value) ? 'is-selected' : '']">
                    <div class="min-w-0">
                      <div class="text-sm font-semibold text-slate-900">{{ option.label }}</div>
                      <div class="text-[11px] text-slate-400 mt-1">{{ option.hint }}</div>
                    </div>
                    <svg v-if="String(state.familyInviteForm.role) === String(option.value)" class="w-4 h-4 text-emerald-600 shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7"></path></svg>
                  </button>
                </div>
              </transition>
            </div>
            <input v-model="state.familyInviteForm.relation" type="text" placeholder="关系 (如: 父亲)" class="m-input">
            <button @click="state.createFamilyInvitation" class="m-btn-primary h-9 text-sm">发送邀请</button>
          </div>
        </div>
        <div>
          <h4 class="font-bold mb-2">配置守护</h4>
          <div class="space-y-2">
            <div class="m-dropdown" data-custom-dropdown>
              <button type="button" @click="state.toggleDropdown('guardian-user')" :class="['m-dropdown-trigger', state.openDropdownKey === 'guardian-user' ? 'is-open' : '']">
                <div class="min-w-0">
                  <div class="text-sm font-semibold text-slate-900">{{ state.getSelectedOptionLabel(state.familyGuardianSelectOptions, state.familyGuardianForm.guardian_user_id, '选择守护人') }}</div>
                  <div class="text-[11px] text-slate-400 mt-1">{{ state.getSelectedOptionHint(state.familyGuardianSelectOptions, state.familyGuardianForm.guardian_user_id, '守护人会收到高风险提醒') }}</div>
                </div>
                <svg :class="['w-4 h-4 text-slate-400 transition-transform', state.openDropdownKey === 'guardian-user' ? 'rotate-180' : '']" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7"></path></svg>
              </button>
              <transition name="fade">
                <div v-if="state.openDropdownKey === 'guardian-user'" class="m-dropdown-menu">
                  <button v-for="option in state.familyGuardianSelectOptions" :key="`guardian-option-${option.value}`" type="button" @click="state.selectDropdownValue('guardian-user', state.familyGuardianForm, 'guardian_user_id', option.value)" :class="['m-dropdown-option', String(state.familyGuardianForm.guardian_user_id) === String(option.value) ? 'is-selected' : '']">
                    <div class="min-w-0">
                      <div class="text-sm font-semibold text-slate-900">{{ option.label }}</div>
                      <div class="text-[11px] text-slate-400 mt-1">{{ option.hint }}</div>
                    </div>
                    <svg v-if="String(state.familyGuardianForm.guardian_user_id) === String(option.value)" class="w-4 h-4 text-emerald-600 shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7"></path></svg>
                  </button>
                </div>
              </transition>
            </div>
            <div class="m-dropdown" data-custom-dropdown>
              <button type="button" @click="state.toggleDropdown('protected-user')" :class="['m-dropdown-trigger', state.openDropdownKey === 'protected-user' ? 'is-open' : '']">
                <div class="min-w-0">
                  <div class="text-sm font-semibold text-slate-900">{{ state.getSelectedOptionLabel(state.familyProtectedSelectOptions, state.familyGuardianForm.member_user_id, '选择被守护人') }}</div>
                  <div class="text-[11px] text-slate-400 mt-1">{{ state.getSelectedOptionHint(state.familyProtectedSelectOptions, state.familyGuardianForm.member_user_id, '被守护成员出现高风险时会触发通知') }}</div>
                </div>
                <svg :class="['w-4 h-4 text-slate-400 transition-transform', state.openDropdownKey === 'protected-user' ? 'rotate-180' : '']" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7"></path></svg>
              </button>
              <transition name="fade">
                <div v-if="state.openDropdownKey === 'protected-user'" class="m-dropdown-menu">
                  <button v-for="option in state.familyProtectedSelectOptions" :key="`protected-option-${option.value}`" type="button" @click="state.selectDropdownValue('protected-user', state.familyGuardianForm, 'member_user_id', option.value)" :class="['m-dropdown-option', String(state.familyGuardianForm.member_user_id) === String(option.value) ? 'is-selected' : '']">
                    <div class="min-w-0">
                      <div class="text-sm font-semibold text-slate-900">{{ option.label }}</div>
                      <div class="text-[11px] text-slate-400 mt-1">{{ option.hint }}</div>
                    </div>
                    <svg v-if="String(state.familyGuardianForm.member_user_id) === String(option.value)" class="w-4 h-4 text-emerald-600 shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7"></path></svg>
                  </button>
                </div>
              </transition>
            </div>
            <button @click="state.createGuardianLink" class="m-btn-primary h-9 text-sm">保存关系</button>
          </div>
        </div>
        <div>
          <h4 class="font-bold mb-2">家庭成员</h4>
          <div class="space-y-2">
            <div v-for="member in state.familyMembers" :key="member.member_id || member.user_id" class="m-card p-3">
              <div class="flex justify-between gap-3">
                <div>
                  <div class="font-bold text-sm">{{ member.username }}</div>
                  <div class="text-xs text-gray-500 mt-1">{{ member.email || '-' }}</div>
                  <div class="text-xs text-gray-400 mt-1">{{ member.relation || member.role }}</div>
                </div>
                <button v-if="state.familyOverview.current_member && state.familyOverview.current_member.role === 'owner' && member.role !== 'owner'" @click="state.deleteFamilyMember(member)" :disabled="state.familyDeletingMembers[member.member_id]" class="text-xs font-bold text-red-600 bg-red-50 px-3 py-1 rounded-lg disabled:opacity-50">
                  移除
                </button>
              </div>
            </div>
            <div v-if="state.familyMembers.length === 0" class="text-center py-4 text-sm text-gray-400">暂无成员</div>
          </div>
        </div>
        <div>
          <h4 class="font-bold mb-3">邀请记录</h4>
          <div class="space-y-3">
            <div v-for="invitation in state.familyInvitations" :key="invitation.id" class="m-card p-4">
              <div class="font-bold text-sm">{{ invitation.invitee_email || invitation.invitee_phone || '未指定目标' }}</div>
              <div class="text-xs text-gray-500 mt-1">角色：{{ invitation.role }} / 关系：{{ invitation.relation || '未填写' }}</div>
              <div class="text-xs text-slate-400 mt-1 font-mono break-all">邀请码：{{ invitation.invite_code }}</div>
              <div class="text-xs text-slate-400 mt-1">状态：{{ invitation.status }} / 截止：{{ state.formatTime(invitation.expires_at) }}</div>
            </div>
            <div v-if="state.familyInvitations.length === 0" class="text-center py-4 text-sm text-gray-400">暂无邀请记录</div>
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
