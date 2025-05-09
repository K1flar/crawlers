import React, { useState, useEffect, useRef } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { Button, Card, Divider, InputNumber, message, Space, Tag, Typography, Table, Spin } from 'antd';
import {
  PlayCircleOutlined,
  PauseCircleOutlined,
  ClockCircleOutlined,
  InfoCircleOutlined,
  CloseCircleOutlined,
  ArrowLeftOutlined
} from '@ant-design/icons';
import dayjs from 'dayjs';
import utc from 'dayjs/plugin/utc';
import Graph from './Graph';
import './TaskPage.css';

const { Title, Text } = Typography;
dayjs.extend(utc);

const TaskPage = () => {
  const { id } = useParams();
  const navigate = useNavigate();
  const [task, setTask] = useState(null);
  const [loading, setLoading] = useState(true);
  const [updating, setUpdating] = useState(false);
  const [tempParams, setTempParams] = useState({
    depthLevel: 0,
    minWeight: 0,
    maxSources: 0,
    maxNeighboursForSource: 0,
  });
  const [sources, setSources] = useState([]);
  const [sourcesLoading, setSourcesLoading] = useState(false);
  const prevStatusRef = useRef(null);

  const statusName = {
    created: 'Создана',
    active: 'Активна',
    in_processing: 'В обработке',
    stopped: 'Остановлена',
    stopped_with_error: 'Остановлена с ошибкой',
  }

  // Получение данных задачи
  const fetchTask = async () => {
    try {
      const response = await fetch('http://localhost:8080/get-task', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ id: parseInt(id) })
      });

      const data = await response.json();
      if (response.ok) {
        setTask(data);
        setTempParams({
          depthLevel: data.depthLevel,
          minWeight: data.minWeight,
          maxSources: data.maxSources,
          maxNeighboursForSource: data.maxNeighboursForSource,
        });
        if (data.status === 'active') {
          await fetchSources(data);
        }
      } else {
        throw new Error(data.error || 'Ошибка загрузки задачи');
      }
    } catch (error) {
      message.error(error.message);
    } finally {
      setLoading(false);
    }
  };

  const fetchTaskStatus = async () => {
    try {
      const response = await fetch('http://localhost:8080/get-task-status', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ id: parseInt(id) })
      });

      const data = await response.json();
      if (response.ok) {
        const prevStatus = prevStatusRef.current;
        const newStatus = data.status;

        if (task && newStatus === task.status) {
          return
        }

        setTask(prev => ({ ...prev, status: newStatus }));

        // Загружаем источники только при переходе из in_processing в active
        if (prevStatus === 'in_processing' && newStatus === 'active') {
          fetchTask();
        }

        prevStatusRef.current = newStatus;
      }
    } catch (error) {
      console.error('Ошибка проверки статуса:', error);
    }
  };

  // Обновление параметров задачи
  const updateTaskParams = async (params) => {
    setUpdating(true);
    try {
      const response = await fetch('http://localhost:8080/update-task', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ id: parseInt(id), ...params })
      });

      const data = await response.json();
      if (response.ok) {
        message.success('Параметры успешно обновлены');
        fetchTask();
      } else {
        throw new Error(data.error || 'Ошибка обновления');
      }
    } catch (error) {
      message.error(error.message);
    } finally {
      setUpdating(false);
    }
  };

  // Управление статусом задачи
  const toggleTaskStatus = async () => {
    const endpoint = task.status === 'active' ? '/stop-task' : '/activate-task';

    try {
      const response = await fetch(`http://localhost:8080${endpoint}`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ id: parseInt(id) })
      });

      const data = await response.json();
      if (response.ok) {
        message.success(`Задача ${task.status === 'active' ? 'остановлена' : 'активирована'}`);
        fetchTask();
      } else {
        throw new Error(data.error || 'Ошибка изменения статуса');
      }
    } catch (error) {
      message.error(error.message);
    }
  };

  // Получение источников
  const fetchSources = async (task) => {
    setSourcesLoading(true);
    try {
      const response = await fetch('http://localhost:8080/get-sources', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ id: parseInt(id) })
      });

      const data = await response.json();
      if (response.ok) {
        const newSources = data.sources.map((e) => ({ ...e, title: e.title || task.query }))
        setSources(newSources || []);
      } else {
        throw new Error(data.error || 'Ошибка загрузки источников');
      }
    } catch (error) {
      console.log(error.message)
      message.error(error.message);
    } finally {
      setSourcesLoading(false);
    }
  };

  // Колонки для таблицы источников
  const sourcesColumns = [
    {
      title: 'ID',
      dataIndex: 'ID',
      key: 'ID',
      width: 80,
      render: (_, record) => record.id || '-'
    },
    {
      title: 'Название',
      dataIndex: 'Title',
      key: 'Title',
      render: (_, record) => (
        <a href={record.url ? record.url : '#'} target="_blank" rel="noopener noreferrer">
          {record.title.length > 50 ? `${record.title.substring(0, 50)}...` : record.title || 'Без названия'}
        </a>
      ),
    },
    {
      title: 'Вес (BM25)',
      dataIndex: 'Weight',
      key: 'Weight',
      render: (_, record) => (record.weight ? record.weight.toFixed(2) : '0'),
      sorter: (a, b) => (a?.weight || 0) - (b?.weight || 0),
      defaultSortOrder: 'descend',
      width: 120,
    },
  ];

  useEffect(() => {
    fetchTask();
  }, [id]);

  // Периодическая проверка статуса
  useEffect(() => {
    if (task?.status) {
      prevStatusRef.current = task.status;
    }

    const intervalId = setInterval(() => {
      if (task) {
        fetchTaskStatus();
      }
    }, 1000);

    return () => clearInterval(intervalId);
  }, [task]);

  if (loading) return <div className="loading-spinner">Загрузка...</div>;
  if (!task) return <div className="error-message">Задача не найдена</div>;

  // Определение иконки статуса
  const statusIcons = {
    created: <InfoCircleOutlined />,
    active: <PlayCircleOutlined />,
    stopped: <PauseCircleOutlined />,
    stopped_with_error: <CloseCircleOutlined />,
    in_processing: <ClockCircleOutlined />
  };

  const statusColors = {
    created: 'blue',
    active: 'green',
    stopped: 'orange',
    stopped_with_error: 'red',
    in_processing: 'gold'
  };

  return (
    <div className="task-page-container">
      <Button
        icon={<ArrowLeftOutlined />}
        onClick={() => navigate('/')}
        className="back-button"
      >
        На главную
      </Button>

      <Card title={`Задача #${id}`} className="task-card">
        {/* Блок основной информации */}
        <div className="task-header">
          <Space size="large">
            <Tag
              icon={statusIcons[task.status] || <ClockCircleOutlined />}
              color={statusColors[task.status] || 'gray'}
              className="status-tag"
            >
              {statusName[task.status] || task.status}
            </Tag>
            <Title level={4} className="task-query">{task.query}</Title>
          </Space>
        </div>

        <Divider orientation="left">Информация</Divider>

        <div className="task-info">
          <div className="info-row">
            <Text strong>Создана:</Text>
            <Text>{prepareDate(task.createdAt)}</Text>
          </div>
          <div className="info-row">
            <Text strong>Обновлена:</Text>
            <Text>{prepareDate(task.updatedAt)}</Text>
          </div>

          {/* Блок управления параметрами */}
          <Divider orientation="left">Параметры задачи</Divider>

          <div className="task-params">
            <div className="param-row">
              <Text strong>Уровень погружения:</Text>
              <InputNumber
                min={1}
                max={10}
                value={tempParams.depthLevel}
                onChange={(value) => setTempParams({ ...tempParams, depthLevel: value })}
              />
            </div>

            <div className="param-row">
              <Text strong>Минимальный вес:</Text>
              <InputNumber
                min={0}
                max={1}
                step={0.01}
                value={tempParams.minWeight}
                onChange={(value) => setTempParams({ ...tempParams, minWeight: value })}
              />
            </div>

            <div className="param-row">
              <Text strong>Макс. источников:</Text>
              <InputNumber
                min={1}
                max={10000}
                value={tempParams.maxSources}
                onChange={(value) => setTempParams({ ...tempParams, maxSources: value })}
              />
            </div>

            <div className="param-row">
              <Text strong>Макс. соседей:</Text>
              <InputNumber
                min={1}
                max={100}
                value={tempParams.maxNeighboursForSource}
                onChange={(value) => setTempParams({ ...tempParams, maxNeighboursForSource: value })}
              />
            </div>
          </div>

          <div className="task-actions">
            <Space>
              <Button
                type="primary"
                onClick={() => updateTaskParams(tempParams)}
                loading={updating}
                disabled={
                  tempParams.depthLevel === task.depthLevel &&
                  tempParams.minWeight === task.minWeight &&
                  tempParams.maxSources === task.maxSources &&
                  tempParams.maxNeighboursForSource === task.maxNeighboursForSource
                }
              >
                Сохранить параметры
              </Button>

              <Button
                icon={task.status === 'active' ? <PauseCircleOutlined /> : <PlayCircleOutlined />}
                onClick={toggleTaskStatus}
                loading={updating}
                danger={task.status === 'active'}
              >
                {task.status === 'active' ? 'Остановить' : 'Активировать'}
              </Button>
            </Space>
          </div>

          {/* Блок информации о запуске */}
          {(task.status === 'active' || task.status === 'stopped' || task.status === 'stopped_with_error') && (
            <>
              <Divider orientation="left">Последний запуск</Divider>
              {task.processedAt && (
                <div className="info-row">
                  <Text strong>Время запуска:</Text>
                  <Text>{prepareDate(task.processedAt)}</Text>
                </div>
              )}
              {task.sourcesViewed !== undefined && task.sourcesViewed !== null && (
                <div className="info-row">
                  <Text strong>Просмотрено источников:</Text>
                  <Text>{task.sourcesViewed}</Text>
                </div>
              )}
              {task.launchDuration !== undefined && task.launchDuration !== null && (
                <div className="info-row">
                  <Text strong>Длительность:</Text>
                  <Text>{(task.launchDuration / 1000000000).toFixed(2)} сек</Text>
                </div>
              )}
              {task.status === 'stopped_with_error' && task.errorMsg && (
                <div className="info-row">
                  <Text strong>Ошибка:</Text>
                  <Text type="danger">{task.errorMsg}</Text>
                </div>
              )}
            </>
          )}
        </div>

        {/* Блок источников */}
        <Divider orientation="left">Найденные источники</Divider>

        {task.status === 'in_processing' ? (
          <div className="sources-loader">
            <Spin tip="Идет поиск источников..." size="large" />
          </div>
        ) : (
          <>
            <Table
              columns={sourcesColumns}
              dataSource={sources}
              rowKey="ID" // Убедитесь, что это поле уникально для каждой записи
              loading={sourcesLoading}
              pagination={{ pageSize: 10 }}
            />

            {sources.length && <>
              <Divider orientation="left">Граф источников</Divider>

              <div style={{ height: 500, border: '1px solid #ddd', borderRadius: 4 }}>
                <Graph sources={sources} />
              </div>
            </>}

          </>
        )}
      </Card>
    </div>
  );
};

export default TaskPage;

function prepareDate(utcString) {
 return dayjs.utc(utcString).format('DD.MM.YYYY HH:mm')
}