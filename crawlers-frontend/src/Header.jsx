import { Button, Layout } from 'antd';
import { HomeOutlined, BookOutlined } from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';

const { Header } = Layout;

const AppHeader = () => {
  const navigate = useNavigate();

  return (
    <Header style={{
      display: 'flex',
      alignItems: 'center',
      background: '#fff',
      boxShadow: '0 2px 8px #f0f1f2',
      zIndex: 1,
      padding: '0 24px'
    }}>
      <Button 
        type="text" 
        icon={<HomeOutlined />} 
        onClick={() => navigate('/')}
        style={{ fontSize: '16px' }}
      >
        Главная
      </Button>

      <Button 
        type="text" 
        icon={<BookOutlined />} 
        onClick={() => navigate('/protocol')}
        style={{ fontSize: '16px' }}
      >
        Протокол
      </Button>
    </Header>
  );
};

export default AppHeader;