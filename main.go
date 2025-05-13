package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"agent-forge/internal/config"
	"agent-forge/internal/logger"

	"github.com/google/uuid"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/sashabaranov/go-openai"
	"go.uber.org/zap"
)

// OpenAI客户端
var openaiClient *openai.Client

// Agent 结构体定义
type Agent struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	CoreTraits  string `json:"core_traits"`
	Personality string `json:"personality"`
	CreatedAt   string `json:"created_at"`
}

// 存储所有生成的智能体
var agents = make(map[string]*Agent)

// ParamsStruct 定义请求参数结构
type ParamsStruct struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments,omitempty"`
	Meta      *struct {
		ProgressToken mcp.ProgressToken `json:"progressToken,omitempty"`
	} `json:"_meta,omitempty"`
}

func init() {
	// 初始化配置
	if _, err := config.LoadConfig(); err != nil {
		fmt.Fprintf(os.Stderr, "配置加载失败: %v\n", err)
		os.Exit(1)
	}

	// 初始化日志系统
	cfg := config.GetConfig()
	if err := logger.InitLogger(cfg); err != nil {
		panic(fmt.Sprintf("初始化日志系统失败: %v", err))
	}

	// 初始化DeepSeek客户端
	if cfg.DeepSeek.APIKey == "" {
		logger.Error("DEEPSEEK_API_KEY 环境变量未设置")
		fmt.Fprintln(os.Stderr, "\n请设置 DEEPSEEK_API_KEY 环境变量后再运行程序。例如：")
		fmt.Fprintln(os.Stderr, "export DEEPSEEK_API_KEY=your_api_key_here")
		os.Exit(1)
	}

	openaiConfig := openai.DefaultConfig(cfg.DeepSeek.APIKey)
	openaiConfig.BaseURL = cfg.DeepSeek.BaseURL
	openaiClient = openai.NewClientWithConfig(openaiConfig)

	logger.Info("系统初始化完成",
		zap.String("host", cfg.Server.Host),
		zap.Int("port", cfg.Server.Port))
}

// 调用DeepSeek API的公共方法
func callOpenAI(ctx context.Context, systemPrompt, userQuestion, context string) (string, error) {
	messages := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: systemPrompt,
		},
	}

	// 如果有背景信息，添加到提示中
	if context != "" {
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleUser,
			Content: context,
		})
	}

	// 添加用户问题
	messages = append(messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: userQuestion,
	})

	resp, err := openaiClient.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model:       "deepseek-chat", // 使用DeepSeek的模型
			Messages:    messages,
			Temperature: 0.7,
		},
	)

	if err != nil {
		return "", fmt.Errorf("DeepSeek API调用失败: %v", err)
	}

	if len(resp.Choices) == 0 {
		return "", errors.New("DeepSeek返回结果为空")
	}

	// 去除返回内容中可能的前后空白字符
	content := strings.TrimSpace(resp.Choices[0].Message.Content)

	// 如果返回的内容是JSON格式，尝试提取纯文本内容
	var jsonResponse map[string]interface{}
	if err := json.Unmarshal([]byte(content), &jsonResponse); err == nil {
		// 如果是JSON格式，尝试获取content字段
		if textContent, ok := jsonResponse["content"].(string); ok {
			content = textContent
		}
	}

	return content, nil
}

func main() {
	cfg := config.GetConfig()
	log := logger.GetLogger()

	// 创建 MCP 服务器
	s := server.NewMCPServer(
		"智能体锻造工具",
		"1.0.0",
	)

	// 创建智能体工具
	createTool := mcp.NewTool(
		"expert_personality_generation",
		mcp.WithDescription("创建新的智能体"),
		mcp.WithString("agent_name",
			mcp.Required(),
			mcp.Description("智能体名称"),
		),
		mcp.WithString("core_traits",
			mcp.Required(),
			mcp.Description("核心特质"),
		),
	)

	// 模拟智能体回答工具
	answerTool := mcp.NewTool(
		"agent_answer",
		mcp.WithDescription(`模拟智能体作答。
参数说明:
- agent_id: 智能体ID,创建智能体时返回的唯一标识
- context: 当前对话的上下文内容，具体包括:
  * 用户的原始问题及补充说明
  * 其他专家智能体的观点
  * 我已经陈述的观点
  * 外部搜索或知识输入
- planned_rounds: 预估需要进行的回答次数
- current_round: 已回答次数
- need_more_rounds: 是否需要新增回答次数,当主持人认为需要新增回答次数时，设置为true，否则设置为false`),
		mcp.WithString("agent_id",
			mcp.Required(),
			mcp.Description("智能体ID"),
		),
		mcp.WithString("context",
			mcp.Description("对话上下文"),
		),
		mcp.WithNumber("planned_rounds",
			mcp.Description("计划回答次数"),
		),
		mcp.WithNumber("current_round",
			mcp.Description("已回答次数"),
		),
		mcp.WithBoolean("need_more_rounds",
			mcp.Description("是否需要新增回答次数"),
		),
	)

	// 获取智能体信息工具
	getTool := mcp.NewTool(
		"get_agent",
		mcp.WithDescription("获取指定智能体的信息"),
		mcp.WithString("agent_id",
			mcp.Required(),
			mcp.Description("智能体ID"),
		),
	)

	// 列出所有智能体工具
	listTool := mcp.NewTool(
		"list_agents",
		mcp.WithDescription("列出所有智能体"),
	)

	// 删除智能体工具
	deleteTool := mcp.NewTool(
		"delete_agent",
		mcp.WithDescription("删除指定的智能体"),
		mcp.WithString("agent_id",
			mcp.Required(),
			mcp.Description("要删除的智能体ID"),
		),
	)

	// 更新智能体工具
	updateTool := mcp.NewTool(
		"update_agent",
		mcp.WithDescription("更新智能体信息"),
		mcp.WithString("agent_id",
			mcp.Required(),
			mcp.Description("智能体ID"),
		),
		mcp.WithString("name",
			mcp.Description("新的智能体名称"),
		),
		mcp.WithString("core_traits",
			mcp.Description("新的核心特质"),
		),
	)

	// 添加工具处理器
	s.AddTool(createTool, createToolHandler)
	s.AddTool(answerTool, answerToolHandler)
	s.AddTool(getTool, getAgentHandler)
	s.AddTool(listTool, listAgentsHandler)
	s.AddTool(deleteTool, deleteAgentHandler)
	s.AddTool(updateTool, updateAgentHandler)

	// 启动服务器
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	log.Info("服务启动",
		zap.String("address", addr),
		zap.Int("rate_limit", cfg.Server.RateLimit))

	if err := server.ServeStdio(s); err != nil {
		log.Fatal("服务启动失败", zap.Error(err))
		os.Exit(1)
	}
}

// 修改getHandlerForModel函数
func createToolHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log := logger.GetLogger()

	// 提取参数
	agentName, ok := request.Params.Arguments["agent_name"].(string)
	if !ok || agentName == "" {
		log.Error("无效的智能体名称")
		return nil, errors.New("无效的智能体名称")
	}

	coreTraits, ok := request.Params.Arguments["core_traits"].(string)
	if !ok || coreTraits == "" {
		log.Error("无效的核心特征")
		return nil, errors.New("无效的核心特征")
	}

	log.Info("创建智能体",
		zap.String("name", agentName),
		zap.String("traits", coreTraits))

	systemPrompt := "你是一个专家人格生成工具，请根据智能体名称和核心特质生成一个专家人格的提示词。注意仅返回提示词，不要包含任何其他内容。"
	question := fmt.Sprintf("请为名为[%s]的智能体生成一个人格描述，核心特质是：[%s]", agentName, coreTraits)

	// 调用OpenAI
	response, err := callOpenAI(ctx, systemPrompt, question, "")
	if err != nil {
		return nil, err
	}

	// 创建新的智能体实例
	agentID := uuid.New().String()
	newAgent := &Agent{
		ID:          agentID,
		Name:        agentName,
		CoreTraits:  coreTraits,
		Personality: response,
		CreatedAt:   time.Now().Format(time.RFC3339),
	}

	// 存储智能体
	agents[agentID] = newAgent

	// 返回处理结果，包含智能体ID
	result := map[string]interface{}{
		"status":   "success",
		"message":  "智能体创建成功",
		"agent_id": agentID,
	}

	jsonResponse, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("marshal response failed: %v", err)
	}

	// 确保返回的是有效的JSON字符串
	return mcp.NewToolResultText(string(jsonResponse)), nil
}

// 获取智能体处理函数
func getAgentHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	agentID, ok := request.Params.Arguments["agent_id"].(string)
	if !ok {
		return nil, errors.New("agent_id must be a string")
	}

	agent, exists := agents[agentID]
	if !exists {
		return nil, fmt.Errorf("agent with ID %s not found", agentID)
	}

	jsonResponse, err := json.Marshal(agent)
	if err != nil {
		return nil, fmt.Errorf("marshal response failed: %v", err)
	}

	return mcp.NewToolResultText(string(jsonResponse)), nil
}

// 列出所有智能体处理函数
func listAgentsHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	agentList := make([]*Agent, 0, len(agents))
	for _, agent := range agents {
		agentList = append(agentList, agent)
	}

	jsonResponse, err := json.Marshal(agentList)
	if err != nil {
		return nil, fmt.Errorf("marshal response failed: %v", err)
	}

	return mcp.NewToolResultText(string(jsonResponse)), nil
}

// 删除智能体处理函数
func deleteAgentHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	agentID, ok := request.Params.Arguments["agent_id"].(string)
	if !ok {
		return nil, errors.New("agent_id must be a string")
	}

	if _, exists := agents[agentID]; !exists {
		return nil, fmt.Errorf("agent with ID %s not found", agentID)
	}

	delete(agents, agentID)

	result := map[string]string{
		"status":  "success",
		"message": fmt.Sprintf("智能体 %s 已成功删除", agentID),
	}

	jsonResponse, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("marshal response failed: %v", err)
	}

	return mcp.NewToolResultText(string(jsonResponse)), nil
}

// 更新智能体处理函数
func updateAgentHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	agentID, ok := request.Params.Arguments["agent_id"].(string)
	if !ok {
		return nil, errors.New("agent_id must be a string")
	}

	agent, exists := agents[agentID]
	if !exists {
		return nil, fmt.Errorf("agent with ID %s not found", agentID)
	}

	// 更新名称（如果提供）
	if newName, ok := request.Params.Arguments["name"].(string); ok && newName != "" {
		agent.Name = newName
	}

	// 更新核心特质（如果提供）
	if newTraits, ok := request.Params.Arguments["core_traits"].(string); ok && newTraits != "" {
		agent.CoreTraits = newTraits
		// 重新生成人格描述
		systemPrompt := "你是一个专家人格生成工具，请根据智能体名称和核心特质生成一个专家人格。"
		question := fmt.Sprintf("请为名为[%s]的智能体生成一个人格描述，核心特质是：[%s]", agent.Name, newTraits)

		newPersonality, err := callOpenAI(ctx, systemPrompt, question, "")
		if err != nil {
			return nil, fmt.Errorf("generate new personality failed: %v", err)
		}
		agent.Personality = newPersonality
	}

	result := map[string]interface{}{
		"status":  "success",
		"message": "智能体更新成功",
		"agent":   agent,
	}

	jsonResponse, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("marshal response failed: %v", err)
	}

	return mcp.NewToolResultText(string(jsonResponse)), nil
}

// 模拟智能体回答处理函数
func answerToolHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// 获取参数
	agentID, ok := request.Params.Arguments["agent_id"].(string)
	if !ok {
		return nil, errors.New("agent_id must be a string")
	}

	context, _ := request.Params.Arguments["context"].(string)
	plannedRounds, _ := request.Params.Arguments["planned_rounds"].(float64)
	currentRound, _ := request.Params.Arguments["current_round"].(float64)
	needMoreRounds, _ := request.Params.Arguments["need_more_rounds"].(bool)

	if currentRound > plannedRounds {
		plannedRounds = currentRound
	}

	// 获取智能体
	agent, exists := agents[agentID]
	if !exists {
		return nil, fmt.Errorf("agent with ID %s not found", agentID)
	}

	// 构建系统提示词
	systemPrompt := fmt.Sprintf("你现在扮演一个%s。%s", agent.Name, agent.Personality)

	// 调用OpenAI生成回答
	response, err := callOpenAI(ctx, systemPrompt, context, "")
	if err != nil {
		return nil, err
	}

	// 创建包含所有信息的响应
	result := map[string]interface{}{
		"content":          response,
		"planned_rounds":   plannedRounds,
		"current_round":    currentRound,
		"need_more_rounds": needMoreRounds,
	}

	jsonResponse, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("marshal response failed: %v", err)
	}

	return mcp.NewToolResultText(string(jsonResponse)), nil
}
