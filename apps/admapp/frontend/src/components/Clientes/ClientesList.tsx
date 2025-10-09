import React, { useState, useEffect } from 'react';
import { Table, Form, Button, InputGroup, Spinner, Alert } from 'react-bootstrap';
import { Cliente, ClientesListState } from '../../types/api';

interface ClientesListProps {
  clientes: Cliente[];
  isLoading: boolean;
  selectedCliente: Cliente | null;
  includeInactive: boolean;
  onClienteSelect: (cliente: Cliente) => void;
  onNewCliente: () => void;
  onIncludeInactiveChange: (include: boolean) => void;
}

const ClientesList: React.FC<ClientesListProps> = ({
  clientes,
  isLoading,
  selectedCliente,
  includeInactive,
  onClienteSelect,
  onNewCliente,
  onIncludeInactiveChange,
}) => {
  const [searchTerm, setSearchTerm] = useState('');
  const [idFilter, setIdFilter] = useState('');
  const [filteredClientes, setFilteredClientes] = useState<Cliente[]>([]);

  // Filter clientes based on search term and ID filter
  useEffect(() => {
    let filtered = clientes;

    // Filter by search term (nombre, rnc, nombre_comercial)
    if (searchTerm) {
      filtered = filtered.filter(
        (cliente) =>
          cliente.nombre.toLowerCase().includes(searchTerm.toLowerCase()) ||
          cliente.numero.toLowerCase().includes(searchTerm.toLowerCase()) ||
          cliente.nombre_comercial.toLowerCase().includes(searchTerm.toLowerCase())
      );
    }

    // Filter by ID
    if (idFilter) {
      const id = parseInt(idFilter);
      if (!isNaN(id)) {
        filtered = filtered.filter((cliente) => cliente.id === id);
      }
    }

    setFilteredClientes(filtered);
  }, [clientes, searchTerm, idFilter]);

  const handleIdFilterChange = (value: string) => {
    setIdFilter(value);
    if (value) {
      const id = parseInt(value);
      if (!isNaN(id)) {
        const cliente = clientes.find((c) => c.id === id);
        if (cliente) {
          onClienteSelect(cliente);
        }
      }
    }
  };

  return (
    <div className="card">
      <div className="card-header">
        <div className="d-flex justify-content-between align-items-center">
          <h5 className="mb-0">Lista de Clientes</h5>
          <Button
            variant="primary"
            onClick={onNewCliente}
            className="btn-primary"
          >
            <i className="bi bi-plus-circle me-2"></i>
            Nuevo
          </Button>
        </div>
      </div>
      <div className="card-body">
        {/* Search and Filters */}
        <div className="row mb-3">
          <div className="col-md-6">
            <InputGroup>
              <InputGroup.Text>
                <i className="bi bi-search"></i>
              </InputGroup.Text>
              <Form.Control
                type="text"
                placeholder="Buscar por nombre, RNC o nombre comercial..."
                value={searchTerm}
                onChange={(e) => setSearchTerm(e.target.value)}
              />
            </InputGroup>
          </div>
          <div className="col-md-3">
            <InputGroup>
              <InputGroup.Text>
                <i className="bi bi-hash"></i>
              </InputGroup.Text>
              <Form.Control
                type="number"
                placeholder="ID del cliente"
                value={idFilter}
                onChange={(e) => handleIdFilterChange(e.target.value)}
              />
            </InputGroup>
          </div>
          <div className="col-md-3">
            <Form.Check
              type="checkbox"
              id="includeInactive"
              label="Incluir inactivos"
              checked={includeInactive}
              onChange={(e) => onIncludeInactiveChange(e.target.checked)}
            />
          </div>
        </div>

        {/* Results Count */}
        <div className="mb-3">
          <small className="text-muted">
            Mostrando {filteredClientes.length} de {clientes.length} clientes
          </small>
        </div>

        {/* Loading State */}
        {isLoading && (
          <div className="text-center py-4">
            <Spinner animation="border" variant="primary" />
            <p className="mt-2 text-muted">Cargando clientes...</p>
          </div>
        )}

        {/* Table */}
        {!isLoading && (
          <div className="table-responsive">
            <Table hover className="mb-0">
              <thead>
                <tr>
                  <th>ID</th>
                  <th>Nombre</th>
                  <th>RNC</th>
                  <th>Nombre Comercial</th>
                  <th>Representante</th>
                  <th>Estado</th>
                </tr>
              </thead>
              <tbody>
                {filteredClientes.length === 0 ? (
                  <tr>
                    <td colSpan={6} className="text-center text-muted py-4">
                      {searchTerm || idFilter
                        ? 'No se encontraron clientes con los criterios de búsqueda'
                        : 'No hay clientes registrados'}
                    </td>
                  </tr>
                ) : (
                  filteredClientes.map((cliente) => (
                    <tr
                      key={cliente.id}
                      className={`cursor-pointer ${
                        selectedCliente?.id === cliente.id ? 'table-active' : ''
                      }`}
                      onClick={() => onClienteSelect(cliente)}
                      style={{ cursor: 'pointer' }}
                    >
                      <td>{cliente.id}</td>
                      <td>{cliente.nombre}</td>
                      <td>{cliente.numero}</td>
                      <td>{cliente.nombre_comercial}</td>
                      <td>{cliente.representante}</td>
                      <td>
                        <span
                          className={`badge ${
                            cliente.activo ? 'bg-success' : 'bg-secondary'
                          }`}
                        >
                          {cliente.activo ? 'Activo' : 'Inactivo'}
                        </span>
                      </td>
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

export default ClientesList;
