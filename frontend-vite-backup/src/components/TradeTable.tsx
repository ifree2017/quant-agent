interface Trade {
  date: string
  symbol: string
  action: string
  price: number
  quantity: number
}

interface TradeTableProps {
  trades: Trade[]
}

export default function TradeTable({ trades }: TradeTableProps) {
  if (!trades || trades.length === 0) {
    return <p style={{ color: '#999', fontSize: '14px' }}>暂无交易记录</p>
  }

  return (
    <table className="table">
      <thead>
        <tr>
          <th>日期</th>
          <th>标的</th>
          <th>操作</th>
          <th>价格</th>
          <th>数量</th>
        </tr>
      </thead>
      <tbody>
        {trades.map((t, i) => (
          <tr key={i}>
            <td>{t.date}</td>
            <td>{t.symbol}</td>
            <td>{t.action}</td>
            <td>{t.price}</td>
            <td>{t.quantity}</td>
          </tr>
        ))}
      </tbody>
    </table>
  )
}
