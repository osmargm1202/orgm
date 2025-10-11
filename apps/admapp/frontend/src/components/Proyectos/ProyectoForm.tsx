import React, { useState, useEffect } from 'react';
import { Card, Form, Button, Row, Col, Alert, Spinner } from 'react-bootstrap';
import { Proyecto, ProyectoFormData, ProyectoFormState } from '../../types/api';

// Importar las funciones de Wails desde el runtime generado
// @ts-ignore - Las funciones se generan en tiempo de compilación
import * as App from '../../../wailsjs/go/main/App';

interface ProyectoFormProps {
  proyecto: Proyecto | null;
  formData: ProyectoFormData;
  isNew: boolean;
  isLoading: boolean;
  errors: Record<string, string>;
  onFormDataChange: (formData: ProyectoFormData) => void;
  onSave: (formData: ProyectoFormData) => Promise<void>;
  onCancel: () => void;
}

const ProyectoForm: React.FC<ProyectoFormProps> = ({
  proyecto,
  formData,
  isNew,
  isLoading,
  errors,
  onFormDataChange,
  onSave,
  onCancel,
}) => {

  const handleInputChange = (field: keyof ProyectoFormData, value: string | number | null) => {
    const newFormData = {
      ...formData,
      [field]: value,
    };
    onFormDataChange(newFormData);
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    await onSave(formData);
  };

  const handleCancel = () => {
    onCancel();
  };

  return (
    <Card className="mt-4" style={{ backgroundColor: '#2d3748' }}>
      <Card.Header>
        <h5 className="mb-0">
          {isNew ? 'Nuevo Proyecto' : `Editar Proyecto ${proyecto?.id}`}
        </h5>
      </Card.Header>
      <Card.Body>
        <Form onSubmit={handleSubmit}>
          <Row>
            <Col md={6}>
              <Form.Group className="mb-3">
                <Form.Label>Nombre del Proyecto *</Form.Label>
                <Form.Control
                  type="text"
                  value={formData.nombre_proyecto}
                  onChange={(e) => handleInputChange('nombre_proyecto', e.target.value)}
                  isInvalid={!!errors.nombre_proyecto}
                  placeholder="Ingrese el nombre del proyecto"
                />
                <Form.Control.Feedback type="invalid">
                  {errors.nombre_proyecto}
                </Form.Control.Feedback>
              </Form.Group>
            </Col>
            <Col md={6}>
              <Form.Group className="mb-3">
                <Form.Label>Ubicación *</Form.Label>
                <Form.Control
                  type="text"
                  value={formData.ubicacion}
                  onChange={(e) => handleInputChange('ubicacion', e.target.value)}
                  isInvalid={!!errors.ubicacion}
                  placeholder="Ingrese la ubicación del proyecto"
                />
                <Form.Control.Feedback type="invalid">
                  {errors.ubicacion}
                </Form.Control.Feedback>
              </Form.Group>
            </Col>
          </Row>

          <Row>
            <Col md={12}>
              <Form.Group className="mb-3">
                <Form.Label>Descripción</Form.Label>
                <Form.Control
                  as="textarea"
                  rows={3}
                  value={formData.descripcion}
                  onChange={(e) => handleInputChange('descripcion', e.target.value)}
                  isInvalid={!!errors.descripcion}
                  placeholder="Ingrese una descripción del proyecto (opcional)"
                />
                <Form.Control.Feedback type="invalid">
                  {errors.descripcion}
                </Form.Control.Feedback>
              </Form.Group>
            </Col>
          </Row>

          <div className="d-flex gap-2">
            <Button
              type="submit"
              variant="primary"
              disabled={isLoading}
            >
              {isLoading ? (
                <>
                  <Spinner animation="border" size="sm" className="me-1" />
                  Guardando...
                </>
              ) : (
                <>
                  <i className="bi bi-check-circle me-1"></i>
                  {isNew ? 'Crear Proyecto' : 'Actualizar Proyecto'}
                </>
              )}
            </Button>
            <Button
              type="button"
              variant="secondary"
              onClick={handleCancel}
              disabled={isLoading}
            >
              <i className="bi bi-x-circle me-1"></i>
              Cancelar
            </Button>
          </div>
        </Form>
      </Card.Body>
    </Card>
  );
};

export default ProyectoForm;
