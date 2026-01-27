import React from 'react'
import { Routes, Route } from 'react-router-dom'
import { Layout } from 'antd'
import { AuthProvider } from './hooks/useAuth'
import Login from './pages/Login'
import Dashboard from './pages/Dashboard'
import Users from './pages/Users'
import Devices from './pages/Devices'
import Traffic from './pages/Traffic'
import NodesManagement from './pages/NodesManagement'
import Settings from './pages/Settings'
import ObfuscationSettings from './pages/ObfuscationSettings'
import Header from './components/common/Header'
import Sidebar from './components/common/Sidebar'
import './App.css'

const { Content } = Layout

function App() {
  return (
    <AuthProvider>
      <Layout style={{ minHeight: '100vh' }}>
        <Sidebar />
        <Layout>
          <Header />
          <Content style={{ margin: '16px', padding: '16px', background: '#fff' }}>
            <Routes>
              <Route path="/login" element={<Login />} />
              <Route path="/" element={<Dashboard />} />
              <Route path="/users" element={<Users />} />
              <Route path="/devices" element={<Devices />} />
              <Route path="/traffic" element={<Traffic />} />
               <Route path="/nodes" element={<NodesManagement />} />
               <Route path="/settings" element={<Settings />} />
               <Route path="/obfuscation" element={<ObfuscationSettings />} />
            </Routes>
          </Content>
        </Layout>
      </Layout>
    </AuthProvider>
  )
}

export default App