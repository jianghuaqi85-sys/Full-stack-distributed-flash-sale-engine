import { useState } from 'react'
import { Input, Button, message, Typography } from 'antd'
import { TagOutlined } from '@ant-design/icons'
import { validatePromoCode, ValidatePromoResponse } from '../api/promo'

const { Text } = Typography

interface PromoCodeInputProps {
  amount: number
  onApply: (discount: number, finalAmount: number, code: string) => void
}

export default function PromoCodeInput({ amount, onApply }: PromoCodeInputProps) {
  const [code, setCode] = useState('')
  const [loading, setLoading] = useState(false)
  const [result, setResult] = useState<ValidatePromoResponse | null>(null)

  const handleValidate = async () => {
    if (!code.trim()) {
      message.warning('请输入促销码')
      return
    }

    setLoading(true)
    try {
      const res = await validatePromoCode(code, amount)
      setResult(res.data)
      onApply(res.data.discount, res.data.final_amount, res.data.code)
      message.success(`优惠 ${res.data.discount.toFixed(2)} 元`)
    } catch (error: unknown) {
      const err = error as { response?: { data?: { error?: string } } }
      message.error(err.response?.data?.error || '促销码无效')
      setResult(null)
    } finally {
      setLoading(false)
    }
  }

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
      <div style={{ display: 'flex', gap: 8 }}>
        <Input
          placeholder="输入促销码"
          value={code}
          onChange={(e) => setCode(e.target.value.toUpperCase())}
          prefix={<TagOutlined />}
          style={{ flex: 1 }}
        />
        <Button onClick={handleValidate} loading={loading}>
          验证
        </Button>
      </div>

      {result && (
        <div style={{ padding: '8px 12px', background: '#f6ffed', borderRadius: 6, border: '1px solid #b7eb8f' }}>
          <Text type="success">
            促销码 {result.code} 有效，优惠 {result.discount.toFixed(2)} 元
          </Text>
        </div>
      )}
    </div>
  )
}
