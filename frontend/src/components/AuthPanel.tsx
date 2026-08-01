import { useState } from 'react'

type AuthPanelProps =
  | { mode: 'loading' }
  | { mode: 'error'; error: string | null; onRetry(): void }
  | { mode: 'login'; error: string | null; onLogin(email: string, password: string): Promise<void> }
  | { mode: 'change-password'; error: string | null; onChangePassword(password: string): Promise<void> }

export function AuthPanel(props: AuthPanelProps) {
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [newPassword, setNewPassword] = useState('')
  const [submitting, setSubmitting] = useState(false)

  async function submitLogin(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault()
    if (props.mode !== 'login') return
    setSubmitting(true)
    try {
      await props.onLogin(email, password)
    } finally {
      setSubmitting(false)
    }
  }

  async function submitPasswordChange(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault()
    if (props.mode !== 'change-password') return
    setSubmitting(true)
    try {
      await props.onChangePassword(newPassword)
    } finally {
      setSubmitting(false)
    }
  }

  if (props.mode === 'loading') {
    return <main className="auth-panel"><p>正在打开远行小屋…</p></main>
  }

  if (props.mode === 'error') {
    return (
      <main className="auth-panel">
        <p role="alert">{props.error}</p>
        <button type="button" onClick={props.onRetry}>重试</button>
      </main>
    )
  }

  if (props.mode === 'change-password') {
    return (
      <main className="auth-panel">
        <h1>设置新密码</h1>
        <p>首次登录前，请先设置新的密码。</p>
        {props.error ? <p role="alert">{props.error}</p> : null}
        <form onSubmit={submitPasswordChange}>
          <label>
            新密码
            <input
              type="password"
              required
              minLength={8}
              value={newPassword}
              onChange={(event) => setNewPassword(event.target.value)}
            />
          </label>
          <button type="submit" disabled={submitting}>{submitting ? '更新中…' : '更新密码'}</button>
        </form>
      </main>
    )
  }

  return (
    <main className="auth-panel">
      <h1>登录远行小屋</h1>
      <p>使用 Agent 注册时获得的邮箱与初始密码登录。</p>
      {props.error ? <p role="alert">{props.error}</p> : null}
      <form onSubmit={submitLogin}>
        <label>
          邮箱
          <input
            type="email"
            required
            autoComplete="email"
            value={email}
            onChange={(event) => setEmail(event.target.value)}
          />
        </label>
        <label>
          密码
          <input
            type="password"
            required
            autoComplete="current-password"
            value={password}
            onChange={(event) => setPassword(event.target.value)}
          />
        </label>
        <button type="submit" disabled={submitting}>{submitting ? '登录中…' : '登录'}</button>
      </form>
    </main>
  )
}
