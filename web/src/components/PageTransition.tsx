import { useEffect, useRef } from 'react'

interface Props {
  children: React.ReactNode
}

export default function PageTransition({ children }: Props) {
  const ref = useRef<HTMLDivElement>(null)
  useEffect(() => {
    ref.current?.classList.add('page-enter')
  }, [])
  return <div ref={ref}>{children}</div>
}
