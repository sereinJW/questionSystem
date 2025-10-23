import { useState } from 'react';
import { Modal, Form, Input, Button, Space, Select, InputNumber, Card, message, Divider } from 'antd';
import { PlusOutlined, DeleteOutlined, ThunderboltOutlined } from '@ant-design/icons';
import type { AIGeneratePaperRequest, AIGeneratePaperSection } from '../api';

const { Option } = Select;

interface AIGeneratePaperModalProps {
  visible: boolean;
  onCancel: () => void;
  onSuccess: (paperId: number) => void;
}

// 题型选项
const typeOptions = [
  { value: 1, label: '单选题' },
  { value: 2, label: '多选题' },
  { value: 3, label: '编程题' },
];

// 难度选项
const difficultyOptions = [
  { value: 1, label: '简单' },
  { value: 2, label: '中等' },
  { value: 3, label: '困难' },
];

// 语言选项
const languageOptions = [
  { value: 'go', label: 'Go' },
  { value: 'javascript', label: 'JavaScript' },
  { value: 'java', label: 'Java' },
  { value: 'python', label: 'Python' },
  { value: 'c++', label: 'C++' },
];

export default function AIGeneratePaperModal({ visible, onCancel, onSuccess }: AIGeneratePaperModalProps) {
  const [form] = Form.useForm();
  const [loading, setLoading] = useState(false);
  const [sections, setSections] = useState<AIGeneratePaperSection[]>([
    {
      name: '一、单选题',
      type: 1,
      count: 10,
      score_each: 2,
      difficulty: 1,
      language: 'go',
      keyword: 'Go语言基础',
    },
  ]);

  // 添加大题
  const handleAddSection = () => {
    const sectionNumber = sections.length + 1;
    const chineseNumbers = ['一', '二', '三', '四', '五', '六', '七', '八', '九', '十'];
    const prefix = sectionNumber <= 10 ? chineseNumbers[sectionNumber - 1] : sectionNumber.toString();
    
    setSections([
      ...sections,
      {
        name: `${prefix}、单选题`,
        type: 1,
        count: 5,
        score_each: 2,
        difficulty: 1,
        language: 'go',
        keyword: '',
      },
    ]);
  };

  // 删除大题
  const handleRemoveSection = (index: number) => {
    if (sections.length === 1) {
      message.warning('至少需要保留一个大题');
      return;
    }
    setSections(sections.filter((_, i) => i !== index));
  };

  // 更新大题配置
  const handleSectionChange = (index: number, field: keyof AIGeneratePaperSection, value: any) => {
    const newSections = [...sections];
    newSections[index] = { ...newSections[index], [field]: value };
    setSections(newSections);
  };

  // 提交生成
  const handleSubmit = async () => {
    try {
      // 验证表单
      const values = await form.validateFields();
      
      // 验证大题配置
      for (let i = 0; i < sections.length; i++) {
        const section = sections[i];
        if (!section.keyword || section.keyword.trim() === '') {
          message.error(`请填写第 ${i + 1} 个大题的关键词`);
          return;
        }
        if (section.count <= 0) {
          message.error(`第 ${i + 1} 个大题的题目数量必须大于0`);
          return;
        }
        if (section.score_each <= 0) {
          message.error(`第 ${i + 1} 个大题的每题分数必须大于0`);
          return;
        }
      }

      // 计算总题数
      const totalQuestions = sections.reduce((sum, s) => sum + s.count, 0);
      if (totalQuestions > 100) {
        message.error('单次生成题目总数不能超过100道');
        return;
      }

      setLoading(true);

      // 构造请求数据
      const request: AIGeneratePaperRequest = {
        title: values.title,
        sections: sections,
      };

      // 调用API
      const { aiGeneratePaper } = await import('../api');
      const res = await aiGeneratePaper(request);

      if (res.code === 0 && res.data) {
        message.success('AI生成试卷成功！');
        form.resetFields();
        setSections([
          {
            name: '一、单选题',
            type: 1,
            count: 10,
            score_each: 2,
            difficulty: 1,
            language: 'go',
            keyword: 'Go语言基础',
          },
        ]);
        onSuccess(res.data.id);
      } else {
        message.error(res.msg || 'AI生成试卷失败');
      }
    } catch (error: any) {
      if (error.errorFields) {
        message.error('请填写完整的试卷信息');
      } else {
        message.error('生成失败，请稍后重试');
        console.error('AI生成试卷错误:', error);
      }
    } finally {
      setLoading(false);
    }
  };

  // 计算总分
  const getTotalScore = () => {
    return sections.reduce((sum, section) => sum + section.count * section.score_each, 0);
  };

  // 计算总题数
  const getTotalQuestions = () => {
    return sections.reduce((sum, section) => sum + section.count, 0);
  };

  return (
    <Modal
      title={
        <Space>
          <ThunderboltOutlined style={{ color: '#1890ff' }} />
          <span>AI智能组卷</span>
        </Space>
      }
      open={visible}
      onCancel={onCancel}
      width={800}
      footer={[
        <Button key="cancel" onClick={onCancel} disabled={loading}>
          取消
        </Button>,
        <Button key="submit" type="primary" loading={loading} onClick={handleSubmit}>
          {loading ? '正在生成中...' : '开始生成'}
        </Button>,
      ]}
    >
      <Form form={form} layout="vertical">
        <Form.Item
          label="试卷标题"
          name="title"
          rules={[{ required: true, message: '请输入试卷标题' }]}
        >
          <Input placeholder="例如：Go语言期末考试试卷" />
        </Form.Item>

        <Divider orientation="left">大题配置</Divider>

        <div style={{ marginBottom: 16 }}>
          <Space>
            <span>总题数：<strong>{getTotalQuestions()}</strong> 道</span>
            <Divider type="vertical" />
            <span>总分：<strong>{getTotalScore()}</strong> 分</span>
          </Space>
        </div>

        {sections.map((section, index) => (
          <Card
            key={index}
            size="small"
            style={{ marginBottom: 16 }}
            title={`第 ${index + 1} 个大题`}
            extra={
              sections.length > 1 && (
                <Button
                  type="text"
                  danger
                  size="small"
                  icon={<DeleteOutlined />}
                  onClick={() => handleRemoveSection(index)}
                >
                  删除
                </Button>
              )
            }
          >
            <Space direction="vertical" style={{ width: '100%' }} size="small">
              <Space wrap style={{ width: '100%' }}>
                <div style={{ width: 200 }}>
                  <div style={{ marginBottom: 4, fontSize: 12, color: '#666' }}>大题名称</div>
                  <Input
                    value={section.name}
                    onChange={(e) => handleSectionChange(index, 'name', e.target.value)}
                    placeholder="例如：一、单选题"
                  />
                </div>

                <div style={{ width: 120 }}>
                  <div style={{ marginBottom: 4, fontSize: 12, color: '#666' }}>题型</div>
                  <Select
                    value={section.type}
                    onChange={(value) => handleSectionChange(index, 'type', value)}
                    style={{ width: '100%' }}
                  >
                    {typeOptions.map((opt) => (
                      <Option key={opt.value} value={opt.value}>
                        {opt.label}
                      </Option>
                    ))}
                  </Select>
                </div>

                <div style={{ width: 100 }}>
                  <div style={{ marginBottom: 4, fontSize: 12, color: '#666' }}>题目数量</div>
                  <InputNumber
                    value={section.count}
                    onChange={(value) => handleSectionChange(index, 'count', value || 1)}
                    min={1}
                    max={50}
                    style={{ width: '100%' }}
                  />
                </div>

                <div style={{ width: 100 }}>
                  <div style={{ marginBottom: 4, fontSize: 12, color: '#666' }}>每题分数</div>
                  <InputNumber
                    value={section.score_each}
                    onChange={(value) => handleSectionChange(index, 'score_each', value || 1)}
                    min={1}
                    max={100}
                    style={{ width: '100%' }}
                  />
                </div>
              </Space>

              <Space wrap style={{ width: '100%' }}>
                <div style={{ width: 120 }}>
                  <div style={{ marginBottom: 4, fontSize: 12, color: '#666' }}>难度</div>
                  <Select
                    value={section.difficulty}
                    onChange={(value) => handleSectionChange(index, 'difficulty', value)}
                    style={{ width: '100%' }}
                  >
                    {difficultyOptions.map((opt) => (
                      <Option key={opt.value} value={opt.value}>
                        {opt.label}
                      </Option>
                    ))}
                  </Select>
                </div>

                <div style={{ width: 120 }}>
                  <div style={{ marginBottom: 4, fontSize: 12, color: '#666' }}>编程语言</div>
                  <Select
                    value={section.language}
                    onChange={(value) => handleSectionChange(index, 'language', value)}
                    style={{ width: '100%' }}
                  >
                    {languageOptions.map((opt) => (
                      <Option key={opt.value} value={opt.value}>
                        {opt.label}
                      </Option>
                    ))}
                  </Select>
                </div>

                <div style={{ flex: 1, minWidth: 200 }}>
                  <div style={{ marginBottom: 4, fontSize: 12, color: '#666' }}>关键词</div>
                  <Input
                    value={section.keyword}
                    onChange={(e) => handleSectionChange(index, 'keyword', e.target.value)}
                    placeholder="例如：Go语言基础、数据结构"
                  />
                </div>
              </Space>

              <div style={{ fontSize: 12, color: '#999' }}>
                小计：{section.count} 题 × {section.score_each} 分 = {section.count * section.score_each} 分
              </div>
            </Space>
          </Card>
        ))}

        <Button
          type="dashed"
          block
          icon={<PlusOutlined />}
          onClick={handleAddSection}
          style={{ marginBottom: 16 }}
        >
          添加大题
        </Button>

        <div style={{ padding: 12, background: '#f0f5ff', borderRadius: 4, fontSize: 12, color: '#666' }}>
          <div style={{ marginBottom: 4 }}>💡 温馨提示：</div>
          <div>• AI将根据您的配置自动生成题目并组成试卷</div>
          <div>• 生成过程可能需要1-3分钟，请耐心等待</div>
          <div>• 单次生成题目总数不超过100道</div>
          <div>• 生成的题目会自动保存到题库中</div>
        </div>
      </Form>
    </Modal>
  );
}
