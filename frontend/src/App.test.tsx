import '@testing-library/jest-dom/vitest'
import { fireEvent, render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'
import App from './App'
import { hasHTTPStatus } from './api/errors'

const travellingGame = {
  frog_id: 'frog-1',
  server_time: '2026-08-01T09:30:00+08:00',
  local_date: '2026-08-01',
  phase: 'travelling',
  journey: {
    template_id: 'willow-pond-reed',
    departed_at: '2026-08-01T08:00:00+08:00',
  },
  events: [{ stage: 'departed', text: '旅人出发了。' }],
  album_postcard_ids: ['cloud-ridge'],
}

function jsonResponse(body: unknown, status = 200) {
  return new Response(JSON.stringify(body), {
    status,
    headers: { 'Content-Type': 'application/json' },
  })
}

afterEach(() => {
  vi.unstubAllGlobals()
  vi.restoreAllMocks()
})

describe('App', () => {
  it('recognizes an unauthorized status even when its error object comes from another runtime', () => {
    expect(hasHTTPStatus({ status: 401 }, 401)).toBe(true)
  })

  it('renders the server-owned journey and removes local departure controls', async () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue(jsonResponse(travellingGame)))

    render(<App />)

    expect(await screen.findByTestId('garden-stage')).toHaveAttribute(
      'data-scene-state',
      'travelling',
    )
    fireEvent.click(screen.getByRole('button', { name: '展开旅行手账' }))
    expect(screen.getByText('旅人已经出发，远行小屋会在归来后更新记录。')).toBeVisible()
    expect(screen.queryByRole('radio')).not.toBeInTheDocument()
    expect(screen.queryByRole('button', { name: '出发' })).not.toBeInTheDocument()

    fireEvent.click(screen.getByRole('button', { name: '展开旅程记录' }))
    expect(screen.getByRole('region', { name: '旅程记录' })).toHaveTextContent(
      '旅人出发了。',
    )
  })

  it('shows the login form when the game snapshot is unauthenticated', async () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue(new Response(null, { status: 401 })))

    render(<App />)

    expect(await screen.findByRole('heading', { name: '登录远行小屋' })).toBeVisible()
    expect(screen.getByLabelText('邮箱')).toBeVisible()
    expect(screen.getByLabelText('密码')).toBeVisible()
  })

  it('shows the password change form when the existing session is restricted', async () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue(new Response(null, { status: 403 })))

    render(<App />)

    expect(await screen.findByRole('heading', { name: '设置新密码' })).toBeVisible()
  })

  it('loads the game after a successful login', async () => {
    const user = userEvent.setup()
    const fetchMock = vi
      .fn()
      .mockResolvedValueOnce(new Response(null, { status: 401 }))
      .mockResolvedValueOnce(jsonResponse({ must_change_password: false }))
      .mockResolvedValueOnce(jsonResponse(travellingGame))
    vi.stubGlobal('fetch', fetchMock)

    render(<App />)

    await user.type(await screen.findByLabelText('邮箱'), 'agent@example.com')
    await user.type(screen.getByLabelText('密码'), 'initial-password')
    await user.click(screen.getByRole('button', { name: '登录' }))

    expect(await screen.findByTestId('garden-stage')).toHaveAttribute(
      'data-scene-state',
      'travelling',
    )
    expect(fetchMock).toHaveBeenNthCalledWith(
      2,
      '/v1/web/login',
      expect.objectContaining({ method: 'POST', credentials: 'include' }),
    )
  })

  it('requires a password change before displaying the game', async () => {
    const user = userEvent.setup()
    const fetchMock = vi
      .fn()
      .mockResolvedValueOnce(new Response(null, { status: 401 }))
      .mockResolvedValueOnce(jsonResponse({ must_change_password: true }))
      .mockResolvedValueOnce(new Response(null, { status: 204 }))
      .mockResolvedValueOnce(jsonResponse(travellingGame))
    vi.stubGlobal('fetch', fetchMock)

    render(<App />)

    await user.type(await screen.findByLabelText('邮箱'), 'agent@example.com')
    await user.type(screen.getByLabelText('密码'), 'initial-password')
    await user.click(screen.getByRole('button', { name: '登录' }))

    expect(await screen.findByRole('heading', { name: '设置新密码' })).toBeVisible()
    await user.type(screen.getByLabelText('新密码'), 'a-new-secure-password')
    await user.click(screen.getByRole('button', { name: '更新密码' }))

    expect(await screen.findByTestId('garden-stage')).toHaveAttribute(
      'data-scene-state',
      'travelling',
    )
    expect(fetchMock).toHaveBeenNthCalledWith(
      3,
      '/v1/web/change-password',
      expect.objectContaining({ method: 'POST', credentials: 'include' }),
    )
  })

  it('shows a retryable error when the authenticated game request fails', async () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue(new Response(null, { status: 500 })))

    render(<App />)

    expect(
      await screen.findByText('暂时无法连接远行小屋，请稍后重试。'),
    ).toBeVisible()
    expect(screen.getByRole('button', { name: '重试' })).toBeVisible()
  })

  it('opens a server-provided postcard from the album without a restart action', async () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue(jsonResponse(travellingGame)))
    render(<App />)

    await screen.findByTestId('garden-stage')
    fireEvent.click(screen.getByRole('button', { name: '展开旅行卡' }))
    fireEvent.click(screen.getByRole('button', { name: '查看旅行卡：云脊坡' }))

    const dialog = screen.getByRole('dialog', { name: '旅行卡：云脊坡' })
    expect(within(dialog).getByRole('button', { name: '关闭旅行卡' })).toBeVisible()
    expect(within(dialog).queryByRole('button', { name: '再次出发' })).not.toBeInTheDocument()
  })
})
