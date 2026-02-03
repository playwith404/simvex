import { useState } from 'react'
import { Link, useLocation, useNavigate } from 'react-router-dom'
import { verifyEmail } from '../services/api'

export const Verify = () => {
  const location = useLocation()
  const navigate = useNavigate()
  const query = new URLSearchParams(location.search)
  const initialEmail = query.get('email') ?? ''

  const [email, setEmail] = useState(initialEmail)
  const [code, setCode] = useState('')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [success, setSuccess] = useState(false)

  const handleSubmit = async (event: React.FormEvent) => {
    event.preventDefault()
    setLoading(true)
    setError(null)
    try {
      await verifyEmail({ email, code })
      setSuccess(true)
      navigate('/login')
    } catch (err) {
      setError(err instanceof Error ? err.message : '인증에 실패했습니다')
    } finally {
      setLoading(false)
    }
  }

  return (
    <section className="auth-page">
      <div className="auth-card">
        <h2>이메일 인증</h2>
        <p className="muted">메일로 받은 인증 코드를 입력하세요.</p>
        {error && <p className="error">{error}</p>}
        {success && <p className="success">인증이 완료되었습니다.</p>}
        <form onSubmit={handleSubmit} className="auth-form">
          <label>
            이메일
            <input value={email} onChange={(e) => setEmail(e.target.value)} type="email" required />
          </label>
          <label>
            인증 코드
            <input value={code} onChange={(e) => setCode(e.target.value)} required />
          </label>
          <button type="submit" disabled={loading}>
            {loading ? '인증 중...' : '인증 완료'}
          </button>
        </form>
        <div className="auth-links">
          <Link to="/login">로그인으로 이동</Link>
        </div>
      </div>
    </section>
  )
}
