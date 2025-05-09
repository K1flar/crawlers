import React, { useState, useEffect } from 'react';
import { Pagination, Select, Input, Button, Spin, message } from 'antd';
import TaskCard from './Task'
import { useNavigate } from 'react-router-dom';

const { Option } = Select;

const HomePage = () => {
  const [tasks, setTasks] = useState([]);
  const [loading, setLoading] = useState(false);
  const [filters, setFilters] = useState({
    status: null,
    query: '',
    limit: 10,
    offset: 0
  });
  const [newTaskQuery, setNewTaskQuery] = useState(''); // Новое состояние для ввода
  const [totalTasks, setTotalTasks] = useState(0);
  const navigate = useNavigate();

  const fetchTasks = async () => {
    setLoading(true);
    try {
      const response = await fetch('http://localhost:8080/get-tasks', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          status: filters.status,
          query: filters.query,
          limit: filters.limit,
          offset: filters.offset
        })
      });

      const data = await response.json();
      if (response.ok) {
        setTasks(data.tasks);
        setTotalTasks(data.total || data.tasks.length);
      } else {
        throw new Error(data.error || 'Failed to fetch tasks');
      }
    } catch (error) {
      message.error(error.message);
    } finally {
      setLoading(false);
    }
  };

  const createNewTask = async () => {
    if (!newTaskQuery.trim()) {
      message.warning('Введите поисковый запрос');
      return;
    }

    try {
      const response = await fetch('http://localhost:8080/create-task', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          query: newTaskQuery
        })
      });

      const data = await response.json();
      if (response.ok) {
        message.success('Задача успешно создана');
        navigate(`/task/${data.id}`); // Редирект на страницу задачи
        setNewTaskQuery('');
        fetchTasks(); // Обновляем список задач
      } else {
        throw new Error(data.error || 'Failed to create task');
      }
    } catch (error) {
      message.error(error.message);
    }
  };

  useEffect(() => {
    fetchTasks();
  }, [filters.offset, filters.status, filters.query]);

  return (
    <div className="home-container">
      <h1>Система администрирования поисковых роботов</h1>
      {/* Блок создания новой задачи */}
      <div className="create-task-section">
        <Input
          placeholder="Введите новый поисковый запрос"
          value={newTaskQuery}
          onChange={(e) => setNewTaskQuery(e.target.value)}
          style={{ width: 400, marginRight: 16 }}
        />
        <Button type="primary" onClick={createNewTask}>
          Создать задачу
        </Button>
      </div>

      {/* Блок фильтрации существующих задач */}
      <div className="filters-section">
        <Select
          placeholder="Статус задачи"
          allowClear
          style={{ width: 200, marginRight: 16 }}
          onChange={(value) => setFilters({ ...filters, status: value, offset: 0 })}
          value={filters.status}
        >
          <Option value="created">Созданные</Option>
          <Option value="active">Активные</Option>
          <Option value="in_processing">В обработке</Option>
          <Option value="stopped">Остановлены</Option>
          <Option value="stopped_with_error">Остановлены с ошибкой</Option>
        </Select>

        <Input
          placeholder="Поиск по задачам"
          style={{ width: 300, marginRight: 16 }}
          value={filters.query}
          onChange={(e) => setFilters({ ...filters, query: e.target.value, offset: 0 })}
        />

        <Button onClick={fetchTasks}>
          Обновить
        </Button>
      </div>

      {/* Список задач */}
      {loading ? (
        <Spin size="large" />
      ) : (
        <div className="tasks-list">
          {tasks.map(task => (
            <TaskCard key={task.id} task={task}/>
          ))}
        </div>
      )}

      {/* Пагинация */}
      <div className="pagination-container">
        <Pagination
          current={filters.offset / filters.limit + 1}
          total={totalTasks}
          pageSize={filters.limit}
          onChange={(page) => setFilters({ ...filters, offset: (page - 1) * filters.limit })}
          showSizeChanger={false}
        />
      </div>
    </div>
  );
};

export default HomePage;