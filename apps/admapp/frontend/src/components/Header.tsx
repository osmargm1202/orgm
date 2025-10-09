import React from 'react';
import { Container, Navbar, Image } from 'react-bootstrap';
import logo from '../assets/images/logo-universal.png';

interface HeaderProps {
  sidebarCollapsed: boolean;
  onToggleSidebar: () => void;
}

const Header: React.FC<HeaderProps> = ({ sidebarCollapsed, onToggleSidebar }) => {
  return (
    <Navbar className="header" fixed="top">
      <Container fluid className="px-3">
        <Navbar.Brand className="d-flex align-items-center">
          <button
            className="btn btn-link text-light me-3 p-0"
            onClick={onToggleSidebar}
            style={{ border: 'none', background: 'none' }}
          >
            <i className="bi bi-list fs-4"></i>
          </button>
          <Image src={logo} alt="Logo" height="40" className="me-3" />
          <span className="fs-4 fw-bold text-light">admAPP</span>
        </Navbar.Brand>
      </Container>
    </Navbar>
  );
};

export default Header;
