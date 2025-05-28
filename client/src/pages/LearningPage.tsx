import { useState, useEffect } from 'react';
import { Card, Spin } from 'antd';
import ReactMarkdown from 'react-markdown';
import 'github-markdown-css/github-markdown.css';

export default function LearningPage() {
  const [markdownContent, setMarkdownContent] = useState('');
  const [loading] = useState(false);

  useEffect(() => {
    
    const learningContent = `

在这次的系统设计中，我对go的使用有了更进一步的了解：1、对gin框架的路由编写有了进一步的掌握，可以更好地进行编写2、对json等类型的数据的处理，能熟练运用，进行格式的转换。我再次阅读了有关ai的库和模型的官方文档，对go进行ai的api调用也有了更好的掌握。在前端的编写中，我利用控制台输出的报错，对代码逐步进行修改。其中印象最深的就是版本兼容问题：react19与ant design v5之间问题，在尝试了降react版本等方法但无果后，我在ant design的官方文档中找到了解决方案。[ant官方文件](https://ant.design/docs/react/v5-for-19)

在这次的系统设计中，我认为我还有一些不足：路由接口可以整合到方法中进行编写，减少在main函数中的代码，后续也希望进一步掌握分模块进行代码编写，不把代码都写在main.go文件中。

总之，我在代码的规范编写上还要更加努力的学习，也要多运用debug和控制台输出来进行问题的排查。同时，也要对git的使用、数据库等进阶知识进行深入学习`;
    
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