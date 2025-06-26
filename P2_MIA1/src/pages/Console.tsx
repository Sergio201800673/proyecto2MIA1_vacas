import { useState } from 'react';
import './Console.css'; // Archivo de estilos que crearemos después

const Console = () => {
  const [inputCode, setInputCode] = useState('');
  const [output, setOutput] = useState('Esperando ejecución...\n');

  const handleExecute = () => {
    // Simulación de ejecución de código
    const newOutput = `> Ejecutando...\n${inputCode}\n> Proceso completado\n`;
    setOutput(prev => prev + newOutput);
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
      <button className="execute-button" onClick={handleExecute}>
        Ejecutar
      </button>
      
      {/* Área de salida */}
      <div className="console-output">
        <pre>{output}</pre>
      </div>
    </div>
  );
};

export default Console;