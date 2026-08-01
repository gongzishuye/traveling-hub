import { describe, expect, it } from 'vitest'
import type { GameSnapshot } from './client'
import { toRemoteGame } from './game'

const travellingSnapshot: GameSnapshot = {
  frog_id: 'frog-1',
  server_time: '2026-08-01T09:30:00+08:00',
  local_date: '2026-08-01',
  phase: 'travelling',
  journey: {
    template_id: 'willow-pond-reed',
    departed_at: '2026-08-01T08:00:00+08:00',
  },
  events: [
    { stage: 'departed', text: '旅人出发了。' },
    { stage: 'midway', text: '芦苇在浅水边给旅人让出一条小路。' },
  ],
  album_postcard_ids: ['cloud-ridge'],
}

describe('toRemoteGame', () => {
  it('maps backend events and postcard ids to the existing visual catalogue', () => {
    expect(toRemoteGame(travellingSnapshot)).toMatchObject({
      phase: 'travelling',
      events: [
        { stage: 'departed', text: '旅人出发了。' },
        { stage: 'midway', text: '芦苇在浅水边给旅人让出一条小路。' },
      ],
      album: [{ id: 'cloud-ridge', title: '云脊坡' }],
      journey: { id: 'willow-pond-reed' },
    })
  })

  it('rejects a backend snapshot with unknown visual content', () => {
    expect(() => toRemoteGame({
      ...travellingSnapshot,
      journey: { template_id: 'unknown-template', departed_at: '2026-08-01T08:00:00+08:00' },
    })).toThrow('无法显示这段旅程')
    expect(() => toRemoteGame({ ...travellingSnapshot, album_postcard_ids: ['unknown-card'] })).toThrow('无法显示这段旅程')
  })
})
