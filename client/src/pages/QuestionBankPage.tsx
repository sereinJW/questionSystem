import { useEffect, useState } from 'react';
import { Table, Button, Input, Select, Space, Modal, message, Tag, Form, Divider, InputNumber } from 'antd';
import type { ColumnsType } from 'antd/es/table';
import type { Topic, AskParams } from '../api';
import { getQuestions, createQuestions, addQuestion, editQuestion, deleteQuestion } from '../api';
import '@ant-design/v5-patch-for-react-19';

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
const languageOptions = [
  { value: 'go', label: 'Go' },
  { value: 'javascript', label: 'JavaScript' },
  { value: 'java', label: 'Java' },
  { value: 'python', label: 'Python' },
  { value: 'c++', label: 'C++' },
];

export default function QuestionBankPage() {
  const [data, setData] = useState<Topic[]>([]);
  const [loading, setLoading] = useState(false);
  const [search, setSearch] = useState('');
  const [type, setType] = useState<number | undefined>();
  const [editModal, setEditModal] = useState(false);
  const [editTopic, setEditTopic] = useState<Topic | null>(null);
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);
  const [form] = Form.useForm();
  const [addModal, setAddModal] = useState(false);
  const [addForm] = Form.useForm();
  const [aiModal, setAiModal] = useState(false);
  const [aiForm] = Form.useForm();
  const [aiLoading, setAiLoading] = useState(false);
  // 添加选中的行状态
  const [selectedRowKeys, setSelectedRowKeys] = useState<React.Key[]>([]);
  const [batchDeleteLoading, setBatchDeleteLoading] = useState(false);
  // 新增AI预览相关状态
  const [aiPreviewVisible, setAiPreviewVisible] = useState(false);
  const [aiPreviewData, setAiPreviewData] = useState<Topic[]>([]);
  const [aiPreviewLoading, setAiPreviewLoading] = useState(false);
  
  // 获取题库数据
  const fetchData = async () => {
    setLoading(true);
    try {
      const res = await getQuestions();
      if (res.code === 0) {
        console.log('获取题库数据成功，数据条数:', res.data?.length);
        console.log('题库数据示例(第一条):', res.data?.[0]);
        setData(res.data || []);
        // 清空选中状态
        setSelectedRowKeys([]);
      } else {
        message.error(res.msg || '获取题库失败');
      }
    } catch (e) {
      message.error('网络错误');
      console.error('获取题库出错:', e);
    }
    setLoading(false);
  };

  useEffect(() => {
    fetchData();
  }, []);

  // 打开编辑弹窗时填充表单
  useEffect(() => {
    if (editModal && editTopic) {
      form.setFieldsValue({
        ...editTopic,
        answers: editTopic.type_id !== 3 ? editTopic.answers?.join(',') || '' : undefined,
        right: editTopic.type_id !== 3 ? editTopic.right?.join(',') || '' : undefined,
      });
    }
  }, [editModal, editTopic, form]);

  // 搜索和筛选
  const filtered = data.filter(item => {
    const matchType = type ? item.type_id === type : true;
    const matchSearch = search ? 
      item.title.toLowerCase().includes(search.toLowerCase()) || 
      item.keyword.toLowerCase().includes(search.toLowerCase()) : 
      true;
    return matchType && matchSearch;
  });

  // 编辑题目
  const handleEdit = (record: Topic) => {
    console.log('准备编辑题目，完整数据:', record);
    if (!record.id) {
      message.error('题目ID不存在，无法编辑');
      return;
    }
    setEditTopic(record);
    setEditModal(true);
  };

  // 删除题目
  const handleDelete = async (record: Topic) => {
    console.log('准备删除题目:', record);
    if (!record.id) {
      message.error('题目ID不存在，无法删除');
      return;
    }
    
    Modal.confirm({
      title: '确认删除该题目？',
      content: '删除后无法恢复，请确认',
      okText: '确认删除',
      cancelText: '取消',
      onOk: async () => {
        try {
          console.log('发送删除请求, 题目数据:', record);
          
          const res = await deleteQuestion(record);
          console.log('删除请求响应:', res);
          if (res.code === 0) {
            message.success('删除成功');
            await fetchData(); // 确保重新获取数据
          } else {
            message.error(res.msg || '删除失败');
          }
        } catch (e) {
          message.error('网络错误');
          console.error('删除出错:', e);
        }
      },
    });
  };

  // 批量删除题目
  const handleBatchDelete = async () => {
    if (selectedRowKeys.length === 0) {
      message.info('请选择要删除的题目');
      return;
    }
    
    console.log('准备批量删除，选中的行：', selectedRowKeys);
    console.log('数据示例 - 第一项ID类型:', data.length > 0 ? typeof data[0].id : '无数据');
    console.log('选中行ID的类型:', typeof selectedRowKeys[0]);
    
    Modal.confirm({
      title: `确认删除选中的 ${selectedRowKeys.length} 道题目？`,
      content: '删除后无法恢复，请确认',
      okText: '确认删除',
      cancelText: '取消',
      onOk: async () => {
        try {
          setBatchDeleteLoading(true);
          
          // 找出所有选中的题目
          const selectedTopics = data.filter(item => 
            item.id && selectedRowKeys.includes(String(item.id))
          );
          console.log('找到要删除的题目：', selectedTopics.length, '条');
          
          // 如果没有找到匹配的题目，提前返回
          if (selectedTopics.length === 0) {
            message.error('未找到选中的题目数据');
            return;
          }
          
          // 记录成功和失败的数量
          let successCount = 0;
          let failCount = 0;
          
          // 串行处理每个删除请求
          for (const topic of selectedTopics) {
            try {
              const res = await deleteQuestion(topic);
              if (res.code === 0) {
                successCount++;
              } else {
                failCount++;
                console.error(`删除题目ID ${topic.id} 失败: ${res.msg}`);
              }
            } catch (e) {
              failCount++;
              console.error(`删除题目ID ${topic.id} 发生错误:`, e);
            }
          }
          
          // 显示结果
          if (successCount > 0 && failCount === 0) {
            message.success(`成功删除 ${successCount} 道题目`);
          } else if (successCount > 0 && failCount > 0) {
            message.warning(`成功删除 ${successCount} 道题目，${failCount} 道题目删除失败`);
          } else {
            message.error('批量删除失败');
          }
          
          // 重新加载数据
          await fetchData();
        } catch (e) {
          message.error('批量删除处理出错');
          console.error('批量删除出错:', e);
        } finally {
          setBatchDeleteLoading(false);
        }
      },
    });
  };

  // 表格行选择配置
  const rowSelection = {
    selectedRowKeys,
    onChange: (newSelectedRowKeys: React.Key[]) => {
      setSelectedRowKeys(newSelectedRowKeys);
    }
  };

  // 保存编辑
  const handleEditOk = async () => {
    try {
      const values = await form.validateFields();
      if (!editTopic || !editTopic.id) {
        message.error('题目ID不存在，无法保存');
        return;
      }
      
      // 确保id字段存在
      let payload: Topic = {
        ...editTopic,
        ...values,
      };
      
      // 根据题型处理答案和选项
      if (values.type_id !== 3) { // 非编程题
        payload.answers = values.answers ? values.answers.split(',').map((s: string) => s.trim()) : [];
        payload.right = values.right ? values.right.split(',').map((s: string) => s.trim()) : [];
      } else { // 编程题
        payload.answers = [];
        payload.right = [];
      }

      console.log('发送编辑请求:', payload);
      const res = await editQuestion(payload);
      console.log('编辑请求响应:', res);
      
      if (res.code === 0) {
        message.success('编辑成功');
        setEditModal(false);
        await fetchData(); // 确保重新获取数据
      } else {
        message.error(res.msg || '编辑失败');
      }
    } catch (e) {
      console.error('表单验证失败:', e);
    }
  };

  // 保存新增
  const handleAddOk = async () => {
    try {
      const values = await addForm.validateFields();
      
      let payload: any = { ...values };
      
      // 根据题型处理答案和选项
      if (values.type_id !== 3) { // 非编程题
        payload.answers = values.answers ? values.answers.split(',').map((s: string) => s.trim()) : [];
        payload.right = values.right ? values.right.split(',').map((s: string) => s.trim()) : [];
      } else { // 编程题
        payload.answers = [];
        payload.right = [];
      }

      const res = await addQuestion(payload);
      if (res.code === 0) {
        message.success('添加成功');
        setAddModal(false);
        addForm.resetFields();
        fetchData();
      } else {
        message.error(res.msg || '添加失败');
      }
    } catch (e) {
      console.error('表单验证失败:', e);
    }
  };

  // AI出题提交
  const handleAiOk = async () => {
    try {
      setAiLoading(true);
      const values = await aiForm.validateFields();
      
      // 确保number是数字类型而不是字符串
      const params: AskParams = {
        ...values,
        number: parseInt(values.number, 10)
      };
      
      // AI出题预览，调用接口获取生成的题目但不保存
      const res = await createQuestions(params);
      if (res.code === 0 && res.data?.length > 0) {
        // 设置预览数据
        setAiPreviewData(res.data);
        // 关闭配置窗口，打开预览窗口
        setAiModal(false);
        setAiPreviewVisible(true);
      } else {
        message.error(res.msg || 'AI出题失败');
      }
    } catch (e) {
      console.error('表单验证失败:', e);
    } finally {
      setAiLoading(false);
    }
  };

  // 保存生成的题目到数据库
  const handleSaveAiQuestions = async () => {
    if (!aiPreviewData || aiPreviewData.length === 0) {
      message.error('没有可保存的题目');
      return;
    }

    setAiPreviewLoading(true);
    try {
      let successCount = 0;
      let failCount = 0;

      // 逐个添加题目
      for (const topic of aiPreviewData) {
        // 构建添加题目的参数，明确设置is_ai=1表示这是AI生成的题目
        const payload = {
          title: topic.title,
          type_id: topic.type_id,
          difficulty: topic.difficulty,
          language: topic.language,
          keyword: topic.keyword,
          answers: Array.isArray(topic.answers) ? topic.answers : [],
          right: Array.isArray(topic.right) ? topic.right : [],
          is_ai: 1  // 设置为AI生成的题目
        };
        
        const res = await addQuestion(payload);
        if (res.code === 0) {
          successCount++;
        } else {
          failCount++;
          console.error(`添加题目失败: ${res.msg}`);
        }
      }

      // 显示结果
      if (successCount > 0 && failCount === 0) {
        message.success(`成功添加 ${successCount} 道题目`);
        // 关闭预览窗口
        setAiPreviewVisible(false);
        // 清空预览数据
        setAiPreviewData([]);
        // 重新获取题库数据
        fetchData();
      } else if (successCount > 0 && failCount > 0) {
        message.warning(`成功添加 ${successCount} 道题目，${failCount} 道题目添加失败`);
        fetchData();
      } else {
        message.error('添加题目失败');
      }
    } catch (e) {
      message.error('保存题目出错');
      console.error('保存题目出错:', e);
    } finally {
      setAiPreviewLoading(false);
    }
  };

  // 表格列
  const columns: ColumnsType<Topic> = [
    { 
      title: '题干', 
      dataIndex: 'title', 
      key: 'title', 
      width: 300,
      ellipsis: true,
      render: text => <div style={{ maxHeight: '60px', overflow: 'hidden', textOverflow: 'ellipsis' }}>{text}</div>
    },
    { 
      title: '类型', 
      dataIndex: 'type_id', 
      key: 'type_id',
      width: 100,
      render: v => <Tag color={v === 3 ? 'purple' : (v === 2 ? 'blue' : 'green')}>{typeOptions.find(t => t.value === v)?.label}</Tag>
    },
    { 
      title: '难度', 
      dataIndex: 'difficulty', 
      key: 'difficulty',
      width: 100,
      render: v => <Tag color={v === 3 ? 'red' : (v === 2 ? 'orange' : 'success')}>{diffOptions.find(d => d.value === v)?.label}</Tag>
    },
    { title: '语言', dataIndex: 'language', key: 'language', width: 100 },
    { title: '关键词', dataIndex: 'keyword', key: 'keyword', width: 100 },
    { 
      title: '来源', 
      dataIndex: 'is_ai', 
      key: 'is_ai',
      width: 100,
      render: v => v ? <Tag color="blue">AI</Tag> : <Tag color="green">手工</Tag>
    },
    {
      title: '操作',
      key: 'action',
      width: 150,
      render: (_, record) => (
        <Space>
          <Button type="link" size="small" onClick={() => handleEdit(record)}>编辑</Button>
          <Button type="link" size="small" danger onClick={() => handleDelete(record)}>删除</Button>
        </Space>
      ),
    },
  ];

  // 处理表单字段显示逻辑
  const renderFormItems = (formInstance: any, topicType?: number) => {
    const currentType = topicType || formInstance.getFieldValue('type_id');
    
    return (
      <>
        <Form.Item 
          name="title" 
          label="题干" 
          rules={[{ required: true, message: '请输入题干' }]}
        >
          <Input.TextArea rows={4} placeholder="请输入题目内容" />
        </Form.Item>
        
        <Form.Item 
          name="type_id" 
          label="题型" 
          rules={[{ required: true, message: '请选择题型' }]}
        >
          <Select 
            options={typeOptions} 
            placeholder="请选择题型"
            onChange={() => formInstance.validateFields(['answers', 'right'])}
          />
        </Form.Item>
        
        <Form.Item 
          name="difficulty" 
          label="难度" 
          rules={[{ required: true, message: '请选择难度' }]}
        >
          <Select options={diffOptions} placeholder="请选择难度" />
        </Form.Item>
        
        <Form.Item 
          name="language" 
          label="语言" 
          rules={[{ required: true, message: '请选择语言' }]}
        >
          <Select options={languageOptions} placeholder="请选择语言" />
        </Form.Item>
        
        <Form.Item 
          name="keyword" 
          label="关键词" 
          rules={[{ required: true, message: '请输入关键词' }]}
        >
          <Input placeholder="请输入关键词，如：数组、排序等" />
        </Form.Item>
        
        {currentType !== 3 && (
          <>
            <Form.Item 
              name="answers" 
              label="选项（用英文逗号分隔）" 
              rules={[{ required: currentType !== 3, message: '请输入选项' }]}
              tooltip="示例：A.选项1,B.选项2,C.选项3,D.选项4"
            >
              <Input.TextArea rows={3} placeholder="请输入选项，用英文逗号分隔，如：A.选项1,B.选项2,C.选项3,D.选项4（若是编程题请填 无）" />
            </Form.Item>
            
            <Form.Item 
              name="right" 
              label="正确答案（用英文逗号分隔）" 
              rules={[{ required: currentType !== 3, message: '请输入正确答案' }]}
              tooltip="示例：单选题填A，多选题填A,B"
            >
              <Input placeholder="请输入正确答案，用英文逗号分隔多个答案（若是编程题请填 无）" />
            </Form.Item>
          </>
        )}
      </>
    );
  };

  return (
    <div>
      <Space style={{ marginBottom: 16 }}>
        <Input.Search 
          placeholder="搜索题干/关键词" 
          allowClear 
          onSearch={setSearch} 
          style={{ width: 250 }} 
        />
        <Select
          placeholder="题型筛选"
          allowClear
          style={{ width: 120 }}
          options={typeOptions}
          onChange={setType}
        />
        <Button type="primary" onClick={() => setAiModal(true)}>AI出题</Button>
        <Button onClick={() => setAddModal(true)}>手工出题</Button>
        <Button onClick={fetchData}>刷新</Button>
        <Button 
          danger 
          disabled={selectedRowKeys.length === 0}
          loading={batchDeleteLoading}
          onClick={handleBatchDelete}
        >
          批量删除({selectedRowKeys.length})
        </Button>
      </Space>
      
      <Table
        rowSelection={rowSelection}
        columns={columns}
        dataSource={filtered}
        rowKey={(record: Topic) => record.id ? String(record.id) : ''}
        loading={loading}
        pagination={{
          current: page,
          pageSize,
          total: filtered.length,
          onChange: setPage,
          onShowSizeChange: (_: number, size: number) => setPageSize(size),
          showSizeChanger: true,
          showTotal: (total: number) => `共 ${total} 道题`,
        }}
      />
      
      {/* 编辑题目弹窗 */}
      <Modal
        open={editModal}
        title="编辑题目"
        onCancel={() => setEditModal(false)}
        onOk={handleEditOk}
        destroyOnClose
        width={700}
      >
        <Form
          form={form}
          layout="vertical"
        >
          {renderFormItems(form, editTopic?.type_id)}
        </Form>
      </Modal>
      
      {/* 手工出题弹窗 */}
      <Modal
        open={addModal}
        title="手工出题"
        onCancel={() => setAddModal(false)}
        onOk={handleAddOk}
        destroyOnClose
        width={700}
      >
        <Form
          form={addForm}
          layout="vertical"
          initialValues={{ type_id: 1, difficulty: 1, language: 'go' }}
        >
          {renderFormItems(addForm)}
        </Form>
      </Modal>
      
      {/* AI出题弹窗 */}
      <Modal
        open={aiModal}
        title="AI出题"
        onCancel={() => setAiModal(false)}
        onOk={handleAiOk}
        confirmLoading={aiLoading}
        destroyOnClose
        width={600}
      >
        <Form
          form={aiForm}
          layout="vertical"
          initialValues={{ number: 1, language: 'go', type: 1, difficulty: 1 }}
        >
          <Form.Item 
            name="number" 
            label="题目数量" 
            rules={[{ required: true, type: 'number', min: 1, max: 10, message: '请输入1-10之间的数字' }]}
          >
            <InputNumber min={1} max={10} style={{ width: '100%' }} />
          </Form.Item>
          
          <Form.Item 
            name="language" 
            label="语言" 
            rules={[{ required: true, message: '请选择语言' }]}
          >
            <Select options={languageOptions} placeholder="请选择语言" />
          </Form.Item>
          
          <Form.Item 
            name="type" 
            label="题型" 
            rules={[{ required: true, message: '请选择题型' }]}
          >
            <Select options={typeOptions} placeholder="请选择题型" />
          </Form.Item>
          
          <Form.Item 
            name="difficulty" 
            label="难度" 
            rules={[{ required: true, message: '请选择难度' }]}
          >
            <Select options={diffOptions} placeholder="请选择难度" />
          </Form.Item>
          
          <Form.Item 
            name="keyword" 
            label="知识点关键词" 
            rules={[{ required: true, message: '请输入关键词' }]}
          >
            <Input placeholder="请输入相关知识点关键词，如：数组、排序等" />
          </Form.Item>
          
          <Divider />
          <p style={{ color: '#999' }}>注意：AI生成可能需要一定时间，请耐心等待。</p>
        </Form>
      </Modal>

      {/* AI出题预览弹窗 */}
      <Modal
        open={aiPreviewVisible}
        title="AI出题预览"
        onCancel={() => {
          setAiPreviewVisible(false);
          setAiPreviewData([]);
        }}
        onOk={handleSaveAiQuestions}
        confirmLoading={aiPreviewLoading}
        okText="保存到题库"
        cancelText="取消"
        width={800}
        destroyOnClose
      >
        <div style={{ marginBottom: 16 }}>
          <div style={{ marginBottom: 8 }}>
            <Tag color="blue">已生成 {aiPreviewData.length} 道题目</Tag>
          </div>
          <Button 
            type="primary" 
            onClick={handleSaveAiQuestions} 
            loading={aiPreviewLoading}
          >
            保存到题库
          </Button>
        </div>

        <div style={{ maxHeight: '500px', overflow: 'auto' }}>
          {aiPreviewData.map((topic, index) => (
            <div key={`preview-${index}`} style={{ 
              border: '1px solid #f0f0f0', 
              borderRadius: '8px', 
              padding: '16px', 
              marginBottom: '16px',
              background: '#fafafa'
            }}>
              <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '8px' }}>
                <Tag color={topic.type_id === 3 ? 'purple' : (topic.type_id === 2 ? 'blue' : 'green')}>
                  {typeOptions.find(t => t.value === topic.type_id)?.label}
                </Tag>
                <Tag color={topic.difficulty === 3 ? 'red' : (topic.difficulty === 2 ? 'orange' : 'success')}>
                  {diffOptions.find(d => d.value === topic.difficulty)?.label}
                </Tag>
              </div>
              
              <div style={{ marginBottom: '12px', fontWeight: 'bold' }}>
                {index + 1}. {topic.title}
              </div>
              
              {/* 选项（如果有） */}
              {topic.type_id !== 3 && topic.answers && topic.answers.length > 0 && (
                <div style={{ marginBottom: '12px' }}>
                  <div style={{ fontWeight: 'bold', marginBottom: '4px' }}>选项：</div>
                  <div style={{ paddingLeft: '16px' }}>
                    {topic.answers.map((answer, ansIndex) => (
                      <div key={`answer-${index}-${ansIndex}`}>{answer}</div>
                    ))}
                  </div>
                </div>
              )}
              
              {/* 正确答案（如果有） */}
              {topic.type_id !== 3 && topic.right && topic.right.length > 0 && (
                <div>
                  <div style={{ fontWeight: 'bold', marginBottom: '4px' }}>正确答案：</div>
                  <div style={{ paddingLeft: '16px', color: '#52c41a' }}>
                    {topic.right.join(', ')}
                  </div>
                </div>
              )}
            </div>
          ))}
        </div>

        <Divider />
        <p style={{ color: '#999' }}>提示：确认保存后，这些题目将添加到题库中</p>
      </Modal>
    </div>
  );
} 