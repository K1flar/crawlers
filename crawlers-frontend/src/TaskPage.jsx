import React from 'react';
import { useParams } from 'react-router';
// import './TaskPage.css';

const TaskPage = () => {
  const { id } = useParams();

  return (
    <div className="task-container">
      <h1>Страница задачи</h1>
      <p>ID задачи: {id}</p>
      <p>Здесь будет детальная информация о задаче и ее результатах</p>
    </div>
  );
};

export default TaskPage;