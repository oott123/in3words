import { useEffect } from 'react'

export function useScrollToTop(trigger: boolean) {
  useEffect(() => {
    if (trigger) {
      setTimeout(() => {
        window.scrollTo({ top: 0, behavior: 'smooth' })
      })
    }
  }, [trigger])
}
