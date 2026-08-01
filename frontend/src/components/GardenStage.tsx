import type { JourneyPhase } from '../domain/travel'
import { sceneVisuals } from '../data/visualCatalog'
import { SceneArt } from './SceneArt'

type GardenStageProps = {
  phase: JourneyPhase
}

const labels: Record<JourneyPhase, string> = {
  home: '旅人在小屋前休息',
  travelling: '旅人正在远行，桌前空着',
  returned: '旅人已经归来',
}

export function GardenStage({ phase }: GardenStageProps) {
  const scene = sceneVisuals[phase]
  const stateLabel = labels[phase]

  return (
    <section
      className={`garden-stage garden-stage--${phase}`}
      data-testid="garden-stage"
      data-scene-state={phase}
      aria-label={`庭院场景：${stateLabel}。${scene.description}`}
    >
      <SceneArt scene={scene} stateLabel={stateLabel} phase={phase} />
      <p className="garden-stage__state-label">当前状态：{stateLabel}</p>
    </section>
  )
}
