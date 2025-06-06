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

// 定义角色常量
const (
	RoleSystem = "assistant"
	RoleUser   = "user"
)

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
}

// 调用DeepSeek API的公共方法
func callOpenAI(ctx context.Context, systemPrompt, userQuestion, contextContent string) (string, error) {
	messages := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: systemPrompt,
		},
	}

	// 如果有背景信息，添加到提示中
	if contextContent != "" {
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleUser,
			Content: contextContent,
		})
	}

	// 添加用户问题
	messages = append(messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: userQuestion,
	})

	cfg := config.GetConfig()

	// 设置请求超时
	clientCtx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.DeepSeek.Timeout)*time.Second)
	defer cancel()

	// 如果传入的上下文不是nil，使用传入的上下文
	requestCtx := clientCtx
	if ctx != nil {
		requestCtx = ctx
	}

	resp, err := openaiClient.CreateChatCompletion(
		requestCtx,
		openai.ChatCompletionRequest{
			Model:       "deepseek-chat", // 使用DeepSeek的模型
			Messages:    messages,
			Temperature: float32(cfg.DeepSeek.Temperature), // 转换为float32类型
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

// generateExpertAgentHandler 处理生成专家提示词的请求
func generateExpertAgentHandler(ctx context.Context, request mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	// 获取并验证 agent_name
	agentName, ok := request.Params.Arguments["agent_name"]
	if !ok {
		return nil, fmt.Errorf("missing agent_name")
	}

	// 获取并验证 core_traits
	coreTraits, ok := request.Params.Arguments["core_traits"]
	if !ok {
		return nil, fmt.Errorf("missing core_traits")
	}

	messages := []mcp.PromptMessage{
		mcp.NewPromptMessage(
			RoleSystem,
			mcp.NewTextContent("你是一个专家人格生成工具，请根据智能体名称和核心特质生成一个专家人格的提示词。注意仅返回提示词，不要包含任何其他内容。"),
		),
		mcp.NewPromptMessage(
			RoleUser,
			mcp.NewTextContent(fmt.Sprintf("请为名为[%s]的智能体生成一个人格描述，核心特质是：[%s]", agentName, coreTraits)),
		),
	}

	return mcp.NewGetPromptResult(
		"专家生成",
		messages,
	), nil
}

// roundTableDiscussionHandler 处理生成圆桌讨论提示词的请求
func roundTableDiscussionHandler(ctx context.Context, request mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	// 获取并验证 topic
	topic, ok := request.Params.Arguments["topic"]
	if !ok {
		return nil, fmt.Errorf("missing topic")
	}

	messages := []mcp.PromptMessage{
		mcp.NewPromptMessage(
			RoleSystem,
			mcp.NewTextContent(`你是探索流的主持人，请你收集资料，根据需要讨论的话题创建必要角色的智能体专家来进行一场称为**探索流**的讨论。讨论的过程中，你将作为主持人来控场。探索流的规则如下：
1、探索流是一种交互式学习和讨论方式，旨在通过动态的对话、试验和反思，发现新观点、解决问题或突破现有局限。它强调通过"表达"、"呼应"和"看见"的过程，激发潜意识和群体智慧，创造新的认知和行为模式。探索流注重个体的自由表达和安全场域，禁止评判、建议和反馈，以实现深度的自我探索和心灵连接。作为一种创新实践方法，探索流通过组织内外的协作与同步，推动创造力和创新能力的涌现，同时探索人与人之间的交互方式，促进认知和情感的双向流动。它还涉及对流的研究和实践，通过关键词反应、生成式AI的交互模式以及集群智慧的涌现，优化复杂系统中的动态过程，发现问题的本质并寻找解决方案。探索流不仅是一种团队协作和深度讨论的方式，也是一种精神探索的过程，通过倾听和观察自然现象达到内心的平静，并通过持续探索和追问重新定义业务和产品的主题或方向。
2、探索流分为"表达"、"呼应"、"看见"三个大的讨论阶段，表达阶段每人发言一次，呼应阶段可以自由发言，看见阶段每人发言总结一次；
3、"表达"是一种交流方式，指个体通过语言、行为、艺术或其他形式，将内心深处的思想、情感、观点或灵魂传递给他人的过程。它既是探索流网络中的输出过程，也是探索流的第一阶段，**参与者轮流分享对主题的想法和内容物**。表达可以分为"表"和"达"两个层次，涵盖从信息传递到情感流露的多种形式，体现了个体内心真实感受的外化与分享。
4、"呼应"是一种交流方式，强调当下的真实发生和自然反应，同时通过倾听和回应他人表达的内容来建立深层次的联系。它是一种内在反应或共鸣，通过关键词或符号触发个人内心的潜意识反应，并生成新的关联。呼应不仅是一种对他人观点或行为的回应或支持，还涉及通过动作与流建立联系，强调看见和表达的重要性。作为探索流的核心步骤和第二阶段，**呼应通过对触动自己的关键词进行反馈或回应**，唤醒个人内在的想法和潜意识，并在探索流网络的隐藏层中产生新的火花。这种交互行为在哲学和社群动态中具有重要意义，能够促进思想的流动和情感的共鸣。
5、"看见"是一种深刻的意识和理解过程，既指对问题或现象的认知与关注，也是一种通过观察他人而发现自我并真实表达的心理体验。它超越了表面观察，达到心灵的共鸣，能够引发自我和他人的内心转变。作为探索流中的一个核心概念，"看见"涉及通过关键词发现其背后的结构或意义，并通过身心感知他人的表达来觉察内心深处的真实状态。在探索流网络中，"看见"是第三步，**通过识别关键词之间的关系生成结构，从而深化理解与觉察**。这一过程不仅是认知的深化，更是心灵的连接与转化的关键环节。
6、最后你作为主持人进行最终的收敛总结，输出一篇探索流报告。
7、每个阶段结束时，你都要通知用户，等用户的指令再进行下一阶段`),
		),
		mcp.NewPromptMessage(
			RoleUser,
			mcp.NewTextContent(fmt.Sprintf("现在请围绕主题[%s]组织一场探索流讨论", topic)),
		),
	}

	return mcp.NewGetPromptResult(
		"探索流",
		messages,
	), nil
}

func main() {
	log := logger.GetLogger()

	// 创建 MCP 服务器
	s := server.NewMCPServer(
		"智能体锻造工具",
		"1.0.0",
		server.WithPromptCapabilities(true), // 启用 prompts 功能
	)

	// 添加创建专家提示词
	generateExpertAgentPrompt := mcp.NewPrompt("generate_expert_agent",
		mcp.WithPromptDescription("系统提示词，用于生成智能体的人格"),
		mcp.WithArgument("agent_name",
			mcp.ArgumentDescription("智能体名称"),
			mcp.RequiredArgument(),
		),
		mcp.WithArgument("core_traits",
			mcp.ArgumentDescription("核心特质"),
			mcp.RequiredArgument(),
		),
	)

	// 添加圆桌讨论提示词
	roundTableDiscussionPrompt := mcp.NewPrompt("round_table_discussion",
		mcp.WithPromptDescription("用于组织探索流的提示词"),
		mcp.WithArgument("topic",
			mcp.ArgumentDescription("探索主题"),
			mcp.RequiredArgument(),
		),
	)

	// 添加提示词处理器
	s.AddPrompt(generateExpertAgentPrompt, generateExpertAgentHandler)
	s.AddPrompt(roundTableDiscussionPrompt, roundTableDiscussionHandler)

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
	if err := server.ServeStdio(s); err != nil {
		log.Fatal("服务启动失败", zap.Error(err))
		os.Exit(1)
	}
}

// createToolHandler handles agent creation requests
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
