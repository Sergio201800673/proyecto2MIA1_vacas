import { useState } from 'react';
import './Console.css'; // Archivo de estilos que crearemos después

const Console = () => {
  const [inputCode, setInputCode] = useState('');
  const [output, setOutput] = useState('\n\nEsperando ejecución...\n');
  const [isLoading, setIsLoading] = useState(false);
  const [pendingDelete, setPendingDelete] = useState<{filename: string, fullPath: string} | null>(null);

  const handleExecute = async () => {
    setIsLoading(true);
    
    try {
      // Enviar el código al backend en el puerto 5000
      const response = await fetch('http://localhost:5000/execute', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json; charset=utf-8',
        },
        body: JSON.stringify({ code: inputCode }),
      });

      if (!response.ok) {
        throw new Error(`Error del servidor: ${response.status}`);
      }

      const result = await response.json();
      
      // Verificar si el resultado contiene una confirmación de eliminación
      if (result.output && result.output.includes('CONFIRM_DELETE:')) {
        const confirmMatch = result.output.match(/CONFIRM_DELETE:([^:]+):(.+)/);
        if (confirmMatch) {
          const filename = confirmMatch[1];
          const fullPath = confirmMatch[2];
          setPendingDelete({ filename, fullPath });
          
          const newOutput = `\n\n> Ejecutando...\n> Resultado:\n⚠️ ¿Está seguro que desea eliminar el disco ${filename}? (s/n): `;
          setOutput(prev => prev + newOutput);
          return;
        }
      }
      
      // Actualizar la salida con el resultado del backend
      const newOutput = `> Ejecutando...\n> Resultado:\n${result.output || result.message}\n`;
      setOutput(prev => prev + newOutput);
      
    } catch (error) {
      console.error('Error al ejecutar el código:', error);
      const errorMessage = error instanceof Error ? error.message : 'Error desconocido';
      const errorOutput = `> Error al ejecutar el código:\n${errorMessage}\n`;
      setOutput(prev => prev + errorOutput);
    } finally {
      setIsLoading(false);
    }
  };

  const handleConfirmDelete = async (confirmed: boolean) => {
    if (!pendingDelete) return;

    if (confirmed) {
      try {
        const response = await fetch('http://localhost:5000/confirm-delete', {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json; charset=utf-8',
          },
          body: JSON.stringify({
            filename: pendingDelete.filename,
            fullPath: pendingDelete.fullPath
          }),
        });

        if (!response.ok) {
          throw new Error(`Error del servidor: ${response.status}`);
        }

        const result = await response.json();
        const newOutput = `${result.message}\n`;
        setOutput(prev => prev + newOutput);
      } catch (error) {
        console.error('Error al confirmar eliminación:', error);
        const errorMessage = error instanceof Error ? error.message : 'Error desconocido';
        const errorOutput = `> Error al confirmar eliminación:\n${errorMessage}\n`;
        setOutput(prev => prev + errorOutput);
      }
    } else {
      const newOutput = `❌ Operación cancelada.\n`;
      setOutput(prev => prev + newOutput);
    }

    setPendingDelete(null);
  };

  const handleInputKeyPress = (e: React.KeyboardEvent<HTMLTextAreaElement>) => {
    if (pendingDelete && (e.key === 's' || e.key === 'S' || e.key === 'n' || e.key === 'N')) {
      e.preventDefault();
      const confirmed = e.key.toLowerCase() === 's';
      handleConfirmDelete(confirmed);
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
        onKeyPress={handleInputKeyPress}
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

      {/* Botones de confirmación (solo visibles cuando hay una eliminación pendiente) */}
      {pendingDelete && (
        <div className="confirmation-buttons">
          <p>¿Confirmar eliminación del disco {pendingDelete.filename}?</p>
          <button 
            className="confirm-button confirm-yes" 
            onClick={() => handleConfirmDelete(true)}
          >
            Sí (s)
          </button>
          <button 
            className="confirm-button confirm-no" 
            onClick={() => handleConfirmDelete(false)}
          >
            No (n)
          </button>
        </div>
      )}
      
      {/* Área de salida */}
      <div className="console-output">
        <pre>{output}</pre>
      </div>
    </div>
  );
};

export default Console;