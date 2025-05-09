import React, { useState, useEffect } from 'react';
import {
  Table,
  Input,
  Button,
  Space,
  Select,
  Tag,
  Typography,
  Divider,
  Tooltip,
  message
} from 'antd';
import { SearchOutlined, LinkOutlined, FileExcelOutlined } from '@ant-design/icons';
import * as XLSX from 'xlsx';

const { Text } = Typography;

const ProtocolPage = () => {
  const [data, setData] = useState([]);
  const [loading, setLoading] = useState(false);
  const [pagination, setPagination] = useState({
    current: 1,
    pageSize: 10,
    total: 0,
  });
  const [filters, setFilters] = useState({
    taskId: null,
    query: null,
    sourceId: null,
    title: null,
    sourceStatus: null,
  });

  const columns = [
    {
      title: 'ID задачи',
      dataIndex: 'taskId',
      key: 'taskId',
      width: 100,
    },
    {
      title: 'Поисковый запрос',
      dataIndex: 'query',
      key: 'query',
      ellipsis: true,
    },
    {
      title: 'ID источника',
      dataIndex: 'sourceId',
      key: 'sourceId',
      width: 120,
    },
    {
      title: 'Название источника',
      dataIndex: 'title',
      key: 'title',
      render: (text, record) => (
        <a
          href={record.url}
          target="_blank"
          rel="noopener noreferrer"
          style={{ display: 'flex', alignItems: 'center' }}
        >
          <LinkOutlined style={{ marginRight: 5 }} />
          {text || 'Без названия'}
        </a>
      ),
    },
    {
      title: 'Создан',
      dataIndex: 'createdAt',
      key: 'createdAt',
      render: (date) => new Date(date).toLocaleString(),
      width: 170,
    },
    {
      title: 'Обновлен',
      dataIndex: 'updatedAt',
      key: 'updatedAt',
      render: (date) => new Date(date).toLocaleString(),
      width: 170,
    },
    {
      title: 'Статус источника',
      dataIndex: 'sourceStatus',
      key: 'sourceStatus',
      render: (status) => (
        <Tag color={status === 'available' ? 'green' : 'red'}>
          {status === 'available' ? 'Доступен' : 'Недоступен'}
        </Tag>
      ),
      width: 140,
    },
    {
      title: 'ID запуска',
      dataIndex: 'launchId',
      key: 'launchId',
      width: 100,
    },
    {
      title: 'Номер запуска',
      dataIndex: 'launchNumber',
      key: 'launchNumber',
      width: 120,
    },
    {
      title: 'Время запуска',
      dataIndex: 'startedAt',
      key: 'startedAt',
      render: (date) => new Date(date).toLocaleString(),
      width: 170,
    },
    {
      title: 'Длительность',
      dataIndex: 'duration',
      key: 'duration',
      render: (duration) => duration ? formatDuration(duration) : '-',
      width: 130,
    },
    {
      title: 'Статус запуска',
      dataIndex: 'launchStatus',
      key: 'launchStatus',
      render: (status) => {
        let color = 'default';
        if (status === 'finished') color = 'green';
        if (status === 'failed') color = 'red';
        if (status === 'running') color = 'blue';

        return (
          <Tag color={color}>
            {status === 'finished' ? 'Завершен' :
              status === 'failed' ? 'Ошибка' :
                status === 'in_progress' ? 'В процессе' : status}
          </Tag>
        );
      },
      width: 140,
    },
    {
      title: 'Ошибка',
      dataIndex: 'launchErrorMsg',
      key: 'launchErrorMsg',
      render: (error) => error || '-',
      ellipsis: true,
    },
  ];

  const fetchProtocol = async (params = {}) => {
    setLoading(true);
    try {
      const { current, pageSize } = pagination;
      const payload = {
        limit: pageSize,
        offset: (current - 1) * pageSize,
        ...filters,
      };

      const response = await fetch('http://localhost:8080/get-protocol', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(payload)
      });

      const data = await response.json();

      const newProtocol = [...data.protocol];
      setData(newProtocol);
      setPagination({
        ...pagination,
        total: newProtocol.length >= pageSize ? current * pageSize + 1 : (current - 1) * pageSize + newProtocol.length,
      });
    } catch (error) {
      message.error('Ошибка загрузки протокола');
      console.error('Error fetching protocol:', error);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchProtocol();
  }, [pagination.current, pagination.pageSize]);

  const handleTableChange = (pagination) => {
    setPagination(pagination);
  };

  const handleFilterChange = (name, value) => {
    setFilters({
      ...filters,
      [name]: value,
    });
  };

  const handleFilterSubmit = () => {
    setPagination({
      ...pagination,
      current: 1,
    });
    fetchProtocol();
  };

  const handleResetFilters = async () => {
    setFilters({
      taskId: null,
      query: null,
      sourceId: null,
      title: null,
      sourceStatus: null,
    });
    setPagination({
      ...pagination,
      current: 1,
    });
    fetchProtocol();
  };

  return (
    <div style={{ padding: 24 }}>
      <h1>Протокол найденных источников</h1>

      <Divider orientation="left">Фильтры</Divider>

      <Space size="middle" style={{ marginBottom: 24 }}>
        <Input
          placeholder="ID задачи"
          value={filters.taskId || ''}
          onChange={(e) => handleFilterChange('taskId', e.target.value ? parseInt(e.target.value) : null)}
          style={{ width: 120 }}
        />

        <Input
          placeholder="Поисковый запрос"
          value={filters.query || ''}
          onChange={(e) => handleFilterChange('query', e.target.value || null)}
          style={{ width: 200 }}
        />

        <Input
          placeholder="ID источника"
          value={filters.sourceId || ''}
          onChange={(e) => handleFilterChange('sourceId', e.target.value ? parseInt(e.target.value) : null)}
          style={{ width: 120 }}
        />

        <Input
          placeholder="Название источника"
          value={filters.title || ''}
          onChange={(e) => handleFilterChange('title', e.target.value || null)}
          style={{ width: 200 }}
        />

        <Select
          placeholder="Статус источника"
          value={filters.sourceStatus}
          onChange={(value) => handleFilterChange('sourceStatus', value)}
          style={{ width: 180 }}
          allowClear
        >
          <Select.Option value="available">Доступен</Select.Option>
          <Select.Option value="unavailable">Недоступен</Select.Option>
        </Select>

        <Button
          type="primary"
          icon={<SearchOutlined />}
          onClick={handleFilterSubmit}
        >
          Поиск
        </Button>

        <Button onClick={handleResetFilters}>
          Сбросить
        </Button>
      </Space>

      <Divider />

      <Table
        columns={columns}
        dataSource={data}
        rowKey={(record) => `${record.taskId}-${record.sourceId}-${record.launchId}`}
        pagination={pagination}
        loading={loading}
        onChange={handleTableChange}
        scroll={{ x: 1800 }}
        bordered
      />

      <Button
        type="primary"
        icon={<FileExcelOutlined />}
        onClick={() => exportToExcel({...filters})}
        style={{ backgroundColor: '#1d6f42', borderColor: '#1d6f42' }}
      >
        Экспорт в Excel
      </Button>
    </div>
  );
};

const exportToExcel = async (payload) => {
  payload.limit = 0
  payload.offset = 0

  const response = await fetch('http://localhost:8080/get-protocol', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(payload)
  });

  const data = await response.json();

  const protocolToExcel = [...data.protocol]

  // Подготовка данных для экспорта
  const excelData = protocolToExcel.map(item => ({
    'ID задачи': item.taskId,
    'Поисковый запрос': item.query,
    'ID источника': item.sourceId,
    'Название источника': item.title,
    'URL источника': item.url,
    'Создан': new Date(item.createdAt).toLocaleString(),
    'Обновлен': new Date(item.updatedAt).toLocaleString(),
    'Статус источника': item.sourceStatus === 'available' ? 'Доступен' : 'Недоступен',
    'ID запуска': item.launchId,
    'Номер запуска': item.launchNumber,
    'Время запуска': new Date(item.startedAt).toLocaleString(),
    'Длительность': item.duration ? formatDuration(item.duration) : '-',
    'Статус запуска': item.launchStatus === 'finished' ? 'Завершен' :
      item.launchStatus === 'failed' ? 'Ошибка' :
        item.launchStatus === 'running' ? 'В процессе' : item.launchStatus,
    'Ошибка': item.launchErrorMsg || '-'
  }));

  // Создание рабочей книги
  const wb = XLSX.utils.book_new();
  const ws = XLSX.utils.json_to_sheet(excelData);

  // Добавление листа в книгу
  XLSX.utils.book_append_sheet(wb, ws, "Протокол");

  // Генерация файла и скачивание
  XLSX.writeFile(wb, `протокол_${new Date().toISOString().slice(0, 10)}.xlsx`);
};

const formatDuration = function(duration) {
  if (!duration) return '0 сек';

  const seconds = Math.floor(duration / 1000000000);
  const minutes = Math.floor(seconds / 60);
  const hours = Math.floor(minutes / 60);
  const days = Math.floor(hours / 24);

  const parts = [];

  if (days > 0) parts.push(`${days} д`);
  if (hours % 24 > 0) parts.push(`${hours % 24} ч`);
  if (minutes % 60 > 0) parts.push(`${minutes % 60} мин`);
  if (seconds % 60 > 0 || parts.length === 0) {
    parts.push(`${seconds % 60} сек`);
  }

  return parts.join(' ');
}

export default ProtocolPage;