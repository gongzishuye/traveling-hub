import type { JourneyTemplate } from '../domain/journeyCatalog'
import type { GamePhase } from '../api/client'

type TravelJournalProps = {
  phase: GamePhase
  journey: JourneyTemplate | null
  localDate: string
}

const messages: Record<GamePhase, string> = {
  home: '旅人正在小屋休息，下一次远行将由世界的节律决定。',
  travelling: '旅人已经出发，远行小屋会在归来后更新记录。',
  returned: '旅人已经回家，新的旅行卡已放进相册。',
}

export function TravelJournal({ phase, journey, localDate }: TravelJournalProps) {
  return (
    <section className="travel-journal" aria-labelledby="travel-journal-title">
      <h2 id="travel-journal-title">旅行手账</h2>
      <section className="travel-journal__summary" aria-live="polite">
        <h3>{localDate} 的手账</h3>
        <p>{messages[phase]}</p>
        {journey ? <p>本次目的地：{journey.postcard.title}</p> : null}
      </section>
    </section>
  )
}
