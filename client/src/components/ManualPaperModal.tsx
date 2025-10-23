import { useState, useEffect } from 'react';
import { Modal, Form, Input, Button, Space, Select, InputNumber, Card, message, Table, Tag, Progress } from 'antd';
import { PlusOutlined, DeleteOutlined, FileTextOutlined } from '@ant-design/icons';
import type { CreatePaperRequest, PaperSection, Topic } from '../api';
import { getQuestions } from '../api';

interface ManualPaperModalProps {
  visible: boolean;
  onCancel: () => void;
  onSuccess: (paperId: number) => void;
}

const typeOptions = [
  { value: 1, label: '单选题' },
  { value: 2, label: '多选题' },
  { value: 3, label: '编程题' },
];

const diffOptions = [
  { value: 1, label: '简单' },
  { value: 2, label: '中等' },
  { value: 3, label: '困难' },
];

export default function ManualPaperModal({ visible, onCancel, onSuccess }: ManualPaperModalProps) {
  const [form] = Form.useForm();
  const [step, setStep] = useState<'config' | 'select'>('config');
  const [paperConfig, setPaperConfig] = useState<{ title: string; sections: PaperSection[] }>({
    title: '',
    sections: [],
  });
  const [currentSectionIndex, setCurrentSectionIndex] = useState(0);
  const [sectionQuestions, setSectionQuestions] = useState<{ [key: number]: number[] }>({});
  const [allQuestions, setAllQuestions] = useState<Topic[]>([]);
  const [loading, setLoading] = useState(false);
  const [searchKeyword, setSearchKeyword] = useState('');
  const [filterDifficulty, setFilterDifficulty] = useState<number | undefined>();

  // 加载题库数据
  useEffect(() => {
    if (step === 'select') {
      loadQuestions();
    }
  }, [step]);

  const loadQuestions = async () => {
    setLoading(true);
    try {
      const res = await getQuestions();
      if (res.code === 0) {
        setAllQuestions(res.data || []);
      } else {
        message.error('加载题库失败');
      }
    } catch (error) {
      message.error('加载题库失败');
    } finally {
      setLoading(false);
    }
  };

  // 配置步骤：确认配置
  const handleConfigOk = async () => {
    try {
      const values = await form.validateFields();
      const config = {
        title: values.title,
        sections: values.sections.map((section: any) => ({
          name: section.name,
          type: section.type,
          count: section.count,
          score_each: section.score_each,
        })),
      };
      setPaperConfig(config);
      setSectionQuestions({});
      setCurrentSectionIndex(0);
      setStep('select');
    } catch (error) {
      console.error('表单验证失败:', error);
    }
  };

  // 选题步骤：处理题目选择
  const handleQuestionSelect = (questionId: number, checked: boolean) => {
    const currentSectionQuestionIds = sectionQuestions[currentSectionIndex] || [];
    const currentSection = paperConfig.sections[currentSectionIndex];

    if (checked) {
      if (currentSectionQuestionIds.length >= currentSection.count) {
        message.warning(`${currentSection.name} 最多只能选择 ${currentSection.count} 道题`);
        return;
      }
      setSectionQuestions({
        ...sectionQuestions,
        [currentSectionIndex]: [...currentSectionQuestionIds, questionId],
      });
    } else {
      setSectionQuestions({
        ...sectionQuestions,
        [currentSectionIndex]: currentSectionQuestionIds.filter((id) => id !== questionId),
      });
    }
  };

  // 检查是否所有题型都选择完毕
  const isSelectionComplete = (): boolean => {
    for (let i = 0; i < paperConfig.sections.length; i++) {
      const section = paperConfig.sections[i];
      const sectionQuestionIds = sectionQuestions[i] || [];
      if (sectionQuestionIds.length !== section.count) {
        return false;
      }
    }
    return true;
  };

  // 获取选题进度
  const getSelectionProgress = () => {
    return paperConfig.sections.map((section, index) => {
      const sectionQuestionIds = sectionQuestions[index] || [];
      const selected = sectionQuestionIds.length;
      const required = section.count;
      return {
        typeName: section.name,
        selected,
        required,
        percentage: required > 0 ? Math.round((selected / required) * 100) : 0,
      };
    });
  };

  // 获取当前大题类型的题目
  const getFilteredQuestions = () => {
    if (paperConfig.sections.length === 0) return [];
    const currentSection = paperConfig.sections[currentSectionIndex];
    return allQuestions.filter((q) => {
      const matchType = q.type_id === currentSection.type;
      const matchKeyword = searchKeyword
        ? q.title.toLowerCase().includes(searchKeyword.toLowerCase()) ||
          q.keyword.toLowerCase().includes(searchKeyword.toLowerCase())
        : true;
      const matchDifficulty = filterDifficulty ? q.difficulty === filterDifficulty : true;
      return matchType && matchKeyword && matchDifficulty;
    });
  };

  // 生成试卷
  const handleGeneratePaper = async () => {
    if (!isSelectionComplete()) {
      message.error('请完成所有题型的题目选择');
      return;
    }

    // 合并所有大题的题目ID
    const allQuestionIds: number[] = [];
    paperConfig.sections.forEach((_, index) => {
      const sectionQuestionIds = sectionQuestions[index] || [];
      allQuestionIds.push(...sectionQuestionIds);
    });

    try {
      setLoading(true);
      const { createPaper } = await import('../api');
      const request: CreatePaperRequest = {
        title: paperConfig.title,
        sections: paperConfig.sections,
        question_ids: allQuestionIds,
      };

      const res = await createPaper(request);
      if (res.code === 0 && res.data) {
        message.success('试卷创建成功');
        handleClose();
        onSuccess(res.data.id);
      } else {
        message.error(res.msg || '创建试卷失败');
      }
    } catch (error) {
      message.error('创建试卷失败');
      console.error('创建试卷错误:', error);
    } finally {
      setLoading(false);
    }
  };

  // 关闭对话框
  const handleClose = () => {
    setStep('config');
    setPaperConfig({ title: '', sections: [] });
    setSectionQuestions({});
    setCurrentSectionIndex(0);
    setSearchKeyword('');
    setFilterDifficulty(undefined);
    form.resetFields();
    onCancel();
  };

  // 返回配置步骤
  const handleBackToConfig = () => {
    Modal.confirm({
      title: '确认返回配置？',
      content: '返回后将丢失当前的选题进度',
      okText: '确认返回',
      cancelText: '继续选题',
      onOk: () => {
        setStep('config');
        setSectionQuestions({});
        setCurrentSectionIndex(0);
      },
    });
  };

  // 表格列
  const columns = [
    {
      title: '选择',
      key: 'select',
      width: 60,
      render: (_: any, record: Topic) => {
        const currentSectionQuestionIds = sectionQuestions[currentSectionIndex] || [];
        const currentSection = paperConfig.sections[currentSectionIndex];
        const isChecked = currentSectionQuestionIds.includes(record.id!);
        const isDisabled = !isChecked && currentSectionQuestionIds.length >= currentSection.count;

        return (
          <input
            type="checkbox"
            checked={isChecked}
            disabled={isDisabled}
            onChange={(e) => handleQuestionSelect(record.id!, e.target.checked)}
          />
        );
      },
    },
    {
      title: '题干',
      dataIndex: 'title',
      key: 'title',
      ellipsis: true,
    },
    {
      title: '难度',
      dataIndex: 'difficulty',
      key: 'difficulty',
      width: 80,
      render: (v: number) => (
        <Tag color={v === 3 ? 'red' : v === 2 ? 'orange' : 'green'}>
          {diffOptions.find((d) => d.value === v)?.label}
        </Tag>
      ),
    },
    {
      title: '关键词',
      dataIndex: 'keyword',
      key: 'keyword',
      width: 120,
    },
    {
      title: '来源',
      dataIndex: 'is_ai',
      key: 'is_ai',
      width: 80,
      render: (v: number) => (v ? <Tag color="blue">AI</Tag> : <Tag color="green">手工</Tag>),
    },
  ];

  return (
    <Modal
      title={
        <Space>
          <FileTextOutlined style={{ color: '#1890ff' }} />
          <span>{step === 'config' ? '手动出卷 - 配置试卷' : '手动出卷 - 选择题目'}</span>
        </Space>
      }
      open={visible}
      onCancel={handleClose}
      width={step === 'config' ? 800 : 1000}
      footer={
        step === 'config'
          ? [
              <Button key="cancel" onClick={handleClose}>
                取消
              </Button>,
              <Button key="submit" type="primary" onClick={handleConfigOk}>
                下一步：选择题目
              </Button>,
            ]
          : [
              <Button key="back" onClick={handleBackToConfig}>
                返回配置
              </Button>,
              <Button key="cancel" onClick={handleClose}>
                取消
              </Button>,
              <Button
                key="submit"
                type="primary"
                disabled={!isSelectionComplete()}
                loading={loading}
                onClick={handleGeneratePaper}
              >
                生成试卷
              </Button>,
            ]
      }
    >
      {step === 'config' ? (
        <Form
          form={form}
          layout="vertical"
          initialValues={{
            sections: [{ name: '一、单选题', type: 1, count: 10, score_each: 2 }],
          }}
        >
          <Form.Item name="title" label="试卷标题" rules={[{ required: true, message: '请输入试卷标题' }]}>
            <Input placeholder="请输入试卷标题，如：Go语言基础知识测试" />
          </Form.Item>

          <Form.List name="sections">
            {(fields, { add, remove }) => (
              <>
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 16 }}>
                  <h4>大题配置</h4>
                  <Button type="dashed" icon={<PlusOutlined />} onClick={() => add()}>
                    添加大题
                  </Button>
                </div>

                {fields.map(({ key, name, ...restField }) => (
                  <Card key={key} size="small" style={{ marginBottom: 16 }}>
                    <div style={{ display: 'flex', gap: 16, alignItems: 'end' }}>
                      <Form.Item
                        {...restField}
                        name={[name, 'name']}
                        label="大题名称"
                        rules={[{ required: true, message: '请输入大题名称' }]}
                        style={{ flex: 1 }}
                      >
                        <Input placeholder="如：一、单选题" />
                      </Form.Item>

                      <Form.Item
                        {...restField}
                        name={[name, 'type']}
                        label="题目类型"
                        rules={[{ required: true, message: '请选择题目类型' }]}
                        style={{ width: 120 }}
                      >
                        <Select options={typeOptions} placeholder="题型" />
                      </Form.Item>

                      <Form.Item
                        {...restField}
                        name={[name, 'count']}
                        label="题目数量"
                        rules={[{ required: true, message: '请输入题目数量' }]}
                        style={{ width: 100 }}
                      >
                        <InputNumber min={1} max={50} placeholder="数量" />
                      </Form.Item>

                      <Form.Item
                        {...restField}
                        name={[name, 'score_each']}
                        label="每题分数"
                        rules={[{ required: true, message: '请输入每题分数' }]}
                        style={{ width: 100 }}
                      >
                        <InputNumber min={1} max={20} placeholder="分数" />
                      </Form.Item>

                      {fields.length > 1 && (
                        <Button
                          type="text"
                          danger
                          icon={<DeleteOutlined />}
                          onClick={() => remove(name)}
                          style={{ marginBottom: 24 }}
                        >
                          删除
                        </Button>
                      )}
                    </div>
                  </Card>
                ))}
              </>
            )}
          </Form.List>
        </Form>
      ) : (
        <>
          {/* 总体进度 */}
          <Card style={{ marginBottom: 16, backgroundColor: '#f6ffed', borderColor: '#b7eb8f' }}>
            <div style={{ marginBottom: 12 }}>
              <h4 style={{ margin: 0, color: '#52c41a' }}>正在为试卷《{paperConfig.title}》选择题目</h4>
            </div>

            <div style={{ display: 'flex', flexWrap: 'wrap', gap: '16px', marginBottom: 16 }}>
              {getSelectionProgress().map((progress, index) => (
                <div key={index} style={{ minWidth: '200px' }}>
                  <div style={{ marginBottom: 4 }}>
                    {progress.typeName}：{progress.selected}/{progress.required} 题
                  </div>
                  <Progress percent={progress.percentage} size="small" status={progress.percentage === 100 ? 'success' : 'active'} />
                </div>
              ))}
            </div>

            {/* 大题切换 */}
            <div style={{ textAlign: 'center' }}>
              <Space wrap>
                {paperConfig.sections.map((section, index) => {
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
                        fontSize: '14px',
                      }}
                      onClick={() => setCurrentSectionIndex(index)}
                    >
                      {section.name}: {sectionQuestionIds.length}/{section.count}
                      {isComplete && ' ✓'}
                    </Tag>
                  );
                })}
              </Space>
            </div>
          </Card>

          {/* 当前大题选择提示 */}
          <Card style={{ marginBottom: 16 }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
              <div>
                <h4 style={{ margin: 0, color: '#1890ff' }}>当前选择：{paperConfig.sections[currentSectionIndex]?.name}</h4>
                <div style={{ marginTop: 8 }}>
                  <Tag color="blue">题型：{typeOptions.find((t) => t.value === paperConfig.sections[currentSectionIndex]?.type)?.label}</Tag>
                  <Tag color="green">每题分数：{paperConfig.sections[currentSectionIndex]?.score_each} 分</Tag>
                  <Tag color="orange">需要选择：{paperConfig.sections[currentSectionIndex]?.count} 道题</Tag>
                  <Tag color="purple">已选择：{(sectionQuestions[currentSectionIndex] || []).length} 道题</Tag>
                </div>
              </div>
            </div>
          </Card>

          {/* 筛选条件 */}
          <Space style={{ marginBottom: 16 }}>
            <Input.Search
              placeholder="搜索题干/关键词"
              allowClear
              style={{ width: 250 }}
              value={searchKeyword}
              onChange={(e) => setSearchKeyword(e.target.value)}
            />
            <Select
              placeholder="难度筛选"
              allowClear
              style={{ width: 120 }}
              options={diffOptions}
              value={filterDifficulty}
              onChange={setFilterDifficulty}
            />
          </Space>

          {/* 题目列表 */}
          <Table
            columns={columns}
            dataSource={getFilteredQuestions()}
            rowKey={(record) => record.id!}
            loading={loading}
            pagination={{
              pageSize: 10,
              showTotal: (total) => `共 ${total} 道题`,
            }}
          />
        </>
      )}
    </Modal>
  );
}
