import React, { useState, useEffect, useRef, useCallback } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import {
  ShoppingCart, Search, X, Package, CheckCircle2, Loader2,
  ArrowRight, ChevronRight, Minus, Plus, Sparkles, Tag,
  RotateCcw, CreditCard, ShieldCheck, Clock,
  Star, Info, Receipt, Store, MapPin, Calendar, Zap,
} from 'lucide-react';

const API_BASE = '/api/v1';
const BPP_ID  = 'bpptest1.remiges.tech';
const BPP_URI = 'https://bpptest.remiges.tech/bpp/receiver';
const NET_ID  = 'ion.id/ion-winroom-0426';

/* ─────────────────────────── helpers ────────────────────────────────────── */

function parseCatalogs(data) {
  const products = [];
  (data?.message?.catalogs || []).forEach(cat => {
    (cat.resources || []).forEach(res => {
      // Find the offer that covers this resource
      const offer = cat.offers?.find(o => o.resourceIds?.includes(res.id)) || null;
      const oa    = offer?.offerAttributes || {};
      const pa    = oa.price || {};
      const cons  = offer?.considerations?.[0]?.considerationAttributes || {};

      // Price: prefer considerations.totalAmount, fallback to offerAttributes.price.value
      let price    = cons.totalAmount ?? pa.value ?? 0;
      let currency = cons.currency || pa.currency || 'INR';

      // Price breakup: prefer considerations.breakup, build from price.components if needed
      let breakup = cons.breakup || [];
      if (!breakup.length && pa.value) {
        breakup = [{ title: res.descriptor?.name || 'Item price', amount: pa.value, type: 'BASE_PRICE' }];
        if (pa.components?.length) {
          pa.components.forEach(c => {
            if (c.type === 'DISCOUNT') {
              breakup.push({ title: `Discount (${c.value}%)`, amount: -(pa.value * c.value / 100), type: 'DISCOUNT' });
            }
          });
          const discountAmt = pa.components.filter(c=>c.type==='DISCOUNT').reduce((s,c)=>s+(pa.value*c.value/100),0);
          price = pa.value - discountAmt;
        }
      }

      // Seller / provider
      const provider      = offer?.provider || {};
      const sellerName    = provider.descriptor?.name || '';
      const sellerAddress = provider.availableAt?.[0]?.address;

      // Policies
      const policies      = oa.policies || {};
      const serviceability= oa.serviceability || {};
      const timing        = serviceability.timing?.[0] || {};
      const distance      = serviceability.distanceConstraint || {};

      products.push({
        id: res.id,
        name: res.descriptor?.name || 'Unknown Product',
        shortDesc: res.descriptor?.shortDesc || '',
        longDesc: res.descriptor?.longDesc || '',
        category: cat.descriptor?.name || 'Retail',
        price,
        currency,
        image: res.descriptor?.mediaFile?.[0]?.uri ||
          'https://images.unsplash.com/photo-1542838132-92c53300491e?w=600&auto=format&fit=crop&q=70',
        // Full offer info for modal
        offerId:    offer?.id || '',
        offerName:  offer?.descriptor?.name || offer?.descriptor?.shortDesc || '',
        breakup,
        sellerName,
        sellerAddress,
        returnPolicy:   policies.returns,
        cancelPolicy:   policies.cancellation,
        codAvailable:   oa.paymentConstraints?.codAvailable,
        deliveryRange:  distance.maxDistance ? `${distance.maxDistance} ${distance.unit || 'KM'}` : '',
        deliveryDays:   timing.daysOfWeek || [],
        deliveryHours:  timing.timeRange ? `${timing.timeRange.start}–${timing.timeRange.end}` : '',
        offerValidity:  offer?.validity?.endDate ? new Date(offer.validity.endDate).toLocaleDateString('en-IN', { day:'numeric', month:'short', year:'numeric' }) : '',
      });
    });
  });
  return products;
}

// Parse the on_select contract from status polling to get the BPP-confirmed quote
function parseOnSelectContract(contract) {
  if (!contract) return null;
  const consideration = contract.consideration?.[0] ?? contract.considerations?.[0];
  const commitment    = contract.commitments?.[0];
  const ca  = consideration?.considerationAttributes || {};
  const pol = commitment?.offer?.offerAttributes?.policies || {};
  return {
    total:          ca.totalAmount,
    currency:       ca.currency || 'INR',
    breakup:        ca.breakup || [],
    paymentMethods: ca.paymentMethods || [],
    codAvailable:   commitment?.offer?.offerAttributes?.paymentConstraints?.codAvailable,
    seller:         commitment?.offer?.provider?.descriptor?.name,
    offerName:      commitment?.offer?.descriptor?.name,
    returnPolicy:   pol.returns,
    cancelPolicy:   pol.cancellation,
  };
}

const fmt = (amount, currency = 'INR') => {
  if (!amount && amount !== 0) return '—';
  const sym = currency === 'IDR' ? 'Rp' : '₹';
  return `${sym}${Number(amount).toLocaleString()}`;
};

const FALLBACK_IMG = 'https://images.unsplash.com/photo-1542838132-92c53300491e?w=600&auto=format&fit=crop&q=70';

/* ─────────────────────────── SkeletonCard ───────────────────────────────── */
function SkeletonCard() {
  return (
    <div className="glass rounded-2xl overflow-hidden">
      <div className="skeleton h-44 w-full" />
      <div className="p-4 space-y-2.5">
        <div className="skeleton h-3 w-1/3" />
        <div className="skeleton h-4 w-4/5" />
        <div className="skeleton h-3 w-2/3" />
        <div className="flex justify-between items-center pt-2">
          <div className="skeleton h-6 w-16" />
          <div className="skeleton h-9 w-20 rounded-xl" />
        </div>
      </div>
    </div>
  );
}

/* ─────────────────────────── OfferInfoPanel ─────────────────────────────── */
// Shows catalog offer details: price breakdown, seller, policies, serviceability
function OfferInfoPanel({ product }) {
  const { breakup, currency, price, sellerName, sellerAddress,
          returnPolicy, cancelPolicy, codAvailable,
          deliveryRange, deliveryHours, deliveryDays, offerName, offerValidity } = product;

  const shortAddr = sellerAddress
    ? [sellerAddress.streetAddress, sellerAddress.addressLocality, sellerAddress.addressRegion].filter(Boolean).join(', ')
    : '';

  return (
    <div className="space-y-3">
      {/* Price section */}
      <div className="rounded-xl border border-white/10 bg-white/[0.03] overflow-hidden">
        <div className="px-4 py-2.5 bg-white/[0.04] border-b border-white/[0.07] flex items-center gap-2">
          <Receipt className="w-3.5 h-3.5 text-emerald-400" />
          <span className="text-xs font-semibold text-slate-300 uppercase tracking-wider">
            {offerName || 'Pricing'}
          </span>
          {offerValidity && (
            <span className="ml-auto text-[10px] text-slate-500 flex items-center gap-1">
              <Calendar className="w-3 h-3" /> Valid till {offerValidity}
            </span>
          )}
        </div>
        <div className="px-4 py-3 space-y-1.5">
          {breakup.length > 0 ? (
            <>
              {breakup.map((b, i) => (
                <div key={i} className="flex justify-between text-sm">
                  <span className="text-slate-400">{b.title}</span>
                  <span className={
                    b.type === 'DISCOUNT' ? 'text-emerald-400 font-medium' :
                    b.type === 'TAX'      ? 'text-amber-400' : 'text-slate-200'
                  }>
                    {b.type === 'DISCOUNT' && b.amount < 0 ? '−' : ''}{fmt(Math.abs(b.amount), currency)}
                  </span>
                </div>
              ))}
              <div className="border-t border-white/10 pt-2 mt-1 flex justify-between items-center">
                <span className="text-sm font-semibold text-slate-300">Total</span>
                <span className="text-xl font-bold text-emerald-300">{fmt(price, currency)}</span>
              </div>
            </>
          ) : (
            <div className="flex justify-between items-center">
              <span className="text-sm text-slate-400">Price</span>
              <span className="text-xl font-bold text-emerald-300">{fmt(price, currency)}</span>
            </div>
          )}
        </div>
      </div>

      {/* Seller */}
      {sellerName && (
        <div className="flex items-start gap-3 px-4 py-3 rounded-xl border border-white/10 bg-white/[0.03]">
          <Store className="w-4 h-4 text-slate-400 mt-0.5 shrink-0" />
          <div>
            <p className="text-sm font-medium text-slate-200">{sellerName}</p>
            {shortAddr && <p className="text-xs text-slate-500 mt-0.5">{shortAddr}</p>}
          </div>
        </div>
      )}

      {/* Delivery & badges */}
      <div className="flex flex-wrap gap-2">
        {deliveryRange && (
          <span className="flex items-center gap-1 text-[10px] font-semibold bg-blue-500/10 text-blue-400 border border-blue-500/20 px-2.5 py-1 rounded-full">
            <MapPin className="w-3 h-3" /> Delivers within {deliveryRange}
          </span>
        )}
        {deliveryHours && (
          <span className="flex items-center gap-1 text-[10px] font-semibold bg-white/5 text-slate-400 border border-white/10 px-2.5 py-1 rounded-full">
            <Clock className="w-3 h-3" /> {deliveryHours}
          </span>
        )}
        {codAvailable && (
          <span className="flex items-center gap-1 text-[10px] font-semibold bg-amber-500/10 text-amber-400 border border-amber-500/20 px-2.5 py-1 rounded-full">
            <CreditCard className="w-3 h-3" /> Cash on Delivery
          </span>
        )}
        {returnPolicy?.allowed && (
          <span className="flex items-center gap-1 text-[10px] font-semibold bg-purple-500/10 text-purple-400 border border-purple-500/20 px-2.5 py-1 rounded-full">
            <RotateCcw className="w-3 h-3" /> {returnPolicy.window?.replace('P','').replace('D',' day')} returns
          </span>
        )}
        {cancelPolicy?.allowed && (
          <span className="flex items-center gap-1 text-[10px] font-semibold bg-rose-500/10 text-rose-400 border border-rose-500/20 px-2.5 py-1 rounded-full">
            <X className="w-3 h-3" /> Free cancellation
          </span>
        )}
      </div>
    </div>
  );
}

/* ─────────────────────────── ConfirmedQuotePanel ────────────────────────── */
// Shows the BPP-confirmed quote from on_select (may differ from catalog price)
function ConfirmedQuotePanel({ quote, onInit, isLoading }) {
  if (!quote) return null;
  const { total, currency, breakup, paymentMethods, codAvailable,
          seller, returnPolicy, cancelPolicy } = quote;
  return (
    <motion.div
      initial={{ opacity: 0, y: 12 }} animate={{ opacity: 1, y: 0 }}
      className="rounded-xl border border-emerald-500/25 bg-emerald-500/5 overflow-hidden"
    >
      <div className="flex items-center gap-2 px-4 py-2.5 bg-emerald-500/10 border-b border-emerald-500/15">
        <CheckCircle2 className="w-3.5 h-3.5 text-emerald-400" />
        <span className="text-xs font-semibold text-emerald-300 uppercase tracking-wider">Confirmed Quote</span>
        {seller && <span className="ml-auto text-[10px] text-slate-400 flex items-center gap-1"><Store className="w-3 h-3" />{seller}</span>}
      </div>
      <div className="px-4 py-3 space-y-2">
        {breakup.length > 0 && (
          <div className="space-y-1.5">
            {breakup.map((b, i) => (
              <div key={i} className="flex justify-between text-sm">
                <span className="text-slate-400">{b.title}</span>
                <span className={b.type === 'TAX' ? 'text-amber-400' : b.type === 'DISCOUNT' ? 'text-emerald-400' : 'text-slate-200'}>
                  {fmt(b.amount, currency)}
                </span>
              </div>
            ))}
            <div className="border-t border-white/10 pt-2 flex justify-between font-bold">
              <span className="text-slate-200">Total</span>
              <span className="text-emerald-300 text-lg">{fmt(total, currency)}</span>
            </div>
          </div>
        )}
        <div className="flex flex-wrap gap-1.5 pt-1">
          {codAvailable && <span className="text-[10px] bg-amber-500/10 text-amber-400 border border-amber-500/20 px-2 py-0.5 rounded-full">COD</span>}
          {paymentMethods?.map(m => <span key={m} className="text-[10px] bg-white/5 text-slate-400 border border-white/10 px-2 py-0.5 rounded-full">{m}</span>)}
          {returnPolicy?.allowed && <span className="text-[10px] bg-purple-500/10 text-purple-400 border border-purple-500/20 px-2 py-0.5 rounded-full">↩ {returnPolicy.window} returns</span>}
        </div>
        <button
          onClick={onInit} disabled={isLoading}
          className="w-full mt-1 bg-emerald-500 hover:bg-emerald-400 disabled:opacity-50 text-slate-900 py-3 rounded-xl font-bold text-sm flex items-center justify-center gap-2 transition-all active:scale-[0.98]"
        >
          {isLoading ? <Loader2 className="w-4 h-4 animate-spin" /> : <><span>Initialize Order</span><ArrowRight className="w-4 h-4" /></>}
        </button>
      </div>
    </motion.div>
  );
}

/* ─────────────────────────── ProductModal ───────────────────────────────── */
function ProductModal({ product, transaction, onClose, onSelect, onInit, onConfirm, isLoading }) {
  const isThisProduct = transaction?.product?.id === product.id;
  const status        = isThisProduct ? transaction.status : null;

  // Parse confirmed on_select contract once QUOTE_RECEIVED
  const confirmedQuote = (status === 'QUOTE_RECEIVED' && transaction?.contract)
    ? parseOnSelectContract(
        typeof transaction.contract === 'string'
          ? JSON.parse(transaction.contract)
          : transaction.contract
      )
    : null;

  const pendingStatuses = ['SELECT_SENT', 'INIT_SENT', 'CONFIRM_SENT'];
  const isPending = pendingStatuses.includes(status);

  const pendingLabel = {
    SELECT_SENT:  'Requesting quote from BPP…',
    INIT_SENT:    'Initializing order…',
    CONFIRM_SENT: 'Confirming your order…',
  }[status];

  return (
    <div className="fixed inset-0 z-50 flex items-end sm:items-center justify-center p-3 sm:p-4">
      <motion.div
        initial={{ opacity: 0 }} animate={{ opacity: 1 }} exit={{ opacity: 0 }}
        className="absolute inset-0 bg-black/70 backdrop-blur-sm"
        onClick={onClose}
      />
      <motion.div
        initial={{ opacity: 0, y: 60, scale: 0.97 }}
        animate={{ opacity: 1, y: 0, scale: 1 }}
        exit={{ opacity: 0, y: 40, scale: 0.97 }}
        transition={{ type: 'spring', damping: 28, stiffness: 300 }}
        className="relative w-full max-w-lg glass rounded-3xl overflow-hidden max-h-[94vh] flex flex-col"
      >
        {/* ── Hero image ─────────────────────────────────────────────── */}
        <div className="relative h-52 shrink-0 bg-slate-800/80 overflow-hidden">
          <img src={product.image} alt={product.name}
            className="w-full h-full object-cover opacity-90"
            onError={e => { e.target.src = FALLBACK_IMG; }} />
          <div className="absolute inset-0 bg-gradient-to-t from-black/80 via-black/10" />
          <button onClick={onClose}
            className="absolute top-3 right-3 w-8 h-8 flex items-center justify-center rounded-full bg-black/50 backdrop-blur text-white hover:bg-black/70 transition-colors z-10">
            <X className="w-4 h-4" />
          </button>
          <div className="absolute bottom-4 left-4 right-4 flex items-end justify-between gap-2">
            <span className="text-[10px] font-bold uppercase tracking-widest bg-black/50 text-emerald-400 border border-emerald-500/20 px-2.5 py-1 rounded-full backdrop-blur">
              <Tag className="inline w-2.5 h-2.5 mr-1 -mt-0.5" />{product.category}
            </span>
          </div>
        </div>

        {/* ── Scrollable body ─────────────────────────────────────────── */}
        <div className="overflow-y-auto flex-grow px-5 py-4 space-y-4">

          {/* Product name & description */}
          <div>
            <h2 className="text-xl leading-snug">{product.name}</h2>
            {product.shortDesc && (
              <p className="text-sm text-slate-400 mt-1.5 leading-relaxed">{product.shortDesc}</p>
            )}
            {product.longDesc && product.longDesc !== product.shortDesc && (
              <p className="text-xs text-slate-500 mt-1 leading-relaxed">{product.longDesc}</p>
            )}
          </div>

          {/* ── Offer info from catalog (always shown) ───────────────── */}
          <OfferInfoPanel product={product} />

          {/* ── Action area ─────────────────────────────────────────── */}

          {/* No transaction yet — show Select button */}
          {!status && (
            <button onClick={() => onSelect(product)} disabled={isLoading}
              className="w-full flex items-center justify-center gap-2 bg-emerald-500 hover:bg-emerald-400 disabled:opacity-50 text-slate-900 py-3.5 rounded-2xl font-bold text-sm transition-all active:scale-[0.98]">
              {isLoading
                ? <><Loader2 className="w-4 h-4 animate-spin" /> Sending select…</>
                : <><Zap className="w-4 h-4" /> Select &amp; Confirm Quote</>}
            </button>
          )}

          {/* Waiting for async response */}
          {isPending && (
            <div className="flex items-center gap-3 px-4 py-3 rounded-xl bg-white/[0.04] border border-white/10">
              <Loader2 className="w-5 h-5 text-emerald-400 animate-spin shrink-0" />
              <div>
                <p className="text-sm text-slate-300">{pendingLabel}</p>
                <p className="text-xs text-slate-500 mt-0.5">Txn {transaction.id.slice(0, 8)}…</p>
              </div>
            </div>
          )}

          {/* on_select confirmed quote */}
          {status === 'QUOTE_RECEIVED' && (
            <ConfirmedQuotePanel quote={confirmedQuote} onInit={onInit} isLoading={isLoading} />
          )}

          {/* Init received — confirm */}
          {status === 'INIT_RECEIVED' && (
            <div className="space-y-3">
              <div className="flex items-center gap-2 text-sm text-slate-300 bg-emerald-500/5 border border-emerald-500/20 rounded-xl px-4 py-3">
                <Info className="w-4 h-4 text-emerald-400 shrink-0" />
                Order initialized. Click below to place &amp; pay.
              </div>
              <button onClick={onConfirm} disabled={isLoading}
                className="w-full flex items-center justify-center gap-2 bg-emerald-500 hover:bg-emerald-400 text-slate-900 py-3.5 rounded-2xl font-bold text-sm transition-all active:scale-[0.98]">
                {isLoading ? <Loader2 className="w-4 h-4 animate-spin" /> : <><ShieldCheck className="w-4 h-4" />Pay &amp; Confirm Order</>}
              </button>
            </div>
          )}

          {/* Confirmed */}
          {status === 'CONFIRMED' && (
            <motion.div initial={{ scale: 0.9, opacity: 0 }} animate={{ scale: 1, opacity: 1 }}
              className="flex flex-col items-center gap-2 py-4">
              <div className="w-16 h-16 rounded-full bg-emerald-500/15 border border-emerald-500/30 flex items-center justify-center">
                <CheckCircle2 className="w-8 h-8 text-emerald-400" />
              </div>
              <p className="font-bold text-emerald-300 text-lg">Order Confirmed!</p>
              <p className="text-sm text-slate-400 text-center">Your order has been placed on the Beckn network.</p>
              <button onClick={onClose} className="mt-1 text-sm text-slate-400 hover:text-white transition-colors underline underline-offset-4">
                Continue Shopping
              </button>
            </motion.div>
          )}
        </div>
      </motion.div>
    </div>
  );
}

/* ─────────────────────────── App ────────────────────────────────────────── */
export default function App() {
  const [catalog, setCatalog]           = useState([]);
  const [isSearching, setIsSearching]   = useState(false);
  const [searchQuery, setSearchQuery]   = useState('');
  const [cart, setCart]                 = useState([]);
  const [transaction, setTransaction]   = useState(null);
  const [isLoading, setIsLoading]       = useState(false);
  const [isCartOpen, setIsCartOpen]     = useState(false);
  const [selectedProduct, setSelectedProduct] = useState(null);
  const [activeStep, setActiveStep]     = useState(0);
  const [error, setError]               = useState(null);
  const debounceRef = useRef(null);

  /* ── Discover ─────────────────────────────────────────────────────────── */
  const discover = useCallback((q = '') => {
    setIsSearching(true);
    setError(null);
    const url = q.trim()
      ? `${API_BASE}/discover?q=${encodeURIComponent(q.trim())}`
      : `${API_BASE}/discover`;
    fetch(url)
      .then(r => { if (!r.ok) throw new Error(`Status ${r.status}`); return r.json(); })
      .then(d => setCatalog(parseCatalogs(d)))
      .catch(e => { setError(e.message); setCatalog([]); })
      .finally(() => setIsSearching(false));
  }, []);

  useEffect(() => { discover(); }, [discover]);

  const handleSearchChange = e => {
    const v = e.target.value;
    setSearchQuery(v);
    clearTimeout(debounceRef.current);
    debounceRef.current = setTimeout(() => discover(v), 500);
  };

  /* ── Cart ─────────────────────────────────────────────────────────────── */
  const updateQty = (id, delta) =>
    setCart(p => p.map(i => i.id === id ? { ...i, qty: i.qty + delta } : i).filter(i => i.qty > 0));

  const addToCart = p => setCart(prev => {
    const ex = prev.find(i => i.id === p.id);
    return ex ? prev.map(i => i.id === p.id ? { ...i, qty: i.qty + 1 } : i) : [...prev, { ...p, qty: 1 }];
  });

  const cartCount = cart.reduce((s, i) => s + i.qty, 0);
  const cartTotal = cart.reduce((s, i) => s + i.price * i.qty, 0);

  /* ── Transaction polling ─────────────────────────────────────────────── */
  useEffect(() => {
    if (!transaction || ['CONFIRMED', 'CANCELLED'].includes(transaction.status)) return;
    const id = setInterval(() => {
      fetch(`${API_BASE}/status/${transaction.id}`)
        .then(r => r.json())
        .then(d => {
          setTransaction(p => ({ ...p, ...d }));
          if (d.status === 'QUOTE_RECEIVED') setActiveStep(1);
          if (d.status === 'INIT_RECEIVED')  setActiveStep(2);
          if (d.status === 'CONFIRMED')      setActiveStep(3);
        })
        .catch(() => {});
    }, 2000);
    return () => clearInterval(id);
  }, [transaction]);

  /* ── Beckn: Select (single product from modal) ───────────────────────── */
  const handleSelectProduct = async (product) => {
    setIsLoading(true);
    try {
      const r = await fetch(`${API_BASE}/select`, {
        method: 'POST', headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          bpp_id: BPP_ID, bpp_uri: BPP_URI, network_id: NET_ID,
          items: [{ id: product.id, quantity: { count: 1 } }],
        }),
      });
      if (!r.ok) throw new Error(await r.text());
      const d = await r.json();
      setTransaction({ id: d.transaction_id, status: 'SELECT_SENT', product });
      setActiveStep(0);
    } catch (e) { alert('Select failed: ' + e.message); }
    finally { setIsLoading(false); }
  };

  /* ── Beckn: Select (from cart) ───────────────────────────────────────── */
  const handleSelectCart = async () => {
    setIsLoading(true);
    try {
      const r = await fetch(`${API_BASE}/select`, {
        method: 'POST', headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          bpp_id: BPP_ID, bpp_uri: BPP_URI, network_id: NET_ID,
          items: cart.map(i => ({ id: i.id, quantity: { count: i.qty } })),
        }),
      });
      if (!r.ok) throw new Error(await r.text());
      const d = await r.json();
      setTransaction({ id: d.transaction_id, status: 'SELECT_SENT' });
      setIsCartOpen(false);
      setActiveStep(0);
    } catch (e) { alert('Select failed: ' + e.message); }
    finally { setIsLoading(false); }
  };

  /* ── Beckn: Init ─────────────────────────────────────────────────────── */
  const handleInit = async () => {
    if (!transaction) return;
    setIsLoading(true);
    try {
      await fetch(`${API_BASE}/init`, {
        method: 'POST', headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          transaction_id: transaction.id,
          bpp_id: BPP_ID, bpp_uri: BPP_URI, network_id: NET_ID,
          billing: { name: 'John Doe', phone: '9999999999', address: '123 Green Street' },
          fulfillments: [{ type: 'Delivery', end: { location: { gps: '12.9716,77.5946' } } }],
        }),
      });
      setTransaction(p => ({ ...p, status: 'INIT_SENT' }));
    } catch (e) { alert('Init failed: ' + e.message); }
    finally { setIsLoading(false); }
  };

  /* ── Beckn: Confirm ──────────────────────────────────────────────────── */
  const handleConfirm = async () => {
    if (!transaction) return;
    setIsLoading(true);
    try {
      await fetch(`${API_BASE}/confirm`, {
        method: 'POST', headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          transaction_id: transaction.id,
          bpp_id: BPP_ID, bpp_uri: BPP_URI, network_id: NET_ID,
        }),
      });
      setTransaction(p => ({ ...p, status: 'CONFIRM_SENT' }));
    } catch (e) { alert('Confirm failed: ' + e.message); }
    finally { setIsLoading(false); }
  };

  const stepLabels = ['Browse', 'Quote', 'Init', 'Confirmed'];

  /* ── Render ───────────────────────────────────────────────────────────── */
  return (
    <div className="min-h-screen">

      {/* ── Nav ──────────────────────────────────────────────────────────── */}
      <header className="sticky top-0 z-30 glass border-b border-white/[0.06]">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 h-16 flex items-center gap-4">
          <div className="flex items-center gap-2 shrink-0">
            <Sparkles className="w-5 h-5 text-emerald-400" />
            <span className="text-lg font-bold gradient-text" style={{ fontFamily: 'Syne, sans-serif' }}>
              NearShop
            </span>
          </div>

          <form onSubmit={e => { e.preventDefault(); clearTimeout(debounceRef.current); discover(searchQuery); }} className="flex-1 max-w-2xl mx-auto">
            <div className="search-glow relative flex items-center bg-white/[0.06] border border-white/10 rounded-xl transition-all duration-300">
              {isSearching
                ? <Loader2 className="absolute left-3 w-4 h-4 text-emerald-400 animate-spin" />
                : <Search className="absolute left-3 w-4 h-4 text-slate-400" />
              }
              <input
                type="text" value={searchQuery} onChange={handleSearchChange}
                placeholder="Search products on the Beckn network…"
                className="w-full bg-transparent pl-9 pr-4 py-2.5 text-sm text-white placeholder-slate-500 outline-none"
              />
              {searchQuery && (
                <button type="button" onClick={() => { setSearchQuery(''); discover(''); }}
                  className="absolute right-3 text-slate-400 hover:text-white transition-colors">
                  <X className="w-4 h-4" />
                </button>
              )}
            </div>
          </form>

          <button
            onClick={() => setIsCartOpen(true)}
            className="relative shrink-0 flex items-center gap-2 bg-emerald-500/10 hover:bg-emerald-500/20 border border-emerald-500/20 text-emerald-400 px-4 py-2 rounded-xl transition-all text-sm font-medium"
          >
            <ShoppingCart className="w-4 h-4" />
            <span className="hidden sm:block">Cart</span>
            {cartCount > 0 && (
              <span className="absolute -top-1.5 -right-1.5 w-5 h-5 bg-rose-500 text-white text-[10px] font-bold rounded-full flex items-center justify-center">
                {cartCount}
              </span>
            )}
          </button>
        </div>
      </header>

      {/* ── Page header ──────────────────────────────────────────────────── */}
      <div className="max-w-7xl mx-auto px-4 sm:px-6 pt-10 pb-6">
        <div className="flex items-end justify-between gap-4 flex-wrap">
          <div>
            <p className="text-xs uppercase tracking-widest text-emerald-400 font-semibold mb-1">
              Beckn Network · {NET_ID}
            </p>
            <h1 className="text-3xl sm:text-4xl gradient-text">
              {searchQuery ? `"${searchQuery}"` : 'Discover Products'}
            </h1>
          </div>
          {catalog.length > 0 && (
            <span className="text-sm text-slate-400">{catalog.length} product{catalog.length !== 1 ? 's' : ''} found</span>
          )}
        </div>
      </div>

      {/* ── Catalog ──────────────────────────────────────────────────────── */}
      <main className="max-w-7xl mx-auto px-4 sm:px-6 pb-40">

        {isSearching && catalog.length === 0 && (
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-5">
            {Array.from({ length: 8 }).map((_, i) => <SkeletonCard key={i} />)}
          </div>
        )}

        {error && !isSearching && (
          <div className="flex flex-col items-center justify-center py-28 gap-4 text-slate-400">
            <Package className="w-14 h-14 opacity-20" />
            <p>Could not reach the network.</p>
            <button onClick={() => discover(searchQuery)}
              className="bg-emerald-500/10 hover:bg-emerald-500/20 border border-emerald-500/20 text-emerald-400 px-5 py-2 rounded-xl text-sm font-medium transition-all">
              Retry
            </button>
          </div>
        )}

        {!isSearching && !error && catalog.length === 0 && (
          <div className="flex flex-col items-center justify-center py-28 gap-3 text-slate-400">
            <Search className="w-14 h-14 opacity-15" />
            <p>No products found{searchQuery ? ` for "${searchQuery}"` : ''}.</p>
            <p className="text-sm opacity-60">Try a different search term.</p>
          </div>
        )}

        {catalog.length > 0 && (
          <motion.div layout className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-5">
            <AnimatePresence>
              {catalog.map((product, i) => {
                const inCart = cart.find(c => c.id === product.id);
                return (
                  <motion.div
                    key={product.id} layout
                    initial={{ opacity: 0, y: 20 }} animate={{ opacity: 1, y: 0 }}
                    exit={{ opacity: 0, scale: 0.95 }}
                    transition={{ duration: 0.22, delay: i * 0.025 }}
                    className="product-card glass rounded-2xl overflow-hidden flex flex-col cursor-pointer group"
                    onClick={() => setSelectedProduct(product)}
                  >
                    {/* Image */}
                    <div className="relative h-44 overflow-hidden bg-slate-800/60 shrink-0">
                      <img
                        src={product.image} alt={product.name}
                        className="w-full h-full object-cover opacity-85 group-hover:opacity-100 group-hover:scale-105 transition-all duration-500"
                        onError={e => { e.target.src = FALLBACK_IMG; }}
                      />
                      {/* View detail hint */}
                      <div className="absolute inset-0 bg-black/0 group-hover:bg-black/20 transition-all duration-300 flex items-center justify-center">
                        <span className="opacity-0 group-hover:opacity-100 transition-opacity text-xs font-semibold text-white bg-black/50 backdrop-blur px-3 py-1.5 rounded-full">
                          View &amp; Select
                        </span>
                      </div>
                      <span className="absolute top-3 left-3 bg-black/50 backdrop-blur-md text-emerald-400 text-[10px] font-bold uppercase tracking-widest px-2.5 py-1 rounded-full border border-emerald-500/20">
                        <Tag className="inline w-2.5 h-2.5 mr-1 -mt-0.5" />
                        {product.category.length > 20 ? product.category.slice(0, 20) + '…' : product.category}
                      </span>
                    </div>

                    {/* Body */}
                    <div className="p-4 flex flex-col flex-grow gap-2">
                      <h3 className="text-sm font-semibold leading-snug line-clamp-2">{product.name}</h3>
                      {product.shortDesc && (
                        <p className="text-xs text-slate-400 line-clamp-2 leading-relaxed">{product.shortDesc}</p>
                      )}
                      <div className="mt-auto pt-3 flex items-center justify-between gap-2">
                        <span className="text-lg font-bold">{fmt(product.price, product.currency)}</span>

                        {/* Cart controls — stop propagation so card click doesn't fire */}
                        {inCart ? (
                          <div
                            onClick={e => e.stopPropagation()}
                            className="flex items-center gap-1 bg-emerald-500/10 border border-emerald-500/20 rounded-xl px-1 py-1"
                          >
                            <button onClick={() => updateQty(product.id, -1)} className="w-7 h-7 flex items-center justify-center rounded-lg hover:bg-white/10 text-emerald-400 transition-colors"><Minus className="w-3 h-3" /></button>
                            <span className="w-6 text-center text-sm font-bold text-emerald-400">{inCart.qty}</span>
                            <button onClick={() => updateQty(product.id, 1)} className="w-7 h-7 flex items-center justify-center rounded-lg hover:bg-white/10 text-emerald-400 transition-colors"><Plus className="w-3 h-3" /></button>
                          </div>
                        ) : (
                          <button
                            onClick={e => { e.stopPropagation(); addToCart(product); }}
                            className="flex items-center gap-1.5 bg-white/[0.08] hover:bg-white/[0.14] border border-white/10 text-slate-200 px-3 py-2 rounded-xl text-xs font-semibold transition-all"
                          >
                            <ShoppingCart className="w-3.5 h-3.5" /> Add
                          </button>
                        )}
                      </div>
                    </div>
                  </motion.div>
                );
              })}
            </AnimatePresence>
          </motion.div>
        )}
      </main>

      {/* ── Product Detail / Quote Modal ─────────────────────────────────── */}
      <AnimatePresence>
        {selectedProduct && (
          <ProductModal
            product={selectedProduct}
            transaction={transaction}
            onClose={() => setSelectedProduct(null)}
            onSelect={handleSelectProduct}
            onInit={handleInit}
            onConfirm={handleConfirm}
            isLoading={isLoading}
          />
        )}
      </AnimatePresence>

      {/* ── Cart Drawer ───────────────────────────────────────────────────── */}
      <AnimatePresence>
        {isCartOpen && (
          <>
            <motion.div initial={{ opacity: 0 }} animate={{ opacity: 1 }} exit={{ opacity: 0 }}
              className="fixed inset-0 bg-black/60 backdrop-blur-sm z-40" onClick={() => setIsCartOpen(false)} />
            <motion.aside
              initial={{ x: '100%' }} animate={{ x: 0 }} exit={{ x: '100%' }}
              transition={{ type: 'spring', damping: 30, stiffness: 300 }}
              className="fixed right-0 top-0 h-full w-full max-w-sm glass z-50 flex flex-col shadow-2xl"
            >
              <div className="flex items-center justify-between px-5 py-4 border-b border-white/[0.07]">
                <div className="flex items-center gap-2">
                  <ShoppingCart className="w-5 h-5 text-emerald-400" />
                  <h2 className="text-lg">Your Cart</h2>
                  {cartCount > 0 && <span className="bg-emerald-500/20 text-emerald-400 text-xs font-bold px-2 py-0.5 rounded-full">{cartCount}</span>}
                </div>
                <button onClick={() => setIsCartOpen(false)} className="text-slate-400 hover:text-white transition-colors"><X className="w-5 h-5" /></button>
              </div>

              <div className="flex-grow overflow-y-auto px-5 py-4 space-y-3">
                {cart.length === 0 ? (
                  <div className="h-full flex flex-col items-center justify-center text-slate-500 gap-3">
                    <ShoppingCart className="w-12 h-12 opacity-20" />
                    <p className="text-sm">Your cart is empty</p>
                    <p className="text-xs opacity-60">Click any product to view &amp; select it</p>
                  </div>
                ) : cart.map(item => (
                  <div key={item.id} className="flex gap-3 items-center p-3 bg-white/[0.03] border border-white/[0.06] rounded-xl">
                    <div className="w-14 h-14 rounded-lg overflow-hidden shrink-0 bg-slate-800">
                      <img src={item.image} alt={item.name} className="w-full h-full object-cover" onError={e => { e.target.src = FALLBACK_IMG; }} />
                    </div>
                    <div className="flex-grow min-w-0">
                      <p className="text-sm font-medium truncate">{item.name}</p>
                      <p className="text-xs text-slate-400 mt-0.5">{fmt(item.price, item.currency)} each</p>
                    </div>
                    <div className="flex items-center gap-1 shrink-0">
                      <button onClick={() => updateQty(item.id, -1)} className="w-6 h-6 flex items-center justify-center rounded-md bg-white/[0.06] hover:bg-white/10 transition-colors"><Minus className="w-3 h-3" /></button>
                      <span className="w-6 text-center text-sm font-bold">{item.qty}</span>
                      <button onClick={() => updateQty(item.id, 1)} className="w-6 h-6 flex items-center justify-center rounded-md bg-white/[0.06] hover:bg-white/10 transition-colors"><Plus className="w-3 h-3" /></button>
                    </div>
                  </div>
                ))}
              </div>

              {cart.length > 0 && (
                <div className="px-5 py-4 border-t border-white/[0.07] space-y-3">
                  <div className="flex items-center justify-between">
                    <span className="text-slate-400 text-sm">Total</span>
                    <span className="text-xl font-bold">₹{cartTotal.toLocaleString()}</span>
                  </div>
                  <button onClick={handleSelectCart} disabled={isLoading}
                    className="w-full bg-emerald-500 hover:bg-emerald-400 disabled:opacity-50 text-slate-900 py-3.5 rounded-xl font-bold text-sm flex items-center justify-center gap-2 transition-all active:scale-[0.98]">
                    {isLoading ? <Loader2 className="w-4 h-4 animate-spin" /> : <><span>Proceed to Checkout</span><ChevronRight className="w-4 h-4" /></>}
                  </button>
                </div>
              )}
            </motion.aside>
          </>
        )}
      </AnimatePresence>

      {/* ── Bottom progress bar (cart checkout / no modal) ────────────────── */}
      <AnimatePresence>
        {transaction && !selectedProduct && (
          <motion.div
            initial={{ y: 120, opacity: 0 }} animate={{ y: 0, opacity: 1 }} exit={{ y: 120, opacity: 0 }}
            className="fixed bottom-5 left-1/2 -translate-x-1/2 z-50 w-[calc(100%-2rem)] max-w-2xl"
          >
            <div className="glass rounded-2xl p-5 shadow-2xl border border-emerald-500/10">
              <div className="flex items-center justify-between mb-4">
                <div className="flex items-center gap-2">
                  <div className="w-2 h-2 rounded-full bg-emerald-400 pulse-dot" />
                  <span className="text-xs font-medium text-slate-300">Txn {transaction.id.slice(0, 8)}…</span>
                </div>
                <span className="text-[10px] font-bold uppercase tracking-widest bg-emerald-500/15 text-emerald-400 px-2.5 py-1 rounded-full border border-emerald-500/20">
                  {(transaction.status || '').replace('_', ' ')}
                </span>
              </div>

              {/* Stepper */}
              <div className="relative flex items-center justify-between mb-4 px-2">
                <div className="absolute inset-x-2 top-3 h-px bg-white/10" />
                <div className="absolute left-2 top-3 h-px bg-emerald-400 transition-all duration-700" style={{ width: `${(activeStep / 3) * 100}%` }} />
                {stepLabels.map((label, i) => (
                  <div key={i} className="relative flex flex-col items-center gap-1.5 z-10">
                    <div className={`w-6 h-6 rounded-full flex items-center justify-center text-[10px] font-bold transition-all duration-500 ${i < activeStep ? 'bg-emerald-400 text-slate-900' : i === activeStep ? 'bg-emerald-400 text-slate-900 ring-4 ring-emerald-400/20' : 'bg-slate-700 text-slate-400'}`}>
                      {i < activeStep ? <CheckCircle2 className="w-3.5 h-3.5" /> : i + 1}
                    </div>
                    <span className={`text-[10px] font-medium whitespace-nowrap ${i <= activeStep ? 'text-emerald-400' : 'text-slate-500'}`}>{label}</span>
                  </div>
                ))}
              </div>

              <div className="flex gap-3 justify-center">
                {transaction.status === 'QUOTE_RECEIVED' && (
                  <button onClick={handleInit} disabled={isLoading}
                    className="flex items-center gap-2 bg-emerald-500 hover:bg-emerald-400 text-slate-900 px-6 py-2.5 rounded-xl text-sm font-bold transition-all active:scale-95">
                    {isLoading ? <Loader2 className="w-4 h-4 animate-spin" /> : <><span>Initialize Order</span><ArrowRight className="w-4 h-4" /></>}
                  </button>
                )}
                {transaction.status === 'INIT_RECEIVED' && (
                  <button onClick={handleConfirm} disabled={isLoading}
                    className="flex items-center gap-2 bg-emerald-500 hover:bg-emerald-400 text-slate-900 px-6 py-2.5 rounded-xl text-sm font-bold transition-all active:scale-95">
                    {isLoading ? <Loader2 className="w-4 h-4 animate-spin" /> : <><span>Pay &amp; Confirm</span><CheckCircle2 className="w-4 h-4" /></>}
                  </button>
                )}
                {transaction.status === 'CONFIRMED' && (
                  <div className="flex items-center gap-2 text-emerald-400 font-bold">
                    <CheckCircle2 className="w-5 h-5" /> Order Confirmed!
                  </div>
                )}
              </div>
            </div>
          </motion.div>
        )}
      </AnimatePresence>
    </div>
  );
}
