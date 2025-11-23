<template>
  <section class="list">
    <header>
      <div>
        <h2>站内信</h2>
        <p v-if="subtitle">{{ subtitle }}</p>
      </div>
      <slot name="actions"></slot>
    </header>

    <div v-if="loading" class="placeholder">加载中...</div>
    <div v-else-if="!messages.length" class="placeholder">暂无数据</div>

    <ul v-else>
      <li v-for="message in messages" :key="message.id" :class="{ unread: !message.isRead }">
        <div class="title">
          <div class="title-text">
            <span>{{ message.title || '（无标题）' }}</span>
            <small>#{{ message.id }}</small>
          </div>
          <span class="pill" :class="message.channel">
            {{ message.channel === 'system' ? '系统通知' : '个人信息' }}
            <template v-if="message.channel === 'system' && message.priority"> · {{ priorityMap[message.priority] }}</template>
          </span>
        </div>
        <p class="content">{{ message.content }}</p>
        <footer>
          <span>发件人：{{ message.senderId }}</span>
          <span>收件人：{{ message.receiverId }}</span>
          <span>{{ formatDate(message.createdAt) }}</span>
          <button v-if="!message.isRead" @click="$emit('mark-read', message.id)">标记已读</button>
        </footer>
      </li>
    </ul>
  </section>
</template>

<script setup lang="ts">
import type { Message } from '@/api/client'

withDefaults(
  defineProps<{
    messages: Message[]
    loading?: boolean
    subtitle?: string
  }>(),
  {
    messages: () => [],
    loading: false,
    subtitle: ''
  }
)

const formatDate = (value?: string) => {
  if (!value) return ''
  return new Intl.DateTimeFormat('zh-CN', {
    dateStyle: 'medium',
    timeStyle: 'short'
  }).format(new Date(value))
}

const priorityMap: Record<string, string> = {
  info: '一般',
  warning: '提醒',
  critical: '紧急'
}
</script>

<style scoped>
.list {
  background: #fff;
  border-radius: 12px;
  padding: 16px;
  box-shadow: 0 2px 4px rgba(15, 23, 42, 0.05);
}

header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}

h2 {
  margin: 0;
}

.placeholder {
  padding: 24px;
  text-align: center;
  color: #94a3b8;
}

ul {
  list-style: none;
  padding: 0;
  margin: 0;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

li {
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  padding: 12px;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

li.unread {
  border-color: #2563eb;
}

.title {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 8px;
}

.title-text {
  display: flex;
  flex-direction: column;
  font-weight: 600;
}

.content {
  margin: 0;
  color: #475467;
}

footer {
  display: flex;
  gap: 12px;
  flex-wrap: wrap;
  align-items: center;
  font-size: 13px;
}

button {
  margin-left: auto;
  border: none;
  background: #22c55e;
  color: #fff;
  padding: 6px 10px;
  border-radius: 6px;
  cursor: pointer;
}

.pill {
  font-size: 12px;
  padding: 4px 8px;
  border-radius: 999px;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  background: #e2e8f0;
  color: #0f172a;
}

.pill.system {
  background: #fee2e2;
  color: #991b1b;
}

.pill.personal {
  background: #dbeafe;
  color: #1d4ed8;
}
</style>
