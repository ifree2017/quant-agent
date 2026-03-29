interface EquityChartProps {
  data?: number[]
  title?: string
}

export default function EquityChart({ data, title = '资金曲线' }: EquityChartProps) {
  if (!data || data.length === 0) {
    return (
      <div className="equity-chart" aria-label={title}>
{`╭──────────────────────────────────────╮
│        ${title}                  │
│   (暂无数据)                           │
│                                      │
│                                      │
│                                      │
╰──────────────────────────────────────╯`}
      </div>
    )
  }

  const min = Math.min(...data)
  const max = Math.max(...data)
  const range = max - min || 1
  const rows = 10
  const cols = Math.min(data.length, 60)

  // Sample data to fit display
  const sampled = data.length > cols
    ? data.filter((_, i) => i % Math.ceil(data.length / cols) === 0)
    : data

  // Build ASCII chart
  const chart: string[][] = Array.from({ length: rows }, () => Array(cols).fill('  '))

  sampled.forEach((v, i) => {
    const col = i
    const row = Math.floor(((v - min) / range) * (rows - 1))
    chart[rows - 1 - row][col] = '█'
  })

  const chartLines = chart.map(line => line.join(''))

  const header = `╭ ${title}  min=${min.toFixed(2)}  max=${max.toFixed(2)}`
  const footer = `╰ 起始:${data[0]?.toFixed(2)} → 结束:${data[data.length - 1]?.toFixed(2)}`

  return (
    <div className="equity-chart" aria-label={title}>
      <pre>{[header, ...chartLines, footer].join('\n')}</pre>
    </div>
  )
}
