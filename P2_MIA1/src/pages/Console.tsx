import { useState } from 'react';
import './Console.css'; // Archivo de estilos que crearemos después

const Console = () => {
  const [inputCode, setInputCode] = useState('');
  const [output, setOutput] = useState('Esperando ejecución...\n');
  const [isLoading, setIsLoading] = useState(false);

  const handleExecute = async () => {
    setIsLoading(true);
    
    try {
      // Enviar el código al backend en el puerto 5000
      const response = await fetch('http://localhost:5000/execute', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ code: inputCode }),
      });

      if (!response.ok) {
        throw new Error(`Error del servidor: ${response.status}`);
      }

      const result = await response.json();
      
      // Actualizar la salida con el resultado del backend
      const newOutput = `> Ejecutando...\n${inputCode}\n> Resultado:\n${result.output || result.message}\n`;
      setOutput(prev => prev + newOutput);
      
    } catch (error) {
      console.error('Error al ejecutar el código:', error);
      const errorOutput = `> Error al ejecutar el código:\n${error.message}\n`;
      setOutput(prev => prev + errorOutput);
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="console-container">
      <h1>Consola de Ejecución</h1>
      
      {/* Área de entrada de texto */}
      <textarea
        className="console-input"
        value={inputCode}
        onChange={(e) => setInputCode(e.target.value)}
        placeholder="Escribe tu código aquí..."
      />
      
      {/* Botón de ejecución */}
      <button 
        className="execute-button" 
        onClick={handleExecute}
        disabled={isLoading}
      >
        {isLoading ? 'Ejecutando...' : 'Ejecutar'}
      </button>
      
      {/* Área de salida */}
      <div className="console-output">
        <pre>{output}</pre>
      </div>
    </div>
  );
};

export default Console;