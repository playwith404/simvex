import type { ReactNode } from 'react'
import { Header } from './Header'

export const Layout = ({ children }: { children: ReactNode }) => {
  return (
    <div className="app-shell">
      <Header />
      <main className="app-main">{children}</main>
      <footer className="app-footer">Copyright © 2026 SIMVEX</footer>
    </div>
  )
}
