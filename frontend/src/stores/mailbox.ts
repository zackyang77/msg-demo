import { defineStore } from 'pinia'
import type { Message } from '@/api/client'
import { messageApi } from '@/api/client'

export const useMailboxStore = defineStore('mailbox', {
  state: () => ({
    loading: false,
    sending: false,
    items: [] as Message[],
    total: 0,
    page: 1,
    size: 20,
    status: 'all' as 'all' | 'unread' | 'sent',
    channel: 'personal' as 'personal' | 'system',
    counts: {
      personal: 0,
      system: 0,
      total: 0
    },
    userId: null as number | null,
    error: '' as string | null,
    pollingTimer: null as number | null
  }),
  actions: {
    async loadMessages(params?: {
      page?: number
      size?: number
      status?: 'all' | 'unread' | 'sent'
      channel?: 'personal' | 'system'
    }) {
      if (!this.userId) {
        return
      }
      this.loading = true
      this.error = null
      try {
        const nextChannel = params?.channel ?? this.channel
        const { items, total, page, size } = await messageApi.list({
          page: params?.page ?? this.page,
          size: params?.size ?? this.size,
          status: params?.status ?? this.status,
          channel: nextChannel
        })
        this.items = items
        this.total = total
        this.page = page
        this.size = size
        this.channel = nextChannel
        this.status = params?.status ?? this.status
      } catch (error) {
        this.error = error instanceof Error ? error.message : '加载站内信失败'
      } finally {
        this.loading = false
      }
    },
    async loadUnreadCount() {
      if (!this.userId) {
        this.counts = { personal: 0, system: 0, total: 0 }
        return
      }
      try {
        this.counts = await messageApi.unreadCount(this.userId)
      } catch (error) {
        console.error('无法获取未读数量', error)
      }
    },
    startAutoUnreadCount(intervalMs = 10000) {
      if (this.pollingTimer) return
      this.loadUnreadCount()
      this.pollingTimer = window.setInterval(() => {
        if (!this.userId) {
          this.stopAutoUnreadCount()
          return
        }
        this.loadUnreadCount()
      }, intervalMs) as unknown as number
    },
    stopAutoUnreadCount() {
      if (this.pollingTimer) {
        clearInterval(this.pollingTimer)
        this.pollingTimer = null
      }
    },
    async sendMessage(payload: {
      channel: 'personal' | 'system'
      senderId?: number
      receiverId: number
      title?: string
      content: string
      priority?: 'info' | 'warning' | 'critical'
    }) {
      this.sending = true
      this.error = null
      try {
        const message = await messageApi.send(payload)
        if (this.channel === payload.channel && this.status !== 'sent') {
          await this.loadMessages({ page: 1, channel: payload.channel })
        }
        await this.loadUnreadCount()
        return message
      } catch (error) {
        this.error = error instanceof Error ? error.message : '发送失败'
        throw error
      } finally {
        this.sending = false
      }
    },
    async markAsRead(id: number) {
      try {
        await messageApi.markRead({ id, channel: this.channel })
        this.items = this.items.map((msg) =>
          msg.id === id
            ? {
                ...msg,
                isRead: true,
                readAt: new Date().toISOString()
              }
            : msg
        )
        await this.loadUnreadCount()
      } catch (error) {
        this.error = error instanceof Error ? error.message : '更新失败'
      }
    },
    setUser(userId: number | null) {
      this.userId = userId
      this.items = []
      this.total = 0
      if (!userId) {
        this.stopAutoUnreadCount()
      }
    }
  }
})
