import type { JourneyPhase } from '../domain/travel'
import crayfishHome from '../assets/characters/crayfish-home-seated-stool-blink.webp'

type CrayfishSpriteProps = {
  phase: JourneyPhase
}

export function CrayfishSprite({ phase }: CrayfishSpriteProps) {
  if (phase !== 'home') {
    return null
  }

  return (
    <div
      className="crayfish-sprite crayfish-sprite--home"
      data-testid="crayfish-sprite"
      aria-hidden="true"
    >
      <img src={crayfishHome} alt="" />
    </div>
  )
}
