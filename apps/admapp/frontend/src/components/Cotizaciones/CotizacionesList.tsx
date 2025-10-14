import React, { useState, useEffect } from 'react';
import { Card, Table, Button, Form, InputGroup, Alert } from 'react-bootstrap';
import { Cotizacion, CotizacionesListState } from '../../types/api';

interface CotizacionesListProps {
  cotizaciones: Cotizacion[];
  isLoading: boolean;
  selectedCotizacion: Cotizacion | null;
  onCotizacionSelect: (cotizacion: Cotizacion) => void;
  onNewCotizacion: () => void;
  onIdFilterChange: (id: string) => void;
  idFilter: string;
}

const CotizacionesList: React.FC<CotizacionesListProps> = ({
  cotizaciones,
  isLoading,
  selectedCotizacion,
  onCotizacionSelect,
  onNewCotizacion,
  onIdFilterChange,
  idFilter,
}) => {
  const [searchTerm, setSearchTerm] = useState('');
  const [filteredCotizaciones, setFilteredCotizaciones] = useState<Cotizacion[]>([]);

  // Filter cotizaciones based on search term and ID filter
  useEffect(() => {
    let filtered = cotizaciones;

    // Filter by search term (cliente, proyecto, servicio)
    if (searchTerm) {
      filtered = filtered.filter(cotizacion =>
        cotizacion.cliente_nombre?.toLowerCase().includes(searchTerm.toLowerCase()) ||
        cotizacion.proyecto_nombre?.toLowerCase().includes(searchTerm.toLowerCase()) ||
        cotizacion.servicio_nombre?.toLowerCase().includes(searchTerm.toLowerCase()) ||
        cotizacion.descripcion?.toLowerCase().includes(searchTerm.toLowerCase())
      );
    }

    // Filter by ID
    if (idFilter) {
      filtered = filtered.filter(cotizacion =>
        cotizacion.id.toString().includes(idFilter)
      );
    }

    // Sort by ID ascending
    filtered = filtered.sort((a, b) => a.id - b.id);

    setFilteredCotizaciones(filtered);
  }, [cotizaciones, searchTerm, idFilter]);

  const handleSearchChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setSearchTerm(e.target.value);
  };

  const handleIdFilterChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    onIdFilterChange(e.target.value);
  };

  const handleRowClick = (cotizacion: Cotizacion) => {
    onCotizacionSelect(cotizacion);
  };

  return (
    <div className="card" style={{ backgroundColor: '#2d3748' }}>
      <div className="card-header">
        <div className="d-flex justify-content-between align-items-center">
          <h5 className="mb-0">Cotizaciones</h5>
          <Button
            variant="primary"
            size="sm"
            onClick={onNewCotizacion}
            disabled={!!selectedCotizacion}
          >
            <i className="bi bi-plus-circle me-1"></i>
            Nueva Cotización
          </Button>
        </div>
      </div>
      <div className="card-body">
        <div className="row mb-3">
          <div className="col-md-6">
            <InputGroup>
              <InputGroup.Text>
                <i className="bi bi-search"></i>
              </InputGroup.Text>
              <Form.Control
                type="text"
                placeholder="Buscar por cliente, proyecto, servicio..."
                value={searchTerm}
                onChange={handleSearchChange}
              />
            </InputGroup>
          </div>
          <div className="col-md-3">
            <Form.Control
              type="text"
              placeholder="ID Cotización"
              value={idFilter}
              onChange={handleIdFilterChange}
            />
          </div>
        </div>

        {isLoading ? (
          <div className="text-center py-4">
            <div className="spinner-border text-primary" role="status">
              <span className="visually-hidden">Cargando...</span>
            </div>
          </div>
        ) : (
          <div className="table-responsive">
            <Table striped hover className="table-dark">
              <thead>
                <tr>
                  <th style={{ width: '80px' }}>ID</th>
                  <th style={{ width: '200px' }}>Cliente</th>
                  <th style={{ width: '200px' }}>Proyecto</th>
                  <th style={{ width: '150px' }}>Servicio</th>
                  <th style={{ width: '100px' }}>Estado</th>
                  <th style={{ width: '100px' }}>Moneda</th>
                  <th style={{ width: '120px' }}>Fecha</th>
                </tr>
              </thead>
              <tbody>
                {filteredCotizaciones.length === 0 ? (
                  <tr>
                    <td colSpan={7} className="text-center text-muted py-4">
                      {cotizaciones.length === 0 
                        ? 'No hay cotizaciones disponibles' 
                        : 'No se encontraron cotizaciones con los filtros aplicados'
                      }
                    </td>
                  </tr>
                ) : (
                  filteredCotizaciones.map((cotizacion) => (
                    <tr
                      key={cotizacion.id}
                      className={selectedCotizacion?.id === cotizacion.id ? 'table-active' : ''}
                      style={{ cursor: 'pointer' }}
                      onClick={() => handleRowClick(cotizacion)}
                    >
                      <td>{cotizacion.id}</td>
                      <td>{cotizacion.cliente_nombre || 'N/A'}</td>
                      <td>{cotizacion.proyecto_nombre || 'N/A'}</td>
                      <td>{cotizacion.servicio_nombre || 'N/A'}</td>
                      <td>
                        <span className={`badge ${
                          cotizacion.estado === 'GENERADA' ? 'bg-warning' :
                          cotizacion.estado === 'APROBADA' ? 'bg-success' :
                          cotizacion.estado === 'RECHAZADA' ? 'bg-danger' :
                          'bg-secondary'
                        }`}>
                          {cotizacion.estado}
                        </span>
                      </td>
                      <td>{cotizacion.moneda}</td>
                      <td>{new Date(cotizacion.fecha).toLocaleDateString()}</td>
                    </tr>
                  ))
                )}
              </tbody>
            </Table>
          </div>
        )}
      </div>
    </div>
  );
};

export default CotizacionesList;
