import { Link } from 'react-router-dom'
import MetricCard from '../components/MetricCard'
import { useEffect, useState } from 'react'
import { getStats } from '../api/client'

interface Stats {
  totalStrategies: number
  totalBacktests: number
  avgReturn: number
  avgSharpe: number
}

export default function Home() {
  const [stats, setStats] = useState<Stats | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  useEffect(() => {
    getStats()
      .then(s => setStats(s))
      .catch(() => setError('加载统计数据失败，请确保后端服务运行中'))
      .finally(() => setLoading(false))
  }, [])

  return (
    <div>
      <h2 style={{ marginBottom: '16px', fontSize: '20px', fontWeight: 600 }}>QuantAgent T12 控制台</h2>

      <div className="metric-grid">
        <MetricCard label="总策略数" value={loading ? '...' : (stats?.totalStrategies ?? '-')} />
        <MetricCard label="总回测次数" value={loading ? '...' : (stats?.totalBacktests ?? '-')} />
        <MetricCard label="平均收益率" value={loading ? '...' : `${((stats?.avgReturn ?? 0) * 100).toFixed(2)}%`} />
        <MetricCard label="平均夏普比率" value={loading ? '...' : (stats?.avgSharpe?.toFixed(2) ?? '-')} />
      </div>

      {error && <p className="error-msg">{error}</p>}

      <div className="card">
        <div className="card-title">快捷入口</div>
        <div className="quick-links">
          <Link to="/style-analyze" className="quick-link">风格分析</Link>
          <Link to="/strategy-generate" className="quick-link">策略生成</Link>
          <Link to="/backtest" className="quick-link">回测</Link>
          <Link to="/strategies" className="quick-link">策略列表</Link>
        </div>
      </div>
    </div>
  )
}
