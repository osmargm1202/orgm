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
