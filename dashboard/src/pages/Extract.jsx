import { useState } from 'react'

function Extract() {
  const [url, setUrl] = useState('')
  const [loading, setLoading] = useState(false)
  const [result, setResult] = useState(null)
  const [error, setError] = useState(null)

  const handleExtract = async (e) => {
    e.preventDefault()
    if (!url.trim()) return

    setLoading(true)
    setError(null)
    setResult(null)

    try {
      const response = await fetch('/api/extract', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ url, format: 'md' })
      })

      const data = await response.json()

      if (data.status === 'error') {
        setError(data.error)
      } else {
        setResult(data)
      }
    } catch (err) {
      setError('Failed to connect to API')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="max-w-3xl mx-auto">
      <h2 className="text-2xl font-bold mb-6">Extract Content</h2>

      {/* URL Input */}
      <form onSubmit={handleExtract} className="mb-6">
        <div className="flex gap-2">
          <input
            type="text"
            value={url}
            onChange={(e) => setUrl(e.target.value)}
            placeholder="Enter URL (https://example.com, YouTube, Reddit, HN, PDF...)"
            className="flex-1 bg-[#1A1A1A] border border-[#262626] rounded-lg px-4 py-3 text-gray-200 placeholder-gray-500 focus:border-orange-500 focus:outline-none"
          />
          <button
            type="submit"
            disabled={loading || !url.trim()}
            className="bg-orange-500 hover:bg-orange-600 disabled:bg-gray-600 disabled:cursor-not-allowed text-white font-semibold px-6 py-3 rounded-lg transition-colors"
          >
            {loading ? 'Squeezing...' : 'Squeeze 🍊'}
          </button>
        </div>
      </form>

      {/* Loading */}
      {loading && (
        <div className="bg-[#1A1A1A] border border-[#262626] rounded-lg p-8 text-center">
          <div className="text-4xl mb-4 animate-spin">🌀</div>
          <p className="text-gray-400">Extracting content...</p>
        </div>
      )}

      {/* Error */}
      {error && (
        <div className="bg-red-900/20 border border-red-500 rounded-lg p-4 text-red-400">
          {error}
        </div>
      )}

      {/* Result */}
      {result && (
        <div className="space-y-4">
          {/* Metadata */}
          <div className="bg-[#1A1A1A] border border-[#262626] rounded-lg p-4 flex items-center justify-between">
            <div>
              <h3 className="font-semibold text-white">{result.title || 'Untitled'}</h3>
              <p className="text-sm text-gray-500">{result.url}</p>
            </div>
            <div className="text-right">
              <span className="text-emerald-400 font-bold">{result.wordCount} words</span>
              <p className="text-xs text-gray-500">{result.outputPath}</p>
            </div>
          </div>

          {/* Preview */}
          <div className="bg-[#1A1A1A] border border-[#262626] rounded-lg overflow-hidden">
            <div className="bg-[#0D0D0D] px-4 py-2 border-b border-[#262626] flex justify-between items-center">
              <span className="text-sm text-gray-400">Preview</span>
              <button
                onClick={() => navigator.clipboard.writeText(result.markdown)}
                className="text-sm text-orange-400 hover:text-orange-300"
              >
                Copy Markdown
              </button>
            </div>
            <pre className="p-4 text-sm text-gray-300 overflow-auto max-h-96 font-mono whitespace-pre-wrap">
              {result.markdown}
            </pre>
          </div>
        </div>
      )}

      {/* Help Text */}
      {!loading && !result && !error && (
        <div className="bg-[#1A1A1A] border border-[#262626] rounded-lg p-4">
          <h4 className="font-semibold text-gray-300 mb-2">Supported Sources</h4>
          <ul className="text-sm text-gray-500 space-y1">
            <li>🌐 <strong>Web pages</strong> - Any URL via defuddle</li>
            <li>▶️ <strong>YouTube</strong> - Videos & Shorts via yt-dlp</li>
            <li>📸 <strong>Instagram</strong> - Reels with transcripts</li>
            <li>🟠 <strong>Reddit</strong> - Posts + top comments</li>
            <li>📊 <strong>Hacker News</strong> - Threads via HN API</li>
            <li>📄 <strong>PDFs</strong> - Text extraction</li>
          </ul>
        </div>
      )}
    </div>
  )
}

export default Extract