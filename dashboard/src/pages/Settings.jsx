import { useState, useEffect } from 'react'

function Settings() {
  const [outputDir, setOutputDir] = useState('')
  const [defaultFormat, setDefaultFormat] = useState('md')
  const [saved, setSaved] = useState(false)

  // Load settings from API
  useEffect(() => {
    fetch('/api/config').then(res => res.json()).then(data => {
      if (data.output_dir) setOutputDir(data.output_dir)
      if (data.default_format) setDefaultFormat(data.default_format)
    }).catch(() => {})
  }, [])

  const handleSave = () => {
    setSaved(true)
    setTimeout(() => setSaved(false), 2000)
  }

  return (
    <div className="max-w-2xl mx-auto">
      <h2 className="text-2xl font-bold mb-6">Settings</h2>

      <div className="space-y-6">
        {/* Output Directory */}
        <div className="bg-[#1A1A1A] border border-[#262626] rounded-lg p-4">
          <label className="block text-sm font-medium text-gray-300 mb-2">
            Output Directory
          </label>
          <input
            type="text"
            value={outputDir}
            onChange={(e) => setOutputDir(e.target.value)}
            placeholder="~/pulp-output"
            className="w-full bg-[#0D0D0D] border border-[#262626] rounded px-3 py-2 text-gray-200 placeholder-gray-600 focus:border-orange-500 focus:outline-none"
          />
          <p className="text-xs text-gray-500 mt-2">
            Where extracted files are saved. Leave empty for stdout.
          </p>
        </div>

        {/* Default Format */}
        <div className="bg-[#1A1A1A] border border-[#262626] rounded-lg p-4">
          <label className="block text-sm font-medium text-gray-300 mb-2">
            Default Format
          </label>
          <select
            value={defaultFormat}
            onChange={(e) => setDefaultFormat(e.target.value)}
            className="w-full bg-[#0D0D0D] border border-[#262626] rounded px-3 py-2 text-gray-200 focus:border-orange-500 focus:outline-none"
          >
            <option value="md">Markdown (.md)</option>
            <option value="skillzip">Skill ZIP (.zip)</option>
            <option value="single">Single file (no frontmatter)</option>
          </select>
        </div>

        {/* Save Button */}
        <button
          onClick={handleSave}
          className="bg-orange-500 hover:bg-orange-600 text-white font-semibold px-6 py-2 rounded-lg transition-colors"
        >
          {saved ? '✓ Saved!' : 'Save Settings'}
        </button>
      </div>
    </div>
  )
}

export default Settings