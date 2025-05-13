# Agent Forge - 智能体锻造工具 (AI Agent Forge Tool)

[English](#english) | [中文](#chinese)

<a name="chinese"></a>
## 中文版

Agent Forge 是一个智能体创建和管理平台，能够创建和管理具有特定性格特征的智能体，并模拟它们对问题的回答。通过agent forge mcp，你可以快速构建起一个类似于[CO-STORM](https://github.com/stanford-oval/storm)的多智能体协作研究项目。

### 功能特点

- 智能体锻造：创建具有特定性格特征的智能体
- 思维模拟：模拟智能体回答问题
- 完整管理：支持智能体的查询、列表、删除等操作
- 多轮对话：支持深度的多轮对话交互
- 自然语言处理：基于 DeepSeek API 的高级语言理解能力

### 系统要求

- Go 1.24.1 或更高版本
- DeepSeek API 密钥

### 安装

1. 克隆仓库：
```bash
git clone https://github.com/HundunOnline/mcp-agent-forge.git
cd agent-forge
```

2. 安装依赖：
```bash
go mod download
```

3. 设置环境变量：
```bash
export DEEPSEEK_API_KEY=your_api_key_here
```

### 使用方法

1. 启动服务：
```bash
go run main.go
```

2. API 端点说明：

- `expert_personality_generation`: 创建新的智能体
- `agent_answer`: 模拟智能体回答问题
- `get_agent`: 获取智能体信息
- `list_agents`: 列出所有智能体
- `delete_agent`: 删除智能体

### 示例

#### 基本用法

```go
// 创建智能体
{
    "name": "expert_personality_generation",
    "arguments": {
        "agent_name": "马斯克思维模型",
        "core_traits": "系统思维,第一性原理,工程思维,风险管理,创新思维"
    }
}

// 智能体回答
{
    "name": "agent_answer",
    "arguments": {
        "agent_id": "your_agent_id",
        "context": "如何看待特斯拉的发展策略？",
        "planned_rounds": 3,
        "current_round": 1,
        "need_more_rounds": false
    }
}
```

### 实际应用案例

我们在 Claude AI 中创建了一个示例应用，展示了如何使用 Agent Forge 创建和管理专家智能体：

[Claude AI Demo](https://claude.ai/public/artifacts/81812f31-e73c-4520-814f-5ad211204d6b)

这个示例展示了：
- 如何创建具有特定专业背景的智能体
- 如何进行多轮对话交互
- 如何利用智能体的专业知识解决问题
- 如何管理和调整智能体的行为

### 贡献指南

欢迎提交 Pull Request 或创建 Issue 来帮助改进这个项目。我们特别欢迎以下方面的贡献：

- 新的智能体模型和特征
- 性能优化
- 文档改进
- Bug 修复
- 新功能建议

### 许可证

本项目采用 MIT 许可证。详见 [LICENSE](LICENSE) 文件。

---

<a name="english"></a>
## English Version

Agent Forge is a platform for creating and managing AI agents with specific personality traits and simulating their responses to questions. Through agent forge mcp, you can quickly build a multi-agent collaboration research project similar to [CO-STORM](https://github.com/stanford-oval/storm).

### Features

- Agent Forging: Create agents with specific personality traits
- Thought Simulation: Simulate agent responses to questions
- Complete Management: Support for agent querying, listing, deletion, and other operations
- Multi-round Dialogue: Support for deep multi-round conversation interactions
- Natural Language Processing: Advanced language understanding capabilities based on DeepSeek API

### System Requirements

- Go 1.24.1 or higher
- DeepSeek API key

### Installation

1. Clone the repository:
```bash
git clone https://github.com/HundunOnline/mcp-agent-forge.git
cd agent-forge
```

2. Install dependencies:
```bash
go mod download
```

3. Set environment variables:
```bash
export DEEPSEEK_API_KEY=your_api_key_here
```

### Usage

1. Start the service:
```bash
go run main.go
```

2. API Endpoints:

- `expert_personality_generation`: Create a new agent
- `agent_answer`: Simulate agent responses
- `get_agent`: Get agent information
- `list_agents`: List all agents
- `delete_agent`: Delete an agent

### Examples

#### Basic Usage

```go
// Create an agent
{
    "name": "expert_personality_generation",
    "arguments": {
        "agent_name": "Elon Musk Thinking Model",
        "core_traits": "Systems Thinking,First Principles,Engineering Mindset,Risk Management,Innovation"
    }
}

// Agent response
{
    "name": "agent_answer",
    "arguments": {
        "agent_id": "your_agent_id",
        "context": "What's your view on Tesla's development strategy?",
        "planned_rounds": 3,
        "current_round": 1,
        "need_more_rounds": false
    }
}
```

### Real Application Case

We created a sample application in Claude AI that demonstrates how to use Agent Forge to create and manage expert agents:

[Claude AI Demo](https://claude.ai/public/artifacts/81812f31-e73c-4520-814f-5ad211204d6b)

This example shows:
- How to create agents with specific professional backgrounds
- How to conduct multi-round dialogue interactions
- How to utilize agents' expertise to solve problems
- How to manage and adjust agent behavior

### Contributing

We welcome Pull Requests or Issues to help improve this project. We especially welcome contributions in the following areas:

- New agent models and traits
- Performance optimizations
- Documentation improvements
- Bug fixes
- New feature suggestions

### License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details. 