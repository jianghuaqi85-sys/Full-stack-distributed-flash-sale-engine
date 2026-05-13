import { Button } from 'antd'
import { SunOutlined, MoonOutlined } from '@ant-design/icons'

interface Props {
  isDark: boolean
  onToggle: () => void
}

export default function ThemeToggle({ isDark, onToggle }: Props) {
  return (
    <Button
      type="text"
      icon={isDark ? <SunOutlined /> : <MoonOutlined />}
      onClick={onToggle}
      style={{ fontSize: 18 }}
    />
  )
}
