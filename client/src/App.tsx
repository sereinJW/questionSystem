import { Layout, Menu } from 'antd';
import { useState } from 'react';
import { BrowserRouter as Router, Routes, Route, useNavigate, useLocation } from 'react-router-dom';
import './App.css';
import LearningPage from './pages/LearningPage';
import QuestionBankPage from './pages/QuestionBankPage';

const { Header, Sider, Content } = Layout;

function App() {
  const [collapsed, setCollapsed] = useState(false);
  const location = useLocation();
  const navigate = useNavigate();
  const selectedKey = location.pathname.startsWith('/questions') ? '2' : '1';
  
  console.log('当前路径:', location.pathname, '选中菜单:', selectedKey);

  return (
    <Layout style={{ minHeight: '100vh' }}>
      <Sider collapsible collapsed={collapsed} onCollapse={setCollapsed}>
        <div className="logo" style={{height: 32, margin: 16, color: '#fff', fontWeight: 'bold', fontSize: 18, textAlign: 'center'}}>
          
        </div>
        <Menu
          theme="dark"
          mode="inline"
          selectedKeys={[selectedKey]}
          items={[
            { key: '1', label: '学习心得', onClick: () => navigate('/') },
            { key: '2', label: '题库管理', onClick: () => navigate('/questions') },
          ]}
        />
      </Sider>
      <Layout>
        <Header style={{ background: '#001529', color: '#fff', fontSize: 20, fontWeight: 'bold', paddingLeft: 24 }}>
          武汉科技大学 徐伽炜 考试出题系统
        </Header>
        <Content style={{ margin: '24px 16px', padding: 24, background: '#fff', minHeight: 280 }}>
          <Routes>
            <Route path="/" element={<LearningPage />} />
            <Route path="/questions" element={<QuestionBankPage />} />
          </Routes>
        </Content>
      </Layout>
    </Layout>
  );
}

export default function AppWithRouter() {
  return (
    <Router>
      <App />
    </Router>
  );
}
