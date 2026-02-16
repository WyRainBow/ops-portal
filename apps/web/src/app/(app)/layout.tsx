'use client'

import { Protected } from '../../components/Protected'
import { Shell } from '../../components/Shell'

export default function AppLayout({ children }: { children: React.ReactNode }) {
  return (
    <Protected>
      <Shell>{children}</Shell>
    </Protected>
  )
}

