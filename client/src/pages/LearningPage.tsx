import { useState, useEffect } from 'react';
import { Card, Spin } from 'antd';
import ReactMarkdown from 'react-markdown';
import 'github-markdown-css/github-markdown.css';

export default function LearningPage() {
  const [markdownContent, setMarkdownContent] = useState('');
  const [loading] = useState(false);

  useEffect(() => {
    
    const learningContent = ` `;
    
    setMarkdownContent(learningContent);
  }, []);

  return (
    <div style={{ maxWidth: 700, margin: '0 auto' }}>
      <Card title="学习心得" type="inner">
        {loading ? (
          <Spin tip="加载中..."/>
        ) : (
          <div className="markdown-body" style={{ padding: '20px' }}>
            <ReactMarkdown>{markdownContent}</ReactMarkdown>
          </div>
        )}
      </Card>
    </div>
  );
} 