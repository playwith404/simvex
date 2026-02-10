import { Link, NavLink, useNavigate } from 'react-router-dom'
import { useAuth } from '../../contexts/AuthContext'

export const Header = () => {
  const { isAuthenticated, isLoading, logout } = useAuth()
  const navigate = useNavigate()

  const handleLogout = async () => {
    await logout()
    navigate('/')
  }

  return (
    <header className="app-header">
      <Link to="/" className="logo">
        <svg width="120" height="28" viewBox="0 0 140 32" fill="none" xmlns="http://www.w3.org/2000/svg" aria-label="SIMVEX">
          <rect x="0" y="4" width="24" height="24" rx="5" fill="#f8c86a" opacity="0.15" />
          <rect x="2" y="6" width="20" height="20" rx="4" stroke="#f8c86a" strokeWidth="1.5" fill="none" />
          <path d="M8 18 L12 12 L16 18" stroke="#f8c86a" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round" fill="none" />
          <circle cx="12" cy="11" r="1.2" fill="#f8c86a" />
          <text x="32" y="22" fontFamily="'Space Grotesk', sans-serif" fontSize="18" fontWeight="600" fill="#f5f3ee" letterSpacing="2">SIM</text>
          <text x="75" y="22" fontFamily="'Space Grotesk', sans-serif" fontSize="18" fontWeight="600" fill="#f8c86a" letterSpacing="2">VEX</text>
        </svg>
      </Link>
      <nav className="nav">
        <NavLink to="/objects" reloadDocument>오브젝트 목록</NavLink>
        <NavLink to="/workflow" reloadDocument>워크플로우</NavLink>
        <NavLink to="/about" reloadDocument>About</NavLink>
        {isLoading ? null : isAuthenticated ? (
          <button onClick={handleLogout} className="nav-logout-btn">로그아웃</button>
        ) : (
          <NavLink to="/login" reloadDocument>로그인</NavLink>
        )}
      </nav>
    </header>
  )
}
