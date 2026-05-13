import { Line } from '@ant-design/charts'
import type { Order } from '../../types'

interface Props {
  orders: Order[]
}

export default function OrderTrendChart({ orders }: Props) {
  // 按日期统计订单数量
  const data: { date: string; count: number }[] = []
  const dateMap = new Map<string, number>()

  orders.forEach((order) => {
    if (order.created_at) {
      const date = new Date(order.created_at).toLocaleDateString('zh-CN')
      dateMap.set(date, (dateMap.get(date) || 0) + 1)
    }
  })

  dateMap.forEach((count, date) => {
    data.push({ date, count })
  })

  // 按日期排序
  data.sort((a, b) => new Date(a.date).getTime() - new Date(b.date).getTime())

  const config = {
    data,
    xField: 'date',
    yField: 'count',
    point: { size: 5, shape: 'diamond' },
    label: { style: { fill: '#aaa' } },
    smooth: true,
  }

  if (data.length === 0) {
    return <div style={{ textAlign: 'center', padding: 40, color: '#999' }}>暂无数据</div>
  }

  return <Line {...config} height={250} />
}
