import { useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { register } from '../services/api'

export const Register = () => {
  const navigate = useNavigate()
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [success, setSuccess] = useState(false)

  const handleSubmit = async (event: React.FormEvent) => {
    event.preventDefault()
    setLoading(true)
    setError(null)
    try {
      await register({ email, password })
      setSuccess(true)
      navigate(`/verify?email=${encodeURIComponent(email)}`)
    } catch (err) {
      setError(err instanceof Error ? err.message : '회원가입에 실패했습니다')
    } finally {
      setLoading(false)
    }
  }

  return (
    <section className="auth-page">
      <div className="auth-card">
        <h2>회원가입</h2>
        <p className="muted">이메일로 계정을 생성하세요.</p>
        {error && <p className="error">{error}</p>}
        {success && <p className="success">인증 코드가 발송되었습니다.</p>}
        <form onSubmit={handleSubmit} className="auth-form">
          <label>
            이메일
            <input value={email} onChange={(e) => setEmail(e.target.value)} type="email" required />
          </label>
          <label>
            비밀번호 (8자 이상)
            <input value={password} onChange={(e) => setPassword(e.target.value)} type="password" minLength={8} required />
          </label>
          <button type="submit" disabled={loading}>
            {loading ? '가입 중...' : '가입하기'}
          </button>
        </form>
        <div className="auth-links">
          <Link to="/login">이미 계정이 있나요?</Link>
        </div>
      </div>
    </section>
  )
}
