import { useEffect, useState } from 'react';
import { useParams, useNavigate, useLocation } from 'react-router-dom';
import { Card, Button, Space, Typography, Divider, Spin, message, Tag, Row, Col } from 'antd';
import { ArrowLeftOutlined, FileTextOutlined } from '@ant-design/icons';
import type { PaperDetail } from '../api';
import { getPaperDetail } from '../api';

const { Title, Text } = Typography;


export default function PaperPreviewPage() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const location = useLocation();
  const [paperDetail, setPaperDetail] = useState<PaperDetail | null>(null);
  const [loading, setLoading] = useState(false);
  const [exportLoading, setExportLoading] = useState<'pdf' | 'docx' | null>(null);

  // 获取试卷详情
  const fetchPaperDetail = async () => {
    if (!id) {
      message.error('试卷ID不存在');
      navigate('/papers');
      return;
    }

    setLoading(true);
    try {
      const res = await getPaperDetail(parseInt(id));
      if (res.code === 0 && res.data) {
        setPaperDetail(res.data);
      } else {
        message.error(res.msg || '获取试卷详情失败');
        navigate('/papers');
      }
    } catch (e) {
      message.error('网络错误');
      console.error('获取试卷详情出错:', e);
      navigate('/papers');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchPaperDetail();
  }, [id]);

  // 导出试卷
  const handleExport = async (format: 'pdf' | 'docx') => {
    if (!paperDetail) return;

    setExportLoading(format);
    try {
      // 创建下载链接
      const downloadUrl = `/api/papers/${paperDetail.paper.id}/export?format=${format}`;
      
      // 创建隐藏的链接元素并触发下载
      const link = document.createElement('a');
      link.href = downloadUrl;
      link.download = `${paperDetail.paper.title}.${format}`;
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);
      
      message.success(`${format.toUpperCase()} 文件导出成功`);
    } catch (e) {
      message.error('导出失败');
      console.error('导出出错:', e);
    } finally {
      setExportLoading(null);
    }
  };

  // 按题型分组题目
  const getQuestionsBySection = () => {
    if (!paperDetail) return [];
    
    const sections: Array<{
      config: any;
      questions: typeof paperDetail.questions;
    }> = [];

    let questionIndex = 0;
    
    for (const sectionConfig of paperDetail.paper.config) {
      const sectionQuestions = paperDetail.questions.slice(questionIndex, questionIndex + sectionConfig.count);
      sections.push({
        config: sectionConfig,
        questions: sectionQuestions
      });
      questionIndex += sectionConfig.count;
    }
    
    return sections;
  };

  // 计算总分
  const getTotalScore = () => {
    if (!paperDetail) return 0;
    return paperDetail.paper.config.reduce((total, section) => {
      return total + (section.count * section.score_each);
    }, 0);
  };

  // 格式化创建时间
  const formatDate = (dateString: string) => {
    try {
      return new Date(dateString).toLocaleString('zh-CN');
    } catch {
      return dateString;
    }
  };

  if (loading) {
    return (
      <div style={{ textAlign: 'center', padding: '100px 0' }}>
        <Spin size="large" tip="加载试卷中..." />
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

  const sections = getQuestionsBySection();
  const totalScore = getTotalScore();

  return (
    <div style={{ maxWidth: '900px', margin: '0 auto' }}>
      {/* 页面头部 */}
      <div style={{ marginBottom: 24 }}>
        <Space>
          <Button 
            icon={<ArrowLeftOutlined />} 
            onClick={() => {
              // 根据来源页面决定返回位置
              const from = location.state?.from;
              if (from === 'papers') {
                navigate('/papers');
              } else {
                navigate('/questions');
              }
            }}
          >
            {location.state?.from === 'papers' ? '返回试卷管理' : '返回题库'}
          </Button>
          
          <Divider type="vertical" />
          
          <Button 
            type="primary"
            icon={<FileTextOutlined />}
            loading={exportLoading === 'docx'}
            onClick={() => handleExport('docx')}
          >
            导出 Word
          </Button>
        </Space>
      </div>

      {/* 试卷信息卡片 */}
      <Card style={{ marginBottom: 24 }}>
        <Row gutter={[16, 16]}>
          <Col span={24}>
            <Title level={2} style={{ textAlign: 'center', margin: 0 }}>
              {paperDetail.paper.title}
            </Title>
          </Col>
          
          <Col span={12}>
            <Text strong>试卷ID：</Text>
            <Text>{paperDetail.paper.id}</Text>
          </Col>
          
          <Col span={12}>
            <Text strong>创建时间：</Text>
            <Text>{formatDate(paperDetail.paper.created_at)}</Text>
          </Col>
          
          <Col span={12}>
            <Text strong>总题数：</Text>
            <Text>{paperDetail.questions.length} 题</Text>
          </Col>
          
          <Col span={12}>
            <Text strong>总分：</Text>
            <Text>{totalScore} 分</Text>
          </Col>
          
          <Col span={24}>
            <Text strong>题型分布：</Text>
            <Space wrap>
              {paperDetail.paper.config.map((section, index) => (
                <Tag key={index} color="blue">
                  {section.name}: {section.count}题 × {section.score_each}分
                </Tag>
              ))}
            </Space>
          </Col>
        </Row>
      </Card>

      {/* 试卷内容 */}
      <Card>
        <div style={{ padding: '20px 0' }}>
          {sections.map((section, sectionIndex) => (
            <div key={sectionIndex} style={{ marginBottom: sectionIndex < sections.length - 1 ? 40 : 0 }}>
              {/* 大题标题 */}
              <div style={{ marginBottom: 20 }}>
                <Title level={3} style={{ margin: 0, color: '#1890ff' }}>
                  {section.config.name}
                  <Text style={{ fontSize: '14px', color: '#666', fontWeight: 'normal', marginLeft: 8 }}>
                    (每题 {section.config.score_each} 分，共 {section.config.count} 题，
                    计 {section.config.count * section.config.score_each} 分)
                  </Text>
                </Title>
              </div>

              {/* 题目列表 */}
              {section.questions.map((questionItem, questionIndex) => {
                const { topic } = questionItem;
                const questionNumber = sectionIndex === 0 ? questionIndex + 1 : 
                  sections.slice(0, sectionIndex).reduce((sum, s) => sum + s.questions.length, 0) + questionIndex + 1;

                return (
                  <div key={questionItem.id} style={{ marginBottom: 24, paddingLeft: 16 }}>
                    {/* 题目标题 */}
                    <div style={{ marginBottom: 12 }}>
                      <Text strong style={{ fontSize: '16px' }}>
                        {questionNumber}. {topic.title}
                      </Text>
                    </div>

                    {/* 选项（非编程题） */}
                    {topic.type_id !== 3 && topic.answers && topic.answers.length > 0 && (
                      <div style={{ marginBottom: 12, paddingLeft: 20 }}>
                        {topic.answers.map((answer, answerIndex) => (
                          <div key={answerIndex} style={{ marginBottom: 6 }}>
                            <Text>{answer}</Text>
                          </div>
                        ))}
                      </div>
                    )}

                    {/* 正确答案（非编程题，仅预览时显示） */}
                    {topic.type_id !== 3 && topic.right && topic.right.length > 0 && (
                      <div style={{ paddingLeft: 20 }}>
                        <Text type="success" strong>
                          正确答案：{topic.right.join(', ')}
                        </Text>
                      </div>
                    )}

                    {/* 编程题提示 */}
                    {topic.type_id === 3 && (
                      <div style={{ paddingLeft: 20, padding: '12px', backgroundColor: '#f6ffed', border: '1px solid #b7eb8f', borderRadius: '6px' }}>
                        <Text type="secondary">
                          这是一道编程题，请在答题时提供完整的代码实现。
                        </Text>
                      </div>
                    )}
                  </div>
                );
              })}
            </div>
          ))}
        </div>
      </Card>

      {/* 页面底部 */}
      <div style={{ textAlign: 'center', padding: '40px 0 20px', color: '#999' }}>
        <Divider />
        <Text type="secondary">
          本试卷由考试出题系统自动生成 | 武汉科技大学 徐伽炜
        </Text>
      </div>
    </div>
  );
} 