import React from 'react';
import { BrowserRouter,Route, Routes, Link } from 'react-router';
import HomePage from './HomePage';
import TaskPage from './TaskPage';

function App() {
  return (
    <BrowserRouter>
        <Routes>
          <Route path="/" element={<HomePage />} />
          <Route path="/task/:id" element={<TaskPage />} />
        </Routes>
    </BrowserRouter>
  );
}

export default App;