import type { JourneyTemplate } from '../domain/journeyCatalog'
import { journeyTemplatesByFood } from '../domain/journeyCatalog'
import type { Postcard } from '../domain/travel'
import type { GamePhase, GameSnapshot } from './client'

export type RemoteJourneyEvent = {
  stage: string
  text: string
}

export type RemoteGame = {
  frogId: string
  serverTime: string
  localDate: string
  phase: GamePhase
  journey: JourneyTemplate | null
  events: RemoteJourneyEvent[]
  album: Postcard[]
}

const allTemplates = Object.values(journeyTemplatesByFood).flat()
const templatesById = new Map(allTemplates.map((template) => [template.id, template]))
const postcardsById = new Map<string, Postcard>(
  allTemplates.map((template) => [template.postcard.id, template.postcard]),
)

export function toRemoteGame(snapshot: GameSnapshot): RemoteGame {
  const journey = snapshot.journey
    ? (templatesById.get(snapshot.journey.template_id) ?? null)
    : null
  if (snapshot.journey && !journey) {
    throw new Error('无法显示这段旅程，请稍后重试。')
  }
  const album = snapshot.album_postcard_ids.map((id) => {
    const postcard = postcardsById.get(id)
    if (!postcard) {
      throw new Error('无法显示这段旅程，请稍后重试。')
    }
    return postcard
  })
  return {
    frogId: snapshot.frog_id,
    serverTime: snapshot.server_time,
    localDate: snapshot.local_date,
    phase: snapshot.phase,
    journey,
    events: snapshot.events,
    album,
  }
}
