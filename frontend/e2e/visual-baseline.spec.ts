import { expect, test } from '@playwright/test'

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

test('garden keeps its expansive travelling scene when driven by a remote snapshot', async ({
  page,
}, testInfo) => {
  await page.emulateMedia({ reducedMotion: 'reduce' })
  await page.route('**/v1/game', (route) => route.fulfill({ json: travellingGame }))
  await page.goto('/')

  const stage = page.getByTestId('garden-stage')
  await expect(stage).toHaveAttribute('data-scene-state', 'travelling')
  await expect(stage).toBeVisible()
  await expect(stage.getByTestId('scene-layer')).toHaveCount(3)
  await expect(stage.getByTestId('scene-detail')).toHaveCount(3)
  await expect(stage.locator('.scene-art__hero')).toBeVisible()

  await page.screenshot({ path: testInfo.outputPath('remote-travelling.png'), animations: 'disabled' })
  await expect(page).toHaveScreenshot('remote-travelling.png', { animations: 'disabled' })
})
