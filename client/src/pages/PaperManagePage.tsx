import React, { useState, useEffect } from 'react';
import { Table, Button, Space, message, Popconfirm, Tag } from 'antd';
import { EditOutlined, DeleteOutlined, DownloadOutlined, EyeOutlined } from '@ant-design/icons';
import { getPapers, deletePaper, type Paper } from '../api';
import { useNavigate } from 'react-router-dom';

const PaperManagePage: React.FC = () => {
  const [papers, setPapers] = useState<Paper[]>([]);
  const [loading, setLoading] = useState(false);

  const navigate = useNavigate();

  // 获取试卷列表
  const loadPapers = async () => {
    setLoading(true);
    try {
      const response = await getPapers();
      if (response.code === 0) {
        setPapers(response.data);
      } else {
        message.error(`获取试卷列表失败: ${response.msg}`);
      }
    } catch (error) {
      message.error('获取试卷列表失败');
    }
    setLoading(false);
  };

  // 删除试卷
  const handleDelete = async (paperId: number) => {
    try {
      const response = await deletePaper(paperId);
      if (response.code === 0) {
        message.success('删除试卷成功');
        loadPapers(); // 重新加载列表
      } else {
        message.error(`删除试卷失败: ${response.msg}`);
      }
    } catch (error) {
      message.error('删除试卷失败');
    }
  };

  // 编辑试卷
  const handleEdit = (paper: Paper) => {
    navigate(`/papers/${paper.id}/edit`);
  };



  // 导出试卷
  const handleExport = (paperId: number, title: string) => {
    const exportUrl = `/api/papers/${paperId}/export?format=docx`;
    const link = document.createElement('a');
    link.href = exportUrl;
    link.download = `${title}.docx`;
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
    message.success('试卷导出中...');
  };

  // 预览试卷
  const handlePreview = (paperId: number) => {
    navigate(`/paper/preview/${paperId}`, { state: { from: 'papers' } });
  };



  // 计算试卷总分
  const getTotalScore = (config: any[]) => {
    return config.reduce((total, section) => total + (section.count * section.score_each), 0);
  };

  // 计算试卷题目总数
  const getTotalQuestions = (config: any[]) => {
    return config.reduce((total, section) => total + section.count, 0);
  };

  useEffect(() => {
    loadPapers();
  }, []);

  const columns = [
    {
      title: '试卷ID',
      dataIndex: 'id',
      key: 'id',
      width: 80,
    },
    {
      title: '试卷标题',
      dataIndex: 'title',
      key: 'title',
      ellipsis: true,
    },
    {
      title: '题目数量',
      key: 'questionCount',
      width: 100,
      render: (_: any, record: Paper) => getTotalQuestions(record.config),
    },
    {
      title: '总分',
      key: 'totalScore',
      width: 80,
      render: (_: any, record: Paper) => getTotalScore(record.config),
    },
    {
      title: '试卷结构',
      key: 'structure',
      render: (_: any, record: Paper) => (
        <div>
          {record.config.map((section, index) => (
            <Tag key={index} color="blue" style={{ margin: '2px' }}>
              {section.name}: {section.count}题×{section.score_each}分
            </Tag>
          ))}
        </div>
      ),
    },
    {
      title: '创建时间',
      dataIndex: 'created_at',
      key: 'created_at',
      width: 180,
      render: (text: string) => new Date(text).toLocaleString('zh-CN'),
    },
    {
      title: '操作',
      key: 'action',
      width: 200,
      render: (_: any, record: Paper) => (
        <Space size="small">
          <Button
            type="primary"
            size="small"
            icon={<EyeOutlined />}
            onClick={() => handlePreview(record.id)}
          >
            预览
          </Button>
          <Button
            size="small"
            icon={<EditOutlined />}
            onClick={() => handleEdit(record)}
          >
            编辑
          </Button>
          <Button
            size="small"
            icon={<DownloadOutlined />}
            onClick={() => handleExport(record.id, record.title)}
          >
            导出
          </Button>
          <Popconfirm
            title="确定要删除这份试卷吗？"
            description="删除后无法恢复！"
            onConfirm={() => handleDelete(record.id)}
            okText="确定"
            cancelText="取消"
          >
            <Button
              size="small"
              danger
              icon={<DeleteOutlined />}
            >
              删除
            </Button>
          </Popconfirm>
        </Space>
      ),
    },
  ];

  return (
    <div>
      <div style={{ marginBottom: 16, display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <h2>试卷管理</h2>
        <Button type="primary" onClick={loadPapers}>
          刷新列表
        </Button>
      </div>

      <Table
        columns={columns}
        dataSource={papers}
        loading={loading}
        rowKey="id"
        pagination={{
          total: papers.length,
          pageSize: 10,
          showSizeChanger: true,
          showQuickJumper: true,
          showTotal: (total) => `共 ${total} 份试卷`,
        }}
        scroll={{ x: 1200 }}
      />


    </div>
  );
};

export default PaperManagePage; 