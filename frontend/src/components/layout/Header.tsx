import { Link, NavLink } from 'react-router-dom'

export const Header = () => {
  return (
    <header className="app-header">
      <div className="logo">
        <Link to="/">SIMVEX</Link>
      </div>
      <nav className="nav">
        <NavLink to="/objects">오브젝트 목록</NavLink>
        <NavLink to="/workflow">워크플로우</NavLink>
        <NavLink to="/about">About</NavLink>
      </nav>
    </header>
  )
}
