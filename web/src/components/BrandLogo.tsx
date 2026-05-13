interface Props {
  size?: number
  collapsed?: boolean
}

export default function BrandLogo({ size = 40, collapsed = false }: Props) {
  if (collapsed) {
    return (
      <svg width={size} height={size} viewBox="0 0 40 40" fill="none">
        <defs>
          <linearGradient id="logo-grad" x1="0" y1="0" x2="40" y2="40">
            <stop offset="0%" stopColor="#5B2FE8" />
            <stop offset="100%" stopColor="#D4A843" />
          </linearGradient>
        </defs>
        <rect x="4" y="8" width="32" height="24" rx="4" fill="url(#logo-grad)" />
        <circle cx="4" cy="20" r="4" fill="#0D0A1A" />
        <circle cx="36" cy="20" r="4" fill="#0D0A1A" />
        <path d="M14 15L18 20L14 25" stroke="white" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" />
        <path d="M22 15L26 20L22 25" stroke="white" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" opacity="0.6" />
      </svg>
    )
  }

  return (
    <div style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
      <svg width={size} height={size} viewBox="0 0 40 40" fill="none">
        <defs>
          <linearGradient id="logo-grad-full" x1="0" y1="0" x2="40" y2="40">
            <stop offset="0%" stopColor="#5B2FE8" />
            <stop offset="100%" stopColor="#D4A843" />
          </linearGradient>
        </defs>
        <rect x="4" y="8" width="32" height="24" rx="4" fill="url(#logo-grad-full)" />
        <circle cx="4" cy="20" r="4" fill="#0D0A1A" />
        <circle cx="36" cy="20" r="4" fill="#0D0A1A" />
        <path d="M14 15L18 20L14 25" stroke="white" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" />
        <path d="M22 15L26 20L22 25" stroke="white" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" opacity="0.6" />
      </svg>
    </div>
  )
}
