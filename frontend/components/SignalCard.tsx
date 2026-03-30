'use client';
import { useState, useEffect } from 'react';

interface SignalData {
  code: string;
  signal: string;
  score: number;
  reason?: string;
}

interface SignalCardProps {
  code: string;
  name: string;
}

export default function SignalCard({ code, name }: SignalCardProps) {
  const [distribution, setDistribution] = useState<SignalData | null>(null);
  const [buyPoint, setBuyPoint] = useState<SignalData | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!code) return;
    setLoading(true);
    setError(null);
    Promise.all([
      fetch(`/api/v1/distribution/${code}`).then(r => r.json()).catch(() => null),
      fetch(`/api/v1/buypoint/${code}`).then(r => r.json()).catch(() => null)
    ]).then(([d, b]) => {
      setDistribution(d);
      setBuyPoint(b);
    }).catch(() => {
      setError('信号加载失败');
    }).finally(() => {
      setLoading(false);
    });
  }, [code]);

  if (loading) {
    return (
      <div className="bg-slate-800/50 rounded-xl border border-slate-700 p-3">
        <p className="text-xs text-slate-500">信号加载中…</p>
      </div>
    );
  }

  if (error) {
    return (
      <div className="bg-slate-800/50 rounded-xl border border-slate-700 p-3">
        <p className="text-xs text-red-400">⚠ {error}</p>
      </div>
    );
  }

  const isStrongDistribution = distribution?.signal === 'STRONG_DISTRIBUTION';
  const isStrongBuy = buyPoint?.signal === 'STRONG_BUY';

  return (
    <div className="bg-slate-800/50 rounded-xl border border-slate-700 p-3">
      <h3 className="text-xs font-semibold mb-2 flex items-center gap-1">
        <span>📊</span> 交易信号 <span className="text-slate-500 font-normal">— {name}({code})</span>
      </h3>
      <div className="grid grid-cols-2 gap-2">
        {/* 出货信号 */}
        <div className={`rounded-lg p-2 border ${isStrongDistribution ? 'bg-red-500/10 border-red-500/40' : 'bg-slate-900/60 border-slate-700'}`}>
          <div className="text-[10px] text-slate-500 mb-1">🚢 出货信号</div>
          <div className={`text-xs font-semibold ${isStrongDistribution ? 'text-red-400' : 'text-slate-400'}`}>
            {distribution?.signal || '—'}
          </div>
          <div className="text-[10px] text-slate-500 mt-0.5">
            评分: {distribution?.score ?? '—'}
          </div>
          {distribution?.reason && (
            <div className="text-[9px] text-slate-600 mt-0.5 truncate">{distribution.reason}</div>
          )}
        </div>
        {/* 买点信号 */}
        <div className={`rounded-lg p-2 border ${isStrongBuy ? 'bg-green-500/10 border-green-500/40' : 'bg-slate-900/60 border-slate-700'}`}>
          <div className="text-[10px] text-slate-500 mb-1">🛒 买点信号</div>
          <div className={`text-xs font-semibold ${isStrongBuy ? 'text-green-400' : 'text-slate-400'}`}>
            {buyPoint?.signal || '—'}
          </div>
          <div className="text-[10px] text-slate-500 mt-0.5">
            评分: {buyPoint?.score ?? '—'}
          </div>
          {buyPoint?.reason && (
            <div className="text-[9px] text-slate-600 mt-0.5 truncate">{buyPoint.reason}</div>
          )}
        </div>
      </div>
    </div>
  );
}
