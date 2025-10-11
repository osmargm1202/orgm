import React, { useState, useEffect } from 'react';
import { Alert, Button } from 'react-bootstrap';
import ClientesList from '../Clientes/ClientesList';
import ProyectosList from './ProyectosList';
import ProyectoForm from './ProyectoForm';
import { Cliente, Proyecto, ProyectoFormData, ProyectoFormState, ProyectosListState, Cotizacion } from '../../types/api';

// Importar las funciones de Wails desde el runtime generado
// @ts-ignore - Las funciones se generan en tiempo de compilación
import * as App from '../../../wailsjs/go/main/App';

const ProyectosPage: React.FC = () => {
  // Client state
  const [clientes, setClientes] = useState<Cliente[]>([]);
  const [selectedCliente, setSelectedCliente] = useState<Cliente | null>(null);

  // Project state
  const [listState, setListState] = useState<ProyectosListState>({
    proyectos: [],
    filteredProyectos: [],
    searchTerm: '',
    idFilter: '',
    isLoading: false,
    selectedProyecto: null,
  });

  // Form state
  const [formState, setFormState] = useState<ProyectoFormState>({
    formData: {
      id: null,
      id_cliente: null,
      nombre_proyecto: '',
      ubicacion: '',
      descripcion: '',
    },
    isNew: false,
    isLoading: false,
    errors: {},
  });

  // UI state
  const [showForm, setShowForm] = useState(false);
  const [alert, setAlert] = useState<{ show: boolean; variant: string; message: string }>({
    show: false,
    variant: '',
    message: '',
  });

  // Load clients on component mount
  useEffect(() => {
    loadClientes();
  }, []);

  // Load projects when client is selected
  useEffect(() => {
    if (selectedCliente) {
      loadProyectos(selectedCliente.id);
    } else {
      setListState(prev => ({ ...prev, proyectos: [], selectedProyecto: null }));
      setShowForm(false);
    }
  }, [selectedCliente]);

  const loadClientes = async () => {
    try {
      const result = await App.GetClientes(false);
      if (result.success) {
        setClientes(result.data);
      } else {
        showAlert('danger', `Error cargando clientes: ${result.error}`);
      }
    } catch (error) {
      showAlert('danger', `Error cargando clientes: ${error}`);
    }
  };

  const loadProyectos = async (idCliente: number) => {
    setListState(prev => ({ ...prev, isLoading: true }));
    try {
      const result = await App.GetProyectos(idCliente, false);
      if (result.success) {
        setListState(prev => ({ ...prev, proyectos: result.data, isLoading: false }));
      } else {
        showAlert('danger', `Error cargando proyectos: ${result.error}`);
        setListState(prev => ({ ...prev, isLoading: false }));
      }
    } catch (error) {
      showAlert('danger', `Error cargando proyectos: ${error}`);
      setListState(prev => ({ ...prev, isLoading: false }));
    }
  };

  const showAlert = (variant: string, message: string) => {
    setAlert({ show: true, variant, message });
    setTimeout(() => {
      setAlert(prev => ({ ...prev, show: false }));
    }, 5000);
  };

  const handleClienteSelect = (cliente: Cliente) => {
    setSelectedCliente(cliente);
    setListState(prev => ({ ...prev, selectedProyecto: null }));
    setShowForm(false);
  };

  const handleProyectoSelect = (proyecto: Proyecto) => {
    setListState(prev => ({ ...prev, selectedProyecto: proyecto }));
    setFormState(prev => ({
      ...prev,
      isNew: false,
      formData: {
        id: proyecto.id,
        id_cliente: proyecto.id_cliente,
        nombre_proyecto: proyecto.nombre_proyecto,
        ubicacion: proyecto.ubicacion,
        descripcion: proyecto.descripcion,
      },
      errors: {},
    }));
    setShowForm(true);
  };

  const handleNewProyecto = () => {
    if (!selectedCliente) return;
    
    setFormState(prev => ({
      ...prev,
      isNew: true,
      formData: {
        id: null,
        id_cliente: selectedCliente.id,
        nombre_proyecto: '',
        ubicacion: '',
        descripcion: '',
      },
      errors: {},
    }));
    setShowForm(true);
  };

  const handleSave = async (formData: ProyectoFormData) => {
    setFormState(prev => ({ ...prev, isLoading: true, errors: {} }));

    try {
      let result;
      if (formState.isNew) {
        result = await App.CreateProyecto(
          formData.id_cliente!,
          formData.nombre_proyecto,
          formData.ubicacion,
          formData.descripcion
        );
      } else {
        result = await App.UpdateProyecto(
          formData.id!,
          formData.nombre_proyecto,
          formData.ubicacion,
          formData.descripcion
        );
      }

      if (result.success) {
        showAlert('success', formState.isNew ? 'Proyecto creado exitosamente' : 'Proyecto actualizado exitosamente');
        await loadProyectos(selectedCliente!.id);
        
        // Reload the proyecto data to keep the form open with updated data
        const savedProyectoId = result.data.id;
        const proyectoResult = await App.GetProyectoByID(savedProyectoId);
        
        if (proyectoResult.success) {
          handleProyectoSelect(proyectoResult.data);
        }
      } else {
        showAlert('danger', `Error: ${result.error}`);
      }
    } catch (error) {
      showAlert('danger', `Error: ${error}`);
    } finally {
      setFormState(prev => ({ ...prev, isLoading: false }));
    }
  };

  const handleCancel = () => {
    setShowForm(false);
    setListState(prev => ({ ...prev, selectedProyecto: null }));
  };

  const handleCancelCliente = () => {
    setSelectedCliente(null);
    setListState(prev => ({ 
      ...prev, 
      selectedProyecto: null,
      proyectos: []
    }));
    setShowForm(false);
  };

  const handleFormDataChange = (formData: ProyectoFormData) => {
    setFormState(prev => ({
      ...prev,
      formData: formData,
    }));
  };

  const handleCreateCotizacion = async () => {
    if (!listState.selectedProyecto) return;

    try {
      const result = await App.CreateCotizacionFromProyecto(listState.selectedProyecto.id, 1);
      if (result.success) {
        showAlert('success', `Cotización creada exitosamente con ID: ${result.data.id}`);
      } else {
        showAlert('danger', `Error creando cotización: ${result.error}`);
      }
    } catch (error) {
      showAlert('danger', `Error creando cotización: ${error}`);
    }
  };

  const handleIdFilterChange = (value: string) => {
    setListState(prev => ({ ...prev, idFilter: value }));
  };

  return (
    <div style={{ maxHeight: '100vh', overflowY: 'auto', paddingBottom: '60px' }}>
      <h2 className="mb-4">
        <i className="bi bi-folder me-2"></i>
        Gestión de Proyectos
      </h2>

      {/* Alert */}
      {alert.show && (
        <Alert
          variant={alert.variant}
          dismissible
          onClose={() => setAlert(prev => ({ ...prev, show: false }))}
        >
          {alert.message}
        </Alert>
      )}

      {/* Client Selection */}
      <div className="mb-4">
        <h5 className="mb-3">Seleccionar Cliente</h5>
        <ClientesList
          clientes={clientes}
          isLoading={false}
          selectedCliente={selectedCliente}
          includeInactive={false}
          onClienteSelect={handleClienteSelect}
          onNewCliente={() => {}}
          onIncludeInactiveChange={() => {}}
          onEdit={() => {}}
        />
      </div>

      {/* Project Management */}
      {selectedCliente && (
        <div className="mb-4">
          <div className="d-flex justify-content-between align-items-center mb-3">
            <div className="d-flex align-items-center">
              <h5 className="mb-0 me-3">Proyectos de {selectedCliente.nombre}</h5>
              <Button
                variant="outline-secondary"
                size="sm"
                onClick={handleCancelCliente}
              >
                <i className="bi bi-x-circle me-1"></i>
                Cancelar
              </Button>
            </div>
            {listState.selectedProyecto && (
              <Button
                variant="success"
                size="sm"
                onClick={handleCreateCotizacion}
              >
                <i className="bi bi-file-text me-1"></i>
                Crear Cotización
              </Button>
            )}
          </div>
          <ProyectosList
            proyectos={listState.proyectos}
            isLoading={listState.isLoading}
            selectedProyecto={listState.selectedProyecto}
            clienteSelected={!!selectedCliente}
            onProyectoSelect={handleProyectoSelect}
            onNewProyecto={handleNewProyecto}
            onIdFilterChange={handleIdFilterChange}
            idFilter={listState.idFilter}
          />
        </div>
      )}

      {/* Project Form */}
      {showForm && (
        <div style={{ marginBottom: '80px' }}>
        <ProyectoForm
          proyecto={listState.selectedProyecto}
          formData={formState.formData}
          isNew={formState.isNew}
          isLoading={formState.isLoading}
          errors={formState.errors}
          onFormDataChange={handleFormDataChange}
          onSave={handleSave}
          onCancel={handleCancel}
        />
        </div>
      )}
    </div>
  );
};

export default ProyectosPage;
