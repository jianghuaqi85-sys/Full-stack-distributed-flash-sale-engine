import { Pie } from '@ant-design/charts'
import type { Order } from '../../types'
import { ORDER_STATUS_CONFIG } from '../../utils/constants'

interface Props {
  orders: Order[]
}

export default function StatusPieChart({ orders }: Props) {
  // 统计各状态数量
  const statusCount = new Map<string, number>()
  orders.forEach((order) => {
    statusCount.set(order.status, (statusCount.get(order.status) || 0) + 1)
  })

  const data = Array.from(statusCount.entries()).map(([status, count]) => ({
    type: ORDER_STATUS_CONFIG[status as keyof typeof ORDER_STATUS_CONFIG]?.label || status,
    value: count,
  }))

  const config = {
    data,
    angleField: 'value',
    colorField: 'type',
    radius: 0.8,
    innerRadius: 0.6,
    label: {
      text: 'type',
      position: 'outside' as const,
    },
    legend: {
      position: 'bottom' as const,
    },
    interactions: [{ type: 'element-active' }],
  }

  if (data.length === 0) {
    return <div style={{ textAlign: 'center', padding: 40, color: '#999' }}>暂无数据</div>
  }

  return <Pie {...config} height={250} />
}
