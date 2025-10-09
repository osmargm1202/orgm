import React, { useState, useEffect } from 'react';
import { Card, Form, Button, Row, Col, Alert, Spinner, Image } from 'react-bootstrap';
import { Cliente, ClienteFormData, ClienteFormState } from '../../types/api';

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

  const handleLogoFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (file) {
      onLogoChange(file);
    }
  };

  return (
    <Card className="mt-4">
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
                      className="logo-preview"
                      fluid
                    />
                  ) : (
                    <div
                      className="logo-preview d-flex align-items-center justify-content-center bg-light"
                      style={{ minHeight: '100px' }}
                    >
                      <i className="bi bi-image text-muted fs-1"></i>
                    </div>
                  )}
                </div>
                <Form.Group>
                  <Form.Label>Logo del Cliente</Form.Label>
                  <Form.Control
                    type="file"
                    accept="image/*"
                    onChange={handleLogoFileChange}
                    size="sm"
                  />
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
