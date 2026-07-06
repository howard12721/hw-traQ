import { OAuth2ResponseType } from '@traptitech/traq'

import {
  buildGazerOAuthAuthorizeURL,
  consumeGazerTokenCallback
} from '/@/lib/gazerOAuth'

describe('gazerOAuth', () => {
  beforeEach(() => {
    sessionStorage.clear()
    history.replaceState(null, '', '/')
  })

  afterEach(() => {
    vi.restoreAllMocks()
  })

  it('builds authorization code flow url with pkce', async () => {
    const url = new URL(
      await buildGazerOAuthAuthorizeURL('client-id', 'state', 'code-verifier')
    )
    expect(url.pathname).toBe('/api/v3/oauth2/authorize')
    expect(url.searchParams.get('client_id')).toBe('client-id')
    expect(url.searchParams.get('response_type')).toBe(OAuth2ResponseType.Code)
    expect(url.searchParams.get('redirect_uri')).toBe(
      `${location.origin}/settings/gazer`
    )
    expect(url.searchParams.get('state')).toBe('state')
    expect(url.searchParams.get('code_challenge')).toBeTruthy()
    expect(url.searchParams.get('code_challenge_method')).toBe('S256')
  })

  it('consumes authorization code callback', () => {
    sessionStorage.setItem('hw-traq:gazer-oauth-state', 'state')
    sessionStorage.setItem('hw-traq:gazer-oauth-code-verifier', 'verifier')
    history.replaceState(
      null,
      '',
      '/settings/gazer?code=oauth-code&state=state'
    )

    expect(consumeGazerTokenCallback()).toEqual({
      code: 'oauth-code',
      codeVerifier: 'verifier',
      redirectUri: `${location.origin}/settings/gazer`
    })
    expect(location.href).toBe(`${location.origin}/settings/gazer`)
  })

  it('rejects oauth callback errors', () => {
    sessionStorage.setItem('hw-traq:gazer-oauth-state', 'state')
    sessionStorage.setItem('hw-traq:gazer-oauth-code-verifier', 'verifier')
    history.replaceState(
      null,
      '',
      '/settings/gazer?error=unsupported_response_type&state=state'
    )

    expect(() => consumeGazerTokenCallback()).toThrow(
      'unsupported_response_type'
    )
    expect(location.href).toBe(`${location.origin}/settings/gazer`)
  })
})
