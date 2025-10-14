import React, { useState, useEffect } from 'react';
import { Card, Form, Row, Col, Button, Alert, InputGroup } from 'react-bootstrap';
import { Cotizacion, CotizacionFormData, Totales, PagoAsignado } from '../../types/api';

interface CotizacionFormProps {
  cotizacion: Cotizacion | null;
  formData: CotizacionFormData;
  isNew: boolean;
  isLoading: boolean;
  errors: Record<string, string>;
  totales: Totales | null;
  pagos: PagoAsignado[];
  onFormDataChange: (formData: CotizacionFormData) => void;
  onSave: (formData: CotizacionFormData) => Promise<void>;
  onCancel: () => void;
  onDelete: () => void;
  onPrintPDF: () => void;
  onCalculateTotales: (descuentop: number, retencionp: number, itbisp: number) => void;
}

const CotizacionForm: React.FC<CotizacionFormProps> = ({
  cotizacion,
  formData,
  isNew,
  isLoading,
  errors,
  totales,
  pagos,
  onFormDataChange,
  onSave,
  onCancel,
  onDelete,
  onPrintPDF,
  onCalculateTotales,
}) => {
  const [hasChanges, setHasChanges] = useState(false);

  // Calculate totales when percentage fields change
  useEffect(() => {
    if (formData.id && (formData.descuentop !== 0 || formData.retencionp !== 0 || formData.itbisp !== 0)) {
      onCalculateTotales(formData.descuentop, formData.retencionp, formData.itbisp);
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [formData.descuentop, formData.retencionp, formData.itbisp, formData.id]);

  const handleInputChange = (field: keyof CotizacionFormData, value: string | number | null) => {
    const newFormData = {
      ...formData,
      [field]: value,
    };
    onFormDataChange(newFormData);
    setHasChanges(true);
  };

  const handleSave = async () => {
    await onSave(formData);
    setHasChanges(false);
  };

  const handlePrintPDF = () => {
    if (hasChanges) {
      if (window.confirm('Tienes cambios sin guardar. ¿Deseas guardar antes de imprimir?')) {
        handleSave().then(() => {
          onPrintPDF();
        });
      } else {
        onPrintPDF();
      }
    } else {
      onPrintPDF();
    }
  };

  const calculatePaymentPercentage = () => {
    if (!totales || totales.total === 0) return 0;
    const totalPaid = pagos.reduce((sum, pago) => sum + pago.monto, 0);
    return (totalPaid / totales.total) * 100;
  };

  return (
    <Card className="mt-4" style={{ backgroundColor: '#2d3748' }}>
      <Card.Header>
        <h5 className="mb-0">
          {isNew ? 'Nueva Cotización' : `Editar Cotización #${cotizacion?.id}`}
        </h5>
      </Card.Header>
      <Card.Body>
        <Row>
          <Col md={8}>
            <Row>
              <Col md={6}>
                <Form.Group className="mb-3">
                  <Form.Label>Moneda</Form.Label>
                  <Form.Select
                    value={formData.moneda}
                    onChange={(e) => handleInputChange('moneda', e.target.value)}
                    isInvalid={!!errors.moneda}
                  >
                    <option value="RD$">RD$</option>
                    <option value="USD">USD</option>
                  </Form.Select>
                  {errors.moneda && <Form.Control.Feedback type="invalid">{errors.moneda}</Form.Control.Feedback>}
                </Form.Group>
              </Col>
              <Col md={6}>
                <Form.Group className="mb-3">
                  <Form.Label>Fecha</Form.Label>
                  <Form.Control
                    type="date"
                    value={formData.fecha}
                    onChange={(e) => handleInputChange('fecha', e.target.value)}
                    isInvalid={!!errors.fecha}
                  />
                  {errors.fecha && <Form.Control.Feedback type="invalid">{errors.fecha}</Form.Control.Feedback>}
                </Form.Group>
              </Col>
            </Row>

            <Row>
              <Col md={6}>
                <Form.Group className="mb-3">
                  <Form.Label>Tiempo de Entrega (días)</Form.Label>
                  <Form.Control
                    type="text"
                    value={formData.tiempo_entrega}
                    onChange={(e) => handleInputChange('tiempo_entrega', e.target.value)}
                    isInvalid={!!errors.tiempo_entrega}
                  />
                  {errors.tiempo_entrega && <Form.Control.Feedback type="invalid">{errors.tiempo_entrega}</Form.Control.Feedback>}
                </Form.Group>
              </Col>
              <Col md={6}>
                <Form.Group className="mb-3">
                  <Form.Label>Avance (%)</Form.Label>
                  <Form.Control
                    type="text"
                    value={formData.avance}
                    onChange={(e) => handleInputChange('avance', e.target.value)}
                    isInvalid={!!errors.avance}
                  />
                  {errors.avance && <Form.Control.Feedback type="invalid">{errors.avance}</Form.Control.Feedback>}
                </Form.Group>
              </Col>
            </Row>

            <Row>
              <Col md={6}>
                <Form.Group className="mb-3">
                  <Form.Label>Validez (días)</Form.Label>
                  <Form.Control
                    type="number"
                    value={formData.validez}
                    onChange={(e) => handleInputChange('validez', parseInt(e.target.value) || 0)}
                    isInvalid={!!errors.validez}
                  />
                  {errors.validez && <Form.Control.Feedback type="invalid">{errors.validez}</Form.Control.Feedback>}
                </Form.Group>
              </Col>
              <Col md={6}>
                <Form.Group className="mb-3">
                  <Form.Label>Estado</Form.Label>
                  <Form.Select
                    value={formData.estado}
                    onChange={(e) => handleInputChange('estado', e.target.value)}
                    isInvalid={!!errors.estado}
                  >
                    <option value="GENERADA">GENERADA</option>
                    <option value="APROBADA">APROBADA</option>
                    <option value="RECHAZADA">RECHAZADA</option>
                    <option value="CANCELADA">CANCELADA</option>
                  </Form.Select>
                  {errors.estado && <Form.Control.Feedback type="invalid">{errors.estado}</Form.Control.Feedback>}
                </Form.Group>
              </Col>
            </Row>

            <Row>
              <Col md={6}>
                <Form.Group className="mb-3">
                  <Form.Label>Idioma</Form.Label>
                  <Form.Select
                    value={formData.idioma}
                    onChange={(e) => handleInputChange('idioma', e.target.value)}
                    isInvalid={!!errors.idioma}
                  >
                    <option value="ES">Español</option>
                    <option value="EN">English</option>
                  </Form.Select>
                  {errors.idioma && <Form.Control.Feedback type="invalid">{errors.idioma}</Form.Control.Feedback>}
                </Form.Group>
              </Col>
              <Col md={6}>
                <Form.Group className="mb-3">
                  <Form.Label>Retención</Form.Label>
                  <Form.Select
                    value={formData.retencion}
                    onChange={(e) => handleInputChange('retencion', e.target.value)}
                    isInvalid={!!errors.retencion}
                  >
                    <option value="NINGUNA">NINGUNA</option>
                    <option value="10%">10%</option>
                    <option value="5%">5%</option>
                    <option value="2%">2%</option>
                  </Form.Select>
                  {errors.retencion && <Form.Control.Feedback type="invalid">{errors.retencion}</Form.Control.Feedback>}
                </Form.Group>
              </Col>
            </Row>

            <Form.Group className="mb-3">
              <Form.Label>Descripción</Form.Label>
              <Form.Control
                as="textarea"
                rows={3}
                value={formData.descripcion}
                onChange={(e) => handleInputChange('descripcion', e.target.value)}
                isInvalid={!!errors.descripcion}
              />
              {errors.descripcion && <Form.Control.Feedback type="invalid">{errors.descripcion}</Form.Control.Feedback>}
            </Form.Group>

            {/* Percentage Fields */}
            <Row>
              <Col md={4}>
                <Form.Group className="mb-3">
                  <Form.Label>Descuento (%)</Form.Label>
                  <Form.Control
                    type="number"
                    step="0.01"
                    value={formData.descuentop}
                    onChange={(e) => handleInputChange('descuentop', parseFloat(e.target.value) || 0)}
                    isInvalid={!!errors.descuentop}
                  />
                  {errors.descuentop && <Form.Control.Feedback type="invalid">{errors.descuentop}</Form.Control.Feedback>}
                </Form.Group>
              </Col>
              <Col md={4}>
                <Form.Group className="mb-3">
                  <Form.Label>Retención (%)</Form.Label>
                  <Form.Control
                    type="number"
                    step="0.01"
                    value={formData.retencionp}
                    onChange={(e) => handleInputChange('retencionp', parseFloat(e.target.value) || 0)}
                    isInvalid={!!errors.retencionp}
                  />
                  {errors.retencionp && <Form.Control.Feedback type="invalid">{errors.retencionp}</Form.Control.Feedback>}
                </Form.Group>
              </Col>
              <Col md={4}>
                <Form.Group className="mb-3">
                  <Form.Label>ITBIS (%)</Form.Label>
                  <Form.Control
                    type="number"
                    step="0.01"
                    value={formData.itbisp}
                    onChange={(e) => handleInputChange('itbisp', parseFloat(e.target.value) || 0)}
                    isInvalid={!!errors.itbisp}
                  />
                  {errors.itbisp && <Form.Control.Feedback type="invalid">{errors.itbisp}</Form.Control.Feedback>}
                </Form.Group>
              </Col>
            </Row>
          </Col>

          <Col md={4}>
            {/* Totales Section */}
            <Card className="mb-3" style={{ backgroundColor: '#1a202c' }}>
              <Card.Header>
                <h6 className="mb-0">Totales</h6>
              </Card.Header>
              <Card.Body>
                {totales ? (
                  <div>
                    <div className="d-flex justify-content-between mb-2">
                      <span>Subtotal:</span>
                      <span>{formData.moneda} {totales.subtotal.toLocaleString()}</span>
                    </div>
                    <div className="d-flex justify-content-between mb-2">
                      <span>Descuento:</span>
                      <span>{formData.moneda} {totales.descuentom.toLocaleString()}</span>
                    </div>
                    <div className="d-flex justify-content-between mb-2">
                      <span>Retención:</span>
                      <span>{formData.moneda} {totales.retencionm.toLocaleString()}</span>
                    </div>
                    <div className="d-flex justify-content-between mb-2">
                      <span>ITBIS:</span>
                      <span>{formData.moneda} {totales.itbism.toLocaleString()}</span>
                    </div>
                    <hr />
                    <div className="d-flex justify-content-between mb-2">
                      <span>Total sin ITBIS:</span>
                      <span>{formData.moneda} {totales.total_sin_itbis.toLocaleString()}</span>
                    </div>
                    <div className="d-flex justify-content-between fw-bold">
                      <span>Total:</span>
                      <span>{formData.moneda} {totales.total.toLocaleString()}</span>
                    </div>
                  </div>
                ) : (
                  <div className="text-muted">Los totales se calcularán automáticamente</div>
                )}
              </Card.Body>
            </Card>

            {/* Pagos Section */}
            <Card style={{ backgroundColor: '#1a202c' }}>
              <Card.Header>
                <h6 className="mb-0">Pagos Asignados</h6>
              </Card.Header>
              <Card.Body>
                {pagos.length > 0 ? (
                  <div>
                    {pagos.map((pago) => (
                      <div key={pago.id} className="d-flex justify-content-between mb-2">
                        <span>Pago #{pago.id_pago}:</span>
                        <span>{formData.moneda} {pago.monto.toLocaleString()}</span>
                      </div>
                    ))}
                    <hr />
                    <div className="d-flex justify-content-between">
                      <span>Total Pagado:</span>
                      <span>{formData.moneda} {pagos.reduce((sum, pago) => sum + pago.monto, 0).toLocaleString()}</span>
                    </div>
                    <div className="d-flex justify-content-between">
                      <span>Porcentaje:</span>
                      <span>{calculatePaymentPercentage().toFixed(1)}%</span>
                    </div>
                  </div>
                ) : (
                  <div className="text-muted">No hay pagos asignados</div>
                )}
              </Card.Body>
            </Card>
          </Col>
        </Row>

        <div className="d-flex justify-content-end gap-2 mt-4">
          <Button variant="outline-secondary" onClick={onCancel} disabled={isLoading}>
            Cancelar
          </Button>
          {!isNew && (
            <Button variant="outline-danger" onClick={onDelete} disabled={isLoading}>
              Eliminar
            </Button>
          )}
          <Button variant="outline-primary" onClick={handlePrintPDF} disabled={isLoading}>
            <i className="bi bi-file-pdf me-1"></i>
            Imprimir PDF
          </Button>
          <Button variant="primary" onClick={handleSave} disabled={isLoading}>
            {isLoading ? (
              <>
                <span className="spinner-border spinner-border-sm me-2" role="status" aria-hidden="true"></span>
                Guardando...
              </>
            ) : (
              'Guardar'
            )}
          </Button>
        </div>
      </Card.Body>
    </Card>
  );
};

export default CotizacionForm;
