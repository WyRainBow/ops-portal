'use client'

export function Card(props: {
  title?: string
  subtitle?: string
  right?: React.ReactNode
  children?: React.ReactNode
  className?: string
}) {
  return (
    <section className={['ops-card rounded-3xl p-5', props.className || ''].join(' ')}>
      {(props.title || props.subtitle || props.right) && (
        <header className="flex items-start justify-between gap-4">
          <div className="min-w-0">
            {props.title && <div className="truncate text-base font-semibold">{props.title}</div>}
            {props.subtitle && <div className="mt-1 text-xs text-[color:var(--muted)]">{props.subtitle}</div>}
          </div>
          {props.right ? <div className="shrink-0">{props.right}</div> : null}
        </header>
      )}
      {props.children ? <div className="mt-4">{props.children}</div> : null}
    </section>
  )
}

export function Badge(props: { tone?: 'ok' | 'warn' | 'bad' | 'neutral'; children: React.ReactNode }) {
  const tone = props.tone || 'neutral'
  const cls =
    tone === 'ok'
      ? 'border-emerald-200 bg-emerald-50 text-emerald-800'
      : tone === 'warn'
        ? 'border-amber-200 bg-amber-50 text-amber-800'
        : tone === 'bad'
          ? 'border-rose-200 bg-rose-50 text-rose-800'
          : 'border-slate-200 bg-slate-50 text-slate-700'
  const dot =
    tone === 'ok'
      ? 'bg-emerald-600'
      : tone === 'warn'
        ? 'bg-amber-600'
        : tone === 'bad'
          ? 'bg-rose-600'
          : 'bg-slate-400'
  return (
    <span className={['inline-flex items-center gap-2 rounded-full border px-2 py-0.5 text-xs', cls].join(' ')}>
      <span className={['h-1.5 w-1.5 rounded-full', dot].join(' ')} />
      <span>{props.children}</span>
    </span>
  )
}

export function Button(props: React.ButtonHTMLAttributes<HTMLButtonElement> & { tone?: 'primary' | 'ghost' | 'danger' }) {
  const tone = props.tone || 'ghost'
  const base =
    'rounded-2xl px-4 py-2 text-sm transition disabled:opacity-50 disabled:cursor-not-allowed focus-visible:outline-none focus-visible:ring-4 focus-visible:ring-slate-200'
  const cls =
    tone === 'primary'
      ? 'border border-slate-900 bg-slate-900 text-white hover:bg-slate-700'
      : tone === 'danger'
        ? 'border border-rose-300 bg-rose-50 text-rose-700 hover:bg-rose-100'
        : 'border border-slate-300 bg-white text-slate-900 hover:bg-slate-100'
  return (
    <button {...props} className={[base, cls, props.className || ''].join(' ')}>
      {props.children}
    </button>
  )
}

export function Input(props: React.InputHTMLAttributes<HTMLInputElement>) {
  return (
    <input
      {...props}
      className={[
        'ops-input w-full rounded-2xl px-3 py-2 text-sm text-slate-900 placeholder:text-slate-400 outline-none',
        props.className || '',
      ].join(' ')}
    />
  )
}

export function TextArea(props: React.TextareaHTMLAttributes<HTMLTextAreaElement>) {
  return (
    <textarea
      {...props}
      className={[
        'ops-input w-full rounded-2xl px-3 py-2 text-sm text-slate-900 placeholder:text-slate-400 outline-none',
        props.className || '',
      ].join(' ')}
    />
  )
}

export function Select(props: React.SelectHTMLAttributes<HTMLSelectElement>) {
  return (
    <select
      {...props}
      className={[
        'ops-input w-full appearance-none rounded-2xl px-3 py-2 text-sm text-slate-900 outline-none',
        props.className || '',
      ].join(' ')}
    />
  )
}
