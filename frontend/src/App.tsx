import { Route, Routes, useLocation } from 'react-router-dom'
import { Layout } from './components/layout/Layout'
import { Landing } from './pages/Landing'
import { ObjectList } from './pages/ObjectList'
import { Viewer } from './pages/Viewer'
import { Workflow } from './pages/Workflow'
import { About } from './pages/About'
import { Login } from './pages/Login'
import { Register } from './pages/Register'
import { Verify } from './pages/Verify'
import { ResetPassword } from './pages/ResetPassword'

const App = () => {
  const location = useLocation()

  return (
    <Layout>
      <Routes location={location} key={location.pathname}>
        <Route path="/" element={<Landing />} />
        <Route path="/login" element={<Login />} />
        <Route path="/register" element={<Register />} />
        <Route path="/verify" element={<Verify />} />
        <Route path="/reset-password" element={<ResetPassword />} />
        <Route path="/objects" element={<ObjectList />} />
        <Route path="/viewer/:objectId" element={<Viewer />} />
        <Route path="/workflow" element={<Workflow />} />
        <Route path="/about" element={<About />} />
      </Routes>
    </Layout>
  )
}

export default App
