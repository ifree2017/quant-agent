import { BrowserRouter, Routes, Route, Link, useLocation } from 'react-router-dom'
import Home from './pages/Home'
import StyleAnalyze from './pages/StyleAnalyze'
import StrategyGenerate from './pages/StrategyGenerate'
import Backtest from './pages/Backtest'
import StrategyList from './pages/StrategyList'
import Compare from './pages/Compare'

const NAV = [
  { path: '/', label: '首页' },
  { path: '/style-analyze', label: '风格分析' },
  { path: '/strategy-generate', label: '策略生成' },
  { path: '/backtest', label: '回测' },
  { path: '/strategies', label: '策略列表' },
  { path: '/compare', label: '⚖️ 策略对比' },
]

function Header() {
  const location = useLocation()
  return (
    <header className="header">
      <div className="header-inner">
        <Link to="/" className="logo">QuantAgent T12</Link>
        <nav className="nav">
          {NAV.map(n => (
            <Link
              key={n.path}
              to={n.path}
              style={location.pathname === n.path ? { background: '#1890ff', color: '#fff' } : {}}
            >
              {n.label}
            </Link>
          ))}
        </nav>
      </div>
    </header>
  )
}

export default function App() {
  return (
    <BrowserRouter>
      <div className="app">
        <Header />
        <main className="main">
          <Routes>
            <Route path="/" element={<Home />} />
            <Route path="/style-analyze" element={<StyleAnalyze />} />
            <Route path="/strategy-generate" element={<StrategyGenerate />} />
            <Route path="/backtest" element={<Backtest />} />
            <Route path="/strategies" element={<StrategyList />} />
            <Route path="/compare" element={<Compare />} />
          </Routes>
        </main>
      </div>
    </BrowserRouter>
  )
}
