import React from 'react';
import { Link, useNavigate, useLocation } from 'react-router-dom';
import { LogOut, Palette } from 'lucide-react';

const Layout = ({ children }) => {
  const navigate = useNavigate();
  const location = useLocation();
  const isAuthPage = ['/login', '/register'].includes(location.pathname);

  const handleLogout = () => {
    localStorage.removeItem('access_token');
    localStorage.removeItem('refresh_token');
    navigate('/login');
  };

  return (
    <div className="layout">
      <header className="header">
        <div className="container header-content">
          <Link to="/" className="logo">
            <Palette className="logo-icon" size={24} />
            <span className="logo-text">Pixly</span>
          </Link>

          {!isAuthPage && (
            <button onClick={handleLogout} className="btn btn-secondary btn-sm">
              <LogOut size={16} style={{ marginRight: '0.5rem' }} />
              Logout
            </button>
          )}
        </div>
      </header>

      <main className="main">
        {children}
      </main>

      <style>{`
        .header {
          background-color: var(--color-surface);
          border-bottom: 1px solid var(--color-border);
          padding: var(--space-3) 0;
          position: sticky;
          top: 0;
          z-index: 10;
        }
        .header-content {
          display: flex;
          align-items: center;
          justify-content: space-between;
        }
        .logo {
          display: flex;
          align-items: center;
          gap: var(--space-2);
          font-weight: 700;
          font-size: 1.25rem;
          color: var(--color-text);
        }
        .logo-icon {
          color: var(--color-primary);
        }
        .main {
          min-height: calc(100vh - 60px);
          padding: var(--space-8) 0;
        }
        .btn-sm {
          padding: var(--space-1) var(--space-3);
          font-size: 0.875rem;
        }
      `}</style>
    </div>
  );
};

export default Layout;
