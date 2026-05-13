import { Form, Input, Button, Typography, message } from 'antd'
import { UserOutlined, LockOutlined, MailOutlined } from '@ant-design/icons'
import { Link, useNavigate } from 'react-router-dom'
import { useAuthStore } from '../../stores/authStore'
import BrandLogo from '../../components/BrandLogo'

const { Text } = Typography

export default function Register() {
  const navigate = useNavigate()
  const { register, loading } = useAuthStore()

  const onFinish = async (values: { username: string; password: string; email: string }) => {
    try {
      await register(values.username, values.password, values.email)
      message.success('注册成功，请登录')
      navigate('/login')
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
            注册账号
          </h1>
          <Text style={{ color: 'rgba(255,255,255,0.6)' }}>创建您的新账号</Text>
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
          <Form.Item name="email" rules={[{ required: true, type: 'email', message: '请输入有效邮箱' }]}>
            <Input
              prefix={<MailOutlined style={{ color: 'rgba(0,0,0,0.45)' }} />}
              placeholder="邮箱"
              style={{
                background: 'rgba(255,255,255,0.08)',
                border: '1px solid rgba(255,255,255,0.15)',
                borderRadius: 12,
              }}
            />
          </Form.Item>
          <Form.Item name="password" rules={[{ required: true, min: 6, message: '密码至少6位' }]}>
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
          <Form.Item
            name="confirmPassword"
            dependencies={['password']}
            rules={[
              { required: true, message: '请确认密码' },
              ({ getFieldValue }) => ({
                validator(_, value) {
                  if (!value || getFieldValue('password') === value) {
                    return Promise.resolve()
                  }
                  return Promise.reject(new Error('两次密码不一致'))
                },
              }),
            ]}
          >
            <Input.Password
              prefix={<LockOutlined style={{ color: 'rgba(0,0,0,0.45)' }} />}
              placeholder="确认密码"
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
              注册
            </Button>
          </Form.Item>
        </Form>
        <div style={{ textAlign: 'center' }}>
          <Text style={{ color: 'rgba(255,255,255,0.5)' }}>已有账号？</Text>
          <Link to="/login" style={{ color: '#D4A843', fontWeight: 500, marginLeft: 4 }}>去登录</Link>
        </div>
      </div>
    </div>
  )
}
