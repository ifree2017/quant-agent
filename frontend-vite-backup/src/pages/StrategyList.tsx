import { useEffect, useState } from 'react'
import { listStrategies, deleteStrategy } from '../api/client'

interface Strategy {
  id: string
  name: string
  style: string
  version: number
  created_at: string
  rules: any
}

export default function StrategyList() {
  const [strategies, setStrategies] = useState<Strategy[]>([])
  const [loading, setLoading] = useState(false)

  useEffect(() => {
    load()
  }, [])

  async function load() {
    setLoading(true)
    try {
      const data = await listStrategies()
      setStrategies(data.strategies || [])
    } catch (e) {
      console.error(e)
    } finally {
      setLoading(false)
    }
  }

  async function handleDelete(id: string) {
    if (!confirm('确认删除？')) return
    await deleteStrategy(id)
    load()
  }

  if (loading) return <div className="loading">加载中...</div>

  return (
    <div className="container">
      <h2>📊 策略库</h2>
      {strategies.length === 0 ? (
        <p>暂无策略，请先 <a href="/strategy">生成策略</a></p>
      ) : (
        <table className="table">
          <thead>
            <tr>
              <th>策略名称</th>
              <th>风格</th>
              <th>版本</th>
              <th>创建时间</th>
              <th>操作</th>
            </tr>
          </thead>
          <tbody>
            {strategies.map((s) => (
              <tr key={s.id}>
                <td>{s.name}</td>
                <td>{s.style}</td>
                <td>v{s.version}</td>
                <td>{new Date(s.created_at).toLocaleDateString()}</td>
                <td>
                  <button onClick={() => window.location.href = `/backtest?strategyId=${s.id}`}>
                    回测
                  </button>
                  <button onClick={() => handleDelete(s.id)} className="btn-danger">
                    删除
                  </button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
    </div>
  )
}
