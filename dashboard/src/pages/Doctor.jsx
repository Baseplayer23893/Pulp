import { useState, useEffect } from 'react'

function Doctor() {
  const [status, setStatus] = useState(null)
  const [checking, setChecking] = useState(true)

  useEffect(() => {
    fetch('/api/health')
      .then(res => res.json())
      .then(data => {
        setStatus(data)
        setChecking(false)
      })
      .catch(() => setChecking(false))
  }, [])

  const checks = [
    { name: 'Pulp CLI', status: checking ? 'checking' : status ? 'ok' : 'error', label: 'CLI installed' },
    { name: 'defuddle', status: 'ok', label: 'npm package' },
    { name: 'yt-dlp', status: 'ok', label: 'Python package' },
  ]

  return (
    <div className="max-w-2xl mx-auto">
      <h2 className="text-2xl font-bold mb-6">Doctor</h2>

      {/* Status */}
      <div className="bg-[#1A1A1A] border border-[#262626] rounded-lg p-4 mb-6">
        <div className="flex items-center gap-3">
          <span className="text-2xl">●</span>
          <div>
            <div className="font-semibold text-emerald-400">System Healthy</div>
            <div className="text-sm text-gray-500">v{status?.version || 'unknown'}</div>
          </div>
        </div>
      </div>

      {/* Checks */}
      <div className="space-y-2">
        {checks.map(check => (
          <div
            key={check.name}
            className="bg-[#1A1A1A] border border-[#262626] rounded-lg p-4 flex justify-between items-center"
          >
            <div>
              <div className="font-medium text-gray-200">{check.name}</div>
              <div className="text-sm text-gray-500">{check.label}</div>
            </div>
            <span className={`px-3 py-1 rounded text-sm font-medium ${
              check.status === 'ok' ? 'bg-emerald-900/30 text-emerald-400' :
              check.status === 'error' ? 'bg-red-900/30 text-red-400' :
              'bg-yellow-900/30 text-yellow-400'
            }`}>
              {check.status === 'ok' ? '✓ OK' :
               check.status === 'error' ? '✗ Missing' :
               '...'}
            </span>
          </div>
        ))}
      </div>
    </div>
  )
}

export default Doctor