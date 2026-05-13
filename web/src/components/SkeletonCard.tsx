interface Props {
  variant?: 'card' | 'list' | 'stat'
  count?: number
}

export default function SkeletonCard({ variant = 'card', count = 1 }: Props) {
  return (
    <div style={{ display: 'flex', gap: 16, flexWrap: 'wrap' }}>
      {Array.from({ length: count }).map((_, i) => (
        <div key={i} style={{ flex: '1 1 250px', maxWidth: variant === 'card' ? 300 : '100%' }}>
          {variant === 'card' && (
            <div style={{ borderRadius: 16, overflow: 'hidden', background: 'var(--color-bg-container)' }}>
              <div className="skeleton" style={{ height: 160, borderRadius: 0 }} />
              <div style={{ padding: 16 }}>
                <div className="skeleton" style={{ height: 20, width: '80%', marginBottom: 8 }} />
                <div className="skeleton" style={{ height: 14, width: '60%', marginBottom: 8 }} />
                <div className="skeleton" style={{ height: 14, width: '40%' }} />
              </div>
            </div>
          )}
          {variant === 'list' && (
            <div style={{ display: 'flex', gap: 12, padding: '12px 0', borderBottom: '1px solid var(--color-border-light)' }}>
              <div className="skeleton" style={{ width: 48, height: 48, borderRadius: 8, flexShrink: 0 }} />
              <div style={{ flex: 1 }}>
                <div className="skeleton" style={{ height: 16, width: '70%', marginBottom: 8 }} />
                <div className="skeleton" style={{ height: 12, width: '50%' }} />
              </div>
            </div>
          )}
          {variant === 'stat' && (
            <div style={{ padding: 20, borderRadius: 16, background: 'var(--color-bg-container)' }}>
              <div className="skeleton" style={{ height: 14, width: '50%', marginBottom: 12 }} />
              <div className="skeleton" style={{ height: 28, width: '40%', marginBottom: 8 }} />
              <div className="skeleton" style={{ height: 12, width: '60%' }} />
            </div>
          )}
        </div>
      ))}
    </div>
  )
}
