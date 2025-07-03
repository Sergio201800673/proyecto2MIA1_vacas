import React, { useEffect, useState } from 'react';

interface Disco {
  nombre: string;
}

const SeleccionDisco: React.FC = () => {
  const [discos, setDiscos] = useState<Disco[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [discoSeleccionado, setDiscoSeleccionado] = useState<string | null>(null);
  const [particiones, setParticiones] = useState<any[]>([]);
  const [cargandoPart, setCargandoPart] = useState(false);

  useEffect(() => {
    fetch('http://3.144.31.35:5000/discos')
      .then(res => res.json())
      .then(data => {
        setDiscos(data);
        setLoading(false);
      })
      .catch(() => {
        setError('No se pudieron cargar los discos');
        setLoading(false);
      });
  }, []);

  const handleSeleccionarDisco = (nombre: string) => {
    setDiscoSeleccionado(nombre);
    setCargandoPart(true);
    fetch(`http://3.144.31.35:5000/particiones?disco=${encodeURIComponent(nombre)}`)
      .then(res => res.json())
      .then(data => {
        setParticiones(data);
        setCargandoPart(false);
      })
      .catch(() => {
        setParticiones([]);
        setCargandoPart(false);
      });
  };

  if (loading) return <div>Cargando discos...</div>;
  if (error) return <div>{error}</div>;

  return (
    <div className="selector-disco">
      <h2>Visualizador del Sistema de Archivos</h2>
      <p>Seleccione el disco que desea visualizar:</p>
      <div className="discos-lista">
        {discos.map((disco, idx) => (
          <button
            key={idx}
            className="disco-btn"
            style={{display: 'flex', flexDirection: 'column', alignItems: 'center', margin: '10px', fontSize: '2.5rem', border: '1px solid #ccc', borderRadius: '8px', padding: '16px', background: '#222', color: '#fff', cursor: 'pointer'}}
            onClick={() => handleSeleccionarDisco(disco.nombre)}
          >
            <span role="img" aria-label="disco" style={{fontSize: '4rem', marginBottom: '8px'}}>💽</span>
            <div style={{fontSize: '1.1rem', marginTop: '4px'}}>{disco.nombre}</div>
          </button>
        ))}
      </div>

      {discoSeleccionado && (
        <div style={{marginTop: '2rem'}}>
          <h3>Particiones de {discoSeleccionado}:</h3>
          {cargandoPart ? (
            <div>Cargando particiones...</div>
          ) : particiones.length === 0 ? (
            <div>No hay particiones activas.</div>
          ) : (
            <ul>
              {particiones.map((p, i) => (
                <li key={i}>
                  <b>{p.nombre.trim()}</b> | Tipo: {p.tipo} | Tamaño: {p.tamano} bytes
                </li>
              ))}
            </ul>
          )}
        </div>
      )}
    </div>
  );
};

export default SeleccionDisco; 