import { useEffect, useRef } from 'react'
import { createPortal } from 'react-dom'
import { postcardVisuals } from '../data/visualCatalog'
import type { Postcard } from '../domain/travel'

type PostcardRevealProps = {
  card: Postcard
  onClose(): void
}

export function PostcardReveal({ card, onClose }: PostcardRevealProps) {
  const closeButton = useRef<HTMLButtonElement>(null)
  const visual = postcardVisuals[card.id]

  useEffect(() => {
    closeButton.current?.focus()
    function handleKeyDown(event: KeyboardEvent) {
      if (event.key === 'Escape') onClose()
    }
    window.addEventListener('keydown', handleKeyDown)
    return () => window.removeEventListener('keydown', handleKeyDown)
  }, [onClose])

  return createPortal(
    <div className="postcard-reveal__backdrop">
      <section className="postcard-reveal__panel" role="dialog" aria-modal="true" aria-label={`旅行卡：${card.title}`}>
        <img src={visual.src} alt={visual.alt} />
        <h2>{card.title}</h2>
        <p>{card.body}</p>
        <button type="button" ref={closeButton} onClick={onClose}>关闭旅行卡</button>
      </section>
    </div>,
    document.body,
  )
}
