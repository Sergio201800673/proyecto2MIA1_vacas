import './App.css'
import Navbar from './components/Navbar'
import { Routes, Route } from 'react-router-dom'
import Home from './pages/Inicio'
import Console from './pages/Console'
import Login from './pages/Login'
import SeleccionDisco from './pages/SeleccionDisco'

function App() {

  return (
    <div className="App">
      <Navbar />
      <div className="content">
        <Routes>
          <Route path="/" element={<Home />} />
          <Route path="/console" element={<Console />} />
          <Route path="/login" element={<Login />} />
          <Route path="/seleccion-disco" element={<SeleccionDisco />} />
        </Routes>
      </div>
    </div>
  )
}

export default App
