import '@testing-library/jest-dom/vitest'
import { fireEvent, render, screen, within } from '@testing-library/react'
import { afterEach, test, vi } from 'vitest'
import { AlbumStrip } from './AlbumStrip'
import type { Postcard } from '../domain/travel'
import { postcardVisuals } from '../data/visualCatalog'

afterEach(() => {
  vi.restoreAllMocks()
})

test('renders repeated album destinations without duplicate-key warnings', () => {
  const consoleError = vi
    .spyOn(console, 'error')
    .mockImplementation(() => undefined)
  const repeatedAlbum: Postcard[] = [
    {
      id: 'willow-pond',
      title: '柳影池',
      body: '水面把云推向更远的地方。',
      alt: '柳树和池水边的旅行卡插画',
    },
    {
      id: 'willow-pond',
      title: '柳影池',
      body: '水面把云推向更远的地方。',
      alt: '柳树和池水边的旅行卡插画',
    },
  ]

  render(<AlbumStrip album={repeatedAlbum} />)

  const album = screen.getByRole('region', { name: '旅行卡' })
  expect(within(album).getAllByRole('listitem')).toHaveLength(2)
  expect(within(album).getAllByRole('heading', { name: '柳影池' })).toHaveLength(
    2,
  )

  const duplicateKeyWarnings = consoleError.mock.calls.filter((call) =>
    call.some((part) => String(part).includes('same key')),
  )
  expect(duplicateKeyWarnings).toHaveLength(0)
})

test('renders each postcard with its catalog thumbnail and a paper empty state', () => {
  const card: Postcard = {
    id: 'cloud-ridge',
    title: '云脊坡',
    body: '风从山脊带来晒过的松香。',
    alt: '云雾山脊上的旅行卡插画',
  }

  const { rerender } = render(<AlbumStrip album={[card]} />)

  expect(
    screen.getByRole('img', { name: postcardVisuals['cloud-ridge'].alt }),
  ).toHaveAttribute('src', postcardVisuals['cloud-ridge'].src)

  rerender(<AlbumStrip album={[]} />)

  expect(screen.getByText('第一段旅程会留在这里').parentElement).toHaveClass(
    'album-strip__empty',
  )
})

test('opens its postcard when a thumbnail card is activated', () => {
  const onOpen = vi.fn()
  const card: Postcard = {
    id: 'willow-pond',
    title: '柳影池',
    body: '水面把云推向更远的地方。',
    alt: '柳树和池水边的旅行卡插画',
  }

  render(<AlbumStrip album={[card]} onOpen={onOpen} />)

  fireEvent.click(screen.getByRole('button', { name: '查看旅行卡：柳影池' }))

  expect(onOpen).toHaveBeenCalledWith(card, expect.any(HTMLButtonElement))
})
