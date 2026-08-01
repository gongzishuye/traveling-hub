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

test('a user logs in with agent-issued credentials and sees the server-owned traveler', async ({ page }) => {
  let gameRequestCount = 0
  let isAuthenticated = false
  await page.route('**/v1/game', async (route) => {
    gameRequestCount += 1
    if (!isAuthenticated) {
      await route.fulfill({ status: 401, json: { error: 'web session required' } })
      return
    }
    await route.fulfill({ json: travellingGame })
  })
  await page.route('**/v1/web/login', async (route) => {
    expect(route.request().method()).toBe('POST')
    expect(route.request().postDataJSON()).toEqual({
      email: 'traveler@example.com',
      password: 'agent-issued-password',
    })
    isAuthenticated = true
    await route.fulfill({ json: { must_change_password: false } })
  })

  await page.goto('/')
  await page.getByLabel('邮箱').fill('traveler@example.com')
  await page.getByLabel('密码').fill('agent-issued-password')
  await page.getByRole('button', { name: '登录' }).click()

  await expect(page.getByTestId('garden-stage')).toHaveAttribute('data-scene-state', 'travelling')
  expect(gameRequestCount).toBeGreaterThanOrEqual(2)
})

test('a restricted first session forces a password change before showing the traveler', async ({ page }) => {
  let canReadGame = false
  await page.route('**/v1/game', async (route) => {
    if (!canReadGame) {
      await route.fulfill({ status: 403, json: { error: 'password change required' } })
      return
    }
    await route.fulfill({ json: travellingGame })
  })
  await page.route('**/v1/web/change-password', async (route) => {
    expect(route.request().method()).toBe('POST')
    expect(route.request().postDataJSON()).toEqual({ password: 'a-new-secure-password' })
    canReadGame = true
    await route.fulfill({ status: 204 })
  })

  await page.goto('/')
  await expect(page.getByRole('heading', { name: '设置新密码' })).toBeVisible()
  await page.getByLabel('新密码').fill('a-new-secure-password')
  await page.getByRole('button', { name: '更新密码' }).click()

  await expect(page.getByTestId('garden-stage')).toHaveAttribute('data-scene-state', 'travelling')
})

test('a transient game loading failure can be retried', async ({ page }) => {
  let shouldSucceed = false
  await page.route('**/v1/game', async (route) => {
    if (!shouldSucceed) {
      await route.fulfill({ status: 500, json: { error: 'temporarily unavailable' } })
      return
    }
    await route.fulfill({ json: travellingGame })
  })

  await page.goto('/')
  await expect(page.getByRole('alert')).toHaveText('暂时无法连接远行小屋，请稍后重试。')
  shouldSucceed = true
  await page.getByRole('button', { name: '重试' }).click()

  await expect(page.getByTestId('garden-stage')).toHaveAttribute('data-scene-state', 'travelling')
})
