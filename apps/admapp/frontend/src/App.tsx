import React from 'react';
import 'bootstrap/dist/css/bootstrap.min.css';
import 'bootstrap-icons/font/bootstrap-icons.css';
import './styles/tokyo-night.css';
import './App.css';
import Layout from './components/Layout';
import ClientesPage from './components/Clientes/ClientesPage';

function App() {
  return (
    <div className="App">
      <Layout>
        <ClientesPage />
      </Layout>
    </div>
  );
}

export default App;
