import type { OrderStatus } from '../types'

export const ORDER_STATUS_CONFIG: Record<OrderStatus, { color: string; label: string; next: OrderStatus[] }> = {
  pending: { color: 'blue', label: '待确认', next: ['confirmed', 'cancelled'] },
  confirmed: { color: 'orange', label: '已确认', next: ['paid', 'cancelled'] },
  paid: { color: 'cyan', label: '已支付', next: ['shipped', 'cancelled'] },
  shipped: { color: 'purple', label: '已发货', next: ['delivered'] },
  delivered: { color: 'green', label: '已送达', next: [] },
  cancelled: { color: 'red', label: '已取消', next: [] },
}

export const ORDER_STATUS_LIST: OrderStatus[] = [
  'pending',
  'confirmed',
  'paid',
  'shipped',
  'delivered',
  'cancelled',
]
