import React, { useState, useEffect } from 'react';
import { Table, Button, InputGroup, Form, Spinner } from 'react-bootstrap';
import { Proyecto, ProyectosListState } from '../../types/api';

interface ProyectosListProps {
  proyectos: Proyecto[];
  isLoading: boolean;
  selectedProyecto: Proyecto | null;
  clienteSelected: boolean;
  onProyectoSelect: (proyecto: Proyecto) => void;
  onNewProyecto: () => void;
  onIdFilterChange: (value: string) => void;
  idFilter: string;
}

const ProyectosList: React.FC<ProyectosListProps> = ({
  proyectos,
  isLoading,
  selectedProyecto,
  clienteSelected,
  onProyectoSelect,
  onNewProyecto,
  onIdFilterChange,
  idFilter,
}) => {
  const [searchTerm, setSearchTerm] = useState('');

  // Filter and sort proyectos
  useEffect(() => {
    let filtered = proyectos.filter(proyecto => proyecto.activo);

    // If a proyecto is selected, show only that proyecto
    if (selectedProyecto) {
      filtered = [selectedProyecto];
    } else {
      // Apply search filter
      if (searchTerm) {
        filtered = filtered.filter(proyecto =>
          proyecto.nombre_proyecto.toLowerCase().includes(searchTerm.toLowerCase()) ||
          proyecto.ubicacion.toLowerCase().includes(searchTerm.toLowerCase()) ||
          proyecto.descripcion.toLowerCase().includes(searchTerm.toLowerCase())
        );
      }

      // Apply ID filter
      if (idFilter) {
        const id = parseInt(idFilter);
        if (!isNaN(id)) {
          filtered = filtered.filter(proyecto => proyecto.id === id);
        }
      }

      // Sort by ID ascending
      filtered = filtered.sort((a, b) => a.id - b.id);
    }
  }, [proyectos, searchTerm, idFilter, selectedProyecto]);

  const handleSearchChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setSearchTerm(e.target.value);
  };

  const handleIdFilterChange = (value: string) => {
    onIdFilterChange(value);
  };

  return (
    <div className="card" style={{ backgroundColor: '#2d3748' }}>
      <div className="card-header">
        <div className="d-flex justify-content-between align-items-center">
          <h5 className="mb-0">Lista de Proyectos</h5>
          <Button
            variant="primary"
            size="sm"
            onClick={onNewProyecto}
            disabled={!clienteSelected || !!selectedProyecto}
          >
            <i className="bi bi-plus-circle me-1"></i>
            Nuevo Proyecto
          </Button>
        </div>
      </div>
      <div className="card-body">
        {/* Search and Filters */}
        <div className="row mb-3">
          <div className="col-md-8">
            <InputGroup>
              <InputGroup.Text>
                <i className="bi bi-search"></i>
              </InputGroup.Text>
              <Form.Control
                type="text"
                placeholder="Buscar por nombre, ubicación o descripción..."
                value={searchTerm}
                onChange={handleSearchChange}
              />
            </InputGroup>
          </div>
          <div className="col-md-4">
            <InputGroup>
              <InputGroup.Text>
                <i className="bi bi-hash"></i>
              </InputGroup.Text>
              <Form.Control
                type="number"
                placeholder="ID del proyecto"
                value={idFilter}
                onChange={(e) => handleIdFilterChange(e.target.value)}
              />
            </InputGroup>
          </div>
        </div>

        {/* Results Count */}
        <div className="mb-3">
          <small className="text-muted">
            Mostrando {proyectos.filter(p => p.activo).length} proyectos
          </small>
        </div>

        {/* Loading State */}
        {isLoading && (
          <div className="text-center py-4">
            <Spinner animation="border" variant="primary" />
            <p className="mt-2 text-muted">Cargando proyectos...</p>
          </div>
        )}

        {/* Table */}
        {!isLoading && (
          <div className="table-container">
            <Table hover striped className="mb-0 table-dark">
              <thead>
                <tr>
                  <th style={{ width: '80px' }}>ID</th>
                  <th style={{ width: '300px' }}>Nombre Proyecto</th>
                  <th style={{ width: '200px' }}>Ubicación</th>
                </tr>
              </thead>
              <tbody>
                {proyectos.filter(p => p.activo).length === 0 ? (
                  <tr>
                    <td colSpan={3} className="text-center text-muted py-4">
                      Seleccione un cliente para ver sus proyectos
                    </td>
                  </tr>
                ) : (
                  proyectos
                    .filter(p => p.activo)
                    .sort((a, b) => a.id - b.id)
                    .map((proyecto) => (
                      <tr
                        key={proyecto.id}
                        onClick={() => onProyectoSelect(proyecto)}
                        style={{
                          cursor: 'pointer',
                          backgroundColor: selectedProyecto?.id === proyecto.id ? 'rgba(122, 162, 247, 0.2)' : undefined,
                        }}
                      >
                        <td>{proyecto.id}</td>
                        <td>{proyecto.nombre_proyecto}</td>
                        <td>{proyecto.ubicacion}</td>
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

export default ProyectosList;
