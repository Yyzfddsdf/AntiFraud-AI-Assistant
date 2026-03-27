<template>
  <div :class="['chat-shell', embedded ? 'chat-shell--embedded' : 'animate-fade-in']">
    <div class="chat-frame">
      <div class="chat-topbar">
        <div class="flex items-center min-w-0">
          <button v-if="embedded || !hideBackAction" @click="handleBack" class="chat-topbar-action">
            {{ embedded ? '返回题目' : '返回全景分析' }}
          </button>
        </div>
        <div class="chat-topbar-title">
          <div class="chat-topbar-eyebrow">{{ embedded ? '用户问题助手回应' : '管理员 AI 助手' }}</div>
          <div class="chat-topbar-name">{{ embedded ? 'Sentinel AI' : '管理员 AI 聊天' }}</div>
        </div>
        <div class="flex items-center justify-end">
          <button @click="clearChatHistory" class="chat-topbar-action">
            清空
          </button>
        </div>
      </div>

      <div id="chat-container" class="chat-stage custom-scrollbar">
        <div class="chat-thread">
          <div v-for="(msg, idx) in chatMessages" :key="idx" :class="['chat-row', msg.type === 'user' ? 'chat-row--user' : msg.type === 'tool' ? 'chat-row--tool' : 'chat-row--ai']">
            <div v-if="msg.type === 'tool'" class="chat-tool-note">
            <span class="animate-pulse mr-1">*</span> {{ msg.content }}
            </div>
            <div v-else :class="['chat-message', msg.type === 'user' ? 'chat-message--user' : msg.type === 'error' ? 'chat-message--error' : 'chat-message--ai']">
              <div v-if="msg.content" class="chat-markdown" v-html="renderMarkdown(msg.content)"></div>
              <div v-if="msg.images && msg.images.length" :class="[msg.content ? 'mt-3' : '', 'grid grid-cols-2 gap-3']">
                <button
                  v-for="(image, imageIdx) in msg.images"
                  :key="`${idx}-${imageIdx}`"
                  type="button"
                  @click="openImage(image)"
                  class="chat-inline-image"
                >
                  <img :src="image" alt="chat image" class="w-full h-28 object-cover block">
                </button>
              </div>
            </div>
          </div>
          <div v-if="isChatting" class="chat-row chat-row--ai">
            <div class="chat-typing">
              <span class="w-2 h-2 bg-slate-400 rounded-full animate-bounce"></span>
              <span class="w-2 h-2 bg-slate-400 rounded-full animate-bounce delay-75"></span>
              <span class="w-2 h-2 bg-slate-400 rounded-full animate-bounce delay-150"></span>
            </div>
          </div>
        </div>
      </div>

      <div class="chat-composer-shell">
        <input :id="chatImageInputId" type="file" accept="image/*" multiple class="hidden" @change="handleChatImageSelect">
        <div v-if="chatImages.length" class="chat-preview-strip">
          <div v-for="(image, idx) in chatImages" :key="`chat-image-${idx}`" class="chat-preview-card">
            <img :src="image" alt="selected chat image" class="w-full h-full object-cover">
            <button type="button" @click="removeChatImage(idx)" class="chat-preview-remove">×</button>
          </div>
        </div>
        <form
          @submit.prevent="sendChatMessage"
          :class="['container-ia-chat', { 'composer-ready': chatInput.trim() || chatImages.length > 0, 'composer-busy': isChatting }]"
        >
          <div class="container-upload-files">
            <label
              :for="isChatting ? null : chatImageInputId"
              :class="['upload-file', { 'upload-file--disabled': isChatting }]"
              title="添加图片"
            >
              <svg
                xmlns="http://www.w3.org/2000/svg"
                width="24"
                height="24"
                viewBox="0 0 24 24"
              >
                <g
                  fill="none"
                  stroke="currentColor"
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="2"
                >
                  <rect width="18" height="18" x="3" y="3" rx="2" ry="2"></rect>
                  <circle cx="9" cy="9" r="2"></circle>
                  <path d="m21 15l-3.086-3.086a2 2 0 0 0-2.828 0L6 21"></path>
                </g>
              </svg>
            </label>
          </div>
          <input
            v-model="chatInput"
            type="text"
            placeholder="Ask Anything..."
            class="input-text"
            :disabled="isChatting"
            :required="chatImages.length === 0"
          >
          <button
            type="submit"
            :disabled="(!chatInput.trim() && chatImages.length === 0) || isChatting"
            class="label-text"
            title="发送"
          >
            <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor">
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M4 11.5L19 4l-4.5 16l-2.5-6l-6-2.5z"
              ></path>
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M11.5 14L19 4"
              ></path>
            </svg>
          </button>
        </form>
        <p class="chat-footnote">内容由 AI 生成，请仔细甄别</p>
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
    },
    hideBackAction: {
      type: Boolean,
      default: false
    }
  },
  setup(props, { emit }) {
    const chatImageInputId = `chat-image-input-${Math.random().toString(36).slice(2, 10)}`;
    const renderMarkdown = (content) => markdown.render(String(content || ''));

    const handleBack = () => {
      if (props.embedded) {
        emit('back');
        return;
      }
      props.app.activeTab = 'admin_stats';
    };

    return {
      ...props.app,
      chatImageInputId,
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

.chat-shell {
  width: 100%;
  height: 100%;
  min-height: 0;
  background: #ffffff;
}

.chat-frame {
  width: min(100%, 980px);
  height: 100%;
  margin: 0 auto;
  display: flex;
  flex-direction: column;
  background: transparent;
}

.chat-shell--embedded .chat-frame {
  width: min(100%, 880px);
}

.chat-topbar {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto minmax(0, 1fr);
  align-items: center;
  gap: 1rem;
  padding: 0.9rem 0 1.1rem;
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

.chat-topbar-action:hover,
.chat-topbar-action:focus-visible {
  color: #0f172a;
  background: transparent;
}

.chat-stage {
  flex: 1;
  overflow-y: auto;
  min-height: 0;
  padding: 2rem 0 1.25rem;
  background: #ffffff;
}

.chat-thread {
  width: min(100%, 720px);
  margin: 0 auto;
  display: flex;
  flex-direction: column;
  gap: 1.75rem;
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
  max-width: min(100%, 46rem);
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
  max-width: min(68%, 24rem);
  padding: 0.7rem 1rem;
  border-radius: 1.15rem;
  border-top-right-radius: 0.45rem;
  background: linear-gradient(180deg, rgba(244, 246, 255, 0.98), rgba(236, 242, 255, 0.98));
  color: #334155;
  box-shadow: 0 10px 24px rgba(148, 163, 184, 0.12);
}

.chat-message--error {
  max-width: min(80%, 30rem);
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
  padding: 1rem 0 0.35rem;
  background: #ffffff;
}

.chat-preview-strip {
  width: min(100%, 720px);
  margin: 0 auto 0.65rem;
  display: flex;
  flex-wrap: wrap;
  gap: 0.65rem;
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
  width: min(100%, 720px);
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
  width: min(100%, 720px);
  margin: 0.45rem auto 0;
  text-align: center;
  font-size: 0.65rem;
  color: #c0c7d2;
  letter-spacing: 0.04em;
}

@media (max-width: 768px) {
  .chat-frame,
  .chat-shell--embedded .chat-frame {
    width: 100%;
  }

  .chat-topbar {
    grid-template-columns: 1fr;
    justify-items: center;
    padding-top: 0.75rem;
  }

  .chat-topbar > div:first-child,
  .chat-topbar > div:last-child {
    width: 100%;
  }

  .chat-topbar > div:last-child {
    display: none;
  }

  .chat-thread,
  .container-ia-chat,
  .chat-preview-strip,
  .chat-footnote {
    width: 100%;
  }

  .chat-message--user,
  .chat-message--error {
    max-width: 82%;
  }
}
</style>
