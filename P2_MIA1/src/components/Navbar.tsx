import { NavLink } from 'react-router-dom';
import './Navbar.css';
import { useState } from 'react';

const Navbar = () => {

  // const [inputCode, setInputCode] = useState('');
  // const [output, setOutput] = useState('Esperando ejecución...\n');
  // const [isLoading, setIsLoading] = useState(false);

  // const handleExecute = () => {
  //   setIsLoading(true);
  //   setOutput(prev => prev + `> Enviando código al servidor...\n`);

  //   try {
  //     const response = await fetch('http://localhost:8080/api/console', {
  //       method: 'POST',
  //       headers: {
  //         'Content-Type': 'application/json',
  //       },
  //       body: JSON.stringify({ text: inputCode }),
  //     });

  //     const data = await response.json();
      
  //     if (data.received) {
  //       setOutput(prev => prev + data.output);
  //     } else {
  //       setOutput(prev => prev + `> Error: ${data.message}\n`);
  //     }
  //   } catch (error) {
  //     setOutput(prev => prev + `> Error de conexión: ${error}\n`);
  //   } finally {
  //     setIsLoading(false);
  //   }
  // }

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