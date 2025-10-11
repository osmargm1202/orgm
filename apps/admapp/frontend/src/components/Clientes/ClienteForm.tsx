import React, { useState, useEffect } from 'react';
import { Card, Form, Button, Row, Col, Alert, Spinner, Image } from 'react-bootstrap';
import { Cliente, ClienteFormData, ClienteFormState } from '../../types/api';

// Importar las funciones de Wails desde el runtime generado
// @ts-ignore - Las funciones se generan en tiempo de compilación
import * as App from '../../../wailsjs/go/main/App';

interface ClienteFormProps {
  cliente: Cliente | null;
  isNew: boolean;
  isLoading: boolean;
  errors: Record<string, string>;
  logoPreview: string | null;
  onSave: (formData: ClienteFormData, logoFile: File | null) => void;
  onCancel: () => void;
  onDelete: (id: number) => void;
  onLogoChange: (file: File) => void;
}

const ClienteForm: React.FC<ClienteFormProps> = ({
  cliente,
  isNew,
  isLoading,
  errors,
  logoPreview,
  onSave,
  onCancel,
  onDelete,
  onLogoChange,
}) => {
  const [formData, setFormData] = useState<ClienteFormData>({
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
  });

  // Logo states
  const [logoUploading, setLogoUploading] = useState(false);
  const [logoMessage, setLogoMessage] = useState<{ type: 'success' | 'danger'; text: string } | null>(null);

  // Update form data when cliente changes
  useEffect(() => {
    if (cliente) {
      setFormData({
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
      });
    } else {
      setFormData({
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
      });
    }
  }, [cliente]);

  // Load logo when editing cliente
  useEffect(() => {
    if (formData.id && !isNew) {
      loadLogoPreview(formData.id);
    }
  }, [formData.id, isNew]);

  const loadLogoPreview = async (clienteId: number) => {
    try {
      const result = await App.GetLogoURL(clienteId);
      if (result.success && result.data?.url) {
        // Logo exists, it will be shown via logoPreview prop
        console.log('Logo cargado para cliente:', clienteId);
      }
    } catch (error) {
      // Logo doesn't exist, that's ok
      console.log('No hay logo para cliente:', clienteId);
    }
  };

  const handleInputChange = (field: keyof ClienteFormData, value: string) => {
    setFormData((prev) => ({
      ...prev,
      [field]: value,
    }));
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    onSave(formData, null); // Logo handling will be implemented separately
  };

  const handleUploadLogo = async () => {
    if (!formData.id) return;
    
    setLogoUploading(true);
    setLogoMessage(null);
    
    try {
      // Use OpenFile to get file path
      const fileResult = await App.OpenFile();
      if (!fileResult.success) {
        setLogoMessage({ type: 'danger', text: 'Error seleccionando archivo' });
        return;
      }
      
      const result = await App.UploadLogo(formData.id, fileResult.data);
      if (result.success) {
        setLogoMessage({ type: 'success', text: 'Logo actualizado exitosamente' });
        // Reload logo preview
        await loadLogoPreview(formData.id);
      } else {
        setLogoMessage({ type: 'danger', text: `Error al subir logo: ${result.error}` });
      }
    } catch (error) {
      setLogoMessage({ type: 'danger', text: `Error al subir logo: ${error}` });
    } finally {
      setLogoUploading(false);
    }
  };

  return (
    <Card className="mt-4" style={{ backgroundColor: '#2d3748' }}>
      <Card.Header>
        <h5 className="mb-0">
          {isNew ? 'Nuevo Cliente' : `Editar Cliente ${cliente?.id}`}
        </h5>
      </Card.Header>
      <Card.Body>
        <Form onSubmit={handleSubmit}>
          <Row>
            {/* Logo Upload */}
            <Col md={3}>
              <div className="text-center">
                <div className="mb-3">
                  {logoPreview ? (
                    <Image
                      src={logoPreview}
                      alt="Logo preview"
                      style={{ width: '300px', height: 'auto' }}
                      onError={(e) => {
                        console.error('❌ Error loading logo image:', logoPreview);
                        console.error('Error event:', e);
                      }}
                      onLoad={() => {
                        console.log('✅ Logo image loaded successfully:', logoPreview);
                      }}
                    />
                  ) : (
                    <div
                      className="d-flex align-items-center justify-content-center bg-light"
                      style={{ width: '300px', minHeight: '200px' }}
                    >
                      <i className="bi bi-image text-muted fs-1"></i>
                    </div>
                  )}
                </div>
                <Form.Group>
                  <Form.Label>Logo del Cliente</Form.Label>
                  <div className="mb-2">
                    <Button 
                      variant="secondary" 
                      size="sm"
                      disabled={logoUploading || !formData.id}
                      onClick={handleUploadLogo}
                    >
                      {logoUploading ? (
                        <>
                          <Spinner animation="border" size="sm" className="me-1" />
                          Subiendo...
                        </>
                      ) : (
                        'Seleccionar y Subir Logo'
                      )}
                    </Button>
                  </div>
                  {logoMessage && (
                    <Alert variant={logoMessage.type} className="py-2 mb-0">
                      <small>{logoMessage.text}</small>
                    </Alert>
                  )}
                </Form.Group>
              </div>
            </Col>

            {/* Form Fields */}
            <Col md={9}>
              <Row>
                <Col md={6}>
                  <Form.Group className="mb-3">
                    <Form.Label>ID</Form.Label>
                    <Form.Control
                      type="number"
                      value={formData.id || ''}
                      onChange={(e) => handleInputChange('id', e.target.value)}
                      disabled={!isNew}
                      isInvalid={!!errors.id}
                    />
                    <Form.Control.Feedback type="invalid">
                      {errors.id}
                    </Form.Control.Feedback>
                  </Form.Group>
                </Col>
                <Col md={6}>
                  <Form.Group className="mb-3">
                    <Form.Label>Tipo de Factura</Form.Label>
                    <Form.Select
                      value={formData.tipo_factura}
                      onChange={(e) => handleInputChange('tipo_factura', e.target.value)}
                      isInvalid={!!errors.tipo_factura}
                    >
                      <option value="NCFC">NCFC</option>
                      <option value="NCF">NCF</option>
                      <option value="NCG">NCG</option>
                      <option value="NCRE">NCRE</option>
                      <option value="NDC">NDC</option>
                      <option value="NDD">NDD</option>
                    </Form.Select>
                    <Form.Control.Feedback type="invalid">
                      {errors.tipo_factura}
                    </Form.Control.Feedback>
                  </Form.Group>
                </Col>
              </Row>

              <Row>
                <Col md={6}>
                  <Form.Group className="mb-3">
                    <Form.Label>Nombre *</Form.Label>
                    <Form.Control
                      type="text"
                      value={formData.nombre}
                      onChange={(e) => handleInputChange('nombre', e.target.value)}
                      placeholder="Nombre de la empresa"
                      isInvalid={!!errors.nombre}
                      required
                    />
                    <Form.Control.Feedback type="invalid">
                      {errors.nombre}
                    </Form.Control.Feedback>
                  </Form.Group>
                </Col>
                <Col md={6}>
                  <Form.Group className="mb-3">
                    <Form.Label>Nombre Comercial</Form.Label>
                    <Form.Control
                      type="text"
                      value={formData.nombre_comercial}
                      onChange={(e) => handleInputChange('nombre_comercial', e.target.value)}
                      placeholder="Nombre comercial"
                      isInvalid={!!errors.nombre_comercial}
                    />
                    <Form.Control.Feedback type="invalid">
                      {errors.nombre_comercial}
                    </Form.Control.Feedback>
                  </Form.Group>
                </Col>
              </Row>

              <Row>
                <Col md={6}>
                  <Form.Group className="mb-3">
                    <Form.Label>RNC *</Form.Label>
                    <Form.Control
                      type="text"
                      value={formData.numero}
                      onChange={(e) => handleInputChange('numero', e.target.value)}
                      placeholder="RNC-123456789"
                      isInvalid={!!errors.numero}
                      required
                    />
                    <Form.Control.Feedback type="invalid">
                      {errors.numero}
                    </Form.Control.Feedback>
                  </Form.Group>
                </Col>
                <Col md={6}>
                  <Form.Group className="mb-3">
                    <Form.Label>Correo</Form.Label>
                    <Form.Control
                      type="email"
                      value={formData.correo}
                      onChange={(e) => handleInputChange('correo', e.target.value)}
                      placeholder="correo@empresa.com"
                      isInvalid={!!errors.correo}
                    />
                    <Form.Control.Feedback type="invalid">
                      {errors.correo}
                    </Form.Control.Feedback>
                  </Form.Group>
                </Col>
              </Row>

              <Row>
                <Col md={12}>
                  <Form.Group className="mb-3">
                    <Form.Label>Dirección</Form.Label>
                    <Form.Control
                      type="text"
                      value={formData.direccion}
                      onChange={(e) => handleInputChange('direccion', e.target.value)}
                      placeholder="Dirección completa"
                      isInvalid={!!errors.direccion}
                    />
                    <Form.Control.Feedback type="invalid">
                      {errors.direccion}
                    </Form.Control.Feedback>
                  </Form.Group>
                </Col>
              </Row>

              <Row>
                <Col md={4}>
                  <Form.Group className="mb-3">
                    <Form.Label>Ciudad</Form.Label>
                    <Form.Control
                      type="text"
                      value={formData.ciudad}
                      onChange={(e) => handleInputChange('ciudad', e.target.value)}
                      placeholder="Ciudad"
                      isInvalid={!!errors.ciudad}
                    />
                    <Form.Control.Feedback type="invalid">
                      {errors.ciudad}
                    </Form.Control.Feedback>
                  </Form.Group>
                </Col>
                <Col md={4}>
                  <Form.Group className="mb-3">
                    <Form.Label>Provincia</Form.Label>
                    <Form.Control
                      type="text"
                      value={formData.provincia}
                      onChange={(e) => handleInputChange('provincia', e.target.value)}
                      placeholder="Provincia"
                      isInvalid={!!errors.provincia}
                    />
                    <Form.Control.Feedback type="invalid">
                      {errors.provincia}
                    </Form.Control.Feedback>
                  </Form.Group>
                </Col>
                <Col md={4}>
                  <Form.Group className="mb-3">
                    <Form.Label>Teléfono</Form.Label>
                    <Form.Control
                      type="tel"
                      value={formData.telefono}
                      onChange={(e) => handleInputChange('telefono', e.target.value)}
                      placeholder="809-555-1234"
                      isInvalid={!!errors.telefono}
                    />
                    <Form.Control.Feedback type="invalid">
                      {errors.telefono}
                    </Form.Control.Feedback>
                  </Form.Group>
                </Col>
              </Row>

              <Row>
                <Col md={6}>
                  <Form.Group className="mb-3">
                    <Form.Label>Representante</Form.Label>
                    <Form.Control
                      type="text"
                      value={formData.representante}
                      onChange={(e) => handleInputChange('representante', e.target.value)}
                      placeholder="Nombre del representante"
                      isInvalid={!!errors.representante}
                    />
                    <Form.Control.Feedback type="invalid">
                      {errors.representante}
                    </Form.Control.Feedback>
                  </Form.Group>
                </Col>
                <Col md={6}>
                  <Form.Group className="mb-3">
                    <Form.Label>Correo del Representante</Form.Label>
                    <Form.Control
                      type="email"
                      value={formData.correo_representante}
                      onChange={(e) => handleInputChange('correo_representante', e.target.value)}
                      placeholder="representante@empresa.com"
                      isInvalid={!!errors.correo_representante}
                    />
                    <Form.Control.Feedback type="invalid">
                      {errors.correo_representante}
                    </Form.Control.Feedback>
                  </Form.Group>
                </Col>
              </Row>
            </Col>
          </Row>

          {/* Action Buttons */}
          <div className="d-flex justify-content-end gap-2 mt-4">
            <Button
              variant="secondary"
              onClick={onCancel}
              disabled={isLoading}
            >
              <i className="bi bi-x-circle me-2"></i>
              Cancelar
            </Button>
            {!isNew && cliente && (
              <Button
                variant="danger"
                onClick={() => onDelete(cliente.id)}
                disabled={isLoading}
              >
                <i className="bi bi-trash me-2"></i>
                Eliminar
              </Button>
            )}
            <Button
              type="submit"
              variant="primary"
              disabled={isLoading}
              className="btn-primary"
            >
              {isLoading ? (
                <>
                  <Spinner size="sm" className="me-2" />
                  Guardando...
                </>
              ) : (
                <>
                  <i className="bi bi-check-circle me-2"></i>
                  Guardar
                </>
              )}
            </Button>
          </div>
        </Form>
      </Card.Body>
    </Card>
  );
};

export default ClienteForm;
