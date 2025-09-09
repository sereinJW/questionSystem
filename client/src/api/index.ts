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

// 试卷大题配置
export interface PaperSection {
  name: string;        // 大题名称，如"一、单选题"
  type: number;        // 题目类型 1:单选 2:多选 3:编程
  count: number;       // 题目数量
  score_each: number;  // 每题分数
}

// 创建试卷的请求体
export interface CreatePaperRequest {
  title: string;          // 试卷标题
  sections: PaperSection[]; // 大题配置
  question_ids: number[];   // 选中的题目ID列表
}

// 试卷
export interface Paper {
  id: number;             // 试卷ID
  title: string;          // 试卷标题
  config: PaperSection[]; // 大题配置
  created_at: string;     // 创建时间
}

// 试卷-题目关联
export interface PaperQuestion {
  id: number;             // 主键
  paper_id: number;       // 试卷ID
  question_id: number;    // 题目ID
  score: number;          // 这道题的分数
  order_in_paper: number; // 题目在试卷中的顺序
}

// 试卷题目（包含完整题目信息）
export interface PaperQuestionWithTopic extends PaperQuestion {
  topic: Topic;           // 完整的题目信息
}

// 试卷详情（包含完整题目信息）
export interface PaperDetail {
  paper: Paper;                          // 试卷基本信息
  questions: PaperQuestionWithTopic[];   // 题目列表（带完整题目信息）
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

// 创建试卷
export const createPaper = async (request: CreatePaperRequest): Promise<ApiResponse<Paper | null>> => {
  try {
    console.log('发起创建试卷请求, 数据:', request);
    const response = await axios.post('/api/papers', request);
    console.log('创建试卷响应:', response.data);
    return response.data;
  } catch (error) {
    console.error('创建试卷失败:', error);
    return { code: -999, msg: '网络错误', data: null };
  }
};

// 获取试卷详情
export const getPaperDetail = async (paperId: number): Promise<ApiResponse<PaperDetail | null>> => {
  try {
    console.log('发起获取试卷详情请求, ID:', paperId);
    const response = await axios.get(`/api/papers/${paperId}`);
    console.log('获取试卷详情响应:', response.data);
    return response.data;
  } catch (error) {
    console.error('获取试卷详情失败:', error);
    return { code: -999, msg: '网络错误', data: null };
  }
};

// 获取试卷列表
export const getPapers = async (): Promise<ApiResponse<Paper[]>> => {
  try {
    console.log('发起获取试卷列表请求');
    const response = await axios.get('/api/papers');
    console.log('获取试卷列表响应:', response.data);
    return response.data;
  } catch (error) {
    console.error('获取试卷列表失败:', error);
    return { code: -999, msg: '网络错误', data: [] };
  }
};

// 删除试卷
export const deletePaper = async (paperId: number): Promise<ApiResponse> => {
  try {
    console.log('发起删除试卷请求, ID:', paperId);
    const response = await axios.delete(`/api/papers/${paperId}`);
    console.log('删除试卷响应:', response.data);
    return response.data;
  } catch (error) {
    console.error('删除试卷失败:', error);
    return { code: -999, msg: '网络错误', data: null };
  }
};

// 编辑试卷标题
export const editPaperTitle = async (paperId: number, title: string): Promise<ApiResponse> => {
  try {
    console.log('发起编辑试卷标题请求, ID:', paperId, 'Title:', title);
    const response = await axios.put(`/api/papers/${paperId}`, { title });
    console.log('编辑试卷标题响应:', response.data);
    return response.data;
  } catch (error) {
    console.error('编辑试卷标题失败:', error);
    return { code: -999, msg: '网络错误', data: null };
  }
};

// 更新试卷题目
export const updatePaperQuestions = async (paperId: number, title: string, questionIds: number[]): Promise<ApiResponse> => {
  try {
    console.log('发起更新试卷题目请求, ID:', paperId, 'Title:', title, 'QuestionIds:', questionIds);
    const response = await axios.put(`/api/papers/${paperId}/questions`, { 
      title, 
      question_ids: questionIds 
    });
    console.log('更新试卷题目响应:', response.data);
    return response.data;
  } catch (error) {
    console.error('更新试卷题目失败:', error);
    return { code: -999, msg: '网络错误', data: null };
  }
}; 