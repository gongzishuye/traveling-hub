import type { RemoteJourneyEvent } from '../api/game'

type JourneyTimelineProps = {
  events: RemoteJourneyEvent[]
}

export function JourneyTimeline({ events }: JourneyTimelineProps) {
  return (
    <section className="journey-timeline" aria-labelledby="journey-timeline-title">
      <h2 id="journey-timeline-title">旅程记录</h2>
      <ol className="journey-timeline__roadbook" aria-live="polite">
        {events.map((event, index) => (
          <li className="journey-timeline__entry" key={`${event.stage}-${index}`}>
            <span className="journey-timeline__node" aria-hidden="true" />
            <p>{event.text}</p>
            <span className="journey-timeline__stamp">旅程 {index + 1}</span>
          </li>
        ))}
      </ol>
    </section>
  )
}
