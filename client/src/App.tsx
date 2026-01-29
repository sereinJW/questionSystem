import { Layout, Menu } from 'antd';
import { useState } from 'react';
import { BrowserRouter as Router, Routes, Route, useNavigate, useLocation } from 'react-router-dom';
import './App.css';
import QuestionBankPage from './pages/QuestionBankPage';
import PaperPreviewPage from './pages/PaperPreviewPage';
import PaperManagePage from './pages/PaperManagePage';
import PaperEditPage from './pages/PaperEditPage';

const { Header, Sider, Content } = Layout;

function App() {
  const [collapsed, setCollapsed] = useState(false);
  const location = useLocation();
  const navigate = useNavigate();
  
  // 根据路径确定选中的菜单项
  const getSelectedKey = () => {
    if (location.pathname.startsWith('/questions')) return '1';
    if (location.pathname.startsWith('/papers')) return '2';
    if (location.pathname.startsWith('/paper')) return '2'; // 试卷相关页面也归属于试卷管理
    return '1';
  };
  
  const selectedKey = getSelectedKey();
  
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
            { key: '1', label: '题库管理', onClick: () => navigate('/questions') },
            { key: '2', label: '试卷管理', onClick: () => navigate('/papers') },
          ]}
        />
      </Sider>
      <Layout>
        <Header style={{ background: '#001529', color: '#fff', fontSize: 20, fontWeight: 'bold', paddingLeft: 24 }}>
          考试出题系统
        </Header>
        <Content style={{ margin: '24px 16px', padding: 24, background: '#fff', minHeight: 280 }}>
          <Routes>
            <Route path="/" element={<QuestionBankPage />} />
            <Route path="/questions" element={<QuestionBankPage />} />
            <Route path="/papers" element={<PaperManagePage />} />
            <Route path="/papers/:id/edit" element={<PaperEditPage />} />
            <Route path="/paper/preview/:id" element={<PaperPreviewPage />} />
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
