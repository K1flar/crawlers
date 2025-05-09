import React from 'react';
import { BrowserRouter, Route, Routes } from 'react-router';
import { Button, Layout } from 'antd';
import HomePage from './HomePage';
import TaskPage from './TaskPage';
import AppHeader from './Header';
import ProtocolPage from './ProtocolPage';

function App() {
  return (
    <BrowserRouter>
      <AppHeader />
      <Routes>
        <Route path="/" element={<HomePage />} />
        <Route path="/task/:id" element={<TaskPage />} />
        <Route path="/protocol" element={<ProtocolPage />} />
      </Routes>
    </BrowserRouter>
  );
}

export default App;