import { Card, Button, Badge, Tag } from 'antd';
import { useNavigate } from 'react-router-dom';
import {
    PlayCircleOutlined,
    PauseCircleOutlined,
    ClockCircleOutlined,
    InfoCircleOutlined,
    CloseCircleOutlined,
    DatabaseOutlined,
} from '@ant-design/icons';

const TaskCard = ({ task }) => {
    const navigate = useNavigate();

    // Определяем цвет бейджа в зависимости от статуса
    const statusColor = {
        created: 'blue',
        active: 'green',
        stopped: 'orange',
        stopped_with_error: 'red',
        in_processing: 'gold'
    }[task.status] || 'gray';

    const statusName = {
        created: "Создана",
        active: 'Активна',
        in_processing: 'В обработке',
        stopped: "Остановлена",
        stopped_with_error: 'Остановлена с ошибкой',
    }[task.status] || 'Создана'

    const statusIcon = {
        created: <InfoCircleOutlined />,
        active: <PlayCircleOutlined />,
        stopped: <PauseCircleOutlined />,
        stopped_with_error: <CloseCircleOutlined />,
        in_processing: <ClockCircleOutlined />
    }[task.status] || <ClockCircleOutlined />

    return (
        <Card
            title={task.query}
            className="task-card"
            actions={[
                <Button type="link" onClick={() => navigate(`/task/${task.id}`)}>
                    Подробнее
                </Button>
            ]}
        >
            <div className="task-meta">
                <Tag
                    icon={statusIcon}
                    color={statusColor || 'gray'}
                    className="status-tag"
                >
                    {statusName}
                </Tag>

                <Tag
                    icon={<DatabaseOutlined />}
                    color={'purple'}
                    className="sources-tag"
                >
                    {`${task.countSources} ${pluralizeSources(task.countSources)}`}
                </Tag>
            </div>
        </Card>
    );
};

export function pluralizeSources(count) {
    const lastDigit = count % 10;
    const lastTwoDigits = count % 100;
  
    if (lastTwoDigits >= 11 && lastTwoDigits <= 19) {
      return 'источников';
    }
    
    if (lastDigit === 1) {
      return 'источник';
    }
    
    if (lastDigit >= 2 && lastDigit <= 4) {
      return 'источника';
    }
    
    return 'источников';
  }

export default TaskCard