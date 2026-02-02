import { Route, Routes } from 'react-router-dom'
import { Layout } from './components/layout/Layout'
import { Landing } from './pages/Landing'
import { ObjectList } from './pages/ObjectList'
import { Viewer } from './pages/Viewer'
import { Workflow } from './pages/Workflow'
import { About } from './pages/About'

const App = () => {
  return (
    <Layout>
      <Routes>
        <Route path="/" element={<Landing />} />
        <Route path="/objects" element={<ObjectList />} />
        <Route path="/viewer/:objectId" element={<Viewer />} />
        <Route path="/workflow" element={<Workflow />} />
        <Route path="/about" element={<About />} />
      </Routes>
    </Layout>
  )
}

export default App
