import { Routes, Route } from 'react-router-dom'

export default function App() {
  return (
    <Routes>
      <Route
        path="*"
        element={
          <div className="min-h-screen bg-gray-950 text-white flex items-center justify-center">
            <h1 className="text-4xl font-bold text-brand-400">Shortly</h1>
          </div>
        }
      />
    </Routes>
  )
}
