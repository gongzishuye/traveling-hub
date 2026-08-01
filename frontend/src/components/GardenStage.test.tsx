import '@testing-library/jest-dom/vitest'
import { render, screen, within } from '@testing-library/react'
import { sceneVisuals, worldDetails } from '../data/visualCatalog'
import { GardenStage } from './GardenStage'

describe('GardenStage', () => {
  it('renders the travelling scene as a labelled illustration with its layers and details', () => {
    render(<GardenStage phase="travelling" />)

    expect(screen.getByTestId('garden-stage')).toHaveAttribute(
      'data-scene-state',
      'travelling',
    )
    expect(screen.getAllByTestId('scene-layer')).toHaveLength(3)
    expect(screen.getAllByTestId('scene-detail')).toHaveLength(3)
    expect(screen.getAllByTestId('ambient-mote')).toHaveLength(6)
    expect(screen.getByAltText(/桌前空着/)).toBeVisible()
  })

  it.each(['home', 'travelling', 'returned'] as const)(
    'uses the catalogue hero and details for the %s scene',
    (phase) => {
      render(<GardenStage phase={phase} />)

      const scene = sceneVisuals[phase]
      const stage = screen.getByTestId('garden-stage')
      const images = within(stage).getAllByRole('img')

      expect(images[0]).toHaveAttribute('src', scene.src)
      scene.detailIds.forEach((detailId, index) => {
        expect(images[index + 1]).toHaveAttribute(
          'src',
          worldDetails[detailId].src,
        )
      })
    },
  )

  it('uses the animated crayfish overlay only in the home scene', () => {
    const { rerender } = render(<GardenStage phase="home" />)

    const sprite = screen.getByTestId('crayfish-sprite')
    expect(sprite).toBeVisible()
    expect(sprite.querySelector('img')).toHaveAttribute(
      'src',
      expect.stringContaining('crayfish-home-seated-stool-blink.webp'),
    )

    rerender(<GardenStage phase="travelling" />)
    expect(screen.queryByTestId('crayfish-sprite')).not.toBeInTheDocument()

    rerender(<GardenStage phase="returned" />)
    expect(screen.queryByTestId('crayfish-sprite')).not.toBeInTheDocument()
  })
})
