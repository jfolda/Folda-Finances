import { useState } from 'react'

function App() {
  const [count, setCount] = useState(0)

  return (
    <div className="min-h-screen bg-gray-50 flex items-center justify-center">
      <div className="text-center">
        <h1 className="text-4xl font-bold text-gray-900 mb-4">
          Folda Finances
        </h1>
        <p className="text-gray-600 mb-8">
          Your personal budgeting companion
        </p>
        <button
          onClick={() => setCount((count) => count + 1)}
          className="bg-blue-600 text-white px-6 py-2 rounded-lg hover:bg-blue-700 transition"
        >
          Count: {count}
        </button>
      </div>
    </div>
  )
}

export default App
