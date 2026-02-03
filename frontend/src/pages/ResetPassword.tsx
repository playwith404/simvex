import { useState } from 'react'
import { Link } from 'react-router-dom'
import { confirmPasswordReset, requestPasswordReset } from '../services/api'

export const ResetPassword = () => {
  const [email, setEmail] = useState('')
  const [code, setCode] = useState('')
  const [newPassword, setNewPassword] = useState('')
  const [loading, setLoading] = useState(false)
  const [message, setMessage] = useState<string | null>(null)
  const [error, setError] = useState<string | null>(null)

  const handleRequest = async (event: React.FormEvent) => {
    event.preventDefault()
    setLoading(true)
    setError(null)
    setMessage(null)
    try {
      await requestPasswordReset({ email })
      setMessage('재설정 코드가 이메일로 발송되었습니다.')
    } catch (err) {
      setError(err instanceof Error ? err.message : '요청에 실패했습니다')
    } finally {
      setLoading(false)
    }
  }

  const handleConfirm = async (event: React.FormEvent) => {
    event.preventDefault()
    setLoading(true)
    setError(null)
    setMessage(null)
    try {
      await confirmPasswordReset({ email, code, newPassword })
      setMessage('비밀번호가 변경되었습니다. 로그인해 주세요.')
    } catch (err) {
      setError(err instanceof Error ? err.message : '비밀번호 변경에 실패했습니다')
    } finally {
      setLoading(false)
    }
  }

  return (
    <section className="auth-page">
      <div className="auth-card">
        <h2>비밀번호 재설정</h2>
        <p className="muted">이메일로 재설정 코드를 받은 뒤 변경하세요.</p>
        {error && <p className="error">{error}</p>}
        {message && <p className="success">{message}</p>}
        <form onSubmit={handleRequest} className="auth-form">
          <label>
            이메일
            <input value={email} onChange={(e) => setEmail(e.target.value)} type="email" required />
          </label>
          <button type="submit" disabled={loading}>
            {loading ? '전송 중...' : '코드 보내기'}
          </button>
        </form>
        <form onSubmit={handleConfirm} className="auth-form">
          <label>
            인증 코드
            <input value={code} onChange={(e) => setCode(e.target.value)} required />
          </label>
          <label>
            새 비밀번호
            <input value={newPassword} onChange={(e) => setNewPassword(e.target.value)} type="password" minLength={8} required />
          </label>
          <button type="submit" disabled={loading}>
            {loading ? '변경 중...' : '비밀번호 변경'}
          </button>
        </form>
        <div className="auth-links">
          <Link to="/login">로그인으로 이동</Link>
        </div>
      </div>
    </section>
  )
}
