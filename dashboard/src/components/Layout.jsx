import { Outlet, NavLink, useLocation } from 'react-router-dom'
import { useState } from 'react'

const navItems = [
  { path: '/', label: 'Home', icon: '🏠' },
  { path: '/extract', label: 'Extract', icon: '🌐' },
  { path: '/history', label: 'History', icon: '📜' },
  { path: '/skills', label: 'Skills', icon: '📦' },
  { path: '/settings', label: 'Settings', icon: '⚙️' },
  { path: '/doctor', label: 'Doctor', icon: '🩺' },
]

function Layout() {
  const [sidebarOpen, setSidebarOpen] = useState(true)
  const location = useLocation()

  const getPageTitle = () => {
    const item = navItems.find(i => i.path === location.pathname)
    return item?.label || 'Pulp'
  }

  return (
    <div className="flex h-screen bg-[#0D0D0D]">
      {/* Sidebar */}
      <aside className={`${sidebarOpen ? 'w-60' : 'w-16'} bg-[#1A1A1A] border-r border-[#262626] flex flex-col transition-all duration-300`}>
        {/* Logo */}
        <div className="p-4 border-b border-[#262626]">
          <h1 className="text-xl font-bold bg-gradient-to-r from-orange-500 to-orange-400 bg-clip-text text-transparent">
            🍊 Pulp
          </h1>
        </div>

        {/* Navigation */}
        <nav className="flex-1 p-2">
          {navItems.map(item => (
            <NavLink
              key={item.path}
              to={item.path}
              className={({ isActive }) => `
                flex items-center gap-3 px-3 py-2.5 rounded-lg mb-1
                transition-colors duration-150
                ${isActive 
                  ? 'bg-orange-500/20 text-orange-400 border-l-2 border-orange-500' 
                  : 'text-gray-400 hover:bg-[#262626] hover:text-gray-200'}
              `}
            >
              <span className="text-lg">{item.icon}</span>
              {sidebarOpen && <span className="text-sm font-medium">{item.label}</span>}
            </NavLink>
          ))}
        </nav>

        {/* Toggle Button */}
        <button
          onClick={() => setSidebarOpen(!sidebarOpen)}
          className="p-4 border-t border-[#262626] text-gray-500 hover:text-gray-300"
        >
          {sidebarOpen ? '◀' : '▶'}
        </button>
      </aside>

      {/* Main Content */}
      <main className="flex-1 flex flex-col overflow-hidden">
        {/* Header */}
        <header className="h-14 bg-[#1A1A1A] border-b border-[#262626] flex items-center px-6">
          <span className="text-gray-400 text-sm">
            pulp <span className="text-orange-500">›</span> {getPageTitle()}
          </span>
        </header>

        {/* Content Area */}
        <div className="flex-1 overflow-auto p-6">
          <Outlet />
        </div>
      </main>
    </div>
  )
}

export default Layout