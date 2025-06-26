import React, { useState } from 'react';
import './Login.css';

const Login = () => {
  const [formData, setFormData] = useState({
    id: '',
    username: '',
    password: ''
  });

  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const { name, value } = e.target;
    setFormData(prev => ({
      ...prev,
      [name]: value
    }));
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    console.log('Datos de login:', formData);
    // Aquí iría la lógica para autenticar al usuario
    alert(`Login enviado para ID: ${formData.id}, Usuario: ${formData.username}`);
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
      </form>
    </div>
  );
};

export default Login;