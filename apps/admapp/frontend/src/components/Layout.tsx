import React, { useState } from 'react';
import Header from './Header';
import Sidebar from './Sidebar';

interface LayoutProps {
  children: React.ReactNode;
}

const Layout: React.FC<LayoutProps> = ({ children }) => {
  const [sidebarCollapsed, setSidebarCollapsed] = useState(false);
  const [activeMenu, setActiveMenu] = useState('clientes');

  const handleToggleSidebar = () => {
    setSidebarCollapsed(!sidebarCollapsed);
  };

  const handleMenuSelect = (menuId: string) => {
    setActiveMenu(menuId);
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
            {children}
          </div>
        </main>
      </div>
    </div>
  );
};

export default Layout;
