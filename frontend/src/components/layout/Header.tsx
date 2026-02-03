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
      <div className="logo">
        <Link to="/">SIMVEX</Link>
      </div>
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
