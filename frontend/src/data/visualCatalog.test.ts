import { describe, expect, it, vi } from 'vitest'
import {
  foodVisuals,
  postcardVisuals,
  preloadSceneHeroes,
  sceneVisuals,
  worldDetails,
} from './visualCatalog'

function expectProjectAsset(src: string, relativePath: string) {
  expect(src).not.toMatch(/^data:/i)
  expect(src).toMatch(
    new RegExp(`/src/assets/${relativePath.replaceAll('.', '\\.')}(?:\\?.*)?$`),
  )
}

describe('visual catalog', () => {
  it('preloads every local scene hero before the journey state can change', () => {
    const assignedSources: string[] = []

    class PreloadImage {
      set src(value: string) {
        assignedSources.push(value)
      }
    }

    vi.stubGlobal('Image', PreloadImage)

    try {
      preloadSceneHeroes()
    } finally {
      vi.unstubAllGlobals()
    }

    expect(assignedSources).toEqual([
      sceneVisuals.home.src,
      sceneVisuals.travelling.src,
      sceneVisuals.returned.src,
    ])
  })

  it('maps every existing food and postcard ID to its required local PNG asset', () => {
    expect(Object.keys(foodVisuals).sort()).toEqual([
      'light-meal',
      'picnic-basket',
    ])
    expect(Object.keys(postcardVisuals).sort()).toEqual([
      'cloud-ridge',
      'firefly-ravine',
      'mist-tea-slope',
      'old-bridge-market',
      'starfall-camp',
      'willow-pond',
    ])

    expectProjectAsset(foodVisuals['light-meal'].src, 'props/light-meal.png')
    expectProjectAsset(foodVisuals['picnic-basket'].src, 'props/picnic-basket.png')
    expectProjectAsset(postcardVisuals['willow-pond'].src, 'postcards/willow-pond.png')
    expectProjectAsset(postcardVisuals['cloud-ridge'].src, 'postcards/cloud-ridge.png')
    expectProjectAsset(postcardVisuals['mist-tea-slope'].src, 'postcards/mist-tea-slope.png')
    expectProjectAsset(postcardVisuals['firefly-ravine'].src, 'postcards/firefly-ravine.png')
    expectProjectAsset(postcardVisuals['old-bridge-market'].src, 'postcards/old-bridge-market.png')
    expectProjectAsset(postcardVisuals['starfall-camp'].src, 'postcards/starfall-camp.png')

    for (const visual of [...Object.values(foodVisuals), ...Object.values(postcardVisuals)]) {
      expect(visual.alt.length).toBeGreaterThanOrEqual(12)
      expect(visual.description.length).toBeGreaterThanOrEqual(12)
    }
  })

  it('maps every journey phase to a complete local PNG hero and exactly three custom details', () => {
    expect(Object.keys(sceneVisuals).sort()).toEqual([
      'home',
      'returned',
      'travelling',
    ])

    expectProjectAsset(
      sceneVisuals.home.src,
      'scenes/garden-travelling-consistent-v2.png',
    )
    expectProjectAsset(
      sceneVisuals.travelling.src,
      'scenes/garden-travelling-consistent-v2.png',
    )
    expectProjectAsset(
      sceneVisuals.returned.src,
      'scenes/garden-returned-consistent-v2.png',
    )

    for (const scene of Object.values(sceneVisuals)) {
      expect(scene.detailIds).toHaveLength(3)
      expect(new Set(scene.detailIds)).toHaveLength(3)
      expect(scene.alt.length).toBeGreaterThanOrEqual(12)
      expect(scene.description.length).toBeGreaterThanOrEqual(12)

      for (const detailId of scene.detailIds) {
        expect(worldDetails[detailId]).toBeDefined()
      }
    }
  })

  it('describes the wide courtyard heroes accurately', () => {
    expect(sceneVisuals.home.alt).toContain('小龙虾')
    expect(sceneVisuals.home.description).toContain('小龙虾')
    expect(sceneVisuals.home.description).toContain('完整远景庭院')
    expect(sceneVisuals.travelling.alt).toContain('无角色')
    expect(sceneVisuals.travelling.alt).toContain('完整远景庭院')
    expect(sceneVisuals.travelling.alt).toContain('空椅')
    expect(sceneVisuals.travelling.description).toContain('打开的手账')
    expect(sceneVisuals.returned.alt).toContain('小龙虾')
    expect(sceneVisuals.returned.description).toContain('小龙虾')
    expect(sceneVisuals.returned.description).toContain('完整远景庭院')
  })

  it('gives every scene-referenced world detail a local source and descriptive text', () => {
    const referencedDetailIds = new Set(
      Object.values(sceneVisuals).flatMap((scene) => scene.detailIds),
    )

    for (const detailId of referencedDetailIds) {
      const detail = worldDetails[detailId]

      expectProjectAsset(detail.src, `illustrations/detail-${detailId}.svg`)
      expect(detail.alt.length).toBeGreaterThanOrEqual(12)
      expect(detail.description.length).toBeGreaterThanOrEqual(12)
    }
  })
})
