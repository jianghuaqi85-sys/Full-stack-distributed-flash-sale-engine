import { Form, Input, Button, Typography, message } from 'antd'
import { UserOutlined, LockOutlined } from '@ant-design/icons'
import { Link, useNavigate } from 'react-router-dom'
import { useAuthStore } from '../../stores/authStore'
import BrandLogo from '../../components/BrandLogo'

const { Text } = Typography

export default function Login() {
  const navigate = useNavigate()
  const { login, loading } = useAuthStore()

  const onFinish = async (values: { username: string; password: string }) => {
    try {
      await login(values.username, values.password)
      message.success('登录成功')
      navigate('/')
    } catch {
      // 错误已由 axios 拦截器处理
    }
  }

  return (
    <div className="login-bg" style={{ display: 'flex', justifyContent: 'center', alignItems: 'center' }}>
      <div className="glass-card animate-slide-up" style={{ width: 420, padding: '40px 36px' }}>
        <div style={{ textAlign: 'center', marginBottom: 32 }}>
          <BrandLogo size={48} />
          <h1 style={{
            margin: '16px 0 4px',
            fontSize: 24,
            fontWeight: 700,
            color: 'var(--color-text-inverse)',
          }}>
            票务系统
          </h1>
          <Text style={{ color: 'rgba(255,255,255,0.6)' }}>请登录您的账号</Text>
        </div>
        <Form onFinish={onFinish} autoComplete="off" size="large">
          <Form.Item name="username" rules={[{ required: true, message: '请输入用户名' }]}>
            <Input
              prefix={<UserOutlined style={{ color: 'rgba(0,0,0,0.45)' }} />}
              placeholder="用户名"
              style={{
                background: 'rgba(255,255,255,0.08)',
                border: '1px solid rgba(255,255,255,0.15)',
                borderRadius: 12,
              }}
            />
          </Form.Item>
          <Form.Item name="password" rules={[{ required: true, message: '请输入密码' }]}>
            <Input.Password
              prefix={<LockOutlined style={{ color: 'rgba(0,0,0,0.45)' }} />}
              placeholder="密码"
              style={{
                background: 'rgba(255,255,255,0.08)',
                border: '1px solid rgba(255,255,255,0.15)',
                borderRadius: 12,
              }}
            />
          </Form.Item>
          <Form.Item>
            <Button
              type="primary"
              htmlType="submit"
              loading={loading}
              block
              style={{
                height: 48,
                borderRadius: 12,
                fontWeight: 600,
                fontSize: 16,
                background: 'linear-gradient(135deg, #5B2FE8, #D4A843)',
                border: 'none',
                boxShadow: '0 4px 20px rgba(91, 47, 232, 0.4)',
              }}
            >
              登录
            </Button>
          </Form.Item>
        </Form>
        <div style={{ textAlign: 'center' }}>
          <Text style={{ color: 'rgba(255,255,255,0.5)' }}>还没有账号？</Text>
          <Link to="/register" style={{ color: '#D4A843', fontWeight: 500, marginLeft: 4 }}>立即注册</Link>
        </div>
      </div>
    </div>
  )
}
