import { useLocation, useNavigation } from 'react-router'

export function useIsPageLoading() {
  const location = useLocation()
  const navigation = useNavigation()

  if (navigation.state !== 'loading') {
    return false
  }

  return (
    navigation.location.pathname !== location.pathname ||
    navigation.location.search !== location.search
  )
}
