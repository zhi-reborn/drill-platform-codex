import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { Notification } from '@/types'
import { notificationApi } from '@/api/modules/notification'

export const useNotificationStore = defineStore('notifications', () => {
  const notifications = ref<Notification[]>([])
  const unreadCount = ref(0)

  const unreadNotifications = computed(() => notifications.value.filter(n => !n.is_read))
  const recentNotifications = computed(() => notifications.value.slice(0, 20))

  async function fetchNotifications(): Promise<void> {
    try {
      const data = await notificationApi.getList({ page: 1, page_size: 50 })
      notifications.value = data.items
      unreadCount.value = data.items.filter(n => !n.is_read).length
    } catch {
      notifications.value = []
    }
  }

  async function markAsRead(id: number): Promise<void> {
    try {
      await notificationApi.markAsRead(id)
      const n = notifications.value.find(n => n.id === id)
      if (n) n.is_read = true
      unreadCount.value = Math.max(0, unreadCount.value - 1)
    } catch { /* ignore */ }
  }

  async function markAllAsRead(): Promise<void> {
    try {
      await notificationApi.markAllAsRead()
      notifications.value.forEach(n => n.is_read = true)
      unreadCount.value = 0
    } catch { /* ignore */ }
  }

  async function deleteNotification(id: number): Promise<void> {
    try {
      await notificationApi.delete(id)
      const index = notifications.value.findIndex(n => n.id === id)
      if (index !== -1) {
        if (!notifications.value[index].is_read) {
          unreadCount.value = Math.max(0, unreadCount.value - 1)
        }
        notifications.value.splice(index, 1)
      }
    } catch { /* ignore */ }
  }

  function addNotification(n: Notification): void {
    notifications.value.unshift(n)
    if (!n.is_read) unreadCount.value++
  }

  return {
    notifications, unreadCount, unreadNotifications, recentNotifications,
    fetchNotifications, markAsRead, markAllAsRead, deleteNotification, addNotification,
  }
})
