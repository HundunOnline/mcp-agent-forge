# Agent Forge API 文档

## API 端点

### 1. 创建智能体 (expert_personality_generation)

创建一个新的智能体实例。

**请求参数：**
```json
{
    "name": "expert_personality_generation",
    "arguments": {
        "agent_name": "string",
        "core_traits": "string"
    }
}
```

| 参数 | 类型 | 描述 | 是否必需 |
|------|------|------|----------|
| agent_name | string | 智能体名称 | 是 |
| core_traits | string | 核心特征，用逗号分隔 | 是 |

**响应：**
```json
{
    "agent_id": "string",
    "name": "string",
    "core_traits": "string",
    "personality": "string",
    "created_at": "string"
}
```

### 2. 智能体回答 (agent_answer)

让智能体对给定问题进行回答。

**请求参数：**
```json
{
    "name": "agent_answer",
    "arguments": {
        "agent_id": "string",
        "context": "string",
        "planned_rounds": number,
        "current_round": number,
        "need_more_rounds": boolean
    }
}
```

| 参数 | 类型 | 描述 | 是否必需 |
|------|------|------|----------|
| agent_id | string | 智能体ID | 是 |
| context | string | 对话上下文 | 是 |
| planned_rounds | number | 计划回答次数 | 是 |
| current_round | number | 当前回答次数 | 是 |
| need_more_rounds | boolean | 是否需要更多回合 | 是 |

**响应：**
```json
{
    "response": "string",
    "needs_followup": boolean
}
```

### 3. 获取智能体信息 (get_agent)

获取指定智能体的详细信息。

**请求参数：**
```json
{
    "name": "get_agent",
    "arguments": {
        "agent_id": "string"
    }
}
```

| 参数 | 类型 | 描述 | 是否必需 |
|------|------|------|----------|
| agent_id | string | 智能体ID | 是 |

**响应：**
```json
{
    "id": "string",
    "name": "string",
    "core_traits": "string",
    "personality": "string",
    "created_at": "string"
}
```

### 4. 列出所有智能体 (list_agents)

获取所有已创建的智能体列表。

**请求参数：**
```json
{
    "name": "list_agents"
}
```

**响应：**
```json
{
    "agents": [
        {
            "id": "string",
            "name": "string",
            "core_traits": "string",
            "personality": "string",
            "created_at": "string"
        }
    ]
}
```

### 5. 删除智能体 (delete_agent)

删除指定的智能体。

**请求参数：**
```json
{
    "name": "delete_agent",
    "arguments": {
        "agent_id": "string"
    }
}
```

| 参数 | 类型 | 描述 | 是否必需 |
|------|------|------|----------|
| agent_id | string | 智能体ID | 是 |

**响应：**
```json
{
    "success": true,
    "message": "智能体已成功删除"
}
```

## 错误处理

所有 API 端点在发生错误时都会返回一个包含错误信息的 JSON 响应：

```json
{
    "error": {
        "code": "string",
        "message": "string"
    }
}
```

## 注意事项

1. 所有请求都需要设置 `Content-Type: application/json`
2. 需要在环境变量中设置 `DEEPSEEK_API_KEY`
3. 智能体的回答可能需要多轮对话才能完成
4. 建议在生产环境中实现适当的速率限制和认证机制 