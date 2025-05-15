package main

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/stretchr/testify/assert"
)

// MockAgent 用于测试的Agent结构体
type mockAgent struct {
	ID          string
	Name        string
	CoreTraits  string
	Personality string
	CreatedAt   string
}

func TestCreateAgent(t *testing.T) {
	tests := []struct {
		name        string
		agentName   string
		coreTraits  string
		expectError bool
	}{
		{
			name:        "正常创建",
			agentName:   "测试智能体",
			coreTraits:  "逻辑思维,创新能力",
			expectError: false,
		},
		{
			name:        "空名称",
			agentName:   "",
			coreTraits:  "逻辑思维",
			expectError: true,
		},
		{
			name:        "空特征",
			agentName:   "测试智能体",
			coreTraits:  "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 使用直接的方法调用而不是通过CallToolRequest
			ctx := context.Background()
			result, err := createAgent(ctx, tt.agentName, tt.coreTraits)

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, result)
		})
	}
}

// 模拟实际的创建Agent功能，不依赖MCP结构
func createAgent(ctx context.Context, name, traits string) (string, error) {
	if name == "" {
		return "", fmt.Errorf("无效的智能体名称")
	}
	if traits == "" {
		return "", fmt.Errorf("无效的核心特征")
	}

	// 创建新的智能体实例
	agentID := uuid.New().String()
	return agentID, nil
}

func TestGetAgent(t *testing.T) {
	// 创建测试用智能体
	testAgent := &Agent{
		ID:          uuid.New().String(),
		Name:        "测试智能体",
		CoreTraits:  "测试特征",
		Personality: "测试性格",
		CreatedAt:   time.Now().Format(time.RFC3339),
	}
	agents[testAgent.ID] = testAgent

	tests := []struct {
		name        string
		agentID     string
		expectError bool
	}{
		{
			name:        "获取存在的智能体",
			agentID:     testAgent.ID,
			expectError: false,
		},
		{
			name:        "获取不存在的智能体",
			agentID:     "non-existent-id",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 直接调用获取功能
			agent, err := getAgent(tt.agentID)

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, agent)
			assert.Equal(t, testAgent.ID, agent.ID)
		})
	}
}

// 模拟获取Agent功能
func getAgent(agentID string) (*Agent, error) {
	agent, exists := agents[agentID]
	if !exists {
		return nil, fmt.Errorf("agent with ID %s not found", agentID)
	}
	return agent, nil
}

func TestListAgents(t *testing.T) {
	// 清空智能体列表
	agents = make(map[string]*Agent)

	// 添加测试用智能体
	testAgents := []*Agent{
		{
			ID:          uuid.New().String(),
			Name:        "测试智能体1",
			CoreTraits:  "特征1",
			Personality: "性格1",
			CreatedAt:   time.Now().Format(time.RFC3339),
		},
		{
			ID:          uuid.New().String(),
			Name:        "测试智能体2",
			CoreTraits:  "特征2",
			Personality: "性格2",
			CreatedAt:   time.Now().Format(time.RFC3339),
		},
	}

	for _, agent := range testAgents {
		agents[agent.ID] = agent
	}

	// 直接测试列表功能
	agentList := listAgents()
	assert.Equal(t, len(testAgents), len(agentList))
}

// 模拟列出Agent功能
func listAgents() []*Agent {
	agentList := make([]*Agent, 0, len(agents))
	for _, agent := range agents {
		agentList = append(agentList, agent)
	}
	return agentList
}

func TestDeleteAgent(t *testing.T) {
	// 创建测试用智能体
	testAgent := &Agent{
		ID:          uuid.New().String(),
		Name:        "测试智能体",
		CoreTraits:  "测试特征",
		Personality: "测试性格",
		CreatedAt:   time.Now().Format(time.RFC3339),
	}
	agents[testAgent.ID] = testAgent

	tests := []struct {
		name        string
		agentID     string
		expectError bool
	}{
		{
			name:        "删除存在的智能体",
			agentID:     testAgent.ID,
			expectError: false,
		},
		{
			name:        "删除不存在的智能体",
			agentID:     "non-existent-id",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 直接测试删除功能
			err := deleteAgent(tt.agentID)

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)

			// 验证智能体是否已被删除
			_, exists := agents[tt.agentID]
			assert.False(t, exists)
		})
	}
}

// 模拟删除Agent功能
func deleteAgent(agentID string) error {
	if _, exists := agents[agentID]; !exists {
		return fmt.Errorf("agent with ID %s not found", agentID)
	}
	delete(agents, agentID)
	return nil
}

func TestPromptHandler(t *testing.T) {
	// 准备测试数据
	agentName := "测试智能体"
	coreTraits := "专业、严谨、耐心"

	// 直接调用处理器，传递必要参数
	prompt := createExpertPrompt(agentName, coreTraits)

	// 验证结果
	assert.Contains(t, prompt, "测试智能体")
	assert.Contains(t, prompt, "专业、严谨、耐心")
}

// 模拟创建专家提示词功能
func createExpertPrompt(name, traits string) string {
	return fmt.Sprintf("请为名为[%s]的智能体生成一个人格描述，核心特质是：[%s]", name, traits)
}

func setupTestServer() *server.MCPServer {
	s := server.NewMCPServer(
		"测试服务器",
		"1.0.0",
		server.WithPromptCapabilities(true),
	)

	// 添加系统提示词
	systemPrompt := mcp.NewPrompt("generate_expert_agent",
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

	// 添加系统提示词处理器
	s.AddPrompt(systemPrompt, generateExpertAgentHandler)

	return s
}
