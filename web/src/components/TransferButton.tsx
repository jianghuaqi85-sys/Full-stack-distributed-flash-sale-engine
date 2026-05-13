import { useState } from 'react'
import { Button, Modal, Input, Form, Radio, message } from 'antd'
import { SwapOutlined } from '@ant-design/icons'
import { requestTransfer, directGift } from '../api/transfer'

interface TransferButtonProps {
  ticketId: number
  disabled?: boolean
  onSuccess?: () => void
}

export default function TransferButton({ ticketId, disabled = false, onSuccess }: TransferButtonProps) {
  const [modalVisible, setModalVisible] = useState(false)
  const [loading, setLoading] = useState(false)
  const [form] = Form.useForm()

  const handleTransfer = async () => {
    try {
      const values = await form.validateFields()
      setLoading(true)

      const payload = {
        ticket_id: ticketId,
        to_user_id: parseInt(values.to_user_id),
        reason: values.reason,
      }

      if (values.transfer_type === 'gift') {
        await directGift(payload)
        message.success('转赠成功！')
      } else {
        await requestTransfer(payload)
        message.success('转让请求已提交，等待管理员审核')
      }

      setModalVisible(false)
      form.resetFields()
      onSuccess?.()
    } catch (error: unknown) {
      const err = error as { response?: { data?: { error?: string } } }
      if (err.response?.data?.error) {
        message.error(err.response.data.error)
      }
    } finally {
      setLoading(false)
    }
  }

  return (
    <>
      <Button
        icon={<SwapOutlined />}
        onClick={() => setModalVisible(true)}
        disabled={disabled}
      >
        转让
      </Button>

      <Modal
        title="转让票务"
        open={modalVisible}
        onOk={handleTransfer}
        onCancel={() => setModalVisible(false)}
        confirmLoading={loading}
      >
        <Form form={form} layout="vertical" initialValues={{ transfer_type: 'gift' }}>
          <Form.Item name="transfer_type" label="转让方式">
            <Radio.Group>
              <Radio value="gift">直接转赠（无需审核，立即生效）</Radio>
              <Radio value="review">申请转让（需管理员审核）</Radio>
            </Radio.Group>
          </Form.Item>
          <Form.Item
            name="to_user_id"
            label="目标用户ID"
            rules={[{ required: true, message: '请输入目标用户ID' }]}
          >
            <Input placeholder="请输入目标用户ID" type="number" />
          </Form.Item>
          <Form.Item name="reason" label="转让原因">
            <Input.TextArea rows={3} placeholder="请输入转让原因（选填）" />
          </Form.Item>
        </Form>
      </Modal>
    </>
  )
}
