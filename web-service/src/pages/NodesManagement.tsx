import React, { useState, useEffect } from 'react';
import {
  Table,
  Button,
  Space,
  Tag,
  Modal,
  Form,
  Input,
  Select,
  message,
  Popconfirm,
  Card,
  Statistic,
  Row,
  Col,
  Tooltip,
  Badge,
  Drawer,
  Descriptions,
  InputNumber,
} from 'antd';
import {
  PlusOutlined,
  EditOutlined,
  DeleteOutlined,
  ReloadOutlined,
  EyeOutlined,
  SettingOutlined,
  PlayCircleOutlined,
  PauseCircleOutlined,
} from '@ant-design/icons';
import { useAuth } from '../hooks/useAuth';
import { api } from '../services/api';

const { Option } = Select;
const { TextArea } = Input;

interface VPSNode {
  id: string;
  name: string;
  hostname: string;
  ip_address: string;
  location: string;
  country: string;
  grpc_port: number;
  status: 'online' | 'offline' | 'maintenance' | 'error';
  version: string;
  capabilities: Record<string, any>;
  created_at: string;
  last_heartbeat?: string;
  metadata: Record<string, any>;
}

interface NodeMetrics {
  id: string;
  node_id: string;
  cpu_usage: number;
  memory_usage: number;
  bandwidth_up: number;
  bandwidth_down: number;
  active_connections: number;
  recorded_at: string;
}

const NodesManagement: React.FC = () => {
  const { user } = useAuth();
  const [nodes, setNodes] = useState<VPSNode[]>([]);
  const [loading, setLoading] = useState(false);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);
  const [selectedNode, setSelectedNode] = useState<VPSNode | null>(null);
  const [nodeMetrics, setNodeMetrics] = useState<NodeMetrics[]>([]);
  const [drawerVisible, setDrawerVisible] = useState(false);
  const [modalVisible, setModalVisible] = useState(false);
  const [editingNode, setEditingNode] = useState<VPSNode | null>(null);
  const [form] = Form.useForm();

  const statusColors = {
    online: 'green',
    offline: 'red',
    maintenance: 'orange',
    error: 'red',
  };

  const statusIcons = {
    online: <PlayCircleOutlined />,
    offline: <PauseCircleOutlined />,
    maintenance: <SettingOutlined />,
    error: <ReloadOutlined spin />,
  };

  const countries = [
    { code: 'US', name: 'United States' },
    { code: 'GB', name: 'United Kingdom' },
    { code: 'DE', name: 'Germany' },
    { code: 'FR', name: 'France' },
    { code: 'NL', name: 'Netherlands' },
    { code: 'SG', name: 'Singapore' },
    { code: 'JP', name: 'Japan' },
    { code: 'AU', name: 'Australia' },
    { code: 'CA', name: 'Canada' },
    { code: 'IN', name: 'India' },
  ];

  useEffect(() => {
    fetchNodes();
  }, [page, pageSize]);

  const fetchNodes = async () => {
    try {
      setLoading(true);
      const response = await api.get('/nodes', {
        params: { page, limit: pageSize },
      });
      setNodes(response.data.nodes);
      setTotal(response.data.total);
    } catch (error) {
      message.error('Не удалось загрузить список узлов');
      console.error('Error fetching nodes:', error);
    } finally {
      setLoading(false);
    }
  };

  const fetchNodeMetrics = async (nodeId: string) => {
    try {
      const response = await api.get(`/nodes/${nodeId}/metrics?limit=50`);
      setNodeMetrics(response.data.metrics);
    } catch (error) {
      message.error('Не удалось загрузить метрики узла');
      console.error('Error fetching node metrics:', error);
    }
  };

  const handleCreateNode = async (values: any) => {
    try {
      const payload = {
        ...values,
        capabilities: values.capabilities ? JSON.parse(values.capabilities) : {},
        metadata: values.metadata ? JSON.parse(values.metadata) : {},
      };

      await api.post('/nodes', payload);
      message.success('Узел успешно создан');
      setModalVisible(false);
      form.resetFields();
      fetchNodes();
    } catch (error: any) {
      message.error(error.response?.data?.error || 'Не удалось создать узел');
    }
  };

  const handleUpdateNode = async (values: any) => {
    if (!editingNode) return;

    try {
      const payload = {
        ...values,
        capabilities: values.capabilities ? JSON.parse(values.capabilities) : editingNode.capabilities,
        metadata: values.metadata ? JSON.parse(values.metadata) : editingNode.metadata,
      };

      await api.put(`/nodes/${editingNode.id}`, payload);
      message.success('Узел успешно обновлен');
      setModalVisible(false);
      setEditingNode(null);
      form.resetFields();
      fetchNodes();
    } catch (error: any) {
      message.error(error.response?.data?.error || 'Не удалось обновить узел');
    }
  };

  const handleDeleteNode = async (nodeId: string) => {
    try {
      await api.delete(`/nodes/${nodeId}`);
      message.success('Узел успешно удален');
      fetchNodes();
    } catch (error: any) {
      message.error(error.response?.data?.error || 'Не удалось удалить узел');
    }
  };

  const handleRestartNode = async (nodeId: string) => {
    try {
      await api.post(`/nodes/${nodeId}/restart`);
      message.success('Команда перезагрузки отправлена на узел');
    } catch (error: any) {
      message.error(error.response?.data?.error || 'Не удалось перезагрузить узел');
    }
  };

  const openNodeDetails = (node: VPSNode) => {
    setSelectedNode(node);
    setDrawerVisible(true);
    fetchNodeMetrics(node.id);
  };

  const openCreateModal = () => {
    setEditingNode(null);
    setModalVisible(true);
    form.resetFields();
  };

  const openEditModal = (node: VPSNode) => {
    setEditingNode(node);
    setModalVisible(true);
    form.setFieldsValue({
      ...node,
      capabilities: JSON.stringify(node.capabilities, null, 2),
      metadata: JSON.stringify(node.metadata, null, 2),
    });
  };

  const formatBytes = (bytes: number) => {
    if (bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
  };

  const columns = [
    {
      title: 'Название',
      dataIndex: 'name',
      key: 'name',
      render: (text: string, record: VPSNode) => (
        <Space>
          <span>{text}</span>
          <Tag color={statusColors[record.status]} icon={statusIcons[record.status]}>
            {record.status}
          </Tag>
        </Space>
      ),
    },
    {
      title: 'Расположение',
      dataIndex: 'location',
      key: 'location',
      render: (text: string, record: VPSNode) => (
        <Space>
          <span>{text}</span>
          {record.country && <Tag>{record.country}</Tag>}
        </Space>
      ),
    },
    {
      title: 'IP адрес',
      dataIndex: 'ip_address',
      key: 'ip_address',
    },
    {
      title: 'Версия',
      dataIndex: 'version',
      key: 'version',
      render: (text: string) => text || 'N/A',
    },
    {
      title: 'Последний heartbeat',
      dataIndex: 'last_heartbeat',
      key: 'last_heartbeat',
      render: (text: string) => text ? new Date(text).toLocaleString() : 'Никогда',
    },
    {
      title: 'Действия',
      key: 'actions',
      render: (_, record: VPSNode) => (
        <Space>
          <Tooltip title="Просмотр деталей">
            <Button
              type="text"
              icon={<EyeOutlined />}
              onClick={() => openNodeDetails(record)}
            />
          </Tooltip>
          <Tooltip title="Редактировать">
            <Button
              type="text"
              icon={<EditOutlined />}
              onClick={() => openEditModal(record)}
            />
          </Tooltip>
          <Tooltip title="Перезагрузить">
            <Button
              type="text"
              icon={<ReloadOutlined />}
              onClick={() => handleRestartNode(record.id)}
            />
          </Tooltip>
          <Popconfirm
            title="Удалить этот узел?"
            description="Это действие нельзя отменить"
            onConfirm={() => handleDeleteNode(record.id)}
            okText="Да"
            cancelText="Нет"
          >
            <Tooltip title="Удалить">
              <Button
                type="text"
                danger
                icon={<DeleteOutlined />}
              />
            </Tooltip>
          </Popconfirm>
        </Space>
      ),
    },
  ];

  const getStats = () => {
    const onlineCount = nodes.filter(n => n.status === 'online').length;
    const offlineCount = nodes.filter(n => n.status === 'offline').length;
    const maintenanceCount = nodes.filter(n => n.status === 'maintenance').length;
    const errorCount = nodes.filter(n => n.status === 'error').length;

    return { onlineCount, offlineCount, maintenanceCount, errorCount };
  };

  const stats = getStats();

  return (
    <div>
      <Card style={{ marginBottom: 16 }}>
        <Row gutter={16}>
          <Col span={6}>
            <Statistic
              title="Всего узлов"
              value={total}
              prefix={<PlusOutlined />}
            />
          </Col>
          <Col span={6}>
            <Statistic
              title="Онлайн"
              value={stats.onlineCount}
              valueStyle={{ color: '#3f8600' }}
              prefix={<Badge status="success" />}
            />
          </Col>
          <Col span={6}>
            <Statistic
              title="Офлайн"
              value={stats.offlineCount}
              valueStyle={{ color: '#cf1322' }}
              prefix={<Badge status="error" />}
            />
          </Col>
          <Col span={6}>
            <Statistic
              title="На обслуживании"
              value={stats.maintenanceCount}
              valueStyle={{ color: '#fa8c16' }}
              prefix={<Badge status="warning" />}
            />
          </Col>
        </Row>
      </Card>

      <Card
        title="Управление VPS узлами"
        extra={
          <Button
            type="primary"
            icon={<PlusOutlined />}
            onClick={openCreateModal}
          >
            Создать узел
          </Button>
        }
      >
        <Table
          columns={columns}
          dataSource={nodes}
          rowKey="id"
          loading={loading}
          pagination={{
            current: page,
            pageSize,
            total,
            showSizeChanger: true,
            showQuickJumper: true,
            showTotal: (total, range) =>
              `${range[0]}-${range[1]} из ${total} узлов`,
            onChange: (newPage, newPageSize) => {
              setPage(newPage);
              setPageSize(newPageSize || 10);
            },
          }}
        />
      </Card>

      {/* Create/Edit Modal */}
      <Modal
        title={editingNode ? 'Редактировать узел' : 'Создать новый узел'}
        open={modalVisible}
        onCancel={() => {
          setModalVisible(false);
          setEditingNode(null);
          form.resetFields();
        }}
        footer={null}
        width={800}
      >
        <Form
          form={form}
          layout="vertical"
          onFinish={editingNode ? handleUpdateNode : handleCreateNode}
        >
          <Row gutter={16}>
            <Col span={12}>
              <Form.Item
                name="name"
                label="Название узла"
                rules={[{ required: true, message: 'Введите название узла' }]}
              >
                <Input placeholder="US-East-1" />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item
                name="hostname"
                label="Hostname"
                rules={[{ required: true, message: 'Введите hostname' }]}
              >
                <Input placeholder="us-east-1.vpn.local" />
              </Form.Item>
            </Col>
          </Row>

          <Row gutter={16}>
            <Col span={12}>
              <Form.Item
                name="ip_address"
                label="IP адрес"
                rules={[{ required: true, message: 'Введите IP адрес' }]}
              >
                <Input placeholder="192.168.1.100" />
              </Form.Item>
            </Col>
            <Col span={6}>
              <Form.Item
                name="grpc_port"
                label="GRPC порт"
                initialValue={50051}
              >
                <InputNumber min={1} max={65535} style={{ width: '100%' }} />
              </Form.Item>
            </Col>
            <Col span={6}>
              <Form.Item name="country" label="Страна">
                <Select placeholder="Выберите страну" allowClear>
                  {countries.map(country => (
                    <Option key={country.code} value={country.code}>
                      {country.name}
                    </Option>
                  ))}
                </Select>
              </Form.Item>
            </Col>
          </Row>

          <Form.Item name="location" label="Расположение">
            <Input placeholder="New York, USA" />
          </Form.Item>

          <Form.Item
            name="capabilities"
            label="Возможности (JSON)"
          >
            <TextArea
              rows={4}
              placeholder='{"max_users": 1000, "bandwidth_gbps": 10}'
            />
          </Form.Item>

          <Form.Item
            name="metadata"
            label="Метаданные (JSON)"
          >
            <TextArea
              rows={3}
              placeholder='{"provider": "AWS", "region": "us-east-1"}'
            />
          </Form.Item>

          <Form.Item style={{ marginBottom: 0, textAlign: 'right' }}>
            <Space>
              <Button onClick={() => setModalVisible(false)}>
                Отмена
              </Button>
              <Button type="primary" htmlType="submit">
                {editingNode ? 'Обновить' : 'Создать'}
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Modal>

      {/* Node Details Drawer */}
      <Drawer
        title={`Детали узла: ${selectedNode?.name}`}
        placement="right"
        size="large"
        onClose={() => {
          setDrawerVisible(false);
          setSelectedNode(null);
          setNodeMetrics([]);
        }}
        open={drawerVisible}
      >
        {selectedNode && (
          <div>
            <Descriptions title="Основная информация" bordered column={1}>
              <Descriptions.Item label="ID">{selectedNode.id}</Descriptions.Item>
              <Descriptions.Item label="Название">{selectedNode.name}</Descriptions.Item>
              <Descriptions.Item label="Hostname">{selectedNode.hostname}</Descriptions.Item>
              <Descriptions.Item label="IP адрес">{selectedNode.ip_address}</Descriptions.Item>
              <Descriptions.Item label="Расположение">
                {selectedNode.location}{' '}
                {selectedNode.country && <Tag>{selectedNode.country}</Tag>}
              </Descriptions.Item>
              <Descriptions.Item label="GRPC порт">{selectedNode.grpc_port}</Descriptions.Item>
              <Descriptions.Item label="Статус">
                <Tag color={statusColors[selectedNode.status]}>
                  {selectedNode.status}
                </Tag>
              </Descriptions.Item>
              <Descriptions.Item label="Версия">{selectedNode.version || 'N/A'}</Descriptions.Item>
              <Descriptions.Item label="Создан">
                {new Date(selectedNode.created_at).toLocaleString()}
              </Descriptions.Item>
              <Descriptions.Item label="Последний heartbeat">
                {selectedNode.last_heartbeat
                  ? new Date(selectedNode.last_heartbeat).toLocaleString()
                  : 'Никогда'}
              </Descriptions.Item>
            </Descriptions>

            {nodeMetrics.length > 0 && (
              <div style={{ marginTop: 24 }}>
                <h4>Последние метрики</h4>
                <Table
                  dataSource={nodeMetrics.slice(0, 10)}
                  rowKey="id"
                  size="small"
                  pagination={false}
                  columns={[
                    {
                      title: 'Время',
                      dataIndex: 'recorded_at',
                      key: 'recorded_at',
                      render: (text: string) => new Date(text).toLocaleString(),
                    },
                    {
                      title: 'CPU %',
                      dataIndex: 'cpu_usage',
                      key: 'cpu_usage',
                      render: (value: number) => `${value.toFixed(2)}%`,
                    },
                    {
                      title: 'RAM %',
                      dataIndex: 'memory_usage',
                      key: 'memory_usage',
                      render: (value: number) => `${value.toFixed(2)}%`,
                    },
                    {
                      title: 'Upload',
                      dataIndex: 'bandwidth_up',
                      key: 'bandwidth_up',
                      render: (value: number) => formatBytes(value),
                    },
                    {
                      title: 'Download',
                      dataIndex: 'bandwidth_down',
                      key: 'bandwidth_down',
                      render: (value: number) => formatBytes(value),
                    },
                    {
                      title: 'Соединения',
                      dataIndex: 'active_connections',
                      key: 'active_connections',
                    },
                  ]}
                />
              </div>
            )}
          </div>
        )}
      </Drawer>
    </div>
  );
};

export default NodesManagement;