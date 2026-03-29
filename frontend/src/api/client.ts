const API = 'http://localhost:8080'

export interface StyleAnalyzeResult {
  style: string
  riskScore: number
  tradeFrequency: string
}

export interface Strategy {
  id: string
  name: string
  style: string
  version: string
  createdAt: string
}

export interface BacktestResult {
  returnRate: number
  sharpeRatio: number
  maxDrawdown: number
  winRate: number
  equity?: number[]
  trades?: any[]
}

export const analyzeStyle = (records: any[]): Promise<StyleAnalyzeResult> =>
  fetch(`${API}/api/style/analyze`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ userID: 'default', records })
  }).then(r => r.json())

export const generateStrategy = (styleProfile: any, symbol: string): Promise<any> =>
  fetch(`${API}/api/strategy/generate`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ userID: 'default', styleProfile, symbol })
  }).then(r => r.json())

export const runBacktest = (params: any): Promise<BacktestResult> =>
  fetch(`${API}/api/backtest`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(params)
  }).then(r => r.json())

export const listStrategies = (): Promise<Strategy[]> =>
  fetch(`${API}/api/strategies`).then(r => r.json())

export const getStats = (): Promise<{ totalStrategies: number; totalBacktests: number; avgReturn: number; avgSharpe: number }> =>
  fetch(`${API}/api/stats`).then(r => r.json())
