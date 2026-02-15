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
      ? 'border-emerald-300/20 bg-emerald-400/10 text-emerald-200'
      : tone === 'warn'
        ? 'border-amber-300/20 bg-amber-400/10 text-amber-100'
        : tone === 'bad'
          ? 'border-rose-300/20 bg-rose-400/10 text-rose-100'
          : 'border-white/15 bg-white/5 text-slate-100/90'
  return <span className={['inline-flex items-center rounded-full border px-2 py-0.5 text-xs', cls].join(' ')}>{props.children}</span>
}

export function Button(props: React.ButtonHTMLAttributes<HTMLButtonElement> & { tone?: 'primary' | 'ghost' | 'danger' }) {
  const tone = props.tone || 'ghost'
  const base = 'rounded-2xl px-4 py-2 text-sm transition disabled:opacity-50 disabled:cursor-not-allowed'
  const cls =
    tone === 'primary'
      ? 'border border-white/20 bg-white/10 hover:bg-white/15'
      : tone === 'danger'
        ? 'border border-rose-300/20 bg-rose-500/10 text-rose-100 hover:bg-rose-500/15'
        : 'border border-[color:var(--stroke)] bg-black/20 hover:bg-black/30'
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
        'ops-input w-full rounded-2xl px-3 py-2 text-sm text-slate-100 placeholder:text-slate-300/40 outline-none focus:border-white/30',
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
        'ops-input w-full rounded-2xl px-3 py-2 text-sm text-slate-100 placeholder:text-slate-300/40 outline-none focus:border-white/30',
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
        'ops-input w-full appearance-none rounded-2xl px-3 py-2 text-sm text-slate-100 outline-none focus:border-white/30',
        props.className || '',
      ].join(' ')}
    />
  )
}

