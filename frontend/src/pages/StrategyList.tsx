import { useEffect, useState } from 'react'
import { listStrategies, Strategy } from '../api/client'

export default function StrategyList() {
  const [strategies, setStrategies] = useState<Strategy[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  useEffect(() => {
    listStrategies()
      .then(data => setStrategies(Array.isArray(data) ? data : []))
      .catch(() => setError('加载失败，请确保后端服务运行中'))
      .finally(() => setLoading(false))
  }, [])

  return (
    <div>
      <h2 style={{ marginBottom: '16px', fontSize: '20px', fontWeight: 600 }}>策略列表</h2>

      {loading && <div className="loading">加载中...</div>}

      {error && <p className="error-msg">{error}</p>}

      {!loading && !error && strategies.length === 0 && (
        <div className="card">
          <p style={{ color: '#999', fontSize: '14px' }}>暂无策略</p>
        </div>
      )}

      <div className="strategy-list">
        {strategies.map((s, i) => (
          <div key={s.id || i} className="strategy-item">
            <div className="strategy-info">
              <span className="strategy-name">{s.name}</span>
              <span className="strategy-meta">风格: {s.style}</span>
              <span className="strategy-meta">版本: {s.version}</span>
            </div>
            <span className="strategy-meta">{s.createdAt}</span>
          </div>
        ))}
      </div>
    </div>
  )
}
