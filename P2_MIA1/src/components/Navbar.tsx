import { NavLink } from 'react-router-dom';
import './Navbar.css';

const Navbar = () => {

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
            <NavLink 
              to="/login" 
              className={({ isActive }: { isActive: boolean }) => 
                isActive ? "navbar-link active" : "navbar-link"
              }
            >
              Login
            </NavLink>
          </li>
        </ul>
      </div>
    </nav>
  );
};

export default Navbar;