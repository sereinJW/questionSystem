import axios from 'axios';

// 接口返回数据类型
export interface ApiResponse<T = any> {
  code: number;
  msg: string;
  data: T;
}

// 题目类型定义
export interface Topic {
  id?: number;
  title: string;
  answers: string[];
  right: string[];
  type_id: number;
  difficulty: number;
  is_ai: number;
  language: string;
  keyword: string;
  active: number;
}

// AI出题请求参数
export interface AskParams {
  number: number;
  language: string;
  type: number;
  difficulty: number;
  keyword: string;
}

// 获取题库列表
export const getQuestions = async (): Promise<ApiResponse<Topic[]>> => {
  try {
    console.log('发起获取题库请求');
    const response = await axios.get('/api/questions');
    console.log('获取题库响应', response.data);
    return response.data;
  } catch (error) {
    console.error('获取题库失败:', error);
    return { code: -999, msg: '网络错误', data: [] };
  }
};

// AI出题
export const createQuestions = async (params: AskParams): Promise<ApiResponse<Topic[]>> => {
  try {
    const response = await axios.post('/api/questions/create', params);
    return response.data;
  } catch (error) {
    console.error('AI出题失败:', error);
    return { code: -999, msg: '网络错误', data: [] };
  }
};

// 添加题目
export const addQuestion = async (topic: Omit<Topic, 'id' | 'active'>): Promise<ApiResponse> => {
  try {
    const response = await axios.post('/api/questions/add', topic);
    return response.data;
  } catch (error) {
    console.error('添加题目失败:', error);
    return { code: -999, msg: '网络错误', data: null };
  }
};

// 编辑题目
export const editQuestion = async (topic: Topic): Promise<ApiResponse> => {
  try {
    console.log('发起编辑题目请求, 数据:', topic);
    const response = await axios.post('/api/questions/edit', topic);
    console.log('编辑题目响应:', response.data);
    return response.data;
  } catch (error) {
    console.error('编辑题目失败:', error);
    return { code: -999, msg: '网络错误', data: null };
  }
};

// 删除题目
export const deleteQuestion = async (topic: Topic): Promise<ApiResponse> => {
  try {
    console.log('发起删除题目请求, 数据:', topic);
    const response = await axios.post('/api/questions/delete', topic);
    console.log('删除题目响应:', response.data);
    return response.data;
  } catch (error) {
    console.error('删除题目失败:', error);
    return { code: -999, msg: '网络错误', data: null };
  }
}; 