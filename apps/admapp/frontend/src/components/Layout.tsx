import React, { useState } from 'react';
import Header from './Header';
import Sidebar from './Sidebar';
import ClientesPage from './Clientes/ClientesPage';
import ProyectosPage from './Proyectos/ProyectosPage';

const Layout: React.FC = () => {
  const [sidebarCollapsed, setSidebarCollapsed] = useState(false);
  const [activeMenu, setActiveMenu] = useState('clientes');

  const handleToggleSidebar = () => {
    setSidebarCollapsed(!sidebarCollapsed);
  };

  const handleMenuSelect = (menuId: string) => {
    setActiveMenu(menuId);
  };

  const renderActivePage = () => {
    switch (activeMenu) {
      case 'clientes':
        return <ClientesPage />;
      case 'proyectos':
        return <ProyectosPage />;
      default:
        return <ClientesPage />;
    }
  };

  return (
    <div className="d-flex">
      <Sidebar
        collapsed={sidebarCollapsed}
        activeMenu={activeMenu}
        onMenuSelect={handleMenuSelect}
      />
      <div className="flex-grow-1">
        <Header
          sidebarCollapsed={sidebarCollapsed}
          onToggleSidebar={handleToggleSidebar}
        />
        <main
          className={`main-content ${sidebarCollapsed ? 'sidebar-collapsed' : ''}`}
        >
          <div className="p-4">
            {renderActivePage()}
          </div>
        </main>
      </div>
    </div>
  );
};

export default Layout;
