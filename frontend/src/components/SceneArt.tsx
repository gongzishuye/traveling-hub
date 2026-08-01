import type { JourneyPhase } from '../domain/travel'
import { worldDetails, type SceneVisual } from '../data/visualCatalog'
import { CrayfishSprite } from './CrayfishSprite'

type SceneArtProps = {
  scene: SceneVisual
  stateLabel: string
  phase: JourneyPhase
}

const layerNames = ['sky', 'middle', 'foreground'] as const
const ambientMotes = ['one', 'two', 'three', 'four', 'five', 'six'] as const

export function SceneArt({ scene, stateLabel, phase }: SceneArtProps) {
  return (
    <div className="scene-art">
      <div
        className="scene-art__layer scene-art__layer--sky"
        data-testid="scene-layer"
        data-scene-layer={layerNames[0]}
        aria-hidden="true"
      />
      <img
        className="scene-art__hero"
        src={scene.src}
        alt={`${scene.alt}。${stateLabel}`}
      />
      <div
        className="scene-art__layer scene-art__layer--middle"
        data-testid="scene-layer"
        data-scene-layer={layerNames[1]}
        aria-hidden="true"
      />
      <div className="scene-art__motes" aria-hidden="true">
        {ambientMotes.map((mote) => (
          <span
            key={mote}
            className={`scene-art__mote scene-art__mote--${mote}`}
            data-testid="ambient-mote"
          />
        ))}
      </div>
      <CrayfishSprite phase={phase} />
      {scene.detailIds.map((detailId, index) => {
        const detail = worldDetails[detailId]

        return (
          <div
            key={detailId}
            className={`scene-art__detail scene-art__detail--${index + 1}`}
            data-testid="scene-detail"
            data-detail-id={detailId}
          >
            <img src={detail.src} alt={detail.alt} />
          </div>
        )
      })}
      <div
        className="scene-art__layer scene-art__layer--foreground"
        data-testid="scene-layer"
        data-scene-layer={layerNames[2]}
        aria-hidden="true"
      />
    </div>
  )
}
