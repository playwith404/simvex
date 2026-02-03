import { createContext, useContext, useState, useEffect, type ReactNode } from 'react'
import { getMe, logout as apiLogout } from '../services/api'

type User = {
  id: string
  email: string
} | null

type AuthContextType = {
  user: User
  isLoading: boolean
  isAuthenticated: boolean
  logout: () => Promise<void>
  refresh: () => Promise<void>
}

const AuthContext = createContext<AuthContextType | null>(null)

export const AuthProvider = ({ children }: { children: ReactNode }) => {
  const [user, setUser] = useState<User>(null)
  const [isLoading, setIsLoading] = useState(true)

  const refresh = async () => {
    try {
      const me = await getMe()
      setUser(me)
    } catch {
      setUser(null)
    } finally {
      setIsLoading(false)
    }
  }

  const logout = async () => {
    try {
      await apiLogout()
    } finally {
      setUser(null)
    }
  }

  useEffect(() => {
    refresh()
  }, [])

  return (
    <AuthContext.Provider
      value={{
        user,
        isLoading,
        isAuthenticated: !!user,
        logout,
        refresh,
      }}
    >
      {children}
    </AuthContext.Provider>
  )
}

export const useAuth = () => {
  const context = useContext(AuthContext)
  if (!context) {
    throw new Error('useAuth must be used within an AuthProvider')
  }
  return context
}
