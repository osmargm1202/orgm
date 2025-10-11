import React from 'react';
import { Nav } from 'react-bootstrap';
import { MenuItem } from '../types/api';

interface SidebarProps {
  collapsed: boolean;
  activeMenu: string;
  onMenuSelect: (menuId: string) => void;
}

const Sidebar: React.FC<SidebarProps> = ({ collapsed, activeMenu, onMenuSelect }) => {
  const menuItems: MenuItem[] = [
    { id: 'clientes', label: 'Clientes', icon: 'bi-people' },
    { id: 'proyectos', label: 'Proyectos', icon: 'bi-folder' },
    { id: 'cotizaciones', label: 'Cotizaciones', icon: 'bi-file-text' },
    { id: 'pagos', label: 'Pagos', icon: 'bi-cash' },
    { id: 'estados', label: 'Estados', icon: 'bi-bar-chart' },
    { id: 'facturas', label: 'Facturas', icon: 'bi-receipt' },
    { id: 'comprobantes', label: 'Comprobantes', icon: 'bi-card-list' },
  ];

  return (
    <div className={`sidebar ${collapsed ? 'collapsed' : 'expanded'} position-fixed`}>
      <Nav className="flex-column px-2 mt-1" style={{ height: '100vh', paddingTop: '80px' }}>
        {menuItems.map((item) => (
          <Nav.Item key={item.id}>
            <Nav.Link
              className={`text-light d-flex align-items-center py-2 px-2 ${
                activeMenu === item.id ? 'active' : ''
              }`}
              onClick={() => onMenuSelect(item.id)}
              style={{
                borderRadius: '8px',
                marginBottom: '4px',
                backgroundColor: activeMenu === item.id ? 'rgba(122, 162, 247, 0.2)' : 'transparent',
                border: activeMenu === item.id ? '1px solid rgba(122, 162, 247, 0.4)' : 'none',
                transition: 'all 0.3s ease',
                cursor: 'pointer',
              }}
              onMouseEnter={(e) => {
                if (activeMenu !== item.id) {
                  e.currentTarget.style.backgroundColor = 'rgba(122, 162, 247, 0.1)';
                }
              }}
              onMouseLeave={(e) => {
                if (activeMenu !== item.id) {
                  e.currentTarget.style.backgroundColor = 'transparent';
                }
              }}
            >
              <i className={`bi ${item.icon} fs-5`} style={{ minWidth: '24px' }}></i>
              {!collapsed && (
                <span className="ms-3" style={{ whiteSpace: 'nowrap' }}>
                  {item.label}
                </span>
              )}
            </Nav.Link>
          </Nav.Item>
        ))}
      </Nav>
    </div>
  );
};

export default Sidebar;
