'use client'

import { useEffect } from 'react'
import { useRouter } from 'next/navigation'
import { getToken } from '../lib/auth'

export function Protected({ children }: { children: React.ReactNode }) {
  const router = useRouter()
  useEffect(() => {
    const t = getToken()
    if (!t) router.replace('/login')
  }, [router])
  return <>{children}</>
}

