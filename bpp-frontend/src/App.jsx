import { useState, useEffect, useCallback } from 'react'
import { motion, AnimatePresence } from 'framer-motion'
import {
  LayoutDashboard,
  ShoppingBag,
  Package,
  Radio,
  MessageSquare,
  ChevronRight,
  TrendingUp,
  Clock,
  CheckCircle2,
  AlertCircle,
  RefreshCw,
  X,
  ChevronLeft,
  ChevronRight as ChevronRightIcon,
  Upload,
  Plus,
  Trash2,
  Eye,
  ArrowUpRight,
  Filter,
  Search,
} from 'lucide-react'

const API_BASE = '/api/v1'

// ─── helpers ────────────────────────────────────────────────────────────────

async function apiFetch(path, opts = {}) {
  const res = await fetch(`${API_BASE}${path}`, {
    headers: { 'Content-Type': 'application/json' },
    ...opts,
  })
  if (!res.ok) {
    const body = await res.text()
    throw new Error(body || res.statusText)
  }
  return res.json()
}

function useApi(path, deps = []) {
  const [data, setData] = useState(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)

  const load = useCallback(async () => {
    if (!path) return
    setLoading(true)
    setError(null)
    try {
      const d = await apiFetch(path)
      setData(d)
    } catch (e) {
      setError(e.message)
    } finally {
      setLoading(false)
    }
  }, [path, ...deps]) // eslint-disable-line react-hooks/exhaustive-deps

  useEffect(() => { load() }, [load])
  return { data, loading, error, reload: load }
}

function fmtDate(ts) {
  if (!ts) return '—'
  return new Date(ts).toLocaleString()
}

function fmtName(obj) {
  if (!obj) return '—'
  if (typeof obj === 'string') {
    try { obj = JSON.parse(obj) } catch { return obj }
  }
  return obj?.name || obj?.code || obj?.descriptor?.name || '—'
}

function StatusBadge({ status }) {
  const map = {
    ACTIVE:    'bg-emerald-500/20 text-emerald-400 border-emerald-500/30',
    DRAFT:     'bg-amber-500/20  text-amber-400  border-amber-500/30',
    CANCELLED: 'bg-red-500/20    text-red-400    border-red-500/30',
    COMPLETE:  'bg-sky-500/20    text-sky-400    border-sky-500/30',
    ACK:       'bg-emerald-500/20 text-emerald-400 border-emerald-500/30',
    NACK:      'bg-red-500/20    text-red-400    border-red-500/30',
  }
  const cls = map[status] || 'bg-slate-500/20 text-slate-400 border-slate-500/30'
  return (
    <span className={`inline-flex items-center px-2 py-0.5 rounded-full text-xs font-semibold border ${cls}`}>
      {status || '—'}
    </span>
  )
}

// ─── layout pieces ──────────────────────────────────────────────────────────

const NAV_ITEMS = [
  { id: 'overview',   label: 'Overview',       icon: LayoutDashboard },
  { id: 'orders',     label: 'Orders',         icon: ShoppingBag },
  { id: 'inventory',  label: 'Inventory',      icon: Package },
  { id: 'publish',    label: 'Publish Catalog',icon: Radio },
  { id: 'messages',   label: 'Message Log',    icon: MessageSquare },
]

function Sidebar({ active, onNav }) {
  return (
    <aside className="fixed inset-y-0 left-0 z-40 flex flex-col w-60 bg-[#0a0d14] border-r border-slate-800">
      <div className="flex items-center gap-3 px-5 py-5 border-b border-slate-800">
        <div className="w-8 h-8 rounded-lg bg-gradient-to-br from-indigo-500 to-violet-600 flex items-center justify-center">
          <span className="text-white font-bold text-sm">B</span>
        </div>
        <div>
          <p className="text-sm font-bold text-white leading-none">BPP Admin</p>
          <p className="text-[10px] text-slate-500 mt-0.5">Seller Dashboard</p>
        </div>
      </div>

      <nav className="flex-1 px-3 py-4 space-y-1">
        {NAV_ITEMS.map(({ id, label, icon: Icon }) => {
          const isActive = active === id
          return (
            <button
              key={id}
              onClick={() => onNav(id)}
              className={`w-full flex items-center gap-3 px-3 py-2.5 rounded-lg text-sm font-medium transition-all
                ${isActive
                  ? 'bg-indigo-600/20 text-indigo-400 border border-indigo-600/30'
                  : 'text-slate-400 hover:text-slate-200 hover:bg-slate-800/60'
                }`}
            >
              <Icon size={16} className={isActive ? 'text-indigo-400' : ''} />
              {label}
              {isActive && <ChevronRight size={14} className="ml-auto text-indigo-500" />}
            </button>
          )
        })}
      </nav>

      <div className="px-5 py-4 border-t border-slate-800">
        <p className="text-[10px] text-slate-600">Beckn Protocol v2.0</p>
      </div>
    </aside>
  )
}

function PageHeader({ title, subtitle, action }) {
  return (
    <div className="flex items-start justify-between mb-6">
      <div>
        <h1 className="text-xl font-bold text-white">{title}</h1>
        {subtitle && <p className="text-sm text-slate-400 mt-0.5">{subtitle}</p>}
      </div>
      {action}
    </div>
  )
}

function Card({ children, className = '' }) {
  return (
    <div className={`bg-[#141821] border border-slate-800 rounded-xl ${className}`}>
      {children}
    </div>
  )
}

function LoadingSpinner() {
  return (
    <div className="flex items-center justify-center py-20">
      <RefreshCw size={20} className="text-indigo-500 animate-spin" />
    </div>
  )
}

function ErrorBox({ message, onRetry }) {
  return (
    <div className="flex items-center gap-3 bg-red-500/10 border border-red-500/20 rounded-lg p-4 text-sm text-red-400">
      <AlertCircle size={16} className="shrink-0" />
      <span className="flex-1">{message}</span>
      {onRetry && (
        <button onClick={onRetry} className="text-xs underline underline-offset-2">
          Retry
        </button>
      )}
    </div>
  )
}

// ─── Overview page ──────────────────────────────────────────────────────────

function StatCard({ label, value, icon: Icon, color = 'indigo', sub, onClick }) {
  const colorMap = {
    indigo:  { bg: 'bg-indigo-500/10',  icon: 'text-indigo-400',  val: 'text-indigo-300',  ring: 'hover:border-indigo-500/40' },
    emerald: { bg: 'bg-emerald-500/10', icon: 'text-emerald-400', val: 'text-emerald-300', ring: 'hover:border-emerald-500/40' },
    amber:   { bg: 'bg-amber-500/10',   icon: 'text-amber-400',   val: 'text-amber-300',   ring: 'hover:border-amber-500/40' },
    sky:     { bg: 'bg-sky-500/10',     icon: 'text-sky-400',     val: 'text-sky-300',     ring: 'hover:border-sky-500/40' },
    violet:  { bg: 'bg-violet-500/10',  icon: 'text-violet-400',  val: 'text-violet-300',  ring: 'hover:border-violet-500/40' },
  }
  const c = colorMap[color] || colorMap.indigo
  const Wrap = onClick ? 'button' : 'div'
  return (
    <Wrap
      onClick={onClick}
      className={`w-full text-left bg-[#141821] border border-slate-800 rounded-xl p-5 transition-all
        ${onClick ? `cursor-pointer ${c.ring} hover:bg-slate-800/40 active:scale-[0.98]` : ''}`}
    >
      <div className="flex items-start justify-between">
        <div>
          <p className="text-xs font-medium text-slate-500 uppercase tracking-wider">{label}</p>
          <p className={`text-3xl font-bold mt-1 ${c.val}`}>
            {value ?? <span className="text-slate-600 animate-pulse">—</span>}
          </p>
          {sub && <p className="text-xs text-slate-500 mt-1">{sub}</p>}
        </div>
        <div className={`w-9 h-9 rounded-lg ${c.bg} flex items-center justify-center shrink-0`}>
          <Icon size={18} className={c.icon} />
        </div>
      </div>
      {onClick && (
        <p className="text-[10px] text-slate-600 mt-3 flex items-center gap-1">
          View details <ChevronRight size={10} />
        </p>
      )}
    </Wrap>
  )
}

function FunnelBar({ funnel }) {
  const steps = [
    { key: 'select',  label: 'Select',  color: 'bg-indigo-500' },
    { key: 'init',    label: 'Init',    color: 'bg-violet-500' },
    { key: 'confirm', label: 'Confirm', color: 'bg-emerald-500' },
  ]
  const max = Math.max(1, ...steps.map(s => funnel?.[s.key] ?? 0))

  return (
    <Card className="p-5">
      <p className="text-xs font-semibold text-slate-400 uppercase tracking-wider mb-4">Order Funnel</p>
      <div className="space-y-3">
        {steps.map(({ key, label, color }) => {
          const val = funnel?.[key] ?? 0
          const pct = (val / max) * 100
          return (
            <div key={key}>
              <div className="flex justify-between text-xs mb-1">
                <span className="text-slate-400">{label}</span>
                <span className="text-slate-200 font-semibold">{val}</span>
              </div>
              <div className="h-2 bg-slate-800 rounded-full overflow-hidden">
                <motion.div
                  className={`h-full ${color} rounded-full`}
                  initial={{ width: 0 }}
                  animate={{ width: `${pct}%` }}
                  transition={{ duration: 0.6, ease: 'easeOut' }}
                />
              </div>
            </div>
          )
        })}
      </div>
    </Card>
  )
}

function RecentOrdersTable({ orders, onView }) {
  if (!orders?.length) {
    return <p className="text-sm text-slate-500 py-4 text-center">No orders yet</p>
  }
  return (
    <table className="w-full text-sm">
      <thead>
        <tr className="text-xs text-slate-500 uppercase tracking-wider border-b border-slate-800">
          <th className="text-left py-2 pr-3 font-medium">Txn ID</th>
          <th className="text-left py-2 pr-3 font-medium">Buyer</th>
          <th className="text-left py-2 pr-3 font-medium">Status</th>
          <th className="text-left py-2 font-medium">Date</th>
          <th />
        </tr>
      </thead>
      <tbody className="divide-y divide-slate-800/60">
        {orders.map(o => (
          <tr key={o.id} className="hover:bg-slate-800/30 transition-colors">
            <td className="py-2.5 pr-3 font-mono text-xs text-slate-400 truncate max-w-[120px]">
              {o.transaction_id?.slice(0, 8)}…
            </td>
            <td className="py-2.5 pr-3 text-slate-300 truncate max-w-[140px]">{o.bap_id}</td>
            <td className="py-2.5 pr-3"><StatusBadge status={o.status} /></td>
            <td className="py-2.5 text-slate-500 text-xs">{fmtDate(o.created_at?.Time || o.created_at)}</td>
            <td className="py-2.5">
              <button onClick={() => onView(o)} className="text-indigo-400 hover:text-indigo-300 transition-colors">
                <Eye size={14} />
              </button>
            </td>
          </tr>
        ))}
      </tbody>
    </table>
  )
}

function OverviewPage({ onNav }) {
  const { data: stats, loading, error, reload } = useApi('/dashboard/stats')
  const { data: ordersData } = useApi('/orders?limit=5')

  return (
    <div>
      <PageHeader
        title="Dashboard Overview"
        subtitle="Real-time snapshot of your BPP activity"
        action={
          <button onClick={reload} className="flex items-center gap-1.5 text-xs text-slate-400 hover:text-slate-200 transition-colors">
            <RefreshCw size={13} className={loading ? 'animate-spin' : ''} />
            Refresh
          </button>
        }
      />

      {error && <ErrorBox message={error} onRetry={reload} />}

      <div className="grid grid-cols-2 xl:grid-cols-5 gap-4 mb-6">
        <StatCard label="Active Orders"  value={stats?.active_orders}  icon={CheckCircle2} color="emerald" onClick={() => onNav('orders')} />
        <StatCard label="Pending Orders" value={stats?.pending_orders} icon={Clock}        color="amber"   onClick={() => onNav('orders')} />
        <StatCard label="Orders Today"   value={stats?.today_orders}   icon={TrendingUp}   color="indigo"  onClick={() => onNav('orders')} />
        <StatCard label="Resources"      value={stats?.resource_count} icon={Package}      color="sky"     onClick={() => onNav('inventory', { initialTab: 'resources' })} />
        <StatCard label="Active Offers"  value={stats?.offer_count}    icon={ArrowUpRight} color="violet"  onClick={() => onNav('inventory', { initialTab: 'offers' })} />
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-4">
        <div className="lg:col-span-2">
          <Card>
            <div className="flex items-center justify-between px-5 py-3 border-b border-slate-800">
              <p className="text-sm font-semibold text-white">Recent Orders</p>
              <button
                onClick={() => onNav('orders')}
                className="text-xs text-indigo-400 hover:text-indigo-300 flex items-center gap-1"
              >
                View all <ChevronRightIcon size={12} />
              </button>
            </div>
            <div className="px-5 py-2">
              <RecentOrdersTable orders={ordersData?.items} onView={() => onNav('orders')} />
            </div>
          </Card>
        </div>
        <FunnelBar funnel={stats?.funnel} />
      </div>
    </div>
  )
}

// ─── Orders page ─────────────────────────────────────────────────────────────

function OrderDetailDrawer({ order, onClose }) {
  const contractId = order?.id
  const { data, loading } = useApi(contractId ? `/orders/${contractId}` : null, [contractId])

  return (
    <AnimatePresence>
      {order && (
        <>
          <motion.div
            className="fixed inset-0 bg-black/50 z-40"
            initial={{ opacity: 0 }} animate={{ opacity: 1 }} exit={{ opacity: 0 }}
            onClick={onClose}
          />
          <motion.div
            className="fixed right-0 top-0 bottom-0 z-50 w-[480px] bg-[#0f1117] border-l border-slate-800 flex flex-col"
            initial={{ x: '100%' }} animate={{ x: 0 }} exit={{ x: '100%' }}
            transition={{ type: 'spring', stiffness: 300, damping: 30 }}
          >
            <div className="flex items-center justify-between px-5 py-4 border-b border-slate-800">
              <div>
                <p className="font-semibold text-white text-sm">Order Detail</p>
                <p className="text-xs text-slate-500 font-mono mt-0.5">
                  {order.transaction_id?.slice(0, 8)}…
                </p>
              </div>
              <button onClick={onClose} className="text-slate-400 hover:text-white transition-colors">
                <X size={18} />
              </button>
            </div>

            <div className="flex-1 overflow-y-auto px-5 py-4 space-y-5">
              {loading ? <LoadingSpinner /> : (
                <>
                  <section>
                    <p className="text-xs font-semibold text-slate-500 uppercase tracking-wider mb-2">Contract</p>
                    <div className="space-y-2">
                      {[
                        ['Status',     <StatusBadge key="s" status={data?.contract?.status} />],
                        ['Buyer',      data?.contract?.bap_id],
                        ['Domain',     data?.contract?.domain || '—'],
                        ['Created',    fmtDate(data?.contract?.created_at?.Time || data?.contract?.created_at)],
                      ].map(([k, v]) => (
                        <div key={k} className="flex items-center justify-between text-sm">
                          <span className="text-slate-500">{k}</span>
                          <span className="text-slate-200">{v}</span>
                        </div>
                      ))}
                    </div>
                  </section>

                  {data?.commitments?.length > 0 && (
                    <section>
                      <p className="text-xs font-semibold text-slate-500 uppercase tracking-wider mb-2">Commitments</p>
                      <div className="space-y-2">
                        {data.commitments.map(cm => (
                          <div key={cm.id} className="bg-slate-800/50 rounded-lg p-3 text-xs space-y-1">
                            <div className="flex justify-between">
                              <span className="text-slate-400 font-mono">{cm.id}</span>
                              <StatusBadge status={cm.status} />
                            </div>
                            {cm.offer_id && (
                              <p className="text-slate-500">Offer: <span className="text-slate-300">{cm.offer_id}</span></p>
                            )}
                          </div>
                        ))}
                      </div>
                    </section>
                  )}

                  {data?.considerations?.length > 0 && (
                    <section>
                      <p className="text-xs font-semibold text-slate-500 uppercase tracking-wider mb-2">Consideration</p>
                      <div className="space-y-2">
                        {data.considerations.map(c => (
                          <div key={c.id} className="bg-slate-800/50 rounded-lg p-3 text-xs space-y-1">
                            <div className="flex justify-between">
                              <span className="text-slate-300 font-semibold">{c.status_code}</span>
                              {c.status_name && <span className="text-slate-500">{c.status_name}</span>}
                            </div>
                            {c.consideration_attributes && (
                              <pre className="text-slate-400 text-[10px] mt-1 overflow-x-auto">
                                {JSON.stringify(c.consideration_attributes, null, 2)}
                              </pre>
                            )}
                          </div>
                        ))}
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
  const [page, setPage] = useState(1)
  const [selected, setSelected] = useState(null)
  const limit = 20
  const { data, loading, error, reload } = useApi(`/orders?page=${page}&limit=${limit}`, [page])

  const total = data?.total ?? 0
  const totalPages = Math.max(1, Math.ceil(total / limit))

  return (
    <div>
      <PageHeader
        title="Orders"
        subtitle={total ? `${total} total contracts` : ''}
        action={
          <button onClick={reload} className="flex items-center gap-1.5 text-xs text-slate-400 hover:text-slate-200">
            <RefreshCw size={13} className={loading ? 'animate-spin' : ''} />
          </button>
        }
      />

      {error && <ErrorBox message={error} onRetry={reload} />}

      <Card>
        {loading ? <LoadingSpinner /> : (
          <table className="w-full text-sm">
            <thead>
              <tr className="text-xs text-slate-500 uppercase tracking-wider border-b border-slate-800">
                <th className="text-left px-4 py-3 font-medium">Txn ID</th>
                <th className="text-left px-4 py-3 font-medium">Buyer</th>
                <th className="text-left px-4 py-3 font-medium">Status</th>
                <th className="text-left px-4 py-3 font-medium">Offer</th>
                <th className="text-left px-4 py-3 font-medium">Created</th>
                <th className="px-4 py-3" />
              </tr>
            </thead>
            <tbody className="divide-y divide-slate-800/60">
              {(data?.items ?? []).map(o => (
                <tr key={o.id} className="hover:bg-slate-800/30 transition-colors">
                  <td className="px-4 py-3 font-mono text-xs text-slate-400">{o.transaction_id?.slice(0, 8)}…</td>
                  <td className="px-4 py-3 text-slate-300 truncate max-w-[160px]">{o.bap_id}</td>
                  <td className="px-4 py-3"><StatusBadge status={o.status} /></td>
                  <td className="px-4 py-3 text-slate-400 text-xs truncate max-w-[140px]">
                    {fmtName(o.offer_descriptor)}
                  </td>
                  <td className="px-4 py-3 text-slate-500 text-xs">{fmtDate(o.created_at?.Time || o.created_at)}</td>
                  <td className="px-4 py-3">
                    <button
                      onClick={() => setSelected(o)}
                      className="text-indigo-400 hover:text-indigo-300 flex items-center gap-1 text-xs"
                    >
                      <Eye size={13} /> View
                    </button>
                  </td>
                </tr>
              ))}
              {!data?.items?.length && (
                <tr><td colSpan={6} className="px-4 py-8 text-center text-slate-500">No orders found</td></tr>
              )}
            </tbody>
          </table>
        )}

        {totalPages > 1 && (
          <div className="flex items-center justify-between px-4 py-3 border-t border-slate-800">
            <span className="text-xs text-slate-500">Page {page} of {totalPages}</span>
            <div className="flex gap-2">
              <button
                disabled={page <= 1}
                onClick={() => setPage(p => p - 1)}
                className="p-1.5 rounded-lg border border-slate-700 text-slate-400 hover:text-white disabled:opacity-30 disabled:cursor-not-allowed"
              >
                <ChevronLeft size={14} />
              </button>
              <button
                disabled={page >= totalPages}
                onClick={() => setPage(p => p + 1)}
                className="p-1.5 rounded-lg border border-slate-700 text-slate-400 hover:text-white disabled:opacity-30 disabled:cursor-not-allowed"
              >
                <ChevronRightIcon size={14} />
              </button>
            </div>
          </div>
        )}
      </Card>

      <OrderDetailDrawer order={selected} onClose={() => setSelected(null)} />
    </div>
  )
}

// ─── Inventory page ──────────────────────────────────────────────────────────

function ResourcesTab() {
  const [page, setPage] = useState(1)
  const limit = 20
  const { data, loading, error, reload } = useApi(`/inventory/resources?page=${page}&limit=${limit}`, [page])
  const total = data?.total ?? 0
  const totalPages = Math.max(1, Math.ceil(total / limit))

  return (
    <div>
      {error && <ErrorBox message={error} onRetry={reload} />}
      <Card className="mt-4">
        {loading ? <LoadingSpinner /> : (
          <table className="w-full text-sm">
            <thead>
              <tr className="text-xs text-slate-500 uppercase tracking-wider border-b border-slate-800">
                <th className="text-left px-4 py-3 font-medium">ID</th>
                <th className="text-left px-4 py-3 font-medium">Name</th>
                <th className="text-left px-4 py-3 font-medium">Description</th>
                <th className="text-left px-4 py-3 font-medium">Catalog</th>
                <th className="text-left px-4 py-3 font-medium">Added</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-slate-800/60">
              {(data?.items ?? []).map(r => (
                <tr key={r.id} className="hover:bg-slate-800/30 transition-colors">
                  <td className="px-4 py-3 font-mono text-xs text-slate-400 max-w-[120px] truncate">{r.id}</td>
                  <td className="px-4 py-3 text-slate-200 font-medium">{r.descriptor_name || '—'}</td>
                  <td className="px-4 py-3 text-slate-400 text-xs max-w-[200px] truncate">{r.descriptor_short_desc || '—'}</td>
                  <td className="px-4 py-3 text-slate-500 text-xs">{r.catalog_id}</td>
                  <td className="px-4 py-3 text-slate-500 text-xs">{fmtDate(r.created_at?.Time || r.created_at)}</td>
                </tr>
              ))}
              {!data?.items?.length && (
                <tr><td colSpan={5} className="px-4 py-8 text-center text-slate-500">No resources found</td></tr>
              )}
            </tbody>
          </table>
        )}
        {totalPages > 1 && (
          <div className="flex items-center justify-between px-4 py-3 border-t border-slate-800">
            <span className="text-xs text-slate-500">Page {page} of {totalPages} · {total} total</span>
            <div className="flex gap-2">
              <button disabled={page <= 1} onClick={() => setPage(p => p - 1)}
                className="p-1.5 rounded-lg border border-slate-700 text-slate-400 hover:text-white disabled:opacity-30">
                <ChevronLeft size={14} />
              </button>
              <button disabled={page >= totalPages} onClick={() => setPage(p => p + 1)}
                className="p-1.5 rounded-lg border border-slate-700 text-slate-400 hover:text-white disabled:opacity-30">
                <ChevronRightIcon size={14} />
              </button>
            </div>
          </div>
        )}
      </Card>
    </div>
  )
}

function OffersTab() {
  const [page, setPage] = useState(1)
  const limit = 20
  const { data, loading, error, reload } = useApi(`/inventory/offers?page=${page}&limit=${limit}`, [page])
  const total = data?.total ?? 0
  const totalPages = Math.max(1, Math.ceil(total / limit))

  return (
    <div>
      {error && <ErrorBox message={error} onRetry={reload} />}
      <Card className="mt-4">
        {loading ? <LoadingSpinner /> : (
          <table className="w-full text-sm">
            <thead>
              <tr className="text-xs text-slate-500 uppercase tracking-wider border-b border-slate-800">
                <th className="text-left px-4 py-3 font-medium">ID</th>
                <th className="text-left px-4 py-3 font-medium">Name</th>
                <th className="text-left px-4 py-3 font-medium">Resources</th>
                <th className="text-left px-4 py-3 font-medium">Valid Until</th>
                <th className="text-left px-4 py-3 font-medium">Added</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-slate-800/60">
              {(data?.items ?? []).map(o => (
                <tr key={o.id} className="hover:bg-slate-800/30 transition-colors">
                  <td className="px-4 py-3 font-mono text-xs text-slate-400 max-w-[120px] truncate">{o.id}</td>
                  <td className="px-4 py-3 text-slate-200 font-medium">{o.descriptor_name || '—'}</td>
                  <td className="px-4 py-3 text-slate-400 text-xs">
                    {o.resource_ids?.length ? (
                      <span className="bg-slate-700/60 px-1.5 py-0.5 rounded text-xs">
                        {o.resource_ids.length} resource{o.resource_ids.length !== 1 ? 's' : ''}
                      </span>
                    ) : '—'}
                  </td>
                  <td className="px-4 py-3 text-slate-500 text-xs">
                    {o.validity_end?.Time ? fmtDate(o.validity_end.Time) : '—'}
                  </td>
                  <td className="px-4 py-3 text-slate-500 text-xs">{fmtDate(o.created_at?.Time || o.created_at)}</td>
                </tr>
              ))}
              {!data?.items?.length && (
                <tr><td colSpan={5} className="px-4 py-8 text-center text-slate-500">No offers found</td></tr>
              )}
            </tbody>
          </table>
        )}
        {totalPages > 1 && (
          <div className="flex items-center justify-between px-4 py-3 border-t border-slate-800">
            <span className="text-xs text-slate-500">Page {page} of {totalPages} · {total} total</span>
            <div className="flex gap-2">
              <button disabled={page <= 1} onClick={() => setPage(p => p - 1)}
                className="p-1.5 rounded-lg border border-slate-700 text-slate-400 hover:text-white disabled:opacity-30">
                <ChevronLeft size={14} />
              </button>
              <button disabled={page >= totalPages} onClick={() => setPage(p => p + 1)}
                className="p-1.5 rounded-lg border border-slate-700 text-slate-400 hover:text-white disabled:opacity-30">
                <ChevronRightIcon size={14} />
              </button>
            </div>
          </div>
        )}
      </Card>
    </div>
  )
}

function InventoryPage({ initialTab = 'resources' }) {
  const [tab, setTab] = useState(initialTab)
  return (
    <div>
      <PageHeader title="Inventory" subtitle="Resources and offers published to the network" />
      <div className="flex gap-1 p-1 bg-slate-800/50 rounded-lg w-fit mb-0">
        {['resources', 'offers'].map(t => (
          <button
            key={t}
            onClick={() => setTab(t)}
            className={`px-4 py-1.5 rounded-md text-sm font-medium transition-all capitalize
              ${tab === t ? 'bg-indigo-600 text-white shadow' : 'text-slate-400 hover:text-slate-200'}`}
          >
            {t}
          </button>
        ))}
      </div>
      {tab === 'resources' ? <ResourcesTab /> : <OffersTab />}
    </div>
  )
}

// ─── Publish page ────────────────────────────────────────────────────────────

const EMPTY_RESOURCE  = () => ({ id: '', name: '', short_desc: '', long_desc: '' })
const EMPTY_OFFER     = () => ({ id: '', name: '', short_desc: '', resource_ids: '', price: '', currency: 'INR', validity_start: '', validity_end: '' })
const CATALOG_TYPES   = ['', 'master', 'regular']

function PublishPage({ onNav }) {
  const [form, setForm] = useState({
    catalog_id:    '',
    catalog_name:  '',
    provider_id:   '',
    provider_name: '',
    catalog_type:  '',
    resources: [EMPTY_RESOURCE()],
    offers:    [EMPTY_OFFER()],
  })
  const [error, setError]     = useState(null)
  const [loading, setLoading] = useState(false)

  const setField = (k, v) => setForm(f => ({ ...f, [k]: v }))
  const setResourceField = (i, k, v) =>
    setForm(f => { const r = [...f.resources]; r[i] = { ...r[i], [k]: v }; return { ...f, resources: r } })
  const setOfferField = (i, k, v) =>
    setForm(f => { const o = [...f.offers]; o[i] = { ...o[i], [k]: v }; return { ...f, offers: o } })

  const toISO = localDT => localDT ? new Date(localDT).toISOString() : undefined

  const handleSubmit = async e => {
    e.preventDefault()
    setLoading(true)
    setError(null)
    try {
      const catalog = {
        id: form.catalog_id,
        descriptor: { name: form.catalog_name },
        provider: { id: form.provider_id, descriptor: { name: form.provider_name } },
        resources: form.resources.map(r => ({
          id: r.id,
          descriptor: { name: r.name, shortDesc: r.short_desc, longDesc: r.long_desc },
        })),
        offers: form.offers.map(o => {
          const start = toISO(o.validity_start)
          const end   = toISO(o.validity_end)
          return {
            id: o.id,
            descriptor: { name: o.name, shortDesc: o.short_desc },
            resourceIds: o.resource_ids.split(',').map(s => s.trim()).filter(Boolean),
            ...(o.price ? {
              considerations: [{
                id: `${o.id}-price`,
                status: { code: 'ACTIVE' },
                considerationAttributes: JSON.stringify({
                  '@type': 'PriceSpecification',
                  price: o.price,
                  currency: o.currency,
                }),
              }],
            } : {}),
            ...(start ? { validity: { startDate: start, endDate: end } } : {}),
          }
        }),
        ...(form.catalog_type ? {
          publishDirectives: { catalogType: form.catalog_type },
        } : {}),
      }
      await apiFetch('/catalog/publish', { method: 'POST', body: JSON.stringify({ catalogs: [catalog] }) })
      // Redirect to inventory so the user can immediately see the new resources/offers
      onNav('inventory')
    } catch (err) {
      setError(err.message)
    } finally {
      setLoading(false)
    }
  }

  const inputCls = 'w-full bg-slate-800/60 border border-slate-700 rounded-lg px-3 py-2 text-sm text-slate-200 placeholder-slate-500 focus:outline-none focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500/50 transition-colors'
  const labelCls = 'block text-xs font-medium text-slate-400 mb-1'

  return (
    <div>
      <PageHeader title="Publish Catalog" subtitle="Push resources and offers to the Beckn network" />

      {error && <ErrorBox message={error} className="mb-5" />}

      <form onSubmit={handleSubmit} className="space-y-5">
        {/* Catalog info */}
        <Card className="p-5">
          <p className="text-sm font-semibold text-white mb-4">Catalog Info</p>
          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className={labelCls}>Catalog ID</label>
              <input className={inputCls} value={form.catalog_id} onChange={e => setField('catalog_id', e.target.value)} placeholder="e.g. cat-winroom-2026" required />
            </div>
            <div>
              <label className={labelCls}>Catalog Name</label>
              <input className={inputCls} value={form.catalog_name} onChange={e => setField('catalog_name', e.target.value)} placeholder="e.g. Winroom Catalog" required />
            </div>
            <div>
              <label className={labelCls}>Provider ID</label>
              <input className={inputCls} value={form.provider_id} onChange={e => setField('provider_id', e.target.value)} placeholder="e.g. provider-001" required />
            </div>
            <div>
              <label className={labelCls}>Provider Name</label>
              <input className={inputCls} value={form.provider_name} onChange={e => setField('provider_name', e.target.value)} placeholder="e.g. Winroom Hotels" required />
            </div>
            <div>
              <label className={labelCls}>Catalog Type <span className="text-slate-600">(optional)</span></label>
              <select className={inputCls} value={form.catalog_type} onChange={e => setField('catalog_type', e.target.value)}>
                {CATALOG_TYPES.map(t => (
                  <option key={t} value={t}>{t || '— none —'}</option>
                ))}
              </select>
            </div>
          </div>
        </Card>

        {/* Resources */}
        <Card className="p-5">
          <div className="flex items-center justify-between mb-4">
            <p className="text-sm font-semibold text-white">Resources</p>
            <button type="button" onClick={() => setForm(f => ({ ...f, resources: [...f.resources, EMPTY_RESOURCE()] }))}
              className="flex items-center gap-1 text-xs text-indigo-400 hover:text-indigo-300">
              <Plus size={13} /> Add Resource
            </button>
          </div>
          <div className="space-y-4">
            {form.resources.map((r, i) => (
              <div key={i} className="bg-slate-800/40 rounded-lg p-4 space-y-3">
                <div className="flex items-center justify-between">
                  <span className="text-xs text-slate-500 font-semibold">Resource {i + 1}</span>
                  {form.resources.length > 1 && (
                    <button type="button" onClick={() => setForm(f => ({ ...f, resources: f.resources.filter((_, j) => j !== i) }))}
                      className="text-red-400 hover:text-red-300"><Trash2 size={13} /></button>
                  )}
                </div>
                <div className="grid grid-cols-2 gap-3">
                  <div>
                    <label className={labelCls}>ID</label>
                    <input className={inputCls} value={r.id} onChange={e => setResourceField(i, 'id', e.target.value)} placeholder="e.g. room-deluxe-001" required />
                  </div>
                  <div>
                    <label className={labelCls}>Name</label>
                    <input className={inputCls} value={r.name} onChange={e => setResourceField(i, 'name', e.target.value)} placeholder="e.g. Deluxe Room" required />
                  </div>
                  <div>
                    <label className={labelCls}>Short Description</label>
                    <input className={inputCls} value={r.short_desc} onChange={e => setResourceField(i, 'short_desc', e.target.value)} placeholder="One-liner" />
                  </div>
                  <div>
                    <label className={labelCls}>Long Description</label>
                    <input className={inputCls} value={r.long_desc} onChange={e => setResourceField(i, 'long_desc', e.target.value)} placeholder="Detailed description" />
                  </div>
                </div>
              </div>
            ))}
          </div>
        </Card>

        {/* Offers */}
        <Card className="p-5">
          <div className="flex items-center justify-between mb-4">
            <p className="text-sm font-semibold text-white">Offers</p>
            <button type="button" onClick={() => setForm(f => ({ ...f, offers: [...f.offers, EMPTY_OFFER()] }))}
              className="flex items-center gap-1 text-xs text-indigo-400 hover:text-indigo-300">
              <Plus size={13} /> Add Offer
            </button>
          </div>
          <div className="space-y-4">
            {form.offers.map((o, i) => (
              <div key={i} className="bg-slate-800/40 rounded-lg p-4 space-y-3">
                <div className="flex items-center justify-between">
                  <span className="text-xs text-slate-500 font-semibold">Offer {i + 1}</span>
                  {form.offers.length > 1 && (
                    <button type="button" onClick={() => setForm(f => ({ ...f, offers: f.offers.filter((_, j) => j !== i) }))}
                      className="text-red-400 hover:text-red-300"><Trash2 size={13} /></button>
                  )}
                </div>
                <div className="grid grid-cols-2 gap-3">
                  <div>
                    <label className={labelCls}>ID</label>
                    <input className={inputCls} value={o.id} onChange={e => setOfferField(i, 'id', e.target.value)} placeholder="e.g. offer-deluxe-bp" required />
                  </div>
                  <div>
                    <label className={labelCls}>Name</label>
                    <input className={inputCls} value={o.name} onChange={e => setOfferField(i, 'name', e.target.value)} placeholder="e.g. Deluxe Room — Bed & Breakfast" required />
                  </div>
                  <div>
                    <label className={labelCls}>Short Description</label>
                    <input className={inputCls} value={o.short_desc} onChange={e => setOfferField(i, 'short_desc', e.target.value)} placeholder="Brief offer description" />
                  </div>
                  <div>
                    <label className={labelCls}>Resource IDs (comma-separated)</label>
                    <input className={inputCls} value={o.resource_ids} onChange={e => setOfferField(i, 'resource_ids', e.target.value)} placeholder="room-deluxe-001, room-deluxe-002" />
                  </div>
                  <div>
                    <label className={labelCls}>Price</label>
                    <input className={inputCls} value={o.price} onChange={e => setOfferField(i, 'price', e.target.value)} placeholder="e.g. 4500" />
                  </div>
                  <div>
                    <label className={labelCls}>Currency</label>
                    <input className={inputCls} value={o.currency} onChange={e => setOfferField(i, 'currency', e.target.value)} placeholder="INR" />
                  </div>
                  <div>
                    <label className={labelCls}>Validity Start <span className="text-slate-600">(optional)</span></label>
                    <input type="datetime-local" className={inputCls} value={o.validity_start} onChange={e => setOfferField(i, 'validity_start', e.target.value)} />
                  </div>
                  <div>
                    <label className={labelCls}>Validity End <span className="text-slate-600">(optional)</span></label>
                    <input type="datetime-local" className={inputCls} value={o.validity_end} onChange={e => setOfferField(i, 'validity_end', e.target.value)} />
                  </div>
                </div>
              </div>
            ))}
          </div>
        </Card>

        <div className="flex justify-end">
          <button
            type="submit"
            disabled={loading}
            className="flex items-center gap-2 px-6 py-2.5 bg-indigo-600 hover:bg-indigo-500 disabled:opacity-60 text-white font-semibold text-sm rounded-lg transition-colors shadow-lg"
          >
            {loading ? <RefreshCw size={15} className="animate-spin" /> : <Upload size={15} />}
            {loading ? 'Publishing…' : 'Publish Catalog'}
          </button>
        </div>
      </form>
    </div>
  )
}

// ─── Messages page ───────────────────────────────────────────────────────────

const ACTION_COLORS = {
  select:     'text-indigo-400',
  on_select:  'text-indigo-300',
  init:       'text-violet-400',
  on_init:    'text-violet-300',
  confirm:    'text-emerald-400',
  on_confirm: 'text-emerald-300',
}

function MessagesPage() {
  const [limit, setLimit] = useState(50)
  const [filter, setFilter] = useState('')
  const { data, loading, error, reload } = useApi(`/messages?limit=${limit}`, [limit])

  const items = (data?.items ?? []).filter(m =>
    !filter || String(m.action).includes(filter.toLowerCase()) || String(m.transaction_id).includes(filter)
  )

  return (
    <div>
      <PageHeader
        title="Message Log"
        subtitle="Recent Beckn protocol messages"
        action={
          <button onClick={reload} className="flex items-center gap-1.5 text-xs text-slate-400 hover:text-slate-200">
            <RefreshCw size={13} className={loading ? 'animate-spin' : ''} />
          </button>
        }
      />

      {error && <ErrorBox message={error} onRetry={reload} />}

      <div className="flex gap-3 mb-4">
        <div className="relative flex-1 max-w-xs">
          <Search size={13} className="absolute left-3 top-1/2 -translate-y-1/2 text-slate-500" />
          <input
            value={filter}
            onChange={e => setFilter(e.target.value)}
            placeholder="Filter by action or txn ID…"
            className="w-full bg-slate-800/60 border border-slate-700 rounded-lg pl-8 pr-3 py-2 text-sm text-slate-200 placeholder-slate-500 focus:outline-none focus:border-indigo-500"
          />
        </div>
        <div className="flex items-center gap-2">
          <Filter size={13} className="text-slate-500" />
          <select
            value={limit}
            onChange={e => setLimit(Number(e.target.value))}
            className="bg-slate-800/60 border border-slate-700 rounded-lg px-3 py-2 text-sm text-slate-300 focus:outline-none focus:border-indigo-500"
          >
            {[25, 50, 100, 200].map(n => <option key={n} value={n}>{n} rows</option>)}
          </select>
        </div>
      </div>

      <Card>
        {loading ? <LoadingSpinner /> : (
          <table className="w-full text-sm">
            <thead>
              <tr className="text-xs text-slate-500 uppercase tracking-wider border-b border-slate-800">
                <th className="text-left px-4 py-3 font-medium">Action</th>
                <th className="text-left px-4 py-3 font-medium">Direction</th>
                <th className="text-left px-4 py-3 font-medium">Status</th>
                <th className="text-left px-4 py-3 font-medium">Txn ID</th>
                <th className="text-left px-4 py-3 font-medium">URL</th>
                <th className="text-left px-4 py-3 font-medium">Time</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-slate-800/60">
              {items.map((m, idx) => (
                <tr key={idx} className="hover:bg-slate-800/30 transition-colors">
                  <td className={`px-4 py-2.5 font-mono text-xs font-semibold ${ACTION_COLORS[m.action] || 'text-slate-400'}`}>
                    {m.action}
                  </td>
                  <td className="px-4 py-2.5 text-xs text-slate-400">{m.direction}</td>
                  <td className="px-4 py-2.5">
                    <StatusBadge status={m.ack_status?.ack_status || m.ack_status} />
                  </td>
                  <td className="px-4 py-2.5 font-mono text-xs text-slate-500">
                    {m.transaction_id ? String(m.transaction_id).slice(0, 8) + '…' : '—'}
                  </td>
                  <td className="px-4 py-2.5 text-xs text-slate-500 max-w-[200px] truncate" title={m.url}>{m.url || '—'}</td>
                  <td className="px-4 py-2.5 text-xs text-slate-500">{fmtDate(m.created_at?.Time || m.created_at)}</td>
                </tr>
              ))}
              {!items.length && (
                <tr><td colSpan={6} className="px-4 py-8 text-center text-slate-500">No messages found</td></tr>
              )}
            </tbody>
          </table>
        )}
      </Card>
    </div>
  )
}

// ─── App shell ───────────────────────────────────────────────────────────────

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

  const onNav = (target, extra = {}) => {
    setPage(target)
    setNavExtra(extra)
  }

  return (
    <div className="flex min-h-screen bg-[#0f1117]">
      <Sidebar active={page} onNav={onNav} />

      <main className="flex-1 ml-60 p-6 lg:p-8 overflow-y-auto min-h-screen">
        <AnimatePresence mode="wait">
          <motion.div
            key={page}
            initial={{ opacity: 0, y: 10 }}
            animate={{ opacity: 1, y: 0 }}
            exit={{ opacity: 0, y: -10 }}
            transition={{ duration: 0.15 }}
          >
            <Page onNav={onNav} {...navExtra} />
          </motion.div>
        </AnimatePresence>
      </main>
    </div>
  )
}
