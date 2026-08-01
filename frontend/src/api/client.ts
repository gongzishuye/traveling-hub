export type GamePhase = 'home' | 'travelling' | 'returned'

export type GameSnapshot = {
  frog_id: string
  server_time: string
  local_date: string
  phase: GamePhase
  journey: {
    template_id: string
    departed_at: string
    return_at?: string
  } | null
  events: Array<{ stage: string; text: string }>
  album_postcard_ids: string[]
}

export type LoginResponse = {
  must_change_password: boolean
}

export class ApiError extends Error {
  readonly status: number

  constructor(
    status: number,
    message: string,
  ) {
    super(message)
    this.name = 'ApiError'
    this.status = status
  }
}

async function request(path: string, init: RequestInit = {}): Promise<Response> {
  const response = await fetch(path, {
    credentials: 'include',
    ...init,
    headers: {
      Accept: 'application/json',
      ...init.headers,
    },
  })

  if (!response.ok) {
    if (response.status === 401) {
      throw new ApiError(401, '请先登录。')
    }
    throw new ApiError(response.status, '暂时无法连接远行小屋，请稍后重试。')
  }

  return response
}

export async function getGameSnapshot(): Promise<GameSnapshot> {
  return (await request('/v1/game')).json() as Promise<GameSnapshot>
}

export async function login(email: string, password: string): Promise<LoginResponse> {
  const response = await request('/v1/web/login', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ email, password }),
  })
  return response.json() as Promise<LoginResponse>
}

export async function changePassword(newPassword: string): Promise<void> {
  await request('/v1/web/change-password', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ password: newPassword }),
  })
}
