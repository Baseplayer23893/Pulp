function Skills() {
  // Placeholder - would list packaged skills from ~/.config/pulp/skills/
  return (
    <div className="max-w-3xl mx-auto">
      <h2 className="text-2xl font-bold mb-6">Skills</h2>
      <div className="bg-[#1A1A1A] border border-[#262626] rounded-lg p-8 text-center">
        <div className="text-4xl mb-4">📦</div>
        <h3 className="text-lg font-semibold text-gray-300 mb-2">No Skills Yet</h3>
        <p className="text-gray-500 mb-4">Package extractions as skills to use in AI IDEs</p>
        <p className="text-sm text-gray-600">
          Run <code className="bg-[#0D0D0D] px-2 py-1 rounded">pulp package &lt;name&gt;</code> to create a skill.zip
        </p>
      </div>
    </div>
  )
}

export default Skills