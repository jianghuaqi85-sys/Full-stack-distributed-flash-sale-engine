import { Form, Input, Button, message } from 'antd'
import { Link, useNavigate } from 'react-router-dom'
import { useAuthStore } from '../../stores/authStore'

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
    <div className="login-container">
      {/* Left — Form */}
      <div className="login-form-side">
        <div style={{ width: 360, maxWidth: '100%' }}>
          <div style={{ marginBottom: 48 }}>
            <span style={{
              fontFamily: 'var(--font-mono)',
              fontSize: 14,
              fontWeight: 700,
              color: 'var(--color-text-primary)',
              letterSpacing: '-0.04em',
            }}>
              TICKET
            </span>
          </div>
          <h1 style={{
            fontSize: 32,
            fontWeight: 700,
            letterSpacing: '-0.03em',
            color: 'var(--color-text-primary)',
            marginBottom: 8,
          }}>
            登录
          </h1>
          <p style={{
            fontSize: 14,
            color: 'var(--color-text-tertiary)',
            marginBottom: 32,
          }}>
            输入您的账号信息继续
          </p>
          <Form onFinish={onFinish} autoComplete="off" layout="vertical" requiredMark={false}>
            <Form.Item
              name="username"
              rules={[{ required: true, message: '请输入用户名' }]}
              style={{ marginBottom: 16 }}
            >
              <Input
                placeholder="用户名"
                size="large"
                style={{
                  height: 48,
                  fontSize: 15,
                }}
              />
            </Form.Item>
            <Form.Item
              name="password"
              rules={[{ required: true, message: '请输入密码' }]}
              style={{ marginBottom: 24 }}
            >
              <Input.Password
                placeholder="密码"
                size="large"
                style={{
                  height: 48,
                  fontSize: 15,
                }}
              />
            </Form.Item>
            <Form.Item style={{ marginBottom: 16 }}>
              <Button
                type="primary"
                htmlType="submit"
                loading={loading}
                block
                style={{
                  height: 48,
                  fontWeight: 600,
                  fontSize: 15,
                }}
              >
                登录
              </Button>
            </Form.Item>
          </Form>
          <div style={{ fontSize: 14, color: 'var(--color-text-tertiary)' }}>
            还没有账号？{' '}
            <Link to="/register" style={{ color: 'var(--color-accent)', fontWeight: 500 }}>
              注册
            </Link>
          </div>
        </div>
      </div>

      {/* Right — Visual */}
      <div className="login-visual-side">
        <h2>TICKET</h2>
        <p>分布式票务系统 — 高性能秒杀、实时排队、安全交易</p>
      </div>
    </div>
  )
}
