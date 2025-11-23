<template>
  <main class="layout">
    <header class="hero">
      <div>
        <p class="eyebrow">站内信系统</p>
        <h1>快速查看与发送信息</h1>
        <p class="muted">
          {{ auth.isAuthenticated ? `你好，${auth.user?.username}` : '请先登录以自动接收消息' }}
        </p>
      </div>
      <div class="badge" v-if="auth.isAuthenticated">
        <div class="badge-total">
          未读：<strong>{{ mailbox.counts.total }}</strong>
        </div>
        <small>个人 {{ mailbox.counts.personal }} · 系统 {{ mailbox.counts.system }}</small>
        <button class="link" @click="logout">退出登录</button>
      </div>
    </header>

    <section v-if="!auth.isAuthenticated" class="auth-card">
      <h2>{{ authMode === 'login' ? '登录账户' : '注册账户' }}</h2>
      <form @submit.prevent="handleAuth">
        <label>
          用户名
          <input v-model="authForm.username" type="text" autocomplete="username" required />
        </label>
        <label>
          密码
          <input v-model="authForm.password" type="password" autocomplete="current-password" required />
        </label>
        <button type="submit" :disabled="auth.loading">
          {{ auth.loading ? '处理中...' : authMode === 'login' ? '登录' : '注册并登录' }}
        </button>
        <p class="auth-switch">
          <span>{{ authMode === 'login' ? '还没有账户？' : '已经有账户？' }}</span>
          <button type="button" class="link" @click="toggleAuthMode">
            {{ authMode === 'login' ? '注册' : '切换到登录' }}
          </button>
        </p>
        <p class="auth-error" v-if="auth.error">{{ auth.error }}</p>
      </form>
    </section>

    <template v-else>
      <MessageComposer
        :sending="mailbox.sending"
        :current-user-id="auth.user?.id ?? null"
        @submit="mailbox.sendMessage"
      />

      <section class="filters">
        <label>类型
          <select v-model="selectedChannel" @change="updateChannel">
            <option value="personal">个人信息</option>
            <option value="system">系统通知</option>
          </select>
        </label>
        <label>状态
          <select v-model="selectedStatus" @change="updateStatus">
            <option value="all">全部</option>
            <option value="unread">未读</option>
            <option value="sent">已发送</option>
          </select>
        </label>
        <button @click="refresh">刷新</button>
      </section>

      <MessageList
        :messages="mailbox.items"
        :loading="mailbox.loading"
        :subtitle="`共 ${mailbox.total} 封`"
        @mark-read="mailbox.markAsRead"
      />
    </template>
  </main>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref, watch } from 'vue'
import MessageComposer from '@/components/MessageComposer.vue'
import MessageList from '@/components/MessageList.vue'
import { useMailboxStore } from '@/stores/mailbox'
import { useAuthStore } from '@/stores/auth'

const mailbox = useMailboxStore()
const auth = useAuthStore()
const selectedStatus = ref(mailbox.status)
const selectedChannel = ref(mailbox.channel)
const authMode = ref<'login' | 'register'>('login')
const authForm = reactive({
  username: '',
  password: ''
})

const refresh = () => mailbox.loadMessages()
const updateStatus = () => mailbox.loadMessages({ status: selectedStatus.value, page: 1 })
const updateChannel = () => mailbox.loadMessages({ channel: selectedChannel.value, page: 1 })
const toggleAuthMode = () => {
  authMode.value = authMode.value === 'login' ? 'register' : 'login'
}
const handleAuth = async () => {
  if (!authForm.username || !authForm.password) return
  if (authMode.value === 'login') {
    await auth.login(authForm)
  } else {
    await auth.register(authForm)
  }
  authForm.password = ''
  await mailbox.loadMessages({ page: 1 })
  await mailbox.loadUnreadCount()
}
const logout = () => {
  auth.logout()
  mailbox.setUser(null)
}

onMounted(() => {
  auth.initFromStorage()
})

watch(
  () => auth.user?.id,
  async (userId) => {
    mailbox.setUser(userId ?? null)
    if (userId) {
      await mailbox.loadMessages({ page: 1 })
      await mailbox.loadUnreadCount()
      mailbox.startAutoUnreadCount(10000)
    } else {
      mailbox.stopAutoUnreadCount()
    }
  },
  { immediate: true }
)
</script>

<style scoped>
.layout {
  max-width: 960px;
  margin: 0 auto;
  padding: 32px 16px 48px;
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.hero {
  display: flex;
  justify-content: space-between;
  align-items: center;
  background: #0f172a;
  color: #fff;
  padding: 20px;
  border-radius: 16px;
}

.badge {
  background: rgba(15, 23, 42, 0.4);
  padding: 10px 16px;
  border-radius: 16px;
  display: flex;
  flex-direction: column;
  gap: 4px;
  text-align: right;
}

.badge-total {
  font-size: 18px;
}

.badge .link {
  background: transparent;
  border: none;
  color: #93c5fd;
  cursor: pointer;
  text-decoration: underline;
}

.auth-card {
  background: #fff;
  border-radius: 16px;
  padding: 24px;
  box-shadow: 0 10px 30px rgba(15, 23, 42, 0.08);
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.auth-card form {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.auth-card label {
  display: flex;
  flex-direction: column;
  gap: 6px;
  font-size: 14px;
  color: #475467;
}

.auth-card input {
  padding: 10px 12px;
  border-radius: 8px;
  border: 1px solid #cbd5f5;
  font-size: 14px;
}

.auth-card button[type='submit'] {
  border: none;
  border-radius: 8px;
  padding: 12px;
  background: #2563eb;
  color: #fff;
  cursor: pointer;
}

.auth-switch {
  display: flex;
  gap: 8px;
  align-items: center;
  font-size: 14px;
}

.link {
  background: none;
  border: none;
  color: #2563eb;
  cursor: pointer;
  padding: 0;
}

.auth-error {
  color: #dc2626;
  font-size: 14px;
}

.filters {
  display: flex;
  gap: 12px;
  align-items: center;
  flex-wrap: wrap;
}

select,
.filters button {
  padding: 8px 12px;
  border-radius: 8px;
  border: 1px solid #cbd5f5;
}

.filters button {
  background: #0f172a;
  color: #fff;
  border: none;
}
</style>
