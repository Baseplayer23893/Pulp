import { Link } from 'react-router-dom'

const features = [
  { icon: '🌐', title: 'Web Pages', desc: 'Extract clean markdown from any URL', path: '/extract' },
  { icon: '▶️', title: 'YouTube', desc: 'Get transcripts from videos', path: '/extract' },
  { icon: '📸', title: 'Instagram', desc: 'Extract reels + captions', path: '/extract' },
  { icon: '🟠', title: 'Reddit', desc: 'Posts + top comments', path: '/extract' },
  { icon: '📊', title: 'Hacker News', desc: 'HN threads as markdown', path: '/extract' },
  { icon: '📄', title: 'PDFs', desc: 'Text extraction', path: '/extract' },
]

function Home() {
  return (
    <div className="max-w-4xl mx-auto">
      {/* Hero */}
      <div className="text-center mb-12">
        <h1 className="text-4xl font-bold mb-4 bg-gradient-to-r from-purple-400 to-purple-600 bg-clip-text text-transparent">
          Pulp Dashboard
        </h1>
        <p className="text-gray-400 text-lg">
          Squeeze the web into clean markdown for AI
        </p>
      </div>

      {/* Quick Actions */}
      <div className="grid grid-cols-2 gap-4 mb-8">
        <Link
          to="/extract"
          className="bg-orange-500 hover:bg-orange-600 text-white font-semibold py-4 px-6 rounded-lg transition-colors text-center"
        >
          🍊 Start Squeezing
        </Link>
        <Link
          to="/history"
          className="bg-[#1A1A1A] hover:bg-[#262626] border border-[#262626] text-gray-300 font-semibold py-4 px-6 rounded-lg transition-colors text-center"
        >
          📜 View History
        </Link>
      </div>

      {/* Features Grid */}
      <div className="grid grid-cols-2 md:grid-cols-3 gap-4">
        {features.map(feature => (
          <Link
            key={feature.title}
            to={feature.path}
            className="bg-[#1A1A1A] border border-[#262626] rounded-lg p-4 hover:border-orange-500/50 transition-colors group"
          >
            <div className="text-2xl mb-2">{feature.icon}</div>
            <h3 className="font-semibold text-gray-200 group-hover:text-orange-400 transition-colors">
              {feature.title}
            </h3>
            <p className="text-sm text-gray-500 mt-1">{feature.desc}</p>
          </Link>
        ))}
      </div>

      {/* Stats */}
      <div className="mt-12 grid grid-cols-3 gap-4">
        <div className="bg-[#1A1A1A] border border-[#262626] rounded-lg p-4 text-center">
          <div className="text-2xl font-bold text-emerald-400">0</div>
          <div className="text-sm text-gray-500">Total Squeezes</div>
        </div>
        <div className="bg-[#1A1A1A] border border-[#262626] rounded-lg p-4 text-center">
          <div className="text-2xl font-bold text-orange-400">0</div>
          <div className="text-sm text-gray-500">Skills Created</div>
        </div>
        <div className="bg-[#1A1A1A] border border-[#262626] rounded-lg p-4 text-center">
          <div className="text-2xl font-bold text-purple-400">●</div>
          <div className="text-sm text-gray-500">System Status</div>
        </div>
      </div>
    </div>
  )
}

export default Home