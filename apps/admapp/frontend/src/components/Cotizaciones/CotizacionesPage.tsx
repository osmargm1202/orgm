import React, { useState, useEffect } from 'react';
import { Alert } from 'react-bootstrap';
import CotizacionesList from './CotizacionesList';
import CotizacionForm from './CotizacionForm';
import { Cotizacion, CotizacionFormData, CotizacionesListState, CotizacionFormState, Totales, PagoAsignado } from '../../types/api';

// Import Wails App
import * as App from '../../../wailsjs/go/main/App';

const CotizacionesPage: React.FC = () => {
  const [listState, setListState] = useState<CotizacionesListState>({
    cotizaciones: [],
    filteredCotizaciones: [],
    searchTerm: '',
    idFilter: '',
    isLoading: false,
    selectedCotizacion: null,
  });

  const [formState, setFormState] = useState<CotizacionFormState>({
    formData: {
      id: null,
      id_cliente: null,
      id_proyecto: null,
      id_servicio: null,
      moneda: 'RD$',
      fecha: new Date().toISOString().split('T')[0],
      tasa_moneda: 1.0,
      tiempo_entrega: '30',
      avance: '60',
      validez: 30,
      estado: 'GENERADA',
      idioma: 'ES',
      descripcion: '',
      retencion: 'NINGUNA',
      descuentop: 0.0,
      retencionp: 0.0,
      itbisp: 0.0,
    },
    isNew: true,
    isLoading: false,
    errors: {},
    totales: null,
    pagos: [],
  });

  const [showForm, setShowForm] = useState(false);
  const [alert, setAlert] = useState<{ type: 'success' | 'danger'; message: string } | null>(null);

  // Load recent cotizaciones on mount
  useEffect(() => {
    loadCotizacionesRecientes();
  }, []);

  const loadCotizacionesRecientes = async () => {
    setListState(prev => ({ ...prev, isLoading: true }));
    try {
      const result: any = await App.GetCotizacionesRecientes(10);
      if (result.success) {
        setListState(prev => ({
          ...prev,
          cotizaciones: result.data,
          isLoading: false,
        }));
      } else {
        showAlert('danger', `Error al cargar cotizaciones: ${result.error}`);
        setListState(prev => ({ ...prev, isLoading: false }));
      }
    } catch (error) {
      showAlert('danger', `Error al cargar cotizaciones: ${error}`);
      setListState(prev => ({ ...prev, isLoading: false }));
    }
  };

  const handleCotizacionSelect = async (cotizacion: Cotizacion) => {
    setListState(prev => ({ ...prev, selectedCotizacion: cotizacion }));
    
    // Load full cotización data
    try {
      const result: any = await App.GetCotizacionFull(cotizacion.id);
      if (result.success) {
        const cotizacionFull = result.data;
        setFormState(prev => ({
          ...prev,
          formData: {
            id: cotizacionFull.cotizacion.id,
            id_cliente: cotizacionFull.cotizacion.id_cliente,
            id_proyecto: cotizacionFull.cotizacion.id_proyecto,
            id_servicio: cotizacionFull.cotizacion.id_servicio,
            moneda: cotizacionFull.cotizacion.moneda,
            fecha: cotizacionFull.cotizacion.fecha,
            tasa_moneda: cotizacionFull.cotizacion.tasa_moneda,
            tiempo_entrega: cotizacionFull.cotizacion.tiempo_entrega,
            avance: cotizacionFull.cotizacion.avance,
            validez: cotizacionFull.cotizacion.validez,
            estado: cotizacionFull.cotizacion.estado,
            idioma: cotizacionFull.cotizacion.idioma,
            descripcion: cotizacionFull.cotizacion.descripcion,
            retencion: cotizacionFull.cotizacion.retencion,
            descuentop: cotizacionFull.cotizacion.descuentop,
            retencionp: cotizacionFull.cotizacion.retencionp,
            itbisp: cotizacionFull.cotizacion.itbisp,
          },
          isNew: false,
          totales: cotizacionFull.totales,
        }));

        // Load pagos
        const pagosResult: any = await App.GetCotizacionPagos(cotizacion.id);
        if (pagosResult.success) {
          setFormState(prev => ({
            ...prev,
            pagos: pagosResult.data,
          }));
        }
      } else {
        showAlert('danger', `Error al cargar cotización: ${result.error}`);
      }
    } catch (error) {
      showAlert('danger', `Error al cargar cotización: ${error}`);
    }

    setShowForm(true);
  };

  const handleNewCotizacion = () => {
    setFormState(prev => ({
      ...prev,
      formData: {
        id: null,
        id_cliente: null,
        id_proyecto: null,
        id_servicio: null,
        moneda: 'RD$',
        fecha: new Date().toISOString().split('T')[0],
        tasa_moneda: 1.0,
        tiempo_entrega: '30',
        avance: '60',
        validez: 30,
        estado: 'GENERADA',
        idioma: 'ES',
        descripcion: '',
        retencion: 'NINGUNA',
        descuentop: 0.0,
        retencionp: 0.0,
        itbisp: 0.0,
      },
      isNew: true,
      errors: {},
      totales: null,
      pagos: [],
    }));
    setListState(prev => ({ ...prev, selectedCotizacion: null }));
    setShowForm(true);
  };

  const handleFormDataChange = (formData: CotizacionFormData) => {
    setFormState(prev => ({
      ...prev,
      formData,
    }));
  };

  const handleSave = async (formData: CotizacionFormData) => {
    setFormState(prev => ({ ...prev, isLoading: true, errors: {} }));

    try {
      let result: any;
      if (formState.isNew) {
        result = await App.CreateCotizacion(
          formData.id_cliente || 0,
          formData.id_proyecto || 0,
          formData.id_servicio || 0,
          formData.moneda,
          formData.fecha,
          formData.tasa_moneda,
          formData.tiempo_entrega,
          formData.avance,
          formData.validez,
          formData.estado,
          formData.idioma,
          formData.descripcion,
          formData.retencion,
          formData.descuentop,
          formData.retencionp,
          formData.itbisp
        );
      } else {
        result = await App.UpdateCotizacion(
          formData.id!,
          formData.moneda,
          formData.fecha,
          formData.tasa_moneda,
          formData.tiempo_entrega,
          formData.avance,
          formData.validez,
          formData.estado,
          formData.idioma,
          formData.descripcion,
          formData.retencion,
          formData.descuentop,
          formData.retencionp,
          formData.itbisp
        );
      }

      if (result.success) {
        showAlert('success', formState.isNew ? 'Cotización creada exitosamente' : 'Cotización actualizada exitosamente');
        
        // Reload cotizaciones
        await loadCotizacionesRecientes();
        
        // If it was a new cotización, reload the form with the saved data
        if (formState.isNew && result.data) {
          await handleCotizacionSelect(result.data);
        }
      } else {
        setFormState(prev => ({
          ...prev,
          isLoading: false,
          errors: { general: result.error },
        }));
        showAlert('danger', `Error al guardar: ${result.error}`);
      }
    } catch (error) {
      setFormState(prev => ({
        ...prev,
        isLoading: false,
        errors: { general: String(error) },
      }));
      showAlert('danger', `Error al guardar: ${error}`);
    }
  };

  const handleCancel = () => {
    setShowForm(false);
    setListState(prev => ({ ...prev, selectedCotizacion: null }));
    setFormState(prev => ({
      ...prev,
      totales: null,
      pagos: [],
    }));
  };

  const handleDelete = async () => {
    if (!formState.formData.id) return;

    if (window.confirm('¿Estás seguro de que deseas eliminar esta cotización?')) {
      setFormState(prev => ({ ...prev, isLoading: true }));
      try {
        const result: any = await App.DeleteCotizacion(formState.formData.id);
        if (result.success) {
          showAlert('success', 'Cotización eliminada exitosamente');
          await loadCotizacionesRecientes();
          handleCancel();
        } else {
          showAlert('danger', `Error al eliminar: ${result.error}`);
          setFormState(prev => ({ ...prev, isLoading: false }));
        }
      } catch (error) {
        showAlert('danger', `Error al eliminar: ${error}`);
        setFormState(prev => ({ ...prev, isLoading: false }));
      }
    }
  };

  const handlePrintPDF = async () => {
    if (!formState.formData.id) return;

    try {
      const result: any = await App.DownloadCotizacionPDF(formState.formData.id, formState.formData.idioma);
      if (result.success) {
        // Convert base64 to blob and download
        const pdfBlob = new Blob([Uint8Array.from(atob(result.data), c => c.charCodeAt(0))], { type: 'application/pdf' });
        const url = URL.createObjectURL(pdfBlob);
        const a = document.createElement('a');
        a.href = url;
        a.download = `cotizacion_${formState.formData.id}.pdf`;
        a.click();
        URL.revokeObjectURL(url);
        showAlert('success', 'PDF descargado exitosamente');
      } else {
        showAlert('danger', `Error al generar PDF: ${result.error}`);
      }
    } catch (error) {
      showAlert('danger', `Error al generar PDF: ${error}`);
    }
  };

  const handleCalculateTotales = async (descuentop: number, retencionp: number, itbisp: number) => {
    if (!formState.formData.id) return;

    try {
      const result: any = await App.CalcularTotalesCotizacion(formState.formData.id, descuentop, retencionp, itbisp);
      if (result.success) {
        setFormState(prev => ({
          ...prev,
          totales: result.data,
        }));
      }
    } catch (error) {
      console.error('Error calculating totales:', error);
    }
  };

  const handleIdFilterChange = (id: string) => {
    setListState(prev => ({ ...prev, idFilter: id }));
  };

  const showAlert = (type: 'success' | 'danger', message: string) => {
    setAlert({ type, message });
    setTimeout(() => setAlert(null), 5000);
  };

  return (
    <div style={{ maxHeight: '100vh', overflowY: 'auto', paddingBottom: '60px' }}>
      <h4 className="mb-4">Gestión de Cotizaciones</h4>

      {alert && (
        <Alert variant={alert.type} dismissible onClose={() => setAlert(null)}>
          {alert.message}
        </Alert>
      )}

      <CotizacionesList
        cotizaciones={listState.cotizaciones}
        isLoading={listState.isLoading}
        selectedCotizacion={listState.selectedCotizacion}
        onCotizacionSelect={handleCotizacionSelect}
        onNewCotizacion={handleNewCotizacion}
        onIdFilterChange={handleIdFilterChange}
        idFilter={listState.idFilter}
      />

      {showForm && (
        <div style={{ marginBottom: '80px' }}>
          <CotizacionForm
            cotizacion={listState.selectedCotizacion}
            formData={formState.formData}
            isNew={formState.isNew}
            isLoading={formState.isLoading}
            errors={formState.errors}
            totales={formState.totales}
            pagos={formState.pagos}
            onFormDataChange={handleFormDataChange}
            onSave={handleSave}
            onCancel={handleCancel}
            onDelete={handleDelete}
            onPrintPDF={handlePrintPDF}
            onCalculateTotales={handleCalculateTotales}
          />
        </div>
      )}
    </div>
  );
};

export default CotizacionesPage;
