import React, { useState } from 'react';
import { Layout, Menu } from 'antd';
import { Link, useLocation } from 'react-router-dom';
import {
  DashboardOutlined,
  UserOutlined,
  DeviceOutlined,
  BarChartOutlined,
  CloudServerOutlined,
  SettingOutlined,
  SecurityOutlined,
  MenuFoldOutlined,
  MenuUnfoldOutlined,
} from '@ant-design/icons';

const { Sider } = Layout;

const Sidebar: React.FC = () => {
  const [collapsed, setCollapsed] = useState(false);
  const location = useLocation();

  const menuItems = [
    {
      key: '/',
      icon: <DashboardOutlined />,
      label: <Link to="/">Панель управления</Link>,
    },
    {
      key: '/users',
      icon: <UserOutlined />,
      label: <Link to="/users">Пользователи</Link>,
    },
    {
      key: '/devices',
      icon: <DeviceOutlined />,
      label: <Link to="/devices">Устройства</Link>,
    },
    {
      key: '/traffic',
      icon: <BarChartOutlined />,
      label: <Link to="/traffic">Трафик</Link>,
    },
    {
      key: '/nodes',
      icon: <CloudServerOutlined />,
      label: <Link to="/nodes">VPS Узлы</Link>,
    },
    {
      key: '/settings',
      icon: <SettingOutlined />,
      label: <Link to="/settings">Настройки</Link>,
    },
    {
      key: '/obfuscation',
      icon: <SecurityOutlined />,
      label: <Link to="/obfuscation">Обфускация DPI</Link>,
    },
  ];

  const getSelectedKey = () => {
    return location.pathname;
  };

  return (
    <Sider
      trigger={null}
      collapsible
      collapsed={collapsed}
      onCollapse={setCollapsed}
      style={{
        overflow: 'auto',
        height: '100vh',
        position: 'fixed',
        left: 0,
        top: 0,
        bottom: 0,
        background: '#001529',
      }}
    >
      <div
        style={{
          height: '64px',
          margin: '16px',
          background: 'rgba(255, 255, 255, 0.1)',
          borderRadius: '6px',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          color: 'white',
          fontSize: collapsed ? '16px' : '20px',
          fontWeight: 'bold',
          marginBottom: '16px',
        }}
      >
        {collapsed ? 'H2' : 'HysteryVPN'}
      </div>

      <Menu
        theme="dark"
        mode="inline"
        selectedKeys={[getSelectedKey()]}
        items={menuItems}
        style={{ border: 'none' }}
      />

      <div
        style={{
          position: 'absolute',
          bottom: '20px',
          left: '50%',
          transform: 'translateX(-50%)',
        }}
      >
        <div
          onClick={() => setCollapsed(!collapsed)}
          style={{
            cursor: 'pointer',
            color: 'white',
            fontSize: '16px',
            padding: '8px',
            borderRadius: '4px',
            background: 'rgba(255, 255, 255, 0.1)',
            textAlign: 'center',
          }}
        >
          {collapsed ? <MenuUnfoldOutlined /> : <MenuFoldOutlined />}
        </div>
      </div>
    </Sider>
  );
};

export default Sidebar;