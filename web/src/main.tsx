import { useEffect, useState } from 'react'
import ReactDOM from 'react-dom/client'
import { Spin } from 'antd'
import App from './App'
import { useAuthStore } from './stores/authStore'
import './index.css'

function Root() {
  const [ready, setReady] = useState(false)
  const loadFromStorage = useAuthStore((s) => s.loadFromStorage)

  useEffect(() => {
    const init = async () => {
      await loadFromStorage()
      setReady(true)
    }
    init()
  }, [loadFromStorage])

  if (!ready) {
    return (
      <div style={{ minHeight: '100vh', display: 'flex', justifyContent: 'center', alignItems: 'center' }}>
        <Spin size="large" tip="加载中..." />
      </div>
    )
  }

  return <App />
}

ReactDOM.createRoot(document.getElementById('root')!).render(<Root />)
