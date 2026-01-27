import React, { useState, useEffect } from 'react'
import {
  Card,
  Switch,
  Select,
  InputNumber,
  Button,
  Form,
  message,
  Row,
  Col,
  Divider,
  Alert,
  Spin,
  Tag,
} from 'antd'
import { SecurityOutlined, CheckCircleOutlined, WarningOutlined } from '@ant-design/icons'
import { api } from '../services/api'

const { Option } = Select

interface ObfuscationConfig {
  advanced_obfuscation_enabled: boolean
  quic_obfuscation: {
    enabled: boolean
    scramble_transform: boolean
    packet_padding: number
    timing_randomization: boolean
  }
  tls_fingerprint: {
    rotation_enabled: boolean
    fingerprints: string[]
  }
  vless_reality: {
    enabled: boolean
    targets: string[]
  }
  traffic_shaping: {
    enabled: boolean
    behavioral_randomization: boolean
  }
  multi_hop_enabled: boolean
}

const ObfuscationSettings: React.FC = () => {
  const [loading, setLoading] = useState(false)
  const [saving, setSaving] = useState(false)
  const [config, setConfig] = useState<ObfuscationConfig | null>(null)
  const [form] = Form.useForm()

  useEffect(() => {
    loadObfuscationConfig()
  }, [])

  const loadObfuscationConfig = async () => {
    setLoading(true)
    try {
      // Load config from all nodes (simplified - get from first node)
      const response = await api.get('/nodes')
      if (response.data.nodes && response.data.nodes.length > 0) {
        const nodeId = response.data.nodes[0].id
        const configResponse = await api.get(`/nodes/${nodeId}/obfuscation`)
        setConfig(configResponse.data)
        form.setFieldsValue(configResponse.data)
      }
    } catch (error) {
      message.error('Не удалось загрузить настройки обфускации')
    } finally {
      setLoading(false)
    }
  }

  const handleSave = async (values: any) => {
    setSaving(true)
    try {
      // Apply to all nodes
      const response = await api.get('/nodes')
      if (response.data.nodes) {
        for (const node of response.data.nodes) {
          await api.post(`/nodes/${node.id}/obfuscation`, values)
        }
      }
      message.success('Настройки обфускации сохранены')
      await loadObfuscationConfig()
    } catch (error) {
      message.error('Не удалось сохранить настройки')
    } finally {
      setSaving(false)
    }
  }

  const enableAdvancedObfuscation = async () => {
    try {
      const response = await api.get('/nodes')
      if (response.data.nodes) {
        for (const node of response.data.nodes) {
          await api.post(`/nodes/${node.id}/obfuscation/enable-advanced`)
        }
      }
      message.success('Расширенная обфускация включена')
      await loadObfuscationConfig()
    } catch (error) {
      message.error('Не удалось включить расширенную обфускацию')
    }
  }

  if (loading) {
    return (
      <div style={{ textAlign: 'center', padding: '50px' }}>
        <Spin size="large" />
        <p>Загрузка настроек обфускации...</p>
      </div>
    )
  }

  return (
    <div>
      <div style={{ marginBottom: '24px' }}>
        <h1 style={{ margin: 0 }}>
          <SecurityOutlined style={{ marginRight: '12px' }} />
          Настройки Обфускации DPI
        </h1>
        <p style={{ color: '#666', marginTop: '8px' }}>
          Конфигурация для обхода систем глубокого анализа пакетов в РФ
        </p>
      </div>

      <Alert
        message="Важно"
        description="Эти настройки помогут обходить блокировки VPN в России. Включение расширенной обфускации может снизить производительность на 15-20%."
        type="info"
        showIcon
        style={{ marginBottom: '24px' }}
      />

      <Row gutter={16}>
        <Col span={24}>
          <Card title="Быстрое включение" bordered={false}>
            <Button
              type="primary"
              size="large"
              onClick={enableAdvancedObfuscation}
              loading={saving}
              icon={<CheckCircleOutlined />}
            >
              Включить Расширенную Обфускацию
            </Button>
            <p style={{ marginTop: '12px', color: '#666' }}>
              Автоматически включает все рекомендуемые настройки для обхода DPI в РФ
            </p>
          </Card>
        </Col>
      </Row>

      <Divider />

      <Form
        form={form}
        layout="vertical"
        onFinish={handleSave}
        initialValues={config || {}}
      >
        <Row gutter={16}>
          <Col span={12}>
            <Card title="QUIC Обфускация" bordered={false}>
              <Form.Item
                name={['quic_obfuscation', 'enabled']}
                label="Включить QUIC обфускацию"
                valuePropName="checked"
              >
                <Switch />
              </Form.Item>

              <Form.Item
                name={['quic_obfuscation', 'scramble_transform']}
                label="Scramble Transform"
                valuePropName="checked"
              >
                <Switch />
              </Form.Item>

              <Form.Item
                name={['quic_obfuscation', 'packet_padding']}
                label="Packet Padding (bytes)"
              >
                <InputNumber min={1200} max={1500} />
              </Form.Item>

              <Form.Item
                name={['quic_obfuscation', 'timing_randomization']}
                label="Timing Randomization"
                valuePropName="checked"
              >
                <Switch />
              </Form.Item>
            </Card>
          </Col>

          <Col span={12}>
            <Card title="TLS Fingerprint Rotation" bordered={false}>
              <Form.Item
                name={['tls_fingerprint', 'rotation_enabled']}
                label="Включить ротацию fingerprints"
                valuePropName="checked"
              >
                <Switch />
              </Form.Item>

              <Form.Item
                name={['tls_fingerprint', 'fingerprints']}
                label="Fingerprints"
              >
                <Select mode="multiple" placeholder="Выберите fingerprints">
                  <Option value="chrome">Chrome</Option>
                  <Option value="firefox">Firefox</Option>
                  <Option value="safari">Safari</Option>
                  <Option value="edge">Edge</Option>
                </Select>
              </Form.Item>
            </Card>

            <Card title="VLESS Reality" bordered={false} style={{ marginTop: '16px' }}>
              <Form.Item
                name={['vless_reality', 'enabled']}
                label="Включить VLESS Reality"
                valuePropName="checked"
              >
                <Switch />
              </Form.Item>

              <Form.Item
                name={['vless_reality', 'targets']}
                label="Target Domains"
              >
                <Select mode="tags" placeholder="Введите домены">
                  <Option value="apple.com">apple.com</Option>
                  <Option value="google.com">google.com</Option>
                  <Option value="microsoft.com">microsoft.com</Option>
                </Select>
              </Form.Item>
            </Card>
          </Col>
        </Row>

        <Row gutter={16} style={{ marginTop: '16px' }}>
          <Col span={12}>
            <Card title="Traffic Shaping" bordered={false}>
              <Form.Item
                name={['traffic_shaping', 'enabled']}
                label="Включить traffic shaping"
                valuePropName="checked"
              >
                <Switch />
              </Form.Item>

              <Form.Item
                name={['traffic_shaping', 'behavioral_randomization']}
                label="Behavioral Randomization"
                valuePropName="checked"
              >
                <Switch />
              </Form.Item>
            </Card>
          </Col>

          <Col span={12}>
            <Card title="Дополнительные настройки" bordered={false}>
              <Form.Item
                name="multi_hop_enabled"
                label="Multi-Hop Routing"
                valuePropName="checked"
              >
                <Switch />
              </Form.Item>

              <Form.Item
                name="advanced_obfuscation_enabled"
                label="Расширенная обфускация"
                valuePropName="checked"
              >
                <Switch />
              </Form.Item>
            </Card>
          </Col>
        </Row>

        <div style={{ marginTop: '24px', textAlign: 'right' }}>
          <Button type="primary" htmlType="submit" loading={saving} size="large">
            Сохранить настройки
          </Button>
        </div>
      </Form>

      {config && (
        <Card title="Статус обфускации" bordered={false} style={{ marginTop: '24px' }}>
          <Row gutter={16}>
            <Col span={6}>
              <div style={{ textAlign: 'center' }}>
                <Tag color={config.advanced_obfuscation_enabled ? 'green' : 'red'}>
                  {config.advanced_obfuscation_enabled ? 'Включена' : 'Отключена'}
                </Tag>
                <p>Расширенная обфускация</p>
              </div>
            </Col>
            <Col span={6}>
              <div style={{ textAlign: 'center' }}>
                <Tag color={config.quic_obfuscation.enabled ? 'green' : 'red'}>
                  {config.quic_obfuscation.enabled ? 'Включена' : 'Отключена'}
                </Tag>
                <p>QUIC обфускация</p>
              </div>
            </Col>
            <Col span={6}>
              <div style={{ textAlign: 'center' }}>
                <Tag color={config.tls_fingerprint.rotation_enabled ? 'green' : 'red'}>
                  {config.tls_fingerprint.rotation_enabled ? 'Включена' : 'Отключена'}
                </Tag>
                <p>TLS Fingerprint</p>
              </div>
            </Col>
            <Col span={6}>
              <div style={{ textAlign: 'center' }}>
                <Tag color={config.vless_reality.enabled ? 'green' : 'red'}>
                  {config.vless_reality.enabled ? 'Включена' : 'Отключена'}
                </Tag>
                <p>VLESS Reality</p>
              </div>
            </Col>
          </Row>
        </Card>
      )}
    </div>
  )
}

export default ObfuscationSettings