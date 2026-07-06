import { OAuth2ResponseType, OAuth2Scope } from '@traptitech/traq'

import { BASE_PATH } from '/@/lib/apis'
import { getGazerOAuthClient } from '/@/lib/internalApi'
import { constructSettingsPath } from '/@/router/settings'

const GAZER_OAUTH_STATE_STORAGE_KEY = 'hw-traq:gazer-oauth-state'
const GAZER_OAUTH_CODE_VERIFIER_STORAGE_KEY =
  'hw-traq:gazer-oauth-code-verifier'
const codeChallengeMethod = 'S256'

export type GazerOAuthCallback = {
  code: string
  codeVerifier: string
  redirectUri: string
}

const getCallbackURL = () =>
  `${location.origin}${constructSettingsPath('settingsGazer')}`

const encodeBase64URL = (bytes: Uint8Array) =>
  btoa(String.fromCharCode(...bytes))
    .replaceAll('+', '-')
    .replaceAll('/', '_')
    .replace(/=+$/, '')

const generateRandomBase64URL = (length: number) => {
  const bytes = new Uint8Array(length)
  crypto.getRandomValues(bytes)
  return encodeBase64URL(bytes)
}

const createCodeChallenge = async (codeVerifier: string) => {
  const digest = await crypto.subtle.digest(
    'SHA-256',
    new TextEncoder().encode(codeVerifier)
  )
  return encodeBase64URL(new Uint8Array(digest))
}

export const buildGazerOAuthAuthorizeURL = async (
  clientId: string,
  state: string,
  codeVerifier: string
) => {
  const codeChallenge = await createCodeChallenge(codeVerifier)
  const url = new URL(`${BASE_PATH}/oauth2/authorize`, location.origin)
  url.searchParams.set('client_id', clientId)
  url.searchParams.set('response_type', OAuth2ResponseType.Code)
  url.searchParams.set('redirect_uri', getCallbackURL())
  url.searchParams.set('scope', OAuth2Scope.Read)
  url.searchParams.set('state', state)
  url.searchParams.set('code_challenge', codeChallenge)
  url.searchParams.set('code_challenge_method', codeChallengeMethod)
  return url.toString()
}

export const startGazerTokenIssue = async () => {
  const { clientId } = await getGazerOAuthClient()
  const state = generateRandomBase64URL(16)
  const codeVerifier = generateRandomBase64URL(32)
  sessionStorage.setItem(GAZER_OAUTH_STATE_STORAGE_KEY, state)
  sessionStorage.setItem(GAZER_OAUTH_CODE_VERIFIER_STORAGE_KEY, codeVerifier)

  location.assign(
    await buildGazerOAuthAuthorizeURL(clientId, state, codeVerifier)
  )
}

export const consumeGazerTokenCallback = () => {
  const params = new URLSearchParams(location.search)
  const code = params.get('code')
  const state = params.get('state')
  const error = params.get('error')
  if (!code && !error) return undefined

  const redirectUri = getCallbackURL()

  history.replaceState(null, document.title, location.pathname + location.hash)

  const expectedState = sessionStorage.getItem(GAZER_OAUTH_STATE_STORAGE_KEY)
  const codeVerifier = sessionStorage.getItem(
    GAZER_OAUTH_CODE_VERIFIER_STORAGE_KEY
  )
  sessionStorage.removeItem(GAZER_OAUTH_STATE_STORAGE_KEY)
  sessionStorage.removeItem(GAZER_OAUTH_CODE_VERIFIER_STORAGE_KEY)
  if (error) {
    throw new Error(error)
  }
  if (!code || !state || state !== expectedState || !codeVerifier) {
    throw new Error('invalid gazer token callback')
  }
  return {
    code,
    codeVerifier,
    redirectUri
  }
}
