<template>
  <section class="composer">
    <form @submit.prevent="handleSubmit">
      <div class="row">
        <label>
          类型
          <select v-model="form.channel">
            <option value="personal">个人联系</option>
            <option value="system">系统通知</option>
          </select>
        </label>
        <label v-if="form.channel === 'system'">
          优先级
          <select v-model="form.priority">
            <option value="info">一般</option>
            <option value="warning">提醒</option>
            <option value="critical">紧急</option>
          </select>
        </label>
      </div>
      <div class="row">
        <label v-if="form.channel === 'system'">
          发件人 ID
          <input v-model.number="form.senderId" type="number" min="0" />
        </label>
        <div v-else class="readonly">
          <span>发件人</span>
          <strong>{{ currentUserId ?? '未登录' }}</strong>
        </div>
        <label>
          收件人 ID
          <input v-model.number="form.receiverId" type="number" min="1" required />
        </label>
      </div>
      <label>
        标题
        <input v-model="form.title" type="text" placeholder="选填" />
      </label>
      <label>
        内容
        <textarea v-model="form.content" rows="4" required placeholder="输入站内信内容" />
      </label>
      <button type="submit" :disabled="sending || !currentUserId">{{ sending ? '发送中...' : '发送站内信' }}</button>
    </form>
  </section>
</template>

<script setup lang="ts">
import { reactive, watch } from 'vue'

const emit = defineEmits<{
  submit: [
    {
      channel: 'personal' | 'system'
      senderId?: number
      receiverId: number
      title?: string
      content: string
      priority?: 'info' | 'warning' | 'critical'
    }
  ]
}>()

const props = defineProps<{ sending: boolean; currentUserId: number | null }>()

const form = reactive({
  channel: 'personal' as 'personal' | 'system',
  priority: 'info' as 'info' | 'warning' | 'critical',
  senderId: props.currentUserId ?? 0,
  receiverId: 0,
  title: '',
  content: ''
})

watch(
  () => props.currentUserId,
  (value) => {
    if (form.channel === 'personal' && value) {
      form.senderId = value
    }
  },
  { immediate: true }
)

watch(
  () => form.channel,
  (value) => {
    if (value === 'personal' && props.currentUserId) {
      form.senderId = props.currentUserId
    }
  }
)

const handleSubmit = () => {
  if (!props.currentUserId) return
  if (!form.content.trim()) return
  emit('submit', {
    channel: form.channel,
    senderId: form.channel === 'personal' ? props.currentUserId : form.senderId,
    receiverId: form.receiverId,
    title: form.title,
    content: form.content,
    priority: form.channel === 'system' ? form.priority : undefined
  })
  form.content = ''
}
</script>

<style scoped>
.composer {
  background: #fff;
  border-radius: 12px;
  padding: 16px;
  box-shadow: 0 2px 4px rgba(15, 23, 42, 0.05);
  margin-bottom: 16px;
}

form {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.row {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(160px, 1fr));
  gap: 12px;
}

label {
  font-size: 14px;
  color: #475467;
  display: flex;
  flex-direction: column;
  gap: 6px;
}

input,
textarea,
select {
  padding: 8px 10px;
  border-radius: 8px;
  border: 1px solid #d0d5dd;
  font-size: 14px;
}

.readonly {
  display: flex;
  flex-direction: column;
  justify-content: center;
  padding: 8px 12px;
  border-radius: 8px;
  border: 1px dashed #d0d5dd;
  background: #f8fafc;
  font-size: 14px;
  color: #334155;
}

.readonly strong {
  font-size: 16px;
}

button {
  align-self: flex-end;
  padding: 10px 16px;
  border: none;
  border-radius: 8px;
  background: #2563eb;
  color: #fff;
  cursor: pointer;
}

button:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}
</style>
