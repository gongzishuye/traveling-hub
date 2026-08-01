import { useCallback, useEffect, useRef, useState } from 'react'
import { AlbumStrip } from './components/AlbumStrip'
import { AuthPanel } from './components/AuthPanel'
import { CollapsiblePanel } from './components/CollapsiblePanel'
import { GardenStage } from './components/GardenStage'
import { JourneyTimeline } from './components/JourneyTimeline'
import { PostcardReveal } from './components/PostcardReveal'
import { TravelJournal } from './components/TravelJournal'
import { changePassword, getGameSnapshot, login } from './api/client'
import { hasHTTPStatus } from './api/errors'
import { toRemoteGame, type RemoteGame } from './api/game'
import type { Postcard } from './domain/travel'
import { preloadSceneHeroes } from './data/visualCatalog'

type PanelId = 'journal' | 'timeline' | 'album'
type Screen = 'loading' | 'login' | 'change-password' | 'game' | 'error'

const POLL_INTERVAL_MS = 60_000

function App() {
  const [screen, setScreen] = useState<Screen>('loading')
  const [game, setGame] = useState<RemoteGame | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [isPostcardOpen, setIsPostcardOpen] = useState(false)
  const [activePostcard, setActivePostcard] = useState<Postcard | null>(null)
  const [expandedPanels, setExpandedPanels] = useState<Record<PanelId, boolean>>({
    journal: false,
    timeline: false,
    album: false,
  })
  const shouldRestorePostcardFocus = useRef(false)
  const postcardOpener = useRef<HTMLButtonElement>(null)
  const albumToggle = useRef<HTMLButtonElement>(null)

  useEffect(() => {
    preloadSceneHeroes()
  }, [])

  const loadGame = useCallback(async () => {
    try {
      const snapshot = await getGameSnapshot()
      setGame(toRemoteGame(snapshot))
      setError(null)
      setScreen('game')
    } catch (requestError) {
      if (hasHTTPStatus(requestError, 401)) {
        setScreen('login')
        return
      }
      if (hasHTTPStatus(requestError, 403)) {
        setScreen('change-password')
        return
      }
      setError(requestError instanceof Error ? requestError.message : '暂时无法连接远行小屋，请稍后重试。')
      setScreen('error')
    }
  }, [])

  useEffect(() => {
    void loadGame()
  }, [loadGame])

  useEffect(() => {
    if (screen !== 'game') {
      return
    }

    const intervalId = window.setInterval(() => {
      void loadGame()
    }, POLL_INTERVAL_MS)
    return () => window.clearInterval(intervalId)
  }, [loadGame, screen])

  useEffect(() => {
    if (shouldRestorePostcardFocus.current && !isPostcardOpen) {
      const focusTarget = postcardOpener.current ?? albumToggle.current
      focusTarget?.focus()
      shouldRestorePostcardFocus.current = false
    }
  }, [isPostcardOpen])

  async function handleLogin(email: string, password: string) {
    try {
      const result = await login(email, password)
      if (result.must_change_password) {
        setScreen('change-password')
        return
      }
      await loadGame()
    } catch (requestError) {
      setError(requestError instanceof Error ? requestError.message : '登录失败，请稍后重试。')
    }
  }

  async function handlePasswordChange(newPassword: string) {
    try {
      await changePassword(newPassword)
      await loadGame()
    } catch (requestError) {
      setError(requestError instanceof Error ? requestError.message : '更新密码失败，请稍后重试。')
    }
  }

  function handleClosePostcard() {
    shouldRestorePostcardFocus.current = true
    setIsPostcardOpen(false)
  }

  function handleOpenPostcard(card: Postcard, opener?: HTMLButtonElement) {
    if (opener) {
      postcardOpener.current = opener
    }
    setActivePostcard(card)
    setIsPostcardOpen(true)
  }

  function togglePanel(panel: PanelId) {
    setExpandedPanels((currentPanels) => ({
      ...currentPanels,
      [panel]: !currentPanels[panel],
    }))
  }

  if (screen === 'loading') {
    return <AuthPanel mode="loading" />
  }

  if (screen === 'login') {
    return <AuthPanel mode="login" error={error} onLogin={handleLogin} />
  }

  if (screen === 'change-password') {
    return <AuthPanel mode="change-password" error={error} onChangePassword={handlePasswordChange} />
  }

  if (screen === 'error') {
    return <AuthPanel mode="error" error={error} onRetry={loadGame} />
  }

  if (!game) {
    return <AuthPanel mode="error" error="暂时无法连接远行小屋，请稍后重试。" onRetry={loadGame} />
  }

  return (
    <main inert={isPostcardOpen ? true : undefined}>
      <h1>远行小屋</h1>
      <div className="single-scene-layout__journal">
        <CollapsiblePanel
          id="travel-journal"
          title="旅行手账"
          expanded={expandedPanels.journal}
          onToggle={() => togglePanel('journal')}
          className="collapsible-panel--journal"
        >
          <TravelJournal phase={game.phase} journey={game.journey} localDate={game.localDate} />
        </CollapsiblePanel>
      </div>
      <GardenStage phase={game.phase} />
      <aside className="single-scene-layout__echoes" aria-label="旅程回响">
        <CollapsiblePanel
          id="journey-timeline"
          title="旅程记录"
          expanded={expandedPanels.timeline}
          onToggle={() => togglePanel('timeline')}
          className="collapsible-panel--timeline"
        >
          <JourneyTimeline events={game.events} />
        </CollapsiblePanel>
        <CollapsiblePanel
          id="travel-album"
          title="旅行卡"
          expanded={expandedPanels.album}
          onToggle={() => togglePanel('album')}
          className="collapsible-panel--album"
          toggleRef={albumToggle}
        >
          <AlbumStrip
            album={game.album}
            onOpen={handleOpenPostcard}
            latestOpenButtonRef={postcardOpener}
          />
        </CollapsiblePanel>
      </aside>
      {isPostcardOpen && activePostcard ? (
        <PostcardReveal card={activePostcard} onClose={handleClosePostcard} />
      ) : null}
      <p className="desktop-only-message">请在桌面浏览器打开</p>
    </main>
  )
}

export default App
