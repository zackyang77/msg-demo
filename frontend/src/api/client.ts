import axios from 'axios'

export interface Message {
  id: number
  senderId: number
  receiverId: number
  title?: string
  content: string
  isRead: boolean
  readAt?: string
  createdAt: string
  channel: 'personal' | 'system'
  priority?: 'info' | 'warning' | 'critical'
}

export interface ListMessagesParams {
  page?: number
  size?: number
  status?: 'all' | 'unread' | 'sent'
  channel?: 'personal' | 'system'
}

export interface ListMessagesResponse {
  items: Message[]
  total: number
  page: number
  size: number
}

export interface User {
  id: number
  username: string
}

export interface AuthRequest {
  username: string
  password: string
}

export interface AuthResponse {
  token: string
  user: User
}

const http = axios.create({
  baseURL: '/api/v1',
  timeout: 10000
})

export const setAuthToken = (token: string | null) => {
  if (token) {
    http.defaults.headers.common.Authorization = `Bearer ${token}`
  } else {
    delete http.defaults.headers.common.Authorization
  }
}

export const authApi = {
  async register(payload: AuthRequest): Promise<AuthResponse> {
    const { data } = await http.post<AuthResponse>('/auth/register', payload)
    return data
  },
  async login(payload: AuthRequest): Promise<AuthResponse> {
    const { data } = await http.post<AuthResponse>('/auth/login', payload)
    return data
  }
}

export const messageApi = {
  async list(params: ListMessagesParams): Promise<ListMessagesResponse> {
    const { data } = await http.get<ListMessagesResponse>('/messages', {
      params
    })
    return data
  },
  async send(payload: {
    channel: 'personal' | 'system'
    senderId?: number
    receiverId: number
    title?: string
    content: string
    priority?: 'info' | 'warning' | 'critical'
  }): Promise<Message> {
    const { data } = await http.post<Message>('/messages', payload)
    return data
  },
  async markRead(params: { id: number; channel: 'personal' | 'system' }): Promise<Message> {
    const { data } = await http.post<Message>(`/messages/${params.id}/read`, {
      channel: params.channel
    })
    return data
  },
  async unreadCount(): Promise<{ personal: number; system: number; total: number }> {
    const { data } = await http.get<{ personal: number; system: number; total: number }>(
      '/messages/unread/count'
    )
    return data
  }
}
