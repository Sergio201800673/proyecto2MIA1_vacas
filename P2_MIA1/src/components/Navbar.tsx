import { NavLink, useNavigate } from 'react-router-dom';
import './Navbar.css';
import { useEffect, useState } from 'react';

const Navbar = () => {
  const [loggedIn, setLoggedIn] = useState(false);
  const navigate = useNavigate();

  useEffect(() => {
    // Revisar si hay sesión guardada en localStorage
    setLoggedIn(localStorage.getItem('loggedIn') === 'true');
  }, []);

  // Escuchar cambios de sesión en otras pestañas
  useEffect(() => {
    const onStorage = () => setLoggedIn(localStorage.getItem('loggedIn') === 'true');
    window.addEventListener('storage', onStorage);
    return () => window.removeEventListener('storage', onStorage);
  }, []);

  const handleLogout = () => {
    localStorage.setItem('loggedIn', 'false');
    setLoggedIn(false);
    navigate('/login');
  };

  return (
    <nav className="navbar">
      <div className="navbar-container">
        <NavLink to="/" className="navbar-logo">
          Inicio
        </NavLink>
        
        <ul className="navbar-menu">
          <li className="navbar-item">
            <NavLink 
              to="/" 
              className={({ isActive }: { isActive: boolean }) => 
                isActive ? "navbar-link active" : "navbar-link"
              }
            >
              Inicio
            </NavLink>
          </li>
          <li className="navbar-item">
            <NavLink 
              to="/console" 
              className={({ isActive }: { isActive: boolean }) => 
                isActive ? "navbar-link active" : "navbar-link"
              }
            >
              Consola
            </NavLink>
          </li>
          <li className="navbar-item">
            {loggedIn ? (
              <button className="navbar-link" onClick={handleLogout} style={{background: 'none', border: 'none', color: 'inherit', cursor: 'pointer'}}>Logout</button>
            ) : (
              <NavLink 
                to="/login" 
                className={({ isActive }: { isActive: boolean }) => 
                  isActive ? "navbar-link active" : "navbar-link"
                }
              >
                Login
              </NavLink>
            )}
          </li>
        </ul>
      </div>
    </nav>
  );
};

export default Navbar;