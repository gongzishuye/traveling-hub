import { describe, expect, it, vi } from 'vitest'
import { ApiError, changePassword, getGameSnapshot } from './client'

describe('getGameSnapshot', () => {
  it('requests the authenticated game snapshot with cookies', async () => {
    const fetchMock = vi.fn().mockResolvedValue(
      new Response(
        JSON.stringify({
          frog_id: 'frog-1',
          server_time: '2026-08-01T08:01:00+08:00',
          local_date: '2026-08-01',
          phase: 'travelling',
          journey: {
            template_id: 'willow-pond-reed',
            departed_at: '2026-08-01T08:00:00+08:00',
          },
          events: [{ stage: 'departed', text: '旅人出发了。' }],
          album_postcard_ids: ['willow-pond'],
        }),
        { status: 200 },
      ),
    )
    vi.stubGlobal('fetch', fetchMock)

    await expect(getGameSnapshot()).resolves.toMatchObject({
      phase: 'travelling',
      journey: { template_id: 'willow-pond-reed' },
    })

    expect(fetchMock).toHaveBeenCalledWith('/v1/game', {
      credentials: 'include',
      headers: { Accept: 'application/json' },
    })
  })

  it('reports an unauthenticated response without attempting to parse a snapshot', async () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue(new Response(null, { status: 401 })))

    await expect(getGameSnapshot()).rejects.toEqual(
      new ApiError(401, '请先登录。'),
    )
  })
})

describe('changePassword', () => {
  it('uses the web session endpoint contract without retaining the initial password', async () => {
    const fetchMock = vi.fn().mockResolvedValue(new Response(null, { status: 204 }))
    vi.stubGlobal('fetch', fetchMock)

    await changePassword('a-new-secure-password')

    expect(fetchMock).toHaveBeenCalledWith('/v1/web/change-password', {
      method: 'POST',
      credentials: 'include',
      headers: { Accept: 'application/json', 'Content-Type': 'application/json' },
      body: JSON.stringify({ password: 'a-new-secure-password' }),
    })
  })
})
