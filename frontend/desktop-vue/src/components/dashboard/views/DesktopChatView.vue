<template>
  <div :class="[embedded ? 'h-full' : 'max-w-6xl mx-auto animate-fade-in']">
    <div class="rounded-sm border border-slate-200 bg-white shadow-sm overflow-hidden h-full flex flex-col">
      <div class="bg-slate-950 px-4 py-2.5 flex items-center justify-between text-white border-b border-white/10 shrink-0">
        <div class="flex items-center min-w-0">
          <button @click="handleBack" class="inline-flex items-center gap-2 rounded-sm border border-white/10 bg-white/5 px-3 py-1.5 text-[11px] font-black uppercase tracking-[0.14em] text-slate-200 transition hover:bg-white/10">
            {{ embedded ? '返回题目' : '返回反诈模拟' }}
          </button>
        </div>
        <button @click="clearChatHistory" class="rounded-sm border border-white/10 bg-white/5 px-3 py-1.5 text-[11px] font-black uppercase tracking-[0.14em] text-slate-200 transition hover:bg-white/10">
          清空
        </button>
      </div>

      <div id="chat-container" class="flex-1 overflow-y-auto px-4 py-4 space-y-4 bg-gradient-to-b from-slate-50 via-white to-slate-100/80 custom-scrollbar">
        <div v-for="(msg, idx) in chatMessages" :key="idx" :class="['flex', msg.type === 'user' ? 'justify-end' : msg.type === 'tool' ? 'justify-center' : 'justify-start']">
          <div v-if="msg.type === 'tool'" class="text-[11px] font-mono text-slate-500 my-2 px-3 py-1.5 bg-white rounded-sm border border-slate-200 shadow-sm">
            <span class="animate-pulse mr-1">*</span> {{ msg.content }}
          </div>
          <div v-else :class="['max-w-[80%] rounded-2xl px-4 py-3 text-[13px] shadow-sm leading-6 select-text', msg.type === 'user' ? 'bg-brand-600 text-white rounded-br-sm' : msg.type === 'error' ? 'bg-rose-50 text-rose-700 border border-rose-200 rounded-bl-sm' : 'bg-white text-slate-700 border border-slate-200 rounded-bl-sm']">
            <div v-if="msg.content" class="chat-markdown" v-html="renderMarkdown(msg.content)"></div>
            <div v-if="msg.images && msg.images.length" :class="[msg.content ? 'mt-3' : '', 'grid grid-cols-2 gap-2']">
              <button
                v-for="(image, imageIdx) in msg.images"
                :key="`${idx}-${imageIdx}`"
                type="button"
                @click="openImage(image)"
                class="overflow-hidden rounded-sm border border-slate-200 bg-white hover:opacity-90 transition-all shadow-sm"
              >
                <img :src="image" alt="chat image" class="w-full h-24 object-cover block">
              </button>
            </div>
          </div>
        </div>

        <div v-if="isChatting" class="flex justify-start">
          <div class="bg-white text-slate-500 border border-slate-200 rounded-2xl px-4 py-3 rounded-bl-sm shadow-sm flex gap-1.5 items-center h-12">
            <span class="w-2 h-2 bg-slate-400 rounded-full animate-bounce"></span>
            <span class="w-2 h-2 bg-slate-400 rounded-full animate-bounce delay-75"></span>
            <span class="w-2 h-2 bg-slate-400 rounded-full animate-bounce delay-150"></span>
          </div>
        </div>
      </div>

      <div class="px-4 py-4 bg-white border-t border-slate-200">
        <input id="chat-image-input" type="file" accept="image/*" multiple class="hidden" @change="handleChatImageSelect">
        <div v-if="chatImages.length" class="mb-3 flex flex-wrap gap-2">
          <div v-for="(image, idx) in chatImages" :key="`chat-image-${idx}`" class="relative w-[60px] h-[60px] rounded-sm overflow-hidden border border-slate-200 bg-white shadow-sm">
            <img :src="image" alt="selected chat image" class="w-full h-full object-cover">
            <button type="button" @click="removeChatImage(idx)" class="absolute top-1 right-1 w-5 h-5 rounded-sm bg-slate-950/60 text-white text-[10px] flex items-center justify-center hover:bg-brand-600">×</button>
          </div>
        </div>
        <form @submit.prevent="sendChatMessage" class="flex gap-3 items-end">
          <button type="button" @click="triggerChatImagePicker" :disabled="isChatting" class="shrink-0 bg-white text-slate-600 p-3 rounded-full border border-slate-200 shadow-sm hover:bg-slate-50 hover:border-slate-300 disabled:opacity-50 transition-all" title="添加图片">
            <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-10h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z"></path></svg>
          </button>
          <input v-model="chatInput" type="text" placeholder="输入您的问题..." class="flex-1 min-h-[46px] bg-white border border-slate-200 rounded-full px-4 py-3 text-sm focus:bg-white focus:ring-2 focus:ring-brand-500 focus:border-brand-400 outline-none transition-all font-medium shadow-sm">
          <button type="submit" :disabled="(!chatInput.trim() && chatImages.length === 0) || isChatting" class="bg-brand-600 text-white p-3 rounded-full hover:bg-brand-700 disabled:opacity-50 transition-all shadow-sm">
            <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 19l9 2-9-18-9 18 9-2zm0 0v-8"></path></svg>
          </button>
        </form>
      </div>
    </div>
  </div>
</template>

<script>
import MarkdownIt from 'markdown-it';

const markdown = new MarkdownIt({
  html: false,
  linkify: true,
  breaks: true
});

const defaultLinkOpen = markdown.renderer.rules.link_open || ((tokens, idx, options, env, self) => self.renderToken(tokens, idx, options));
markdown.renderer.rules.link_open = (tokens, idx, options, env, self) => {
  const targetIndex = tokens[idx].attrIndex('target');
  const relIndex = tokens[idx].attrIndex('rel');
  if (targetIndex < 0) {
    tokens[idx].attrPush(['target', '_blank']);
  } else {
    tokens[idx].attrs[targetIndex][1] = '_blank';
  }
  if (relIndex < 0) {
    tokens[idx].attrPush(['rel', 'noopener noreferrer']);
  } else {
    tokens[idx].attrs[relIndex][1] = 'noopener noreferrer';
  }
  return defaultLinkOpen(tokens, idx, options, env, self);
};

export default {
  name: 'DesktopChatView',
  emits: ['back'],
  props: {
    app: {
      type: Object,
      required: true
    },
    embedded: {
      type: Boolean,
      default: false
    }
  },
  setup(props, { emit }) {
    const renderMarkdown = (content) => markdown.render(String(content || ''));

    const handleBack = () => {
      if (props.embedded) {
        emit('back');
        return;
      }
      props.app.activeTab = 'simulation_quiz';
    };

    return {
      ...props.app,
      handleBack,
      renderMarkdown
    };
  }
};
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
</style>
