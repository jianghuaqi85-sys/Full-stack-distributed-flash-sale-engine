interface Props {
  collapsed?: boolean
}

// Pure text logo — no icons, no gradients
export default function BrandLogo({ collapsed = false }: Props) {
  return (
    <span style={{
      fontFamily: 'var(--font-mono)',
      fontSize: collapsed ? 14 : 18,
      fontWeight: 700,
      color: '#FFFFFF',
      letterSpacing: '-0.04em',
      lineHeight: 1,
    }}>
      {collapsed ? 'T' : 'TICKET'}
    </span>
  )
}
