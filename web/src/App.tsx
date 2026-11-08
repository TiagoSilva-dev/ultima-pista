import { Navigate, Route, Routes } from 'react-router-dom'
import { HomePage } from './pages/HomePage'
import { JoinPage } from './pages/JoinPage'
import { RoomPage } from './pages/RoomPage'
import { TelaoPage } from './pages/TelaoPage'

export default function App() {
  return (
    <Routes>
      <Route path="/" element={<HomePage />} />
      <Route path="/entrar" element={<JoinPage />} />
      <Route path="/entrar/:code" element={<JoinPage />} />
      <Route path="/sala/:code" element={<RoomPage />} />
      <Route path="/telao/:code" element={<TelaoPage />} />
      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  )
}
