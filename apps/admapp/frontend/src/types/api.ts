// TypeScript interfaces for API models

export interface Cliente {
  id: number;
  id_tenant: number;
  nombre: string;
  nombre_comercial: string;
  numero: string; // RNC
  correo: string;
  direccion: string;
  ciudad: string;
  provincia: string;
  telefono: string;
  representante: string;
  correo_representante: string;
  tipo_factura: string;
  activo: boolean;
  created_at: string;
  updated_at: string;
}

export interface CreateClienteRequest {
  nombre: string;
  nombre_comercial: string;
  numero: string;
  correo: string;
  direccion: string;
  ciudad: string;
  provincia: string;
  telefono: string;
  representante: string;
  correo_representante: string;
  tipo_factura: string;
}

export interface UpdateClienteRequest {
  nombre: string;
  nombre_comercial: string;
  numero: string;
  correo: string;
  direccion: string;
  ciudad: string;
  provincia: string;
  telefono: string;
  representante: string;
  correo_representante: string;
  tipo_factura: string;
}

export interface LogoResponse {
  path: string;
}

export interface ApiResponse<T = any> {
  success: boolean;
  data?: T;
  error?: string;
}

// Menu item interface for sidebar
export interface MenuItem {
  id: string;
  label: string;
  icon: string;
  active?: boolean;
}

// Form state interfaces
export interface ClienteFormData {
  id: number | null;
  nombre: string;
  nombre_comercial: string;
  numero: string;
  correo: string;
  direccion: string;
  ciudad: string;
  provincia: string;
  telefono: string;
  representante: string;
  correo_representante: string;
  tipo_factura: string;
}

export interface ClienteFormState {
  formData: ClienteFormData;
  isNew: boolean;
  isLoading: boolean;
  errors: Record<string, string>;
  logoFile: File | null;
  logoPreview: string | null;
}

// List state interfaces
export interface ClientesListState {
  clientes: Cliente[];
  filteredClientes: Cliente[];
  searchTerm: string;
  idFilter: string;
  includeInactive: boolean;
  isLoading: boolean;
  selectedCliente: Cliente | null;
}

// Page state interface
export interface ClientesPageState {
  listState: ClientesListState;
  formState: ClienteFormState;
  sidebarCollapsed: boolean;
}

// Proyecto interfaces
export interface Proyecto {
  id: number;
  id_tenant: number;
  id_cliente: number;
  nombre_proyecto: string;
  ubicacion: string;
  descripcion: string;
  activo: boolean;
  created_at: string;
  updated_at: string;
}

export interface CreateProyectoRequest {
  id_cliente: number;
  nombre_proyecto: string;
  ubicacion: string;
  descripcion: string;
}

export interface UpdateProyectoRequest {
  nombre_proyecto: string;
  ubicacion: string;
  descripcion: string;
}

export interface ProyectoFormData {
  id: number | null;
  id_cliente: number | null;
  nombre_proyecto: string;
  ubicacion: string;
  descripcion: string;
}

export interface ProyectoFormState {
  formData: ProyectoFormData;
  isNew: boolean;
  isLoading: boolean;
  errors: Record<string, string>;
}

export interface ProyectosListState {
  proyectos: Proyecto[];
  filteredProyectos: Proyecto[];
  searchTerm: string;
  idFilter: string;
  isLoading: boolean;
  selectedProyecto: Proyecto | null;
}

export interface ProyectosPageState {
  clientes: Cliente[];
  selectedCliente: Cliente | null;
  listState: ProyectosListState;
  formState: ProyectoFormState;
  sidebarCollapsed: boolean;
}

// Cotizacion interfaces
export interface Cotizacion {
  id: number;
  id_tenant: number;
  id_cliente: number;
  id_proyecto: number;
  id_servicio: number;
  moneda: string;
  fecha: string;
  tasa_moneda: number;
  tiempo_entrega: string;
  avance: string;
  validez: number;
  estado: string;
  idioma: string;
  descripcion: string;
  retencion: string;
  descuentop: number;
  retencionp: number;
  itbisp: number;
  activo: boolean;
  created_at: string;
  updated_at: string;
  // Joined fields
  cliente_nombre?: string;
  proyecto_nombre?: string;
  servicio_nombre?: string;
}

export interface CotizacionFormData {
  id: number | null;
  id_cliente: number | null;
  id_proyecto: number | null;
  id_servicio: number | null;
  moneda: string;
  fecha: string;
  tasa_moneda: number;
  tiempo_entrega: string;
  avance: string;
  validez: number;
  estado: string;
  idioma: string;
  descripcion: string;
  retencion: string;
  descuentop: number;
  retencionp: number;
  itbisp: number;
}

export interface Totales {
  subtotal: number;
  descuentom: number;
  retencionm: number;
  itbism: number;
  total_sin_itbis: number;
  total: number;
}

export interface PagoAsignado {
  id: number;
  id_pago: number;
  monto: number;
  fecha: string;
}

export interface CotizacionFormState {
  formData: CotizacionFormData;
  isNew: boolean;
  isLoading: boolean;
  errors: Record<string, string>;
  totales: Totales | null;
  pagos: PagoAsignado[];
}

export interface CotizacionesListState {
  cotizaciones: Cotizacion[];
  filteredCotizaciones: Cotizacion[];
  searchTerm: string;
  idFilter: string;
  isLoading: boolean;
  selectedCotizacion: Cotizacion | null;
}

export interface CotizacionesPageState {
  listState: CotizacionesListState;
  formState: CotizacionFormState;
  showForm: boolean;
}
