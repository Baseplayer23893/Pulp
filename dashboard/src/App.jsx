import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import Layout from './components/Layout'
import Home from './pages/Home'
import Extract from './pages/Extract'
import History from './pages/History'
import Skills from './pages/Skills'
import Settings from './pages/Settings'
import Doctor from './pages/Doctor'

function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<Layout />}>
          <Route index element={<Home />} />
          <Route path="extract" element={<Extract />} />
          <Route path="history" element={<History />} />
          <Route path="skills" element={<Skills />} />
          <Route path="settings" element={<Settings />} />
          <Route path="doctor" element={<Doctor />} />
        </Route>
      </Routes>
    </BrowserRouter>
  )
}

export default App