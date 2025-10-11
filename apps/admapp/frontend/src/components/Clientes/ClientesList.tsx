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
  onEdit: (cliente: Cliente) => void;
}

const ClientesList: React.FC<ClientesListProps> = ({
  clientes,
  isLoading,
  selectedCliente,
  includeInactive,
  onClienteSelect,
  onNewCliente,
  onIncludeInactiveChange,
  onEdit,
}) => {
  const [searchTerm, setSearchTerm] = useState('');
  const [idFilter, setIdFilter] = useState('');
  const [filteredClientes, setFilteredClientes] = useState<Cliente[]>([]);

  // Filter clientes based on search term, ID filter, and selected cliente
  useEffect(() => {
    let filtered = clientes;

    // If a cliente is selected, show only that cliente
    if (selectedCliente) {
      filtered = [selectedCliente];
    } else {
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

      // Sort by ID ascending
      filtered = filtered.sort((a, b) => a.id - b.id);
    }

    setFilteredClientes(filtered);
  }, [clientes, searchTerm, idFilter, selectedCliente]);

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
    <div className="card" style={{ backgroundColor: '#2d3748' }}>
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
          <div className="col-md-6">
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
          <div className="table-container">
            <Table hover striped className="mb-0 table-dark">
              <thead>
                <tr>
                  <th style={{ width: '80px' }}>ID</th>
                  <th style={{ width: '200px' }}>Nombre</th>
                  <th style={{ width: '150px' }}>RNC</th>
                  <th style={{ width: '200px' }}>Nombre Comercial</th>
                  <th style={{ width: '200px' }}>Representante</th>
                </tr>
              </thead>
              <tbody>
                {filteredClientes.length === 0 ? (
                  <tr>
                    <td colSpan={5} className="text-center text-muted py-4">
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
