import React, { useState } from 'react';
import './Login.css';
import { useNavigate } from 'react-router-dom';

const Login = () => {
  const [formData, setFormData] = useState({
    id: '',
    username: '',
    password: ''
  });
  const [mensaje, setMensaje] = useState('');
  const navigate = useNavigate();

  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const { name, value } = e.target;
    setFormData(prev => ({
      ...prev,
      [name]: value
    }));
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setMensaje(''); // Limpiar mensaje anterior

    try {
      const response = await fetch('http://3.144.31.35:5000/login', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          id: formData.id,
          user: formData.username,
          pass: formData.password
        })
      });
      const data = await response.json();
      setMensaje(data.mensaje);
      if (data.mensaje && data.mensaje.startsWith('✅')) {
        localStorage.setItem('loggedIn', 'true');
        setTimeout(() => navigate('/seleccion-disco'), 1000);
      }
    } catch (error) {
      setMensaje('Error de conexión con el servidor');
    }
  };

  return (
    <div className="login-container">
      <h1>Inicio de Sesión</h1>
      <form onSubmit={handleSubmit} className="login-form">
        <div className="form-group">
          <label htmlFor="id">ID de Partición:</label>
          <input
            type="text"
            id="id"
            name="id"
            value={formData.id}
            onChange={handleChange}
            required
            placeholder="Ingrese su ID"
          />
        </div>

        <div className="form-group">
          <label htmlFor="username">Usuario:</label>
          <input
            type="text"
            id="username"
            name="username"
            value={formData.username}
            onChange={handleChange}
            required
            placeholder="Ingrese su nombre de usuario"
          />
        </div>

        <div className="form-group">
          <label htmlFor="password">Contraseña:</label>
          <input
            type="password"
            id="password"
            name="password"
            value={formData.password}
            onChange={handleChange}
            required
            placeholder="Ingrese su contraseña"
          />
        </div>

        <button type="submit" className="login-button">
          Iniciar Sesión
        </button>
        {mensaje && <div className="login-message">{mensaje}</div>}
      </form>
    </div>
  );
};

export default Login;