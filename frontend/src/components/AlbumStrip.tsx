import type { RefObject } from 'react'
import type { Postcard } from '../domain/travel'
import { postcardVisuals } from '../data/visualCatalog'

type AlbumStripProps = {
  album: Postcard[]
  onOpen?(card: Postcard, opener: HTMLButtonElement): void
  latestOpenButtonRef?: RefObject<HTMLButtonElement | null>
}

export function AlbumStrip({
  album,
  onOpen,
  latestOpenButtonRef,
}: AlbumStripProps) {
  return (
    <section className="album-strip" aria-labelledby="album-strip-title">
      <h2 id="album-strip-title">旅行卡</h2>
      {album.length === 0 ? (
        <div className="album-strip__empty">
          <p>第一段旅程会留在这里</p>
        </div>
      ) : (
        <ol className="album-strip__cards">
          {album.map((postcard, index) => {
            const visual = postcardVisuals[postcard.id]

            return (
              <li key={`${postcard.id}-${index}`}>
                <button
                  type="button"
                  className="album-strip__card"
                  aria-label={`查看旅行卡：${postcard.title}`}
                  onClick={(event) => onOpen?.(postcard, event.currentTarget)}
                  ref={index === album.length - 1 ? latestOpenButtonRef : undefined}
                >
                  <img src={visual.src} alt={visual.alt} />
                  <div>
                    <h3>{postcard.title}</h3>
                    <p>{postcard.body}</p>
                  </div>
                </button>
              </li>
            )
          })}
        </ol>
      )}
    </section>
  )
}
