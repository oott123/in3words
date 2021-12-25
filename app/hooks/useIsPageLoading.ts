import { useLocation, useTransition } from 'remix'

export function useIsPageLoading() {
  const location = useLocation()
  const transition = useTransition()

  if (transition.state !== 'loading') {
    return false
  }

  return (
    transition.location.pathname !== location.pathname ||
    transition.location.search !== location.search
  )
}
