import { ref, reactive } from 'vue'

// Global reactive toast state
const toasts = reactive([])
let toastId = 0

/**
 * Show a toast notification.
 * @param {'success'|'error'|'info'|'warning'} type - Toast type
 * @param {string} message - Message to display
 * @param {number} duration - Duration in ms before auto-dismiss (default 3000)
 */
export function showToast(type, message, duration = 3000) {
  const id = ++toastId
  toasts.push({ id, type, message, duration })
  if (duration > 0) {
    setTimeout(() => {
      removeToast(id)
    }, duration)
  }
}

/**
 * Remove a toast by id.
 * @param {number} id
 */
export function removeToast(id) {
  const index = toasts.findIndex(t => t.id === id)
  if (index !== -1) {
    toasts.splice(index, 1)
  }
}

/**
 * Convenience methods
 */
export function toastSuccess(message, duration) {
  showToast('success', message, duration)
}

export function toastError(message, duration) {
  showToast('error', message, duration)
}

export function toastInfo(message, duration) {
  showToast('info', message, duration)
}

export function toastWarning(message, duration) {
  showToast('warning', message, duration)
}

/**
 * Vue composable for using the toast system.
 */
export function useToast() {
  return {
    toasts,
    showToast,
    removeToast,
    toastSuccess,
    toastError,
    toastInfo,
    toastWarning,
  }
}
