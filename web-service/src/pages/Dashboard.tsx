import React from 'react';
import { Row, Col, Card, Statistic, Progress } from 'antd';
import { UserOutlined, WifiOutlined, DatabaseOutlined, ClockCircleOutlined } from '@ant-design/icons';
import RealTimeTraffic from '../components/common/RealTimeTraffic';

const Dashboard: React.FC = () => {
  // Mock data - in real app this would come from API
  const stats = {
    totalUsers: 42,
    activeUsers: 28,
    totalTraffic: 1.2, // GB
    onlineDevices: 15,
  };

  return (
    <div>
      <h1 style={{ marginBottom: 24 }}>Dashboard</h1>

      {/* Statistics Cards */}
      <Row gutter={[16, 16]} style={{ marginBottom: 24 }}>
        <Col xs={24} sm={12} md={6}>
          <Card>
            <Statistic
              title="Total Users"
              value={stats.totalUsers}
              prefix={<UserOutlined />}
              valueStyle={{ color: '#1890ff' }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={6}>
          <Card>
            <Statistic
              title="Active Users"
              value={stats.activeUsers}
              prefix={<WifiOutlined />}
              valueStyle={{ color: '#52c41a' }}
              suffix={`/ ${stats.totalUsers}`}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={6}>
          <Card>
            <Statistic
              title="Total Traffic"
              value={stats.totalTraffic}
              prefix={<DatabaseOutlined />}
              valueStyle={{ color: '#722ed1' }}
              suffix="GB"
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={6}>
          <Card>
            <Statistic
              title="Online Devices"
              value={stats.onlineDevices}
              prefix={<ClockCircleOutlined />}
              valueStyle={{ color: '#fa8c16' }}
            />
          </Card>
        </Col>
      </Row>

      {/* Charts and Real-time Data */}
      <Row gutter={[16, 16]}>
        <Col xs={24} lg={12}>
          <Card title="System Health" style={{ height: '100%' }}>
            <div style={{ marginBottom: 16 }}>
              <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 8 }}>
                <span>CPU Usage</span>
                <span>45%</span>
              </div>
              <Progress percent={45} strokeColor="#1890ff" showInfo={false} />
            </div>

            <div style={{ marginBottom: 16 }}>
              <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 8 }}>
                <span>Memory Usage</span>
                <span>62%</span>
              </div>
              <Progress percent={62} strokeColor="#52c41a" showInfo={false} />
            </div>

            <div>
              <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 8 }}>
                <span>Disk Usage</span>
                <span>28%</span>
              </div>
              <Progress percent={28} strokeColor="#fa8c16" showInfo={false} />
            </div>
          </Card>
        </Col>

        <Col xs={24} lg={12}>
          <RealTimeTraffic />
        </Col>
      </Row>
    </div>
  );
};

export default Dashboard;