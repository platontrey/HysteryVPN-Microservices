import React, { useEffect } from 'react';
import { Card, List, Typography, Tag, Space, Badge } from 'antd';
import { WifiOutlined, ClockCircleOutlined } from '@ant-design/icons';
import { useWebSocket, TrafficUpdate } from '../hooks/useWebSocket';
import { useAuth } from '../hooks/useAuth';

const { Text, Title } = Typography;

interface RealTimeTrafficProps {
  className?: string;
}

const RealTimeTraffic: React.FC<RealTimeTrafficProps> = ({ className }) => {
  const { user } = useAuth();
  const { isConnected, trafficUpdates, connectionStatus, subscribeToTraffic } = useWebSocket(
    'localhost:8080/ws',
    localStorage.getItem('token') || ''
  );

  useEffect(() => {
    if (isConnected) {
      subscribeToTraffic();
    }
  }, [isConnected, subscribeToTraffic]);

  const formatBytes = (bytes: number): string => {
    if (bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
  };

  const formatTime = (timestamp: string): string => {
    const date = new Date(timestamp);
    const now = new Date();
    const diff = now.getTime() - date.getTime();

    if (diff < 60000) return 'Just now';
    if (diff < 3600000) return `${Math.floor(diff / 60000)}m ago`;
    if (diff < 86400000) return `${Math.floor(diff / 3600000)}h ago`;
    return date.toLocaleDateString();
  };

  const getConnectionStatusColor = (status: string): string => {
    switch (status) {
      case 'connected': return 'green';
      case 'connecting': return 'orange';
      case 'disconnected': return 'red';
      case 'error': return 'red';
      default: return 'gray';
    }
  };

  return (
    <Card
      className={className}
      title={
        <Space>
          <WifiOutlined />
          <span>Real-Time Traffic</span>
          <Badge
            status={isConnected ? 'success' : 'error'}
            text={
              <Text style={{ fontSize: '12px', color: getConnectionStatusColor(connectionStatus) }}>
                {connectionStatus.toUpperCase()}
              </Text>
            }
          />
        </Space>
      }
      extra={
        <Text type="secondary" style={{ fontSize: '12px' }}>
          <ClockCircleOutlined style={{ marginRight: 4 }} />
          Live Updates
        </Text>
      }
    >
      {!isConnected && (
        <div style={{ textAlign: 'center', padding: '20px', color: '#ff4d4f' }}>
          <WifiOutlined style={{ fontSize: '24px', marginBottom: '8px' }} />
          <div>WebSocket Disconnected</div>
          <Text type="secondary">Trying to reconnect...</Text>
        </div>
      )}

      {trafficUpdates.length === 0 ? (
        <div style={{ textAlign: 'center', padding: '40px', color: '#999' }}>
          <WifiOutlined style={{ fontSize: '32px', marginBottom: '16px', opacity: 0.5 }} />
          <Title level={5} style={{ margin: 0, color: '#999' }}>
            No Traffic Updates Yet
          </Title>
          <Text type="secondary">
            Traffic updates will appear here in real-time
          </Text>
        </div>
      ) : (
        <List
          dataSource={trafficUpdates}
          renderItem={(item: TrafficUpdate) => (
            <List.Item
              style={{
                padding: '12px 0',
                borderBottom: '1px solid #f0f0f0',
              }}
            >
              <List.Item.Meta
                title={
                  <Space direction="vertical" size={2}>
                    <Space>
                      <Text strong>Traffic Update</Text>
                      {item.device_id && (
                        <Tag size="small" color="blue">
                          Device: {item.device_id.slice(0, 8)}...
                        </Tag>
                      )}
                    </Space>
                    <Text type="secondary" style={{ fontSize: '12px' }}>
                      {formatTime(item.recorded_at)}
                    </Text>
                  </Space>
                }
                description={
                  <Space direction="vertical" size={4}>
                    <Space size="large">
                      <div>
                        <Text type="secondary" style={{ fontSize: '12px' }}>Upload</Text>
                        <br />
                        <Text strong style={{ color: '#52c41a' }}>
                          ↑ {formatBytes(item.upload)}
                        </Text>
                      </div>
                      <div>
                        <Text type="secondary" style={{ fontSize: '12px' }}>Download</Text>
                        <br />
                        <Text strong style={{ color: '#1890ff' }}>
                          ↓ {formatBytes(item.download)}
                        </Text>
                      </div>
                      <div>
                        <Text type="secondary" style={{ fontSize: '12px' }}>Total</Text>
                        <br />
                        <Text strong style={{ color: '#722ed1' }}>
                          {formatBytes(item.total)}
                        </Text>
                      </div>
                    </Space>
                  </Space>
                }
              />
            </List.Item>
          )}
        />
      )}

      {trafficUpdates.length > 0 && (
        <div style={{ marginTop: 16, padding: '12px', background: '#f6ffed', border: '1px solid #b7eb8f', borderRadius: '4px' }}>
          <Text strong style={{ color: '#52c41a' }}>
            ✓ Receiving {trafficUpdates.length} real-time update{trafficUpdates.length !== 1 ? 's' : ''}
          </Text>
        </div>
      )}
    </Card>
  );
};

export default RealTimeTraffic;