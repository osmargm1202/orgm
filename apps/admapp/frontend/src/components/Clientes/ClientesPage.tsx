import React, { useState, useEffect } from 'react';
import { Alert, Spinner } from 'react-bootstrap';
import ClientesList from './ClientesList';
import ClienteForm from './ClienteForm';
import { Cliente, ClienteFormData, ClienteFormState, ClientesListState } from '../../types/api';

// Importar las funciones de Wails desde el runtime generado
// @ts-ignore - Las funciones se generan en tiempo de compilación
import * as App from '../../../wailsjs/go/main/App';

const ClientesPage: React.FC = () => {
  // List state
  const [listState, setListState] = useState<ClientesListState>({
    clientes: [],
    filteredClientes: [],
    searchTerm: '',
    idFilter: '',
    includeInactive: false,
    isLoading: false,
    selectedCliente: null,
  });

  // Form state
  const [formState, setFormState] = useState<ClienteFormState>({
    formData: {
      id: null,
      nombre: '',
      nombre_comercial: '',
      numero: '',
      correo: '',
      direccion: '',
      ciudad: '',
      provincia: '',
      telefono: '',
      representante: '',
      correo_representante: '',
      tipo_factura: 'NCFC',
    },
    isNew: true,
    isLoading: false,
    errors: {},
    logoFile: null,
    logoPreview: null,
  });

  // Alert state
  const [alert, setAlert] = useState<{ show: boolean; variant: string; message: string }>({
    show: false,
    variant: 'success',
    message: '',
  });

  // Show form state
  const [showForm, setShowForm] = useState(false);

  // Load clientes on component mount
  useEffect(() => {
    loadClientes();
  }, [listState.includeInactive]);

  const loadClientes = async () => {
    console.log('🔄 Cargando clientes...', { includeInactive: listState.includeInactive });
    setListState(prev => ({ ...prev, isLoading: true }));
    try {
      const result = await App.GetClientes(listState.includeInactive);
      console.log('📡 Respuesta de GetClientes:', result);
      if (result.success) {
        console.log('✅ Clientes cargados exitosamente:', result.data);
        setListState(prev => ({
          ...prev,
          clientes: result.data || [],
          isLoading: false,
        }));
      } else {
        console.error('❌ Error al cargar clientes:', result.error);
        showAlert('danger', `Error al cargar clientes: ${result.error}`);
        setListState(prev => ({ ...prev, isLoading: false }));
      }
    } catch (error) {
      console.error('💥 Excepción al cargar clientes:', error);
      showAlert('danger', `Error al cargar clientes: ${error}`);
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
    setShowForm(true);
    setListState(prev => ({ ...prev, selectedCliente: cliente }));
    setFormState(prev => ({
      ...prev,
      isNew: false,
      formData: {
        id: cliente.id,
        nombre: cliente.nombre,
        nombre_comercial: cliente.nombre_comercial,
        numero: cliente.numero,
        correo: cliente.correo,
        direccion: cliente.direccion,
        ciudad: cliente.ciudad,
        provincia: cliente.provincia,
        telefono: cliente.telefono,
        representante: cliente.representante,
        correo_representante: cliente.correo_representante,
        tipo_factura: cliente.tipo_factura,
      },
      errors: {},
      logoPreview: null, // Clear previous logo
    }));
    loadLogoPreview(cliente.id);
  };

  const handleNewCliente = () => {
    setShowForm(true);
    setListState(prev => ({ ...prev, selectedCliente: null }));
    setFormState({
      formData: {
        id: null,
        nombre: '',
        nombre_comercial: '',
        numero: '',
        correo: '',
        direccion: '',
        ciudad: '',
        provincia: '',
        telefono: '',
        representante: '',
        correo_representante: '',
        tipo_factura: 'NCFC',
      },
      isNew: true,
      isLoading: false,
      errors: {},
      logoFile: null,
      logoPreview: null,
    });
  };

  const handleIncludeInactiveChange = (include: boolean) => {
    setListState(prev => ({ ...prev, includeInactive: include }));
  };

  const handleEditCliente = (cliente: Cliente) => {
    handleClienteSelect(cliente);
  };

  const loadLogoPreview = async (clienteId: number) => {
    try {
      console.log('🔍 Loading logo for cliente:', clienteId);
      const result = await App.GetLogoURL(clienteId);
      console.log('📦 GetLogoURL result:', result);
      if (result.success && result.data?.url) {
        console.log('✅ Logo URL:', result.data.url);
        setFormState(prev => ({
          ...prev,
          logoPreview: result.data.url,
        }));
      } else {
        console.log('⚠️ No logo URL in response:', result);
      }
    } catch (error) {
      // Logo not found or error - this is not critical
      console.log('❌ Error loading logo for cliente:', clienteId, error);
    }
  };

  const handleSave = async (formData: ClienteFormData, logoFile: File | null) => {
    setFormState(prev => ({ ...prev, isLoading: true, errors: {} }));

    try {
      let result;
      if (formState.isNew) {
        // Create new cliente
        result = await App.CreateCliente(
          formData.nombre,
          formData.nombre_comercial,
          formData.numero,
          formData.correo,
          formData.direccion,
          formData.ciudad,
          formData.provincia,
          formData.telefono,
          formData.representante,
          formData.correo_representante,
          formData.tipo_factura
        );
      } else {
        // Update existing cliente
        result = await App.UpdateCliente(
          formData.id!,
          formData.nombre,
          formData.nombre_comercial,
          formData.numero,
          formData.correo,
          formData.direccion,
          formData.ciudad,
          formData.provincia,
          formData.telefono,
          formData.representante,
          formData.correo_representante,
          formData.tipo_factura
        );
      }

      if (result.success) {
        showAlert('success', formState.isNew ? 'Cliente creado exitosamente' : 'Cliente actualizado exitosamente');
        await loadClientes();
        
        // Reload the cliente data to keep the form open with updated data
        const savedClienteId = result.data.id;
        const clienteResult = await App.GetClienteByID(savedClienteId);
        
        if (clienteResult.success) {
          handleClienteSelect(clienteResult.data);
        }
      } else {
        showAlert('danger', `Error al guardar: ${result.error}`);
      }
    } catch (error) {
      showAlert('danger', `Error al guardar: ${error}`);
    } finally {
      setFormState(prev => ({ ...prev, isLoading: false }));
    }
  };

  const handleCancel = () => {
    setShowForm(false);
    setListState(prev => ({ ...prev, selectedCliente: null }));
    setFormState(prev => ({
      ...prev,
      logoPreview: null,
    }));
  };

  const handleDelete = async (id: number) => {
    if (window.confirm('¿Está seguro de que desea eliminar este cliente?')) {
      try {
        const result = await App.DeleteCliente(id);
        if (result.success) {
          showAlert('success', 'Cliente eliminado exitosamente');
          await loadClientes();
          handleNewCliente();
        } else {
          showAlert('danger', `Error al eliminar: ${result.error}`);
        }
      } catch (error) {
        showAlert('danger', `Error al eliminar: ${error}`);
      }
    }
  };

  const handleLogoChange = (file: File) => {
    setFormState(prev => ({
      ...prev,
      logoFile: file,
      logoPreview: URL.createObjectURL(file),
    }));
  };

  return (
    <div style={{ maxHeight: '100vh', overflowY: 'auto' }}>
      <h2 className="mb-4">
        <i className="bi bi-people me-2"></i>
        Gestión de Clientes
      </h2>

      {/* Alert */}
      {alert.show && (
        <Alert
          variant={alert.variant}
          dismissible
          onClose={() => setAlert(prev => ({ ...prev, show: false }))}
          className="mb-4"
        >
          {alert.message}
        </Alert>
      )}

      {/* Clientes List */}
      <ClientesList
        clientes={listState.clientes}
        isLoading={listState.isLoading}
        selectedCliente={listState.selectedCliente}
        includeInactive={listState.includeInactive}
        onClienteSelect={handleClienteSelect}
        onNewCliente={handleNewCliente}
        onIncludeInactiveChange={handleIncludeInactiveChange}
        onEdit={handleEditCliente}
      />

      {/* Cliente Form - Only show when showForm is true */}
      {showForm && (
        <ClienteForm
          cliente={listState.selectedCliente}
          isNew={formState.isNew}
          isLoading={formState.isLoading}
          errors={formState.errors}
          logoPreview={formState.logoPreview}
          onSave={handleSave}
          onCancel={handleCancel}
          onDelete={handleDelete}
          onLogoChange={handleLogoChange}
        />
      )}
    </div>
  );
};

export default ClientesPage;
