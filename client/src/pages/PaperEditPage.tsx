import React, { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { 
  Card, 
  Button, 
  Space, 
  message, 
  Input, 
  Table, 
  Checkbox, 
  Typography, 
  Row, 
  Col, 
  Tag, 
  Spin,
  Divider
} from 'antd';
import { ArrowLeftOutlined, SaveOutlined } from '@ant-design/icons';
import { 
  getPaperDetail, 
  getQuestions, 
  updatePaperQuestions,
  type PaperDetail, 
  type Topic 
} from '../api';

const { Text } = Typography;

const PaperEditPage: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState(false);
  const [paperDetail, setPaperDetail] = useState<PaperDetail | null>(null);
  const [allQuestions, setAllQuestions] = useState<Topic[]>([]);

  const [paperTitle, setPaperTitle] = useState('');
  const [currentSectionIndex, setCurrentSectionIndex] = useState(0);
  const [sectionQuestions, setSectionQuestions] = useState<{[key: number]: number[]}>({});

  // 获取试卷详情
  const fetchPaperDetail = async () => {
    if (!id) {
      message.error('试卷ID不存在');
      navigate('/papers');
      return;
    }

    setLoading(true);
    try {
      const [paperRes, questionsRes] = await Promise.all([
        getPaperDetail(parseInt(id)),
        getQuestions()
      ]);

      if (paperRes.code === 0 && paperRes.data) {
        setPaperDetail(paperRes.data);
        setPaperTitle(paperRes.data.paper.title);
        // 按照大题分配现有题目
        const currentQuestionIds = paperRes.data.questions.map(q => q.question_id);
        
        // 将现有题目按大题分组
        const sectionQuestionsMap: {[key: number]: number[]} = {};
        let questionIndex = 0;
        paperRes.data.paper.config.forEach((section, sectionIndex) => {
          const sectionQuestionIds = currentQuestionIds.slice(questionIndex, questionIndex + section.count);
          sectionQuestionsMap[sectionIndex] = sectionQuestionIds;
          questionIndex += section.count;
        });
        setSectionQuestions(sectionQuestionsMap);
      } else {
        message.error(paperRes.msg || '获取试卷详情失败');
        navigate('/papers');
      }

      if (questionsRes.code === 0) {
        setAllQuestions(questionsRes.data);
      } else {
        message.error('获取题库失败');
      }
    } catch (error) {
      message.error('网络错误');
      console.error('获取数据失败:', error);
      navigate('/papers');
    } finally {
      setLoading(false);
    }
  };

  // 保存修改
  const handleSave = async () => {
    if (!paperDetail || !paperTitle.trim()) {
      message.error('请输入试卷标题');
      return;
    }

    // 验证每个大题的题目数量
    const config = paperDetail.paper.config;
    for (let i = 0; i < config.length; i++) {
      const section = config[i];
      const sectionQuestionIds = sectionQuestions[i] || [];
      if (sectionQuestionIds.length !== section.count) {
        message.error(`${section.name} 需要选择 ${section.count} 道题目，当前已选择 ${sectionQuestionIds.length} 道`);
        return;
      }
    }

    // 合并所有大题的题目ID
    const allQuestionIds: number[] = [];
    config.forEach((_, index) => {
      const sectionQuestionIds = sectionQuestions[index] || [];
      allQuestionIds.push(...sectionQuestionIds);
    });

    setSaving(true);
    try {
      const response = await updatePaperQuestions(
        parseInt(id!), 
        paperTitle.trim(), 
        allQuestionIds
      );
      
      if (response.code === 0) {
        message.success('试卷保存成功');
        navigate('/papers');
      } else {
        message.error(`保存失败: ${response.msg}`);
      }
    } catch (error) {
      message.error('保存失败');
      console.error('保存试卷失败:', error);
    } finally {
      setSaving(false);
    }
  };

  // 处理题目选择
  const handleQuestionSelect = (questionId: number, checked: boolean) => {
    const currentSectionQuestionIds = sectionQuestions[currentSectionIndex] || [];
    if (checked) {
      const newSectionQuestions = [...currentSectionQuestionIds, questionId];
      setSectionQuestions({
        ...sectionQuestions,
        [currentSectionIndex]: newSectionQuestions
      });
    } else {
      const newSectionQuestions = currentSectionQuestionIds.filter(id => id !== questionId);
      setSectionQuestions({
        ...sectionQuestions,
        [currentSectionIndex]: newSectionQuestions
      });
    }
  };

  // 全选/取消全选当前大题的题目
  const handleSelectAll = (checked: boolean) => {
    if (!paperDetail) return;
    const currentSection = paperDetail.paper.config[currentSectionIndex];
    const filteredQuestions = getFilteredQuestions();
    
    if (checked) {
      const questionIds = filteredQuestions.slice(0, currentSection.count).map(q => q.id!);
      setSectionQuestions({
        ...sectionQuestions,
        [currentSectionIndex]: questionIds
      });
    } else {
      setSectionQuestions({
        ...sectionQuestions,
        [currentSectionIndex]: []
      });
    }
  };

  // 获取当前大题类型的题目
  const getFilteredQuestions = () => {
    if (!paperDetail) return [];
    const currentSection = paperDetail.paper.config[currentSectionIndex];
    return allQuestions.filter(q => q.type_id === currentSection.type);
  };

  // 切换大题
  const handleSectionChange = (sectionIndex: number) => {
    setCurrentSectionIndex(sectionIndex);
  };

  // 获取题目类型文本
  const getTypeText = (type: number) => {
    const types = ['', '单选题', '多选题', '编程题'];
    return types[type] || '未知';
  };

  // 获取难度文本和颜色
  const getDifficultyDisplay = (difficulty: number) => {
    const configs = [
      { text: '', color: 'default' },
      { text: '简单', color: 'green' },
      { text: '中等', color: 'orange' },
      { text: '困难', color: 'red' }
    ];
    return configs[difficulty] || configs[0];
  };

  useEffect(() => {
    fetchPaperDetail();
  }, [id]);

  if (loading) {
    return (
      <div style={{ textAlign: 'center', padding: '100px 0' }}>
        <Spin size="large" tip="加载中..." />
      </div>
    );
  }

  if (!paperDetail) {
    return (
      <div style={{ textAlign: 'center', padding: '100px 0' }}>
        <Text type="secondary">试卷不存在</Text>
      </div>
    );
  }

  const currentSection = paperDetail.paper.config[currentSectionIndex];
  const filteredQuestions = getFilteredQuestions();
  const currentSectionQuestionIds = sectionQuestions[currentSectionIndex] || [];
  const isAllSelected = filteredQuestions.length > 0 && currentSectionQuestionIds.length === Math.min(filteredQuestions.length, currentSection.count);
  const isIndeterminate = currentSectionQuestionIds.length > 0 && currentSectionQuestionIds.length < Math.min(filteredQuestions.length, currentSection.count);
  
  // 计算总体进度
  const getTotalSelectedCount = () => {
    return Object.values(sectionQuestions).reduce((sum, questions) => sum + questions.length, 0);
  };
  const getTotalRequiredCount = () => {
    return paperDetail.paper.config.reduce((sum, section) => sum + section.count, 0);
  };

  const columns = [
    {
      title: (
        <Checkbox
          indeterminate={isIndeterminate}
          onChange={(e) => handleSelectAll(e.target.checked)}
          checked={isAllSelected}
        >
          全选
        </Checkbox>
      ),
      dataIndex: 'select',
      key: 'select',
      width: 60,
      render: (_: any, record: Topic) => (
        <Checkbox
          checked={currentSectionQuestionIds.includes(record.id!)}
          onChange={(e) => handleQuestionSelect(record.id!, e.target.checked)}
          disabled={!currentSectionQuestionIds.includes(record.id!) && currentSectionQuestionIds.length >= currentSection.count}
        />
      ),
    },
    {
      title: 'ID',
      dataIndex: 'id',
      key: 'id',
      width: 60,
    },
    {
      title: '题目',
      dataIndex: 'title',
      key: 'title',
      ellipsis: true,
    },
    {
      title: '类型',
      dataIndex: 'type_id',
      key: 'type_id',
      width: 80,
      render: (type: number) => getTypeText(type),
    },
    {
      title: '难度',
      dataIndex: 'difficulty',
      key: 'difficulty',
      width: 80,
      render: (difficulty: number) => {
        const { text, color } = getDifficultyDisplay(difficulty);
        return <Tag color={color}>{text}</Tag>;
      },
    },
    {
      title: '语言',
      dataIndex: 'language',
      key: 'language',
      width: 80,
    },
    {
      title: '关键词',
      dataIndex: 'keyword',
      key: 'keyword',
      width: 100,
      ellipsis: true,
    },
  ];

  return (
    <div style={{ maxWidth: '1200px', margin: '0 auto' }}>
      {/* 页面头部 */}
      <div style={{ marginBottom: 24 }}>
        <Space>
          <Button 
            icon={<ArrowLeftOutlined />} 
            onClick={() => navigate('/papers')}
          >
            返回试卷管理
          </Button>
          
          <Divider type="vertical" />
          
          <Button 
            type="primary"
            icon={<SaveOutlined />}
            loading={saving}
            onClick={handleSave}
          >
            保存试卷
          </Button>
        </Space>
      </div>

      {/* 试卷基本信息 */}
      <Card title="试卷基本信息" style={{ marginBottom: 24 }}>
        <Row gutter={[16, 16]}>
          <Col span={12}>
            <div>
              <label style={{ display: 'block', marginBottom: 8, fontWeight: 'bold' }}>
                试卷标题：
              </label>
              <Input
                value={paperTitle}
                onChange={(e) => setPaperTitle(e.target.value)}
                placeholder="请输入试卷标题"
                maxLength={100}
              />
            </div>
          </Col>
          <Col span={12}>
            <div>
              <Text strong>试卷ID：</Text>
              <Text>{paperDetail.paper.id}</Text>
            </div>
            <div style={{ marginTop: 8 }}>
              <Text strong>创建时间：</Text>
              <Text>{new Date(paperDetail.paper.created_at).toLocaleString('zh-CN')}</Text>
            </div>
          </Col>
          <Col span={24}>
            <div>
              <Text strong>试卷结构：</Text>
              <div style={{ marginTop: 8 }}>
                {paperDetail.paper.config.map((section, index) => (
                  <Tag key={index} color="blue" style={{ margin: '2px' }}>
                    {section.name}: {section.count}题×{section.score_each}分
                  </Tag>
                ))}
              </div>
            </div>
          </Col>
        </Row>
      </Card>

            {/* 总体进度 */}
      <Card style={{ marginBottom: 24 }}>
        <div style={{ textAlign: 'center' }}>
          <Text style={{ fontSize: '16px' }}>
            总体进度：已选择 <Text style={{ color: '#1890ff', fontWeight: 'bold' }}>{getTotalSelectedCount()}</Text> / 
            需要 <Text style={{ color: '#1890ff', fontWeight: 'bold' }}>{getTotalRequiredCount()}</Text> 道题
          </Text>
          <div style={{ marginTop: 16 }}>
            <Space wrap>
              {paperDetail.paper.config.map((section, index) => {
                const sectionQuestionIds = sectionQuestions[index] || [];
                const isComplete = sectionQuestionIds.length === section.count;
                const isCurrent = index === currentSectionIndex;
                return (
                  <Tag 
                    key={index} 
                    color={isComplete ? 'green' : isCurrent ? 'blue' : 'default'}
                    style={{ 
                      cursor: 'pointer',
                      padding: '4px 8px',
                      fontSize: '14px'
                    }}
                    onClick={() => handleSectionChange(index)}
                  >
                    {section.name}: {sectionQuestionIds.length}/{section.count}
                    {isComplete && ' ✓'}
                  </Tag>
                );
              })}
            </Space>
          </div>
        </div>
      </Card>

      {/* 当前大题选择 */}
      <Card 
        title={
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
            <span>{currentSection.name} - 选择题目</span>
            <Space>
              <Text>
                已选择 <Text style={{ color: '#1890ff' }}>{currentSectionQuestionIds.length}</Text> / 
                需要 <Text style={{ color: '#1890ff' }}>{currentSection.count}</Text> 道题
              </Text>
              {currentSectionQuestionIds.length !== currentSection.count && (
                <Text type="warning">
                  {currentSectionQuestionIds.length > currentSection.count ? '选择过多' : '选择不足'}
                </Text>
              )}
              {currentSectionQuestionIds.length === currentSection.count && (
                <Text style={{ color: '#52c41a' }}>✓ 完成</Text>
              )}
            </Space>
          </div>
        }
        extra={
          <Space>
            <Button 
              size="small"
              disabled={currentSectionIndex === 0}
              onClick={() => handleSectionChange(currentSectionIndex - 1)}
            >
              上一题型
            </Button>
            <Button 
              size="small"
              disabled={currentSectionIndex === paperDetail.paper.config.length - 1}
              onClick={() => handleSectionChange(currentSectionIndex + 1)}
            >
              下一题型
            </Button>
          </Space>
        }
      >
        <div style={{ marginBottom: 16 }}>
          <Text strong>题型要求：</Text>
          <Tag color="blue" style={{ marginLeft: 8 }}>
            {getTypeText(currentSection.type)} - 每题 {currentSection.score_each} 分
          </Tag>
          {filteredQuestions.length === 0 && (
            <Text type="secondary" style={{ marginLeft: 16 }}>
              暂无此类型题目
            </Text>
          )}
        </div>
        
        {filteredQuestions.length > 0 && (
          <Table
            columns={columns}
            dataSource={filteredQuestions}
            rowKey="id"
            pagination={{
              total: filteredQuestions.length,
              pageSize: 10,
              showSizeChanger: true,
              showQuickJumper: true,
              showTotal: (total) => `共 ${total} 道${getTypeText(currentSection.type)}`,
            }}
            scroll={{ x: 800 }}
          />
        )}
      </Card>
    </div>
  );
};

export default PaperEditPage; 