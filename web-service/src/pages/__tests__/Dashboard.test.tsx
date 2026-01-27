import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import { Card, Statistic, Progress } from 'antd';
import Dashboard from '../pages/Dashboard';

// Mock antd components to avoid rendering issues
vi.mock('antd', () => ({
  Row: ({ children, ...props }: any) => <div data-testid="row" {...props}>{children}</div>,
  Col: ({ children, ...props }: any) => <div data-testid="col" {...props}>{children}</div>,
  Card: ({ children, title, ...props }: any) => <div data-testid="card" data-title={title} {...props}>{children}</div>,
  Statistic: ({ title, value, prefix, suffix, valueStyle }: any) => (
    <div data-testid="statistic" data-title={title} data-value={value} data-prefix={prefix} data-suffix={suffix} style={valueStyle}>
      {title}: {value}
    </div>
  ),
  Progress: ({ percent, strokeColor, showInfo }: any) => (
    <div data-testid="progress" data-percent={percent} data-stroke-color={strokeColor} data-show-info={showInfo}>
      Progress: {percent}%
    </div>
  ),
}));

// Mock icons
vi.mock('@ant-design/icons', () => ({
  UserOutlined: () => <span data-testid="user-icon">User</span>,
  WifiOutlined: () => <span data-testid="wifi-icon">Wifi</span>,
  DatabaseOutlined: () => <span data-testid="database-icon">Database</span>,
  ClockCircleOutlined: () => <span data-testid="clock-icon">Clock</span>,
}));

// Mock the RealTimeTraffic component
vi.mock('../components/common/RealTimeTraffic', () => ({
  default: () => <div data-testid="real-time-traffic">Real Time Traffic Component</div>,
}));

describe('Dashboard', () => {
  it('renders the dashboard title', () => {
    render(<Dashboard />);
    expect(screen.getByText('Dashboard')).toBeInTheDocument();
  });

  it('displays total users statistic', () => {
    render(<Dashboard />);
    const totalUsersStat = screen.getByTestId('statistic');
    expect(totalUsersStat).toHaveAttribute('data-title', 'Total Users');
    expect(totalUsersStat).toHaveAttribute('data-value', '42');
    expect(totalUsersStat).toHaveAttribute('data-prefix', '[object Object]');
  });

  it('displays active users statistic', () => {
    render(<Dashboard />);
    const stats = screen.getAllByTestId('statistic');
    const activeUsersStat = stats.find(stat => stat.getAttribute('data-title') === 'Active Users');
    expect(activeUsersStat).toBeInTheDocument();
    expect(activeUsersStat).toHaveAttribute('data-value', '28');
    expect(activeUsersStat).toHaveAttribute('data-suffix', '/ 42');
  });

  it('displays total traffic statistic', () => {
    render(<Dashboard />);
    const stats = screen.getAllByTestId('statistic');
    const trafficStat = stats.find(stat => stat.getAttribute('data-title') === 'Total Traffic');
    expect(trafficStat).toBeInTheDocument();
    expect(trafficStat).toHaveAttribute('data-value', '1.2');
    expect(trafficStat).toHaveAttribute('data-suffix', 'GB');
  });

  it('displays online devices statistic', () => {
    render(<Dashboard />);
    const stats = screen.getAllByTestId('statistic');
    const devicesStat = stats.find(stat => stat.getAttribute('data-title') === 'Online Devices');
    expect(devicesStat).toBeInTheDocument();
    expect(devicesStat).toHaveAttribute('data-value', '15');
  });

  it('renders system health card with progress bars', () => {
    render(<Dashboard />);
    const systemHealthCard = screen.getByTestId('card');
    expect(systemHealthCard).toHaveAttribute('data-title', 'System Health');

    const progressBars = screen.getAllByTestId('progress');
    expect(progressBars).toHaveLength(3);

    // CPU Usage
    expect(progressBars[0]).toHaveAttribute('data-percent', '45');
    expect(progressBars[0]).toHaveAttribute('data-stroke-color', '#1890ff');

    // Memory Usage
    expect(progressBars[1]).toHaveAttribute('data-percent', '62');
    expect(progressBars[1]).toHaveAttribute('data-stroke-color', '#52c41a');

    // Disk Usage
    expect(progressBars[2]).toHaveAttribute('data-percent', '28');
    expect(progressBars[2]).toHaveAttribute('data-stroke-color', '#fa8c16');
  });

  it('renders real-time traffic component', () => {
    render(<Dashboard />);
    expect(screen.getByTestId('real-time-traffic')).toBeInTheDocument();
    expect(screen.getByText('Real Time Traffic Component')).toBeInTheDocument();
  });

  it('displays system health labels correctly', () => {
    render(<Dashboard />);
    expect(screen.getByText('CPU Usage')).toBeInTheDocument();
    expect(screen.getByText('45%')).toBeInTheDocument();
    expect(screen.getByText('Memory Usage')).toBeInTheDocument();
    expect(screen.getByText('62%')).toBeInTheDocument();
    expect(screen.getByText('Disk Usage')).toBeInTheDocument();
    expect(screen.getByText('28%')).toBeInTheDocument();
  });

  it('renders using responsive grid layout', () => {
    render(<Dashboard />);
    const rows = screen.getAllByTestId('row');
    expect(rows).toHaveLength(2); // Statistics row and charts row

    const cols = screen.getAllByTestId('col');
    expect(cols.length).toBeGreaterThan(0);
  });
});