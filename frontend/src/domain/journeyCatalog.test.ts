import { describe, expect, it } from 'vitest'
import type { FoodId } from './travel'
import { journeyTemplatesByFood } from './journeyCatalog'

const expectedTemplateIdsByFood: Record<FoodId, readonly string[]> = {
  'light-meal': [
    'willow-pond-reed',
    'willow-pond-boat',
    'willow-pond-rain',
    'mist-tea-slope-sprout',
    'mist-tea-slope-kettle',
    'mist-tea-slope-bell',
    'firefly-ravine-fern',
    'firefly-ravine-lantern',
    'firefly-ravine-song',
  ],
  'picnic-basket': [
    'cloud-ridge-pine',
    'cloud-ridge-cloud',
    'cloud-ridge-stone',
    'old-bridge-market-spice',
    'old-bridge-market-ribbon',
    'old-bridge-market-cat',
    'starfall-camp-tent',
    'starfall-camp-map',
    'starfall-camp-comet',
  ],
}

describe('journey catalog', () => {
  it.each(Object.entries(expectedTemplateIdsByFood) as [FoodId, readonly string[]][]) (
    'provides exactly nine stable templates for %s',
    (food, expectedIds) => {
      const templates = journeyTemplatesByFood[food]

      expect(templates).toHaveLength(9)
      expect(templates.map((template) => template.id)).toEqual(expectedIds)
      expect(templates.every((template) => template.food === food)).toBe(true)
      expect(new Set(templates.map((template) => template.postcard.id)).size).toBe(3)
    },
  )

  it('keeps every itinerary as three original chronological event records', () => {
    for (const templates of Object.values(journeyTemplatesByFood)) {
      for (const template of templates) {
        expect(template.events).toHaveLength(3)
        expect(template.events.every((event) => event.length >= 8)).toBe(true)
        expect(template.events[2]).toMatch(/带回|带来了|送回/)
      }
    }
  })

  it('shares one travel card per destination while keeping its text accessible', () => {
    const allTemplates = Object.values(journeyTemplatesByFood).flat()
    const expectedPostcardIds = [
      'willow-pond',
      'mist-tea-slope',
      'firefly-ravine',
      'cloud-ridge',
      'old-bridge-market',
      'starfall-camp',
    ]

    expect([...new Set(allTemplates.map((template) => template.postcard.id))].sort()).toEqual(
      expectedPostcardIds.sort(),
    )

    for (const template of allTemplates) {
      expect(template.postcard.title.length).toBeGreaterThanOrEqual(3)
      expect(template.postcard.body.length).toBeGreaterThanOrEqual(10)
      expect(template.postcard.alt.length).toBeGreaterThanOrEqual(12)
    }
  })
})
