import { Form, Input, Button, message } from 'antd'
import { Link, useNavigate } from 'react-router-dom'
import { useAuthStore } from '../../stores/authStore'

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
            注册
          </h1>
          <p style={{
            fontSize: 14,
            color: 'var(--color-text-tertiary)',
            marginBottom: 32,
          }}>
            创建新账号开始使用
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
                style={{ height: 48, fontSize: 15 }}
              />
            </Form.Item>
            <Form.Item
              name="email"
              rules={[{ required: true, type: 'email', message: '请输入有效邮箱' }]}
              style={{ marginBottom: 16 }}
            >
              <Input
                placeholder="邮箱"
                size="large"
                style={{ height: 48, fontSize: 15 }}
              />
            </Form.Item>
            <Form.Item
              name="password"
              rules={[{ required: true, min: 6, message: '密码至少6位' }]}
              style={{ marginBottom: 16 }}
            >
              <Input.Password
                placeholder="密码"
                size="large"
                style={{ height: 48, fontSize: 15 }}
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
              style={{ marginBottom: 24 }}
            >
              <Input.Password
                placeholder="确认密码"
                size="large"
                style={{ height: 48, fontSize: 15 }}
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
                注册
              </Button>
            </Form.Item>
          </Form>
          <div style={{ fontSize: 14, color: 'var(--color-text-tertiary)' }}>
            已有账号？{' '}
            <Link to="/login" style={{ color: 'var(--color-accent)', fontWeight: 500 }}>
              登录
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
