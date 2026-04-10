import { useState, useEffect, useCallback } from 'react'
import { motion, AnimatePresence } from 'framer-motion'
import {
  LayoutDashboard, ShoppingBag, Package, Radio, MessageSquare,
  ChevronRight, TrendingUp, Clock, CheckCircle2, AlertCircle,
  RefreshCw, X, ChevronLeft, ChevronRight as ChevronRightIcon,
  Upload, Plus, Trash2, Eye, ArrowUpRight, Filter, Search,
  Activity, Zap, Store, Calendar, Tag, BarChart2, Inbox,
  ArrowRight, Circle,
} from 'lucide-react'

const API_BASE = '/api/v1'
const BRAND    = 'linear-gradient(135deg,#00b8e6 0%,#1e2fa0 100%)'
const BRAND2   = 'linear-gradient(135deg,#e6006e 0%,#7c3aed 100%)'

// ─── helpers ────────────────────────────────────────────────────────────────

async function apiFetch(path, opts = {}) {
  const res = await fetch(`${API_BASE}${path}`, {
    headers: { 'Content-Type': 'application/json' }, ...opts,
  })
  if (!res.ok) { const b = await res.text(); throw new Error(b || res.statusText) }
  return res.json()
}

function useApi(path, deps = []) {
  const [data, setData]       = useState(null)
  const [loading, setLoading] = useState(true)
  const [error, setError]     = useState(null)
  const load = useCallback(async () => {
    if (!path) return
    setLoading(true); setError(null)
    try { setData(await apiFetch(path)) }
    catch (e) { setError(e.message) }
    finally { setLoading(false) }
  }, [path, ...deps]) // eslint-disable-line react-hooks/exhaustive-deps
  useEffect(() => { load() }, [load])
  return { data, loading, error, reload: load }
}

const fmtDate = ts => {
  if (!ts) return '—'
  const d = new Date(ts)
  return d.toLocaleDateString('en-IN', { day: '2-digit', month: 'short', year: 'numeric' })
    + ' ' + d.toLocaleTimeString('en-IN', { hour: '2-digit', minute: '2-digit' })
}

const fmtDateShort = ts => {
  if (!ts) return '—'
  return new Date(ts).toLocaleDateString('en-IN', { day: '2-digit', month: 'short', year: 'numeric' })
}

function fmtName(obj) {
  if (!obj) return '—'
  if (typeof obj === 'string') { try { obj = JSON.parse(obj) } catch { return obj } }
  return obj?.name || obj?.code || obj?.descriptor?.name || '—'
}

// ─── base components ─────────────────────────────────────────────────────────

function StatusBadge({ status }) {
  const map = {
    ACTIVE:    { cls: 'bg-emerald-500/15 text-emerald-400 border-emerald-500/25', dot: 'bg-emerald-400' },
    DRAFT:     { cls: 'bg-amber-500/15  text-amber-400  border-amber-500/25',    dot: 'bg-amber-400' },
    CANCELLED: { cls: 'bg-red-500/15    text-red-400    border-red-500/25',      dot: 'bg-red-400' },
    COMPLETE:  { cls: 'bg-sky-500/15    text-sky-400    border-sky-500/25',      dot: 'bg-sky-400' },
    ACK:       { cls: 'bg-emerald-500/15 text-emerald-400 border-emerald-500/25', dot: 'bg-emerald-400' },
    NACK:      { cls: 'bg-red-500/15    text-red-400    border-red-500/25',      dot: 'bg-red-400' },
  }
  const s = map[status] || { cls: 'bg-slate-500/15 text-slate-400 border-slate-500/25', dot: 'bg-slate-400' }
  return (
    <span className={`inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-full text-xs font-semibold border ${s.cls}`}>
      <span className={`w-1.5 h-1.5 rounded-full ${s.dot}`} />
      {status || '—'}
    </span>
  )
}

function Card({ children, className = '', style }) {
  return (
    <div className={`bg-[#111827] border border-white/[0.07] rounded-2xl ${className}`} style={style}>
      {children}
    </div>
  )
}

function LoadingSpinner({ label }) {
  return (
    <div className="flex flex-col items-center justify-center py-16 gap-3">
      <div className="w-8 h-8 rounded-full border-2 border-slate-700 border-t-cyan-400 animate-spin" />
      {label && <p className="text-xs text-slate-500">{label}</p>}
    </div>
  )
}

function ErrorBox({ message, onRetry }) {
  return (
    <div className="flex items-center gap-3 bg-red-500/10 border border-red-500/20 rounded-xl p-4 text-sm text-red-400">
      <AlertCircle size={16} className="shrink-0" />
      <span className="flex-1 text-xs">{message}</span>
      {onRetry && <button onClick={onRetry} className="text-xs font-semibold underline underline-offset-2 shrink-0">Retry</button>}
    </div>
  )
}

function SectionTitle({ children }) {
  return (
    <p className="text-[10px] font-bold uppercase tracking-widest text-slate-500 mb-3">{children}</p>
  )
}

function Pagination({ page, totalPages, onPage, total, limit }) {
  if (totalPages <= 1) return null
  return (
    <div className="flex items-center justify-between px-5 py-3 border-t border-white/[0.06]">
      <span className="text-xs text-slate-500">{total} total · page {page}/{totalPages}</span>
      <div className="flex gap-1.5">
        <button disabled={page <= 1} onClick={() => onPage(p => p - 1)}
          className="p-1.5 rounded-lg border border-white/[0.08] text-slate-400 hover:text-white disabled:opacity-30 transition-colors">
          <ChevronLeft size={14} />
        </button>
        <button disabled={page >= totalPages} onClick={() => onPage(p => p + 1)}
          className="p-1.5 rounded-lg border border-white/[0.08] text-slate-400 hover:text-white disabled:opacity-30 transition-colors">
          <ChevronRightIcon size={14} />
        </button>
      </div>
    </div>
  )
}

// ─── Sidebar ─────────────────────────────────────────────────────────────────

const NAV_ITEMS = [
  { id: 'overview',  label: 'Overview',        icon: LayoutDashboard, desc: 'Stats & insights' },
  { id: 'orders',    label: 'Orders',           icon: ShoppingBag,     desc: 'Manage contracts' },
  { id: 'inventory', label: 'Inventory',        icon: Package,         desc: 'Resources & offers' },
  { id: 'publish',   label: 'Publish Catalog',  icon: Radio,           desc: 'Push to network' },
  { id: 'messages',  label: 'Message Log',      icon: MessageSquare,   desc: 'Protocol audit' },
]

function Sidebar({ active, onNav }) {
  return (
    <aside className="fixed inset-y-0 left-0 z-40 flex flex-col w-64 border-r border-white/[0.06]"
      style={{ background: '#0a0e1a' }}>
      {/* Logo */}
      <div className="px-5 pt-5 pb-4 border-b border-white/[0.06]">
        <img
          src="https://remiges.tech/wp-content/uploads/2024/04/Remiges-logo-2048x403.png"
          alt="Remiges" className="h-6 w-auto"
          style={{ filter: 'brightness(0) invert(1)', opacity: 0.85 }}
        />
        <div className="mt-3 flex items-center gap-2">
          <span className="w-2 h-2 rounded-full bg-emerald-400 animate-pulse" />
          <p className="text-[10px] text-slate-500 font-medium">BPP Admin · ION Network</p>
        </div>
      </div>

      {/* Nav */}
      <nav className="flex-1 px-3 py-4 space-y-0.5 overflow-y-auto">
        {NAV_ITEMS.map(({ id, label, icon: Icon, desc }) => {
          const isActive = active === id
          return (
            <button key={id} onClick={() => onNav(id)}
              className={`w-full flex items-center gap-3 px-3 py-2.5 rounded-xl text-sm font-medium transition-all group ${
                isActive ? 'text-white' : 'text-slate-400 hover:text-slate-200 hover:bg-white/[0.04]'
              }`}
              style={isActive ? { background: 'rgba(0,184,230,0.12)', border: '1px solid rgba(0,184,230,0.25)' } : {}}
            >
              <div className={`w-8 h-8 rounded-lg flex items-center justify-center shrink-0 transition-all ${
                isActive ? '' : 'bg-white/[0.04] group-hover:bg-white/[0.07]'
              }`}
                style={isActive ? { background: 'rgba(0,184,230,0.2)' } : {}}>
                <Icon size={15} style={isActive ? { color: '#00b8e6' } : {}} />
              </div>
              <div className="flex-1 min-w-0 text-left">
                <p className="text-sm leading-none">{label}</p>
                <p className={`text-[10px] mt-0.5 truncate ${isActive ? 'text-cyan-400/60' : 'text-slate-600'}`}>{desc}</p>
              </div>
              {isActive && <ChevronRight size={13} style={{ color: '#00b8e6', flexShrink: 0 }} />}
            </button>
          )
        })}
      </nav>

      {/* Footer */}
      <div className="px-4 py-3 border-t border-white/[0.06]">
        <p className="text-[10px] text-slate-600">Beckn Protocol v2.0 · ION Retail</p>
      </div>
    </aside>
  )
}

// ─── PageHeader ───────────────────────────────────────────────────────────────

function PageHeader({ title, subtitle, action, breadcrumb }) {
  return (
    <div className="flex items-start justify-between mb-7">
      <div>
        {breadcrumb && <p className="text-[10px] font-semibold uppercase tracking-widest mb-1.5" style={{ color: '#00b8e6' }}>{breadcrumb}</p>}
        <h1 className="text-2xl font-bold text-white leading-tight">{title}</h1>
        {subtitle && <p className="text-sm text-slate-400 mt-1">{subtitle}</p>}
      </div>
      {action}
    </div>
  )
}

// ─── Overview ────────────────────────────────────────────────────────────────

function StatCard({ label, value, icon: Icon, gradient, sub, onClick, loading }) {
  const Wrap = onClick ? 'button' : 'div'
  return (
    <Wrap onClick={onClick}
      className={`relative w-full text-left rounded-2xl p-5 border border-white/[0.07] overflow-hidden transition-all
        ${onClick ? 'cursor-pointer hover:border-white/[0.14] hover:shadow-lg active:scale-[0.98]' : ''}`}
      style={{ background: '#111827' }}
    >
      {/* Glow blob */}
      <div className="absolute -right-4 -top-4 w-24 h-24 rounded-full opacity-10 blur-xl"
        style={{ background: gradient }} />

      <div className="relative flex items-start justify-between gap-3">
        <div>
          <p className="text-[11px] font-semibold uppercase tracking-widest text-slate-500">{label}</p>
          <p className="text-4xl font-bold text-white mt-1.5 leading-none">
            {loading ? <span className="inline-block w-12 h-8 rounded bg-slate-700/60 animate-pulse" /> : (value ?? '—')}
          </p>
          {sub && <p className="text-xs text-slate-500 mt-2">{sub}</p>}
        </div>
        <div className="w-10 h-10 rounded-xl flex items-center justify-center shrink-0"
          style={{ background: gradient + '33' }}>
          <Icon size={18} style={{ color: 'white', opacity: 0.9 }} />
        </div>
      </div>
      {onClick && (
        <div className="relative mt-4 flex items-center gap-1 text-[11px] font-semibold"
          style={{ color: '#00b8e6' }}>
          View details <ArrowRight size={11} />
        </div>
      )}
    </Wrap>
  )
}

function FunnelCard({ funnel, loading }) {
  const steps = [
    { key: 'select',  label: 'Select',  color: '#00b8e6', emoji: '🔍' },
    { key: 'init',    label: 'Init',    color: '#7c3aed', emoji: '📋' },
    { key: 'confirm', label: 'Confirm', color: '#10b981', emoji: '✅' },
  ]
  const max = Math.max(1, ...steps.map(s => funnel?.[s.key] ?? 0))
  const selectCount = funnel?.select ?? 0
  const confirmCount = funnel?.confirm ?? 0
  const convRate = selectCount > 0 ? Math.round((confirmCount / selectCount) * 100) : 0

  return (
    <Card className="p-5">
      <div className="flex items-center justify-between mb-4">
        <SectionTitle>Conversion Funnel</SectionTitle>
        {!loading && (
          <span className="text-xs font-bold px-2.5 py-1 rounded-full"
            style={{ background: 'rgba(16,185,129,0.15)', color: '#10b981' }}>
            {convRate}% conversion
          </span>
        )}
      </div>
      {loading ? <LoadingSpinner /> : (
        <div className="space-y-4">
          {steps.map(({ key, label, color, emoji }, idx) => {
            const val = funnel?.[key] ?? 0
            const pct = (val / max) * 100
            const drop = idx > 0 ? steps[idx - 1].key : null
            const dropVal = drop ? (funnel?.[drop] ?? 0) : null
            const dropPct = dropVal > 0 ? Math.round(((dropVal - val) / dropVal) * 100) : null
            return (
              <div key={key}>
                <div className="flex justify-between items-center text-xs mb-1.5">
                  <span className="text-slate-300 font-medium flex items-center gap-1.5">
                    {emoji} {label}
                  </span>
                  <div className="flex items-center gap-2">
                    {dropPct !== null && dropPct > 0 && (
                      <span className="text-red-400 text-[10px]">−{dropPct}%</span>
                    )}
                    <span className="font-bold text-white">{val}</span>
                  </div>
                </div>
                <div className="h-2.5 bg-slate-800 rounded-full overflow-hidden">
                  <motion.div
                    className="h-full rounded-full"
                    style={{ background: color }}
                    initial={{ width: 0 }}
                    animate={{ width: `${pct}%` }}
                    transition={{ duration: 0.8, ease: 'easeOut', delay: idx * 0.1 }}
                  />
                </div>
              </div>
            )
          })}
        </div>
      )}
    </Card>
  )
}

function ActivityRow({ o, onView }) {
  return (
    <div className="flex items-center gap-3 py-3 border-b border-white/[0.05] last:border-0 group">
      <div className="w-8 h-8 rounded-lg flex items-center justify-center shrink-0 text-sm"
        style={{ background: 'rgba(0,184,230,0.1)' }}>
        {o.status === 'ACTIVE' ? '✅' : o.status === 'DRAFT' ? '⏳' : o.status === 'CANCELLED' ? '❌' : '📦'}
      </div>
      <div className="flex-1 min-w-0">
        <p className="text-sm font-medium text-slate-200 truncate">{o.bap_id}</p>
        <p className="text-[11px] text-slate-500 font-mono">{o.transaction_id?.slice(0, 12)}…</p>
      </div>
      <div className="flex flex-col items-end gap-1 shrink-0">
        <StatusBadge status={o.status} />
        <p className="text-[10px] text-slate-600">{fmtDateShort(o.created_at?.Time || o.created_at)}</p>
      </div>
      <button onClick={() => onView(o)} className="opacity-0 group-hover:opacity-100 transition-opacity ml-1 text-cyan-400">
        <Eye size={14} />
      </button>
    </div>
  )
}

function OverviewPage({ onNav }) {
  const { data: stats, loading: sLoad, error: sErr, reload } = useApi('/dashboard/stats')
  const { data: orders } = useApi('/orders?limit=6')

  return (
    <div>
      <PageHeader
        breadcrumb="BPP Admin · Dashboard"
        title="Overview"
        subtitle="Live snapshot of your seller activity on the ION network"
        action={
          <button onClick={reload}
            className="flex items-center gap-1.5 text-xs text-slate-400 hover:text-white transition-colors px-3 py-1.5 rounded-lg border border-white/[0.07] hover:border-white/20">
            <RefreshCw size={12} className={sLoad ? 'animate-spin' : ''} />
            Refresh
          </button>
        }
      />

      {sErr && <div className="mb-4"><ErrorBox message={sErr} onRetry={reload} /></div>}

      {/* Stat cards */}
      <div className="grid grid-cols-2 xl:grid-cols-5 gap-4 mb-6">
        <StatCard label="Active Orders"  value={stats?.active_orders}  icon={CheckCircle2} gradient="linear-gradient(135deg,#10b981,#059669)" loading={sLoad} onClick={() => onNav('orders')} />
        <StatCard label="Pending Orders" value={stats?.pending_orders} icon={Clock}        gradient="linear-gradient(135deg,#f59e0b,#d97706)" loading={sLoad} onClick={() => onNav('orders')} />
        <StatCard label="Orders Today"   value={stats?.today_orders}   icon={TrendingUp}   gradient="linear-gradient(135deg,#00b8e6,#1e2fa0)" loading={sLoad} onClick={() => onNav('orders')} />
        <StatCard label="Resources"      value={stats?.resource_count} icon={Package}      gradient="linear-gradient(135deg,#06b6d4,#0891b2)" loading={sLoad} onClick={() => onNav('inventory', { initialTab: 'resources' })} />
        <StatCard label="Active Offers"  value={stats?.offer_count}    icon={Tag}          gradient="linear-gradient(135deg,#e6006e,#7c3aed)" loading={sLoad} onClick={() => onNav('inventory', { initialTab: 'offers' })} />
      </div>

      {/* Bottom section */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-5">
        {/* Recent orders */}
        <div className="lg:col-span-2">
          <Card>
            <div className="flex items-center justify-between px-5 py-4 border-b border-white/[0.06]">
              <div>
                <p className="text-sm font-bold text-white">Recent Orders</p>
                <p className="text-[11px] text-slate-500 mt-0.5">Latest incoming contracts</p>
              </div>
              <button onClick={() => onNav('orders')}
                className="flex items-center gap-1 text-xs font-semibold transition-colors"
                style={{ color: '#00b8e6' }}>
                View all <ChevronRight size={12} />
              </button>
            </div>
            <div className="px-5 py-1">
              {!orders?.items?.length
                ? <p className="text-sm text-slate-500 py-8 text-center">No orders yet</p>
                : orders.items.map(o => (
                    <ActivityRow key={o.id} o={o} onView={() => onNav('orders')} />
                  ))
              }
            </div>
          </Card>
        </div>

        {/* Funnel */}
        <FunnelCard funnel={stats?.funnel} loading={sLoad} />
      </div>
    </div>
  )
}

// ─── Orders ───────────────────────────────────────────────────────────────────

function OrderDetailDrawer({ order, onClose }) {
  const { data, loading } = useApi(order?.id ? `/orders/${order.id}` : null, [order?.id])
  return (
    <AnimatePresence>
      {order && (
        <>
          <motion.div className="fixed inset-0 bg-black/60 backdrop-blur-sm z-40"
            initial={{ opacity: 0 }} animate={{ opacity: 1 }} exit={{ opacity: 0 }}
            onClick={onClose} />
          <motion.div
            className="fixed right-0 top-0 bottom-0 z-50 w-[500px] flex flex-col border-l border-white/[0.07]"
            style={{ background: '#0a0e1a' }}
            initial={{ x: '100%' }} animate={{ x: 0 }} exit={{ x: '100%' }}
            transition={{ type: 'spring', stiffness: 280, damping: 28 }}
          >
            {/* Header */}
            <div className="flex items-center justify-between px-6 py-4 border-b border-white/[0.07] shrink-0">
              <div>
                <p className="font-bold text-white">Order Detail</p>
                <p className="text-xs text-slate-500 font-mono mt-0.5">{order.transaction_id}</p>
              </div>
              <button onClick={onClose}
                className="w-8 h-8 rounded-lg flex items-center justify-center text-slate-400 hover:text-white hover:bg-white/[0.07] transition-all">
                <X size={16} />
              </button>
            </div>

            <div className="flex-1 overflow-y-auto px-6 py-5 space-y-6">
              {loading ? <LoadingSpinner label="Loading contract…" /> : (
                <>
                  {/* Contract info */}
                  <section>
                    <SectionTitle>Contract</SectionTitle>
                    <Card className="divide-y divide-white/[0.05]">
                      {[
                        ['Status',  <StatusBadge key="s" status={data?.contract?.status} />],
                        ['Buyer',   data?.contract?.bap_id],
                        ['Domain',  data?.contract?.domain || '—'],
                        ['Created', fmtDate(data?.contract?.created_at?.Time || data?.contract?.created_at)],
                        ['Updated', fmtDate(data?.contract?.updated_at?.Time || data?.contract?.updated_at)],
                      ].map(([k, v]) => (
                        <div key={k} className="flex items-center justify-between px-4 py-2.5 text-sm">
                          <span className="text-slate-500 shrink-0 w-20">{k}</span>
                          <span className="text-slate-200 text-right">{v}</span>
                        </div>
                      ))}
                    </Card>
                  </section>

                  {/* Commitments */}
                  {data?.commitments?.length > 0 && (
                    <section>
                      <SectionTitle>Commitments ({data.commitments.length})</SectionTitle>
                      <div className="space-y-2">
                        {data.commitments.map(cm => (
                          <Card key={cm.id} className="p-4 space-y-2">
                            <div className="flex items-center justify-between">
                              <p className="text-xs font-mono text-slate-400 truncate">{cm.id}</p>
                              <StatusBadge status={cm.status} />
                            </div>
                            {cm.offer_id && (
                              <div className="flex items-center gap-2 text-xs">
                                <Tag size={11} className="text-slate-500" />
                                <span className="text-slate-300">{cm.offer_id}</span>
                              </div>
                            )}
                          </Card>
                        ))}
                      </div>
                    </section>
                  )}

                  {/* Considerations */}
                  {data?.considerations?.length > 0 && (
                    <section>
                      <SectionTitle>Consideration</SectionTitle>
                      <div className="space-y-2">
                        {data.considerations.map((c, i) => {
                          let attrs = c.consideration_attributes
                          try { if (typeof attrs === 'string') attrs = JSON.parse(attrs) } catch {}
                          const price    = attrs?.price || attrs?.totalAmount
                          const currency = attrs?.currency || 'INR'
                          return (
                            <Card key={i} className="p-4">
                              <div className="flex items-center justify-between mb-2">
                                <span className="text-xs font-semibold text-slate-300">{c.status_code}</span>
                                {price && (
                                  <span className="text-lg font-bold" style={{ color: '#00b8e6' }}>
                                    {currency === 'INR' ? '₹' : ''}{Number(price).toLocaleString()}
                                  </span>
                                )}
                              </div>
                              {attrs && (
                                <pre className="text-[10px] text-slate-500 bg-black/30 rounded-lg p-2 overflow-x-auto mt-1">
                                  {JSON.stringify(attrs, null, 2)}
                                </pre>
                              )}
                            </Card>
                          )
                        })}
                      </div>
                    </section>
                  )}
                </>
              )}
            </div>
          </motion.div>
        </>
      )}
    </AnimatePresence>
  )
}

function OrdersPage() {
  const [page, setPage]     = useState(1)
  const [selected, setSelected] = useState(null)
  const [search, setSearch] = useState('')
  const limit = 20
  const { data, loading, error, reload } = useApi(`/orders?page=${page}&limit=${limit}`, [page])
  const total      = data?.total ?? 0
  const totalPages = Math.max(1, Math.ceil(total / limit))
  const items      = (data?.items ?? []).filter(o =>
    !search || o.bap_id?.includes(search) || o.transaction_id?.includes(search) || o.status?.includes(search.toUpperCase())
  )

  return (
    <div>
      <PageHeader breadcrumb="Orders" title="Order Management"
        subtitle={total ? `${total} total contracts on the network` : 'No orders yet'}
        action={
          <button onClick={reload}
            className="flex items-center gap-1.5 text-xs text-slate-400 hover:text-white transition-colors px-3 py-1.5 rounded-lg border border-white/[0.07]">
            <RefreshCw size={12} className={loading ? 'animate-spin' : ''} /> Refresh
          </button>
        }
      />

      {error && <div className="mb-4"><ErrorBox message={error} onRetry={reload} /></div>}

      {/* Search */}
      <div className="relative mb-4 max-w-sm">
        <Search size={13} className="absolute left-3 top-1/2 -translate-y-1/2 text-slate-500" />
        <input value={search} onChange={e => setSearch(e.target.value)}
          placeholder="Search buyer, txn ID or status…"
          className="w-full bg-white/[0.04] border border-white/[0.08] rounded-xl pl-9 pr-3 py-2 text-sm text-slate-200 placeholder-slate-500 focus:outline-none focus:border-cyan-500/40 transition-colors" />
      </div>

      <Card>
        {loading ? <LoadingSpinner label="Loading orders…" /> : (
          <table className="w-full text-sm">
            <thead>
              <tr className="text-[10px] font-bold text-slate-500 uppercase tracking-widest border-b border-white/[0.06]">
                <th className="text-left px-5 py-3">Txn ID</th>
                <th className="text-left px-5 py-3">Buyer</th>
                <th className="text-left px-5 py-3">Status</th>
                <th className="text-left px-5 py-3">Offer</th>
                <th className="text-left px-5 py-3">Created</th>
                <th className="px-5 py-3 w-16" />
              </tr>
            </thead>
            <tbody className="divide-y divide-white/[0.04]">
              {items.map(o => (
                <tr key={o.id} className="hover:bg-white/[0.025] transition-colors group">
                  <td className="px-5 py-3 font-mono text-xs text-slate-500">{o.transaction_id?.slice(0, 8)}…</td>
                  <td className="px-5 py-3 text-slate-300 font-medium max-w-[180px] truncate">{o.bap_id}</td>
                  <td className="px-5 py-3"><StatusBadge status={o.status} /></td>
                  <td className="px-5 py-3 text-slate-400 text-xs max-w-[140px] truncate">{fmtName(o.offer_descriptor)}</td>
                  <td className="px-5 py-3 text-slate-500 text-xs">{fmtDateShort(o.created_at?.Time || o.created_at)}</td>
                  <td className="px-5 py-3">
                    <button onClick={() => setSelected(o)}
                      className="opacity-0 group-hover:opacity-100 flex items-center gap-1 text-xs font-semibold transition-all"
                      style={{ color: '#00b8e6' }}>
                      <Eye size={13} /> View
                    </button>
                  </td>
                </tr>
              ))}
              {!items.length && (
                <tr><td colSpan={6} className="px-5 py-12 text-center text-slate-500">No orders found</td></tr>
              )}
            </tbody>
          </table>
        )}
        <Pagination page={page} totalPages={totalPages} onPage={setPage} total={total} limit={limit} />
      </Card>

      <OrderDetailDrawer order={selected} onClose={() => setSelected(null)} />
    </div>
  )
}

// ─── Inventory ────────────────────────────────────────────────────────────────

function ResourcesTab() {
  const [page, setPage] = useState(1)
  const limit = 20
  const { data, loading, error, reload } = useApi(`/inventory/resources?page=${page}&limit=${limit}`, [page])
  const total = data?.total ?? 0

  return (
    <div>
      {error && <div className="mb-3"><ErrorBox message={error} onRetry={reload} /></div>}
      <Card className="mt-4">
        {loading ? <LoadingSpinner label="Loading resources…" /> : (
          <table className="w-full text-sm">
            <thead>
              <tr className="text-[10px] font-bold text-slate-500 uppercase tracking-widest border-b border-white/[0.06]">
                <th className="text-left px-5 py-3">Resource ID</th>
                <th className="text-left px-5 py-3">Name</th>
                <th className="text-left px-5 py-3">Description</th>
                <th className="text-left px-5 py-3">Catalog</th>
                <th className="text-left px-5 py-3">Added</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-white/[0.04]">
              {(data?.items ?? []).map(r => (
                <tr key={r.id} className="hover:bg-white/[0.025] transition-colors">
                  <td className="px-5 py-3 font-mono text-[11px] text-slate-500 max-w-[140px] truncate" title={r.id}>{r.id}</td>
                  <td className="px-5 py-3 text-slate-200 font-semibold">{r.descriptor_name || '—'}</td>
                  <td className="px-5 py-3 text-slate-400 text-xs max-w-[220px] truncate">{r.descriptor_short_desc || '—'}</td>
                  <td className="px-5 py-3">
                    <span className="text-[11px] bg-white/[0.06] border border-white/[0.08] px-2 py-0.5 rounded-full text-slate-400">{r.catalog_id}</span>
                  </td>
                  <td className="px-5 py-3 text-slate-500 text-xs">{fmtDateShort(r.created_at?.Time || r.created_at)}</td>
                </tr>
              ))}
              {!data?.items?.length && (
                <tr><td colSpan={5} className="px-5 py-12 text-center text-slate-500">No resources published yet</td></tr>
              )}
            </tbody>
          </table>
        )}
        <Pagination page={page} totalPages={Math.max(1, Math.ceil(total / limit))} onPage={setPage} total={total} limit={limit} />
      </Card>
    </div>
  )
}

function OffersTab() {
  const [page, setPage] = useState(1)
  const limit = 20
  const { data, loading, error, reload } = useApi(`/inventory/offers?page=${page}&limit=${limit}`, [page])
  const total = data?.total ?? 0

  return (
    <div>
      {error && <div className="mb-3"><ErrorBox message={error} onRetry={reload} /></div>}
      <Card className="mt-4">
        {loading ? <LoadingSpinner label="Loading offers…" /> : (
          <table className="w-full text-sm">
            <thead>
              <tr className="text-[10px] font-bold text-slate-500 uppercase tracking-widest border-b border-white/[0.06]">
                <th className="text-left px-5 py-3">Offer ID</th>
                <th className="text-left px-5 py-3">Name</th>
                <th className="text-left px-5 py-3">Resources</th>
                <th className="text-left px-5 py-3">Valid Until</th>
                <th className="text-left px-5 py-3">Added</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-white/[0.04]">
              {(data?.items ?? []).map(o => (
                <tr key={o.id} className="hover:bg-white/[0.025] transition-colors">
                  <td className="px-5 py-3 font-mono text-[11px] text-slate-500 max-w-[140px] truncate" title={o.id}>{o.id}</td>
                  <td className="px-5 py-3 text-slate-200 font-semibold">{o.descriptor_name || '—'}</td>
                  <td className="px-5 py-3">
                    {o.resource_ids?.length
                      ? <span className="text-xs px-2 py-0.5 rounded-full border border-cyan-500/25 font-semibold" style={{ color: '#00b8e6', background: 'rgba(0,184,230,0.1)' }}>
                          {o.resource_ids.length} resource{o.resource_ids.length !== 1 ? 's' : ''}
                        </span>
                      : <span className="text-slate-600">—</span>
                    }
                  </td>
                  <td className="px-5 py-3 text-slate-500 text-xs">
                    {o.validity_end?.Time ? fmtDateShort(o.validity_end.Time) : '—'}
                  </td>
                  <td className="px-5 py-3 text-slate-500 text-xs">{fmtDateShort(o.created_at?.Time || o.created_at)}</td>
                </tr>
              ))}
              {!data?.items?.length && (
                <tr><td colSpan={5} className="px-5 py-12 text-center text-slate-500">No offers published yet</td></tr>
              )}
            </tbody>
          </table>
        )}
        <Pagination page={page} totalPages={Math.max(1, Math.ceil(total / limit))} onPage={setPage} total={total} limit={limit} />
      </Card>
    </div>
  )
}

function InventoryPage({ initialTab = 'resources' }) {
  const [tab, setTab] = useState(initialTab)
  return (
    <div>
      <PageHeader breadcrumb="Inventory" title="Your Inventory"
        subtitle="Resources and offers published to the ION network" />
      <div className="flex gap-1 p-1 rounded-xl w-fit" style={{ background: 'rgba(255,255,255,0.04)', border: '1px solid rgba(255,255,255,0.07)' }}>
        {['resources', 'offers'].map(t => (
          <button key={t} onClick={() => setTab(t)}
            className={`px-5 py-2 rounded-lg text-sm font-semibold transition-all capitalize ${
              tab !== t ? 'text-slate-400 hover:text-slate-200' : 'text-white shadow-lg'
            }`}
            style={tab === t ? { background: BRAND } : {}}>
            {t === 'resources' ? '📦 Resources' : '🏷️ Offers'}
          </button>
        ))}
      </div>
      {tab === 'resources' ? <ResourcesTab /> : <OffersTab />}
    </div>
  )
}

// ─── Publish ──────────────────────────────────────────────────────────────────

const EMPTY_RESOURCE = () => ({ id: '', name: '', short_desc: '', long_desc: '' })
const EMPTY_OFFER    = () => ({ id: '', name: '', short_desc: '', resource_ids: '', price: '', currency: 'INR', validity_start: '', validity_end: '' })
const CATALOG_TYPES  = ['', 'master', 'regular']

function PublishPage({ onNav }) {
  const [step, setStep]       = useState(0) // 0=catalog, 1=resources, 2=offers, 3=review
  const [form, setForm]       = useState({
    catalog_id: '', catalog_name: '', provider_id: '', provider_name: '', catalog_type: '',
    resources: [EMPTY_RESOURCE()], offers: [EMPTY_OFFER()],
  })
  const [error, setError]     = useState(null)
  const [loading, setLoading] = useState(false)

  const STEPS = ['Catalog Info', 'Resources', 'Offers', 'Review & Publish']

  const setField        = (k, v) => setForm(f => ({ ...f, [k]: v }))
  const setResourceField = (i, k, v) => setForm(f => { const r = [...f.resources]; r[i] = { ...r[i], [k]: v }; return { ...f, resources: r } })
  const setOfferField   = (i, k, v) => setForm(f => { const o = [...f.offers]; o[i] = { ...o[i], [k]: v }; return { ...f, offers: o } })
  const toISO           = dt => dt ? new Date(dt).toISOString() : undefined

  const handleSubmit = async () => {
    setLoading(true); setError(null)
    try {
      const catalog = {
        id: form.catalog_id,
        descriptor: { name: form.catalog_name },
        provider: { id: form.provider_id, descriptor: { name: form.provider_name } },
        resources: form.resources.map(r => ({ id: r.id, descriptor: { name: r.name, shortDesc: r.short_desc, longDesc: r.long_desc } })),
        offers: form.offers.map(o => {
          const start = toISO(o.validity_start), end = toISO(o.validity_end)
          return {
            id: o.id, descriptor: { name: o.name, shortDesc: o.short_desc },
            resourceIds: o.resource_ids.split(',').map(s => s.trim()).filter(Boolean),
            ...(o.price ? { considerations: [{ id: `${o.id}-price`, status: { code: 'ACTIVE' },
              considerationAttributes: JSON.stringify({ '@type': 'PriceSpecification', price: o.price, currency: o.currency }) }] } : {}),
            ...(start ? { validity: { startDate: start, endDate: end } } : {}),
          }
        }),
        ...(form.catalog_type ? { publishDirectives: { catalogType: form.catalog_type } } : {}),
      }
      await apiFetch('/catalog/publish', { method: 'POST', body: JSON.stringify({ catalogs: [catalog] }) })
      onNav('inventory')
    } catch (err) { setError(err.message) }
    finally { setLoading(false) }
  }

  const inputCls = 'w-full bg-white/[0.04] border border-white/[0.08] rounded-xl px-3.5 py-2.5 text-sm text-slate-200 placeholder-slate-600 focus:outline-none focus:border-cyan-500/40 focus:bg-white/[0.06] transition-all'
  const labelCls = 'block text-xs font-semibold text-slate-400 mb-1.5'

  return (
    <div className="max-w-2xl">
      <PageHeader breadcrumb="Publish Catalog" title="Publish to Network"
        subtitle="Push your resources and offers to the ION network" />

      {/* Stepper */}
      <div className="flex items-center mb-8">
        {STEPS.map((s, i) => (
          <div key={i} className="flex items-center flex-1 last:flex-none">
            <button onClick={() => i < step && setStep(i)}
              className={`flex items-center gap-2 text-xs font-semibold transition-all ${
                i < step ? 'cursor-pointer' : 'cursor-default'
              }`}>
              <div className={`w-7 h-7 rounded-full flex items-center justify-center text-[11px] font-bold transition-all ${
                i < step ? 'bg-emerald-500 text-white' :
                i === step ? 'text-white' : 'bg-slate-800 text-slate-500'
              }`}
                style={i === step ? { background: BRAND } : {}}>
                {i < step ? <CheckCircle2 size={14} /> : i + 1}
              </div>
              <span className={i === step ? 'text-white' : i < step ? 'text-emerald-400' : 'text-slate-500'}>{s}</span>
            </button>
            {i < STEPS.length - 1 && (
              <div className="flex-1 mx-3 h-px" style={{ background: i < step ? '#10b981' : 'rgba(255,255,255,0.08)' }} />
            )}
          </div>
        ))}
      </div>

      {error && <div className="mb-4"><ErrorBox message={error} /></div>}

      <AnimatePresence mode="wait">
        <motion.div key={step}
          initial={{ opacity: 0, x: 20 }} animate={{ opacity: 1, x: 0 }} exit={{ opacity: 0, x: -20 }}
          transition={{ duration: 0.18 }}>

          {/* Step 0: Catalog Info */}
          {step === 0 && (
            <Card className="p-6 space-y-4">
              <p className="text-sm font-bold text-white mb-4">Catalog Details</p>
              <div className="grid grid-cols-2 gap-4">
                <div><label className={labelCls}>Catalog ID *</label>
                  <input className={inputCls} value={form.catalog_id} onChange={e => setField('catalog_id', e.target.value)} placeholder="cat-winroom-2026" required />
                </div>
                <div><label className={labelCls}>Catalog Name *</label>
                  <input className={inputCls} value={form.catalog_name} onChange={e => setField('catalog_name', e.target.value)} placeholder="Winroom Hotel Catalog" required />
                </div>
                <div><label className={labelCls}>Provider ID *</label>
                  <input className={inputCls} value={form.provider_id} onChange={e => setField('provider_id', e.target.value)} placeholder="prov-winroom-001" required />
                </div>
                <div><label className={labelCls}>Provider Name *</label>
                  <input className={inputCls} value={form.provider_name} onChange={e => setField('provider_name', e.target.value)} placeholder="Winroom Hotels Pvt Ltd" required />
                </div>
                <div className="col-span-2">
                  <label className={labelCls}>Catalog Type <span className="text-slate-600 font-normal">(optional)</span></label>
                  <select className={inputCls} value={form.catalog_type} onChange={e => setField('catalog_type', e.target.value)}>
                    {CATALOG_TYPES.map(t => <option key={t} value={t}>{t || '— select type —'}</option>)}
                  </select>
                </div>
              </div>
            </Card>
          )}

          {/* Step 1: Resources */}
          {step === 1 && (
            <div className="space-y-3">
              {form.resources.map((r, i) => (
                <Card key={i} className="p-5">
                  <div className="flex items-center justify-between mb-4">
                    <div className="flex items-center gap-2">
                      <div className="w-6 h-6 rounded-md flex items-center justify-center text-[11px] font-bold text-white"
                        style={{ background: BRAND }}>
                        {i + 1}
                      </div>
                      <p className="text-sm font-bold text-white">Resource {i + 1}</p>
                    </div>
                    {form.resources.length > 1 && (
                      <button onClick={() => setForm(f => ({ ...f, resources: f.resources.filter((_, j) => j !== i) }))}
                        className="text-xs text-red-400 hover:text-red-300 flex items-center gap-1 transition-colors">
                        <Trash2 size={12} /> Remove
                      </button>
                    )}
                  </div>
                  <div className="grid grid-cols-2 gap-3">
                    <div><label className={labelCls}>Resource ID *</label>
                      <input className={inputCls} value={r.id} onChange={e => setResourceField(i, 'id', e.target.value)} placeholder="room-deluxe-001" required />
                    </div>
                    <div><label className={labelCls}>Name *</label>
                      <input className={inputCls} value={r.name} onChange={e => setResourceField(i, 'name', e.target.value)} placeholder="Deluxe Room" required />
                    </div>
                    <div><label className={labelCls}>Short Description</label>
                      <input className={inputCls} value={r.short_desc} onChange={e => setResourceField(i, 'short_desc', e.target.value)} placeholder="Cozy deluxe room with city view" />
                    </div>
                    <div><label className={labelCls}>Long Description</label>
                      <input className={inputCls} value={r.long_desc} onChange={e => setResourceField(i, 'long_desc', e.target.value)} placeholder="Full description…" />
                    </div>
                  </div>
                </Card>
              ))}
              <button onClick={() => setForm(f => ({ ...f, resources: [...f.resources, EMPTY_RESOURCE()] }))}
                className="w-full py-3 rounded-xl border border-dashed border-white/[0.12] text-sm text-slate-400 hover:text-slate-200 hover:border-white/20 flex items-center justify-center gap-2 transition-all">
                <Plus size={14} /> Add Another Resource
              </button>
            </div>
          )}

          {/* Step 2: Offers */}
          {step === 2 && (
            <div className="space-y-3">
              {form.offers.map((o, i) => (
                <Card key={i} className="p-5">
                  <div className="flex items-center justify-between mb-4">
                    <div className="flex items-center gap-2">
                      <div className="w-6 h-6 rounded-md flex items-center justify-center text-[11px] font-bold text-white"
                        style={{ background: BRAND2 }}>
                        {i + 1}
                      </div>
                      <p className="text-sm font-bold text-white">Offer {i + 1}</p>
                    </div>
                    {form.offers.length > 1 && (
                      <button onClick={() => setForm(f => ({ ...f, offers: f.offers.filter((_, j) => j !== i) }))}
                        className="text-xs text-red-400 hover:text-red-300 flex items-center gap-1 transition-colors">
                        <Trash2 size={12} /> Remove
                      </button>
                    )}
                  </div>
                  <div className="grid grid-cols-2 gap-3">
                    <div><label className={labelCls}>Offer ID *</label>
                      <input className={inputCls} value={o.id} onChange={e => setOfferField(i, 'id', e.target.value)} placeholder="offer-deluxe-bb" required />
                    </div>
                    <div><label className={labelCls}>Offer Name *</label>
                      <input className={inputCls} value={o.name} onChange={e => setOfferField(i, 'name', e.target.value)} placeholder="Deluxe – Bed & Breakfast" required />
                    </div>
                    <div className="col-span-2"><label className={labelCls}>Short Description</label>
                      <input className={inputCls} value={o.short_desc} onChange={e => setOfferField(i, 'short_desc', e.target.value)} placeholder="Brief offer description" />
                    </div>
                    <div className="col-span-2"><label className={labelCls}>Resource IDs <span className="text-slate-600 font-normal">(comma-separated)</span></label>
                      <input className={inputCls} value={o.resource_ids} onChange={e => setOfferField(i, 'resource_ids', e.target.value)} placeholder="room-deluxe-001, room-deluxe-002" />
                    </div>
                    <div><label className={labelCls}>Price</label>
                      <div className="relative">
                        <span className="absolute left-3.5 top-1/2 -translate-y-1/2 text-slate-500 text-sm">₹</span>
                        <input className={inputCls + ' pl-7'} value={o.price} onChange={e => setOfferField(i, 'price', e.target.value)} placeholder="4500" />
                      </div>
                    </div>
                    <div><label className={labelCls}>Currency</label>
                      <input className={inputCls} value={o.currency} onChange={e => setOfferField(i, 'currency', e.target.value)} placeholder="INR" />
                    </div>
                    <div><label className={labelCls}>Validity Start <span className="text-slate-600 font-normal">(optional)</span></label>
                      <input type="datetime-local" className={inputCls} value={o.validity_start} onChange={e => setOfferField(i, 'validity_start', e.target.value)} />
                    </div>
                    <div><label className={labelCls}>Validity End <span className="text-slate-600 font-normal">(optional)</span></label>
                      <input type="datetime-local" className={inputCls} value={o.validity_end} onChange={e => setOfferField(i, 'validity_end', e.target.value)} />
                    </div>
                  </div>
                </Card>
              ))}
              <button onClick={() => setForm(f => ({ ...f, offers: [...f.offers, EMPTY_OFFER()] }))}
                className="w-full py-3 rounded-xl border border-dashed border-white/[0.12] text-sm text-slate-400 hover:text-slate-200 hover:border-white/20 flex items-center justify-center gap-2 transition-all">
                <Plus size={14} /> Add Another Offer
              </button>
            </div>
          )}

          {/* Step 3: Review */}
          {step === 3 && (
            <div className="space-y-4">
              <Card className="p-5">
                <SectionTitle>Catalog</SectionTitle>
                <div className="space-y-2 text-sm">
                  {[['ID', form.catalog_id], ['Name', form.catalog_name], ['Provider', `${form.provider_name} (${form.provider_id})`], ['Type', form.catalog_type || 'Not set']].map(([k, v]) => (
                    <div key={k} className="flex justify-between">
                      <span className="text-slate-500">{k}</span>
                      <span className="text-slate-200 font-medium">{v || '—'}</span>
                    </div>
                  ))}
                </div>
              </Card>
              <Card className="p-5">
                <SectionTitle>Resources ({form.resources.length})</SectionTitle>
                <div className="space-y-2">
                  {form.resources.map((r, i) => (
                    <div key={i} className="flex items-center gap-3 text-sm">
                      <Package size={13} className="text-slate-500 shrink-0" />
                      <span className="text-slate-200 font-medium">{r.name || '—'}</span>
                      <span className="text-slate-500 text-xs ml-auto">{r.id}</span>
                    </div>
                  ))}
                </div>
              </Card>
              <Card className="p-5">
                <SectionTitle>Offers ({form.offers.length})</SectionTitle>
                <div className="space-y-2">
                  {form.offers.map((o, i) => (
                    <div key={i} className="flex items-center gap-3 text-sm">
                      <Tag size={13} className="text-slate-500 shrink-0" />
                      <span className="text-slate-200 font-medium">{o.name || '—'}</span>
                      {o.price && <span className="text-xs font-bold ml-auto" style={{ color: '#00b8e6' }}>₹{o.price} {o.currency}</span>}
                    </div>
                  ))}
                </div>
              </Card>
            </div>
          )}
        </motion.div>
      </AnimatePresence>

      {/* Navigation buttons */}
      <div className="flex items-center justify-between mt-6 pt-5 border-t border-white/[0.06]">
        <button onClick={() => step > 0 && setStep(s => s - 1)}
          className={`flex items-center gap-2 px-4 py-2 rounded-xl text-sm font-semibold transition-all
            ${step === 0 ? 'opacity-0 pointer-events-none' : 'text-slate-400 hover:text-white border border-white/[0.08] hover:border-white/20'}`}>
          <ChevronLeft size={15} /> Back
        </button>
        {step < 3 ? (
          <button onClick={() => setStep(s => s + 1)}
            className="flex items-center gap-2 px-5 py-2.5 rounded-xl text-sm font-bold text-white transition-all active:scale-95 shadow-lg"
            style={{ background: BRAND }}>
            Continue <ArrowRight size={15} />
          </button>
        ) : (
          <button onClick={handleSubmit} disabled={loading}
            className="flex items-center gap-2 px-6 py-2.5 rounded-xl text-sm font-bold text-white disabled:opacity-60 transition-all active:scale-95 shadow-lg"
            style={{ background: BRAND }}>
            {loading ? <><RefreshCw size={14} className="animate-spin" /> Publishing…</> : <><Upload size={14} /> Publish Catalog</>}
          </button>
        )}
      </div>
    </div>
  )
}

// ─── Messages ─────────────────────────────────────────────────────────────────

const ACTION_META = {
  select:     { color: '#00b8e6', bg: 'rgba(0,184,230,0.12)',  label: 'SELECT' },
  on_select:  { color: '#67e8f9', bg: 'rgba(103,232,249,0.1)', label: 'ON_SELECT' },
  init:       { color: '#a78bfa', bg: 'rgba(167,139,250,0.12)',label: 'INIT' },
  on_init:    { color: '#c4b5fd', bg: 'rgba(196,181,253,0.1)', label: 'ON_INIT' },
  confirm:    { color: '#34d399', bg: 'rgba(52,211,153,0.12)', label: 'CONFIRM' },
  on_confirm: { color: '#6ee7b7', bg: 'rgba(110,231,183,0.1)', label: 'ON_CONFIRM' },
}

function MessagesPage() {
  const [limit, setLimit]   = useState(50)
  const [filter, setFilter] = useState('')
  const [dirFilter, setDirFilter] = useState('ALL')
  const { data, loading, error, reload } = useApi(`/messages?limit=${limit}`, [limit])

  const items = (data?.items ?? []).filter(m => {
    const matchText = !filter || String(m.action).includes(filter.toLowerCase()) || String(m.transaction_id).includes(filter)
    const matchDir  = dirFilter === 'ALL' || m.direction === dirFilter
    return matchText && matchDir
  })

  const ackCount  = items.filter(m => m.ack_status?.ack_status === 'ACK'  || m.ack_status === 'ACK').length
  const nackCount = items.filter(m => m.ack_status?.ack_status === 'NACK' || m.ack_status === 'NACK').length

  return (
    <div>
      <PageHeader breadcrumb="Protocol Audit" title="Message Log"
        subtitle={`${items.length} recent protocol messages`}
        action={
          <button onClick={reload}
            className="flex items-center gap-1.5 text-xs text-slate-400 hover:text-white transition-colors px-3 py-1.5 rounded-lg border border-white/[0.07]">
            <RefreshCw size={12} className={loading ? 'animate-spin' : ''} /> Refresh
          </button>
        }
      />

      {/* Quick stats */}
      {!loading && items.length > 0 && (
        <div className="grid grid-cols-3 gap-3 mb-5">
          <Card className="p-4 flex items-center gap-3">
            <div className="w-8 h-8 rounded-lg flex items-center justify-center text-sm" style={{ background: 'rgba(0,184,230,0.15)' }}>📨</div>
            <div><p className="text-[11px] text-slate-500">Total</p><p className="text-lg font-bold text-white">{items.length}</p></div>
          </Card>
          <Card className="p-4 flex items-center gap-3">
            <div className="w-8 h-8 rounded-lg flex items-center justify-center text-sm" style={{ background: 'rgba(16,185,129,0.15)' }}>✅</div>
            <div><p className="text-[11px] text-slate-500">ACK</p><p className="text-lg font-bold text-emerald-400">{ackCount}</p></div>
          </Card>
          <Card className="p-4 flex items-center gap-3">
            <div className="w-8 h-8 rounded-lg flex items-center justify-center text-sm" style={{ background: 'rgba(239,68,68,0.15)' }}>❌</div>
            <div><p className="text-[11px] text-slate-500">NACK</p><p className="text-lg font-bold text-red-400">{nackCount}</p></div>
          </Card>
        </div>
      )}

      {error && <div className="mb-4"><ErrorBox message={error} onRetry={reload} /></div>}

      {/* Filters */}
      <div className="flex gap-3 mb-4 flex-wrap">
        <div className="relative flex-1 min-w-[200px] max-w-sm">
          <Search size={13} className="absolute left-3 top-1/2 -translate-y-1/2 text-slate-500" />
          <input value={filter} onChange={e => setFilter(e.target.value)}
            placeholder="Filter by action or txn ID…"
            className="w-full bg-white/[0.04] border border-white/[0.08] rounded-xl pl-9 pr-3 py-2 text-sm text-slate-200 placeholder-slate-500 focus:outline-none focus:border-cyan-500/40 transition-colors" />
        </div>
        <div className="flex gap-1 p-0.5 rounded-xl" style={{ background: 'rgba(255,255,255,0.04)', border: '1px solid rgba(255,255,255,0.07)' }}>
          {['ALL', 'INBOUND', 'OUTBOUND'].map(d => (
            <button key={d} onClick={() => setDirFilter(d)}
              className={`px-3 py-1.5 rounded-lg text-xs font-semibold transition-all ${dirFilter === d ? 'text-white' : 'text-slate-400 hover:text-slate-200'}`}
              style={dirFilter === d ? { background: BRAND } : {}}>
              {d}
            </button>
          ))}
        </div>
        <select value={limit} onChange={e => setLimit(Number(e.target.value))}
          className="bg-white/[0.04] border border-white/[0.08] rounded-xl px-3 py-2 text-sm text-slate-300 focus:outline-none focus:border-cyan-500/40 cursor-pointer">
          {[25, 50, 100, 200].map(n => <option key={n} value={n}>{n} rows</option>)}
        </select>
      </div>

      <Card>
        {loading ? <LoadingSpinner label="Loading messages…" /> : (
          <table className="w-full text-sm">
            <thead>
              <tr className="text-[10px] font-bold text-slate-500 uppercase tracking-widest border-b border-white/[0.06]">
                <th className="text-left px-5 py-3">Action</th>
                <th className="text-left px-5 py-3">Direction</th>
                <th className="text-left px-5 py-3">Status</th>
                <th className="text-left px-5 py-3">Txn ID</th>
                <th className="text-left px-5 py-3">URL</th>
                <th className="text-left px-5 py-3">Time</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-white/[0.04]">
              {items.map((m, idx) => {
                const meta = ACTION_META[m.action] || { color: '#94a3b8', bg: 'rgba(148,163,184,0.08)', label: m.action?.toUpperCase() }
                const ack  = m.ack_status?.ack_status || m.ack_status
                return (
                  <tr key={idx} className="hover:bg-white/[0.025] transition-colors">
                    <td className="px-5 py-2.5">
                      <span className="inline-flex items-center px-2.5 py-1 rounded-lg text-[11px] font-bold font-mono"
                        style={{ background: meta.bg, color: meta.color }}>
                        {meta.label}
                      </span>
                    </td>
                    <td className="px-5 py-2.5">
                      <span className={`text-xs font-medium ${m.direction === 'INBOUND' ? 'text-cyan-400' : 'text-violet-400'}`}>
                        {m.direction === 'INBOUND' ? '↓ IN' : '↑ OUT'}
                      </span>
                    </td>
                    <td className="px-5 py-2.5"><StatusBadge status={ack} /></td>
                    <td className="px-5 py-2.5 font-mono text-[11px] text-slate-500">
                      {m.transaction_id ? String(m.transaction_id).slice(0, 8) + '…' : '—'}
                    </td>
                    <td className="px-5 py-2.5 text-xs text-slate-500 max-w-[180px] truncate" title={m.url}>{m.url || '—'}</td>
                    <td className="px-5 py-2.5 text-xs text-slate-500">{fmtDate(m.created_at?.Time || m.created_at)}</td>
                  </tr>
                )
              })}
              {!items.length && (
                <tr><td colSpan={6} className="px-5 py-12 text-center text-slate-500">No messages found</td></tr>
              )}
            </tbody>
          </table>
        )}
      </Card>
    </div>
  )
}

// ─── App shell ────────────────────────────────────────────────────────────────

const PAGES = {
  overview:  OverviewPage,
  orders:    OrdersPage,
  inventory: InventoryPage,
  publish:   PublishPage,
  messages:  MessagesPage,
}

export default function App() {
  const [page, setPage]         = useState('overview')
  const [navExtra, setNavExtra] = useState({})
  const Page = PAGES[page] || OverviewPage

  const onNav = (target, extra = {}) => { setPage(target); setNavExtra(extra) }

  return (
    <div className="flex min-h-screen" style={{ background: '#080d1e' }}>
      <Sidebar active={page} onNav={onNav} />

      <main className="flex-1 ml-64 p-7 lg:p-9 overflow-y-auto min-h-screen">
        <AnimatePresence mode="wait">
          <motion.div key={page}
            initial={{ opacity: 0, y: 12 }}
            animate={{ opacity: 1, y: 0 }}
            exit={{ opacity: 0, y: -8 }}
            transition={{ duration: 0.18 }}>
            <Page onNav={onNav} {...navExtra} />
          </motion.div>
        </AnimatePresence>
      </main>
    </div>
  )
}
