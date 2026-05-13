import { useEffect, useState, useCallback } from 'react'
import { Card, Table, Button, Tag, Space, Modal, Form, InputNumber, Input, Select, message, Tabs, Empty } from 'antd'
import { ShoppingOutlined, PlusOutlined, EyeOutlined } from '@ant-design/icons'
import { useNavigate } from 'react-router-dom'
import { getActiveListings, getMyListings, getMyPurchases, createListing, buyListing, cancelListing, MarketplaceListing } from '../../api/marketplace'
import { getMyTickets } from '../../api/tickets'
import { useAuthStore } from '../../stores/authStore'

interface MyTicket {
  id: number
  order_no: string
  ticket_name: string
  status: string
}

export default function Marketplace() {
  const { user } = useAuthStore()
  const navigate = useNavigate()
  const [listings, setListings] = useState<MarketplaceListing[]>([])
  const [myListings, setMyListings] = useState<MarketplaceListing[]>([])
  const [myPurchases, setMyPurchases] = useState<MarketplaceListing[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [loading, setLoading] = useState(false)
  const [sellModalVisible, setSellModalVisible] = useState(false)
  const [form] = Form.useForm()
  const [myTickets, setMyTickets] = useState<MyTicket[]>([])

  const fetchListings = useCallback(() => {
    setLoading(true)
    getActiveListings(page, 10)
      .then((res) => {
        setListings(res.data.data || [])
        setTotal(res.data.total)
      })
      .finally(() => setLoading(false))
  }, [page])

  const fetchMyData = useCallback(() => {
    getMyListings()
      .then((res) => setMyListings(res.data.data || []))
      .catch(() => {})
    getMyPurchases()
      .then((res) => setMyPurchases(res.data.data || []))
      .catch(() => {})
  }, [])

  const fetchMyTickets = useCallback(() => {
    getMyTickets(1, 100)
      .then((res) => {
        const paid = (res.data.data || []).filter((t: MyTicket) => t.status === 'paid')
        setMyTickets(paid)
      })
      .catch(() => {})
  }, [])

  useEffect(() => {
    fetchListings()
  }, [fetchListings])

  useEffect(() => {
    fetchMyData()
    fetchMyTickets()
  }, [fetchMyData, fetchMyTickets])

  const handleBuy = async (id: number) => {
    try {
      await buyListing(id)
      message.success('购买成功')
      fetchListings()
      fetchMyData()
    } catch {
      // handled by interceptor
    }
  }

  const handleCancel = async (id: number) => {
    try {
      await cancelListing(id)
      message.success('下架成功')
      fetchMyData()
    } catch {
      // handled by interceptor
    }
  }

  const handleSell = async () => {
    try {
      const values = await form.validateFields()
      await createListing(values)
      message.success('上架成功')
      setSellModalVisible(false)
      form.resetFields()
      fetchMyData()
      fetchMyTickets()
    } catch {
      // validation error
    }
  }

  const columns = [
    {
      title: '票种',
      dataIndex: 'ticket_name',
      key: 'ticket_name',
    },
    {
      title: '售价',
      dataIndex: 'price',
      key: 'price',
      render: (price: number) => <span style={{ color: 'var(--color-gold)', fontWeight: 600 }}>¥{price}</span>,
    },
    {
      title: '描述',
      dataIndex: 'description',
      key: 'description',
      ellipsis: true,
    },
    {
      title: '发布时间',
      dataIndex: 'created_at',
      key: 'created_at',
    },
    {
      title: '操作',
      key: 'action',
      render: (_: unknown, record: MarketplaceListing) => (
        <Space>
          <Button size="small" icon={<EyeOutlined />} onClick={() => navigate(`/marketplace/${record.id}`)}>
            详情
          </Button>
          {record.seller_id !== user?.id && (
            <Button type="primary" size="small" icon={<ShoppingOutlined />} onClick={() => handleBuy(record.id)}>
              购买
            </Button>
          )}
          {record.seller_id === user?.id && record.status === 'active' && (
            <Button danger size="small" onClick={() => handleCancel(record.id)}>
              下架
            </Button>
          )}
        </Space>
      ),
    },
  ]

  const myListingColumns = [
    {
      title: '票种',
      dataIndex: 'ticket_name',
      key: 'ticket_name',
    },
    {
      title: '售价',
      dataIndex: 'price',
      key: 'price',
      render: (price: number) => <span style={{ color: 'var(--color-gold)', fontWeight: 600 }}>¥{price}</span>,
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      render: (status: string) => (
        <Tag color={status === 'active' ? 'success' : status === 'sold' ? 'default' : 'warning'}>
          {status === 'active' ? '在售' : status === 'sold' ? '已售' : '已下架'}
        </Tag>
      ),
    },
    {
      title: '操作',
      key: 'action',
      render: (_: unknown, record: MarketplaceListing) => (
        record.status === 'active' && (
          <Button danger size="small" onClick={() => handleCancel(record.id)}>下架</Button>
        )
      ),
    },
  ]

  return (
    <div>
      <Card title={<><ShoppingOutlined /> 二手票市场</>}>
        <Tabs
          defaultActiveKey="market"
          items={[
            {
              key: 'market',
              label: '在售票务',
              children: (
                <Table
                  columns={columns}
                  dataSource={listings}
                  rowKey="id"
                  loading={loading}
                  pagination={{
                    current: page,
                    total,
                    pageSize: 10,
                    onChange: setPage,
                    showTotal: (t) => `共 ${t} 件`,
                  }}
                  locale={{ emptyText: <Empty description="暂无在售票务" /> }}
                />
              ),
            },
            {
              key: 'my-sell',
              label: '我的上架',
              children: (
                <>
                  <Button
                    type="primary"
                    icon={<PlusOutlined />}
                    style={{ marginBottom: 16 }}
                    onClick={() => {
                      fetchMyTickets()
                      setSellModalVisible(true)
                    }}
                  >
                    上架票务
                  </Button>
                  <Table
                    columns={myListingColumns}
                    dataSource={myListings}
                    rowKey="id"
                    pagination={false}
                    locale={{ emptyText: <Empty description="暂无上架记录" /> }}
                  />
                </>
              ),
            },
            {
              key: 'my-buy',
              label: '我的购买',
              children: (
                <Table
                  columns={[
                    { title: '票种', dataIndex: 'ticket_name', key: 'ticket_name' },
                    {
                      title: '价格',
                      dataIndex: 'price',
                      key: 'price',
                      render: (price: number) => <span style={{ color: 'var(--color-gold)', fontWeight: 600 }}>¥{price}</span>,
                    },
                    { title: '购买时间', dataIndex: 'created_at', key: 'created_at' },
                  ]}
                  dataSource={myPurchases}
                  rowKey="id"
                  pagination={false}
                  locale={{ emptyText: <Empty description="暂无购买记录" /> }}
                />
              ),
            },
          ]}
        />
      </Card>

      <Modal
        title="上架票务"
        open={sellModalVisible}
        onOk={handleSell}
        onCancel={() => setSellModalVisible(false)}
        width={500}
      >
        <Form form={form} layout="vertical">
          <Form.Item name="ticket_id" label="选择票务" rules={[{ required: true, message: '请选择要上架的票务' }]}>
            <Select
              placeholder="请选择要上架的票务"
              options={myTickets.map((t) => ({
                value: t.id,
                label: `${t.ticket_name} (${t.order_no})`,
              }))}
            />
          </Form.Item>
          <Form.Item name="price" label="出售价格" rules={[{ required: true, message: '请输入价格' }]}>
            <InputNumber min={0.01} step={0.01} prefix="¥" style={{ width: '100%' }} />
          </Form.Item>
          <Form.Item name="description" label="描述">
            <Input.TextArea rows={3} placeholder="票务描述（可选）" />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}
