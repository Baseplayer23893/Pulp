import { useState, useEffect } from 'react'

function History() {
  const [items, setItems] = useState([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    fetch('/api/history')
      .then(res => res.json())
      .then(data => {
        setItems(data)
        setLoading(false)
      })
      .catch(() => setLoading(false))
  }, [])

  if (loading) {
    return <div className="text-center text-gray-400">Loading...</div>
  }

  if (items.length === 0) {
    return (
      <div className="text-center">
        <h2 className="text-2xl font-bold mb-4">History</h2>
        <p className="text-gray-500">No squeezes yet. Start extracting!</p>
      </div>
    )
  }

  return (
    <div className="max-w-3xl mx-auto">
      <h2 className="text-2xl font-bold mb-6">History</h2>
      <div className="space-y-2">
        {items.map((item, idx) => (
          <div
            key={idx}
            className="bg-[#1A1A1A] border border-[#262626] rounded-lg p-4 hover:border-orange-500/30 transition-colors"
          >
            <div className="flex justify-between items-start">
              <div>
                <h3 className="font-medium text-white">{item.title}</h3>
                <p className="text-sm text-gray-500 truncate">{item.url}</p>
              </div>
              <span className="text-xs text-gray-500">{item.timestamp}</span>
            </div>
            <div className="mt-2 flex gap-4 text-sm text-gray-400">
              <span>{item.wordCount} words</span>
              <span className="capitalize">{item.format}</span>
            </div>
          </div>
        ))}
      </div>
    </div>
  )
}

export default History