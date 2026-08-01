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

test('shows the server-owned traveler without local departure controls', async ({ page }) => {
  await page.route('**/v1/game', (route) => route.fulfill({ json: travellingGame }))
  await page.goto('/')

  await expect(page.getByTestId('garden-stage')).toHaveAttribute(
    'data-scene-state',
    'travelling',
  )
  await page.getByRole('button', { name: '展开旅行手账' }).click()
  await expect(page.getByText('旅人已经出发，远行小屋会在归来后更新记录。')).toBeVisible()
  await expect(page.getByRole('radio')).toHaveCount(0)
  await expect(page.getByRole('button', { name: '出发' })).toHaveCount(0)

  await page.getByRole('button', { name: '展开旅行卡' }).click()
  await page.getByRole('button', { name: '查看旅行卡：云脊坡' }).click()
  await expect(page.getByRole('dialog', { name: '旅行卡：云脊坡' })).toBeVisible()
  await expect(page.getByRole('button', { name: '再次出发' })).toHaveCount(0)
})
