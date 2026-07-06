import { OAuth2ResponseType, OAuth2Scope } from '@traptitech/traq'

import { BASE_PATH } from '/@/lib/apis'
import { getGazerOAuthClient } from '/@/lib/internalApi'
import { constructSettingsPath } from '/@/router/settings'

const GAZER_OAUTH_STATE_STORAGE_KEY = 'hw-traq:gazer-oauth-state'

const getCallbackURL = () =>
  `${location.origin}${constructSettingsPath('settingsGazer')}`

const generateState = () => {
  const bytes = new Uint8Array(16)
  crypto.getRandomValues(bytes)
  return Array.from(bytes, byte => byte.toString(16).padStart(2, '0')).join('')
}

export const startGazerTokenIssue = async () => {
  const { clientId } = await getGazerOAuthClient()
  const state = generateState()
  sessionStorage.setItem(GAZER_OAUTH_STATE_STORAGE_KEY, state)

  const url = new URL(`${BASE_PATH}/oauth2/authorize`, location.origin)
  url.searchParams.set('client_id', clientId)
  url.searchParams.set('response_type', OAuth2ResponseType.Token)
  url.searchParams.set('redirect_uri', getCallbackURL())
  url.searchParams.set('scope', OAuth2Scope.Read)
  url.searchParams.set('state', state)
  location.assign(url.toString())
}

export const consumeGazerTokenCallback = () => {
  if (!location.hash) return undefined

  const params = new URLSearchParams(location.hash.slice(1))
  const accessToken = params.get('access_token')
  const state = params.get('state')
  const error = params.get('error')
  if (!accessToken && !error) return undefined

  history.replaceState(
    null,
    document.title,
    location.pathname + location.search
  )

  const expectedState = sessionStorage.getItem(GAZER_OAUTH_STATE_STORAGE_KEY)
  sessionStorage.removeItem(GAZER_OAUTH_STATE_STORAGE_KEY)
  if (error) {
    throw new Error(error)
  }
  if (!accessToken || !state || state !== expectedState) {
    throw new Error('invalid gazer token callback')
  }
  return accessToken
}
